package store

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/practable/book/internal/filter"
	"github.com/practable/book/internal/interval"
)

const (
	OperationalAlways                 = "always"
	OperationalOutsideOperatingWindow = "outside_operating_window"
	OperationalBeforeBooking          = "before_booking"
	OperationalAfterBooking           = "after_booking"
)

// OperationalWorkflow is an owner-approved task contract. Manifests contain
// identifiers and duration bounds, never executable commands or webhook URLs.
type OperationalWorkflow struct {
	Description      string        `json:"description" yaml:"description"`
	ExpectedDuration time.Duration `json:"expected_duration" yaml:"expected_duration"`
	MaximumDuration  time.Duration `json:"maximum_duration" yaml:"maximum_duration"`
}

type OperationalGuard struct {
	Workflow    string        `json:"workflow" yaml:"workflow"`
	Duration    time.Duration `json:"duration" yaml:"duration"`
	Applies     string        `json:"applies" yaml:"applies"`
	Reclaimable bool          `json:"reclaimable,omitempty" yaml:"reclaimable,omitempty"`
}

type OperationalProfile struct {
	OperatingWindow string             `json:"operating_window,omitempty" yaml:"operating_window,omitempty"`
	BeforeBooking   []OperationalGuard `json:"before_booking,omitempty" yaml:"before_booking,omitempty"`
	AfterBooking    []OperationalGuard `json:"after_booking,omitempty" yaml:"after_booking,omitempty"`
}

type OperationalGuardPlan struct {
	Workflow    string            `json:"workflow" yaml:"workflow"`
	Kind        string            `json:"kind" yaml:"kind"`
	When        interval.Interval `json:"when" yaml:"when"`
	DueAt       time.Time         `json:"due_at" yaml:"due_at"`
	Reclaimable bool              `json:"reclaimable" yaml:"reclaimable"`
}

// PlanOperationalGuards deterministically plans one booking against its
// immediate resource neighbours. Persistence and dispatch remain separate.
func PlanOperationalGuards(manifest Manifest, resourceName string, booking interval.Interval, previous, next *interval.Interval) ([]OperationalGuardPlan, error) {
	resource, ok := manifest.Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource %s not found", resourceName)
	}
	if !booking.End.After(booking.Start) {
		return nil, errors.New("booking end must be after start")
	}
	profile := resource.Operations
	outside := false
	if profile.OperatingWindow != "" {
		window, ok := manifest.Windows[profile.OperatingWindow]
		if !ok {
			return nil, fmt.Errorf("operating window %s not found", profile.OperatingWindow)
		}
		expanded, err := expandWindow(window)
		if err != nil {
			return nil, err
		}
		f := filter.New()
		if err := f.SetAllowed(expanded.Allowed); err != nil {
			return nil, err
		}
		if err := f.SetDenied(expanded.Denied); err != nil {
			return nil, err
		}
		outside = !f.Allowed(booking)
	}
	result := make([]OperationalGuardPlan, 0, len(profile.BeforeBooking)+len(profile.AfterBooking))
	appendRules := func(kind string, rules []OperationalGuard) {
		for _, rule := range rules {
			if rule.Applies == OperationalOutsideOperatingWindow && !outside {
				continue
			}
			var when interval.Interval
			if kind == OperationalBeforeBooking {
				when = interval.Interval{Start: booking.Start.Add(-rule.Duration), End: booking.Start}
				if rule.Reclaimable && previous != nil && previous.End.After(when.Start) {
					continue
				}
			} else {
				when = interval.Interval{Start: booking.End, End: booking.End.Add(rule.Duration)}
				if rule.Reclaimable && next != nil && next.Start.Before(when.End) {
					continue
				}
			}
			result = append(result, OperationalGuardPlan{Workflow: rule.Workflow, Kind: kind, When: when, DueAt: when.Start, Reclaimable: rule.Reclaimable})
		}
	}
	appendRules(OperationalBeforeBooking, profile.BeforeBooking)
	appendRules(OperationalAfterBooking, profile.AfterBooking)
	sort.Slice(result, func(i, j int) bool {
		if result[i].When.Start.Equal(result[j].When.Start) {
			return result[i].Workflow < result[j].Workflow
		}
		return result[i].When.Start.Before(result[j].When.Start)
	})
	return result, nil
}

func validateOperationalManifest(m Manifest) []string {
	var messages []string
	for name, workflow := range m.OperationalWorkflows {
		if workflow.Description == "" {
			messages = append(messages, "missing description field in operational workflow "+name)
		}
		if workflow.ExpectedDuration <= 0 {
			messages = append(messages, "operational workflow "+name+" expected_duration must be positive")
		}
		if workflow.MaximumDuration < workflow.ExpectedDuration {
			messages = append(messages, "operational workflow "+name+" maximum_duration must be at least expected_duration")
		}
	}
	for resourceName, resource := range m.Resources {
		profile := resource.Operations
		if profile.OperatingWindow != "" {
			if _, ok := m.Windows[profile.OperatingWindow]; !ok {
				messages = append(messages, "resource "+resourceName+" references non-existent operating window: "+profile.OperatingWindow)
			}
		}
		for _, rules := range [][]OperationalGuard{profile.BeforeBooking, profile.AfterBooking} {
			for _, rule := range rules {
				workflow, ok := m.OperationalWorkflows[rule.Workflow]
				if !ok {
					messages = append(messages, "resource "+resourceName+" references non-existent operational workflow: "+rule.Workflow)
					continue
				}
				if rule.Duration <= 0 || rule.Duration > workflow.MaximumDuration {
					messages = append(messages, "resource "+resourceName+" guard duration for "+rule.Workflow+" must be positive and within workflow maximum")
				}
				if rule.Applies != OperationalAlways && rule.Applies != OperationalOutsideOperatingWindow {
					messages = append(messages, "resource "+resourceName+" guard for "+rule.Workflow+" has invalid applies value")
				}
				if rule.Applies == OperationalOutsideOperatingWindow && profile.OperatingWindow == "" {
					messages = append(messages, "resource "+resourceName+" uses outside_operating_window without an operating_window")
				}
			}
		}
	}
	return messages
}
