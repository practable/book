package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/practable/book/internal/filter"
	"github.com/practable/book/internal/interval"
	"github.com/practable/book/internal/operations"
)

const (
	OperationalAlways                 = "always"
	OperationalOutsideOperatingWindow = "outside_operating_window"
	OperationalBeforeBooking          = "before_booking"
	OperationalAfterBooking           = "after_booking"
	OperationalConflictRequire        = "require"
	OperationalConflictSkip           = "skip"
)

// OperationalWorkflow is an owner-approved task contract. Manifests contain
// identifiers and duration bounds, never executable commands or webhook URLs.
type OperationalWorkflow struct {
	Description      string        `json:"description" yaml:"description"`
	ExpectedDuration time.Duration `json:"expected_duration" yaml:"expected_duration"`
	MaximumDuration  time.Duration `json:"maximum_duration" yaml:"maximum_duration"`
}

type OperationalSchedule struct {
	Slot       string                `json:"slot" yaml:"slot"`
	Workflow   string                `json:"workflow" yaml:"workflow"`
	Duration   time.Duration         `json:"duration" yaml:"duration"`
	Conflict   string                `json:"conflict" yaml:"conflict"`
	Recurrence OperationalRecurrence `json:"recurrence" yaml:"recurrence"`
}

type OperationalRecurrence struct {
	Timezone   string   `json:"timezone" yaml:"timezone"`
	StartDate  string   `json:"start_date" yaml:"start_date"`
	EndDate    string   `json:"end_date" yaml:"end_date"`
	Weekdays   []string `json:"weekdays" yaml:"weekdays"`
	Time       string   `json:"time" yaml:"time"`
	Exceptions []string `json:"exceptions,omitempty" yaml:"exceptions,omitempty"`
}

type OperationalOccurrence struct {
	Schedule string
	Slot     string
	Workflow string
	Conflict string
	When     interval.Interval
}

type OperationalScheduleSummary struct {
	Planned   int
	Skipped   int
	Conflicts int
	Missed    int
	Existing  int
}

func (s *Store) MaterializeOperationalSchedules(ctx context.Context, from, until time.Time) (OperationalScheduleSummary, error) {
	s.Lock()
	defer s.Unlock()
	repository, ok := s.repository.(OperationalScheduleRepository)
	if !ok {
		return OperationalScheduleSummary{}, errors.New("operational schedules require durable persistence")
	}
	manifest := s.exportManifestLocked()
	occurrences, err := MaterializeOperationalSchedules(manifest, from.UTC(), until.UTC())
	if err != nil {
		return OperationalScheduleSummary{}, err
	}
	summary := OperationalScheduleSummary{}
	for _, occurrence := range occurrences {
		slot, ok := s.Slots[occurrence.Slot]
		if !ok {
			return summary, fmt.Errorf("operational schedule %s slot %s not found", occurrence.Schedule, occurrence.Slot)
		}
		identity := occurrence.Schedule + "\x00" + occurrence.When.Start.UTC().Format(time.RFC3339Nano) + "\x00" + fmt.Sprint(s.manifestVersion)
		jobID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("scheduled-job\x00"+identity)).String()
		bookingID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("scheduled-booking\x00"+identity)).String()
		deliveryID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("scheduled-delivery\x00"+identity)).String()
		idempotencyKey := "scheduled-operation:" + jobID
		command := operations.Command{Version: 1, JobID: jobID, Workflow: occurrence.Workflow, Resource: slot.Resource,
			Kind: "scheduled", StartsAt: occurrence.When.Start.UTC(), EndsAt: occurrence.When.End.UTC(), BookingName: bookingID,
			PlanRevision: 1, IdempotencyKey: idempotencyKey}
		body, err := json.Marshal(command)
		if err != nil {
			return summary, err
		}
		booking := Booking{Name: bookingID, User: "__operations__", Policy: "__operations__", Slot: occurrence.Slot,
			Maintenance: true, When: occurrence.When}
		item := OperationalReservation{
			Request: CreateBookingRequest{Booking: booking, Resource: slot.Resource, ResourceConstrained: true,
				Now: s.now(), ManifestVersion: s.manifestVersion, Maintenance: true, Actor: "operational-scheduler"},
			Job: operations.Job{ID: jobID, Resource: slot.Resource, Workflow: occurrence.Workflow, Kind: "scheduled", State: "reserved",
				DueAt: occurrence.When.Start.UTC(), StartsAt: occurrence.When.Start.UTC(), EndsAt: occurrence.When.End.UTC(),
				TriggeringBookingName: "schedule:" + occurrence.Schedule, ManifestVersion: s.manifestVersion, PlanRevision: 1,
				IdempotencyKey: idempotencyKey, Payload: body},
			Delivery: operations.Delivery{ID: deliveryID, JobID: jobID, Body: body},
		}
		result, err := repository.CreateScheduledOperation(ctx, occurrence.Schedule, occurrence.When.Start, occurrence.Conflict, item)
		if err != nil {
			return summary, err
		}
		if !result.Created {
			summary.Existing++
			continue
		}
		switch result.State {
		case "planned":
			summary.Planned++
		case "skipped":
			summary.Skipped++
		case "conflict":
			summary.Conflicts++
		case "missed":
			summary.Missed++
		}
	}
	if summary.Planned > 0 {
		if err := s.refreshFromRepositoryLocked(ctx); err != nil {
			return summary, err
		}
	}
	return summary, nil
}

func MaterializeOperationalSchedules(manifest Manifest, from, until time.Time) ([]OperationalOccurrence, error) {
	result := make([]OperationalOccurrence, 0)
	for name, schedule := range manifest.OperationalSchedules {
		rule := WeeklyRecurrence{Timezone: schedule.Recurrence.Timezone, StartDate: schedule.Recurrence.StartDate,
			EndDate: schedule.Recurrence.EndDate, Weekdays: schedule.Recurrence.Weekdays,
			StartTime: schedule.Recurrence.Time, EndTime: schedule.Recurrence.Time, Exceptions: schedule.Recurrence.Exceptions}
		starts, err := materializeWeekly(rule)
		if err != nil {
			return nil, fmt.Errorf("operational schedule %s: %w", name, err)
		}
		for _, occurrence := range starts {
			start := occurrence.Start
			if start.Before(from) || !start.Before(until) {
				continue
			}
			result = append(result, OperationalOccurrence{Schedule: name, Slot: schedule.Slot, Workflow: schedule.Workflow,
				Conflict: schedule.Conflict, When: interval.Interval{Start: start, End: start.Add(schedule.Duration)}})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].When.Start.Equal(result[j].When.Start) {
			return result[i].Schedule < result[j].Schedule
		}
		return result[i].When.Start.Before(result[j].When.Start)
	})
	return result, nil
}

func (s *Store) plannedOperationalReservationsLocked(slotName, bookingName, excludeBookingName string, when interval.Interval) ([]OperationalReservation, error) {
	slot := s.Slots[slotName]
	resource := s.Resources[slot.Resource]
	if len(resource.Operations.BeforeBooking) == 0 && len(resource.Operations.AfterBooking) == 0 {
		return nil, nil
	}
	var previous, next *interval.Interval
	for _, existing := range s.Bookings {
		if existing.Cancelled || existing.Name == excludeBookingName || strings.HasPrefix(existing.Policy, "__operations") {
			continue
		}
		existingSlot, ok := s.Slots[existing.Slot]
		if !ok || existingSlot.Resource != slot.Resource {
			continue
		}
		candidate := existing.When
		if !candidate.End.After(when.Start) && (previous == nil || candidate.End.After(previous.End)) {
			copy := candidate
			previous = &copy
		}
		if !candidate.Start.Before(when.End) && (next == nil || candidate.Start.Before(next.Start)) {
			copy := candidate
			next = &copy
		}
	}
	plans, err := PlanOperationalGuards(s.exportManifestLocked(), slot.Resource, when, previous, next)
	if err != nil {
		return nil, err
	}
	result := make([]OperationalReservation, 0, len(plans))
	for _, plan := range plans {
		identity := bookingName + "\x00" + plan.Kind + "\x00" + plan.Workflow + "\x00" +
			plan.When.Start.UTC().Format(time.RFC3339Nano) + "\x00" + plan.When.End.UTC().Format(time.RFC3339Nano)
		jobID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("job\x00"+identity)).String()
		reservationName := uuid.NewSHA1(uuid.NameSpaceURL, []byte("reservation\x00"+identity)).String()
		deliveryID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("delivery\x00"+identity)).String()
		idempotencyKey := "booking-operation:" + jobID
		jobKind := "setup"
		if plan.Kind == OperationalAfterBooking {
			jobKind = "teardown"
		}
		command := operations.Command{Version: 1, JobID: jobID, Workflow: plan.Workflow, Resource: slot.Resource,
			Kind: jobKind, StartsAt: plan.When.Start.UTC(), EndsAt: plan.When.End.UTC(), BookingName: reservationName,
			PlanRevision: 1, IdempotencyKey: idempotencyKey}
		body, err := json.Marshal(command)
		if err != nil {
			return nil, err
		}
		policy := "__operations__"
		if plan.Reclaimable {
			policy = "__operations_reclaimable__"
		}
		reservation := Booking{Name: reservationName, User: "__operations__", Policy: policy, Slot: slotName, Maintenance: true, When: plan.When}
		result = append(result, OperationalReservation{
			Request: CreateBookingRequest{Booking: reservation, Resource: slot.Resource, ResourceConstrained: true,
				Now: s.now(), ManifestVersion: s.manifestVersion, Maintenance: true, Actor: "operational-planner"},
			Job: operations.Job{ID: jobID, Resource: slot.Resource, Workflow: plan.Workflow, Kind: jobKind, State: "reserved",
				DueAt: plan.DueAt.UTC(), StartsAt: plan.When.Start.UTC(), EndsAt: plan.When.End.UTC(), TriggeringBookingName: bookingName,
				ManifestVersion: s.manifestVersion, PlanRevision: 1, IdempotencyKey: idempotencyKey, Payload: body},
			Delivery: operations.Delivery{ID: deliveryID, JobID: jobID, Body: body},
		})
	}
	return result, nil
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
		cursor := booking.End
		first, last, step := 0, len(rules), 1
		if kind == OperationalBeforeBooking {
			cursor = booking.Start
			first, last, step = len(rules)-1, -1, -1
		}
		for index := first; index != last; index += step {
			rule := rules[index]
			if rule.Applies == OperationalOutsideOperatingWindow && !outside {
				continue
			}
			var when interval.Interval
			if kind == OperationalBeforeBooking {
				when = interval.Interval{Start: cursor.Add(-rule.Duration), End: cursor}
				if rule.Reclaimable && previous != nil && previous.End.After(when.Start) {
					continue
				}
				cursor = when.Start
			} else {
				when = interval.Interval{Start: cursor, End: cursor.Add(rule.Duration)}
				if rule.Reclaimable && next != nil && next.Start.Before(when.End) {
					continue
				}
				cursor = when.End
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
	for name, schedule := range m.OperationalSchedules {
		workflow, workflowOK := m.OperationalWorkflows[schedule.Workflow]
		if !workflowOK {
			messages = append(messages, "operational schedule "+name+" references non-existent workflow: "+schedule.Workflow)
		}
		if _, ok := m.Slots[schedule.Slot]; !ok {
			messages = append(messages, "operational schedule "+name+" references non-existent slot: "+schedule.Slot)
		}
		if schedule.Duration <= 0 || (workflowOK && schedule.Duration > workflow.MaximumDuration) {
			messages = append(messages, "operational schedule "+name+" duration must be positive and within workflow maximum")
		}
		if schedule.Conflict != OperationalConflictRequire && schedule.Conflict != OperationalConflictSkip {
			messages = append(messages, "operational schedule "+name+" has invalid conflict mode")
		}
		probe := m
		probe.OperationalSchedules = map[string]OperationalSchedule{name: schedule}
		if _, err := MaterializeOperationalSchedules(probe, time.Time{}, time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
			messages = append(messages, err.Error())
		}
	}
	return messages
}
