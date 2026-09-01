package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
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
	Kind             string        `json:"kind,omitempty" yaml:"kind,omitempty"`
	ExpectedDuration time.Duration `json:"expected_duration" yaml:"expected_duration"`
	MaximumDuration  time.Duration `json:"maximum_duration" yaml:"maximum_duration"`
}

type OperationalRetryPolicy struct {
	Attempts       int           `json:"attempts" yaml:"attempts"`
	InitialDelay   time.Duration `json:"initial_delay" yaml:"initial_delay"`
	Backoff        float64       `json:"backoff" yaml:"backoff"`
	MaximumDelay   time.Duration `json:"maximum_delay" yaml:"maximum_delay"`
	TotalTimeout   time.Duration `json:"total_timeout" yaml:"total_timeout"`
	RetryableCodes []string      `json:"retryable_codes,omitempty" yaml:"retryable_codes,omitempty"`
}

type OperationalProgressMessages struct {
	Initial string `json:"initial,omitempty" yaml:"initial,omitempty"`
	Retry   string `json:"retry,omitempty" yaml:"retry,omitempty"`
}

type OperationalFailureAction struct {
	Type  string `json:"type" yaml:"type"`
	Label string `json:"label" yaml:"label"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
}

type OperationalFailureGuidance struct {
	Title   string                     `json:"title" yaml:"title"`
	Message string                     `json:"message" yaml:"message"`
	Actions []OperationalFailureAction `json:"actions,omitempty" yaml:"actions,omitempty"`
}

type OperationalJobTemplate struct {
	Workflow         string                      `json:"workflow" yaml:"workflow"`
	Timeout          time.Duration               `json:"timeout" yaml:"timeout"`
	Parameters       map[string]string           `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	AllowedOverrides map[string]string           `json:"allowed_overrides,omitempty" yaml:"allowed_overrides,omitempty"`
	Retry            OperationalRetryPolicy      `json:"retry,omitempty" yaml:"retry,omitempty"`
	ProgressMessages OperationalProgressMessages `json:"progress_messages,omitempty" yaml:"progress_messages,omitempty"`
	FailureGuidance  OperationalFailureGuidance  `json:"failure_guidance,omitempty" yaml:"failure_guidance,omitempty"`
}

type OperationalPipelineStage struct {
	Name        string        `json:"name" yaml:"name"`
	JobTemplate string        `json:"job_template" yaml:"job_template"`
	WaitAfter   time.Duration `json:"wait_after,omitempty" yaml:"wait_after,omitempty"`
}

type OperationalPipelineTemplate struct {
	Stages  []OperationalPipelineStage `json:"stages" yaml:"stages"`
	Cleanup []OperationalPipelineStage `json:"cleanup,omitempty" yaml:"cleanup,omitempty"`
}

type OperationalParameterBinding struct {
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
	From  string `json:"from,omitempty" yaml:"from,omitempty"`
}

type OperationalStreamBinding struct {
	ActivationPipeline string                                 `json:"activation_pipeline" yaml:"activation_pipeline"`
	Parameters         map[string]OperationalParameterBinding `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

type ResolvedOperationalStage struct {
	Name             string
	Template         string
	Workflow         string
	Timeout          time.Duration
	WaitAfter        time.Duration
	Parameters       map[string]string
	Retry            OperationalRetryPolicy
	ProgressMessages OperationalProgressMessages
	FailureGuidance  OperationalFailureGuidance
}

func ResolveOperationalPipeline(manifest Manifest, resourceName, streamName string) ([]ResolvedOperationalStage, []ResolvedOperationalStage, error) {
	resource, ok := manifest.Resources[resourceName]
	if !ok {
		return nil, nil, fmt.Errorf("resource %s not found", resourceName)
	}
	binding, ok := resource.StreamOperations[streamName]
	if !ok {
		return nil, nil, fmt.Errorf("resource %s stream %s has no activation pipeline", resourceName, streamName)
	}
	pipeline, ok := manifest.OperationalPipelineTemplates[binding.ActivationPipeline]
	if !ok {
		return nil, nil, fmt.Errorf("operational pipeline %s not found", binding.ActivationPipeline)
	}
	resolve := func(stages []OperationalPipelineStage) ([]ResolvedOperationalStage, error) {
		result := make([]ResolvedOperationalStage, 0, len(stages))
		for _, stage := range stages {
			template, ok := manifest.OperationalJobTemplates[stage.JobTemplate]
			if !ok {
				return nil, fmt.Errorf("operational job template %s not found", stage.JobTemplate)
			}
			parameters := make(map[string]string, len(template.Parameters)+len(binding.Parameters))
			for name, value := range template.Parameters {
				parameters[name] = value
			}
			for name, source := range binding.Parameters {
				if _, allowed := template.AllowedOverrides[name]; !allowed {
					continue
				}
				value := source.Value
				if source.From != "" {
					value = resource.Properties[strings.TrimPrefix(source.From, "resource.properties.")]
				}
				parameters[name] = value
			}
			result = append(result, ResolvedOperationalStage{Name: stage.Name, Template: stage.JobTemplate, Workflow: template.Workflow,
				Timeout: template.Timeout, WaitAfter: stage.WaitAfter, Parameters: parameters, Retry: template.Retry,
				ProgressMessages: template.ProgressMessages, FailureGuidance: template.FailureGuidance})
		}
		return result, nil
	}
	stages, err := resolve(pipeline.Stages)
	if err != nil {
		return nil, nil, err
	}
	cleanup, err := resolve(pipeline.Cleanup)
	return stages, cleanup, err
}

// BeginBookingActivation resolves the active manifest into an immutable plan
// and atomically persists that plan with its first runner delivery. Repeated
// requests with the same key return the original activation run.
func (s *Store) BeginBookingActivation(ctx context.Context, bookingName, streamName, idempotencyKey string) (operations.ActivationRun, bool, error) {
	s.Lock()
	defer s.Unlock()
	if err := s.expireAndRefreshLocked(ctx); err != nil {
		return operations.ActivationRun{}, false, err
	}
	repository, ok := s.repository.(BookingActivationRepository)
	if !ok {
		return operations.ActivationRun{}, false, errors.New("booking activation requires durable persistence")
	}
	booking, ok := s.Bookings[bookingName]
	if !ok {
		return operations.ActivationRun{}, false, ErrPersistentNotFound
	}
	slot, ok := s.Slots[booking.Slot]
	if !ok {
		return operations.ActivationRun{}, false, fmt.Errorf("slot %s not found", booking.Slot)
	}
	manifest := s.exportManifestLocked()
	stages, cleanup, err := ResolveOperationalPipeline(manifest, slot.Resource, streamName)
	if err != nil {
		return operations.ActivationRun{}, false, err
	}
	if len(stages) == 0 {
		return operations.ActivationRun{}, false, errors.New("activation pipeline has no stages")
	}
	now := s.now().UTC()
	runID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("booking-activation\x00"+bookingName+"\x00"+idempotencyKey)).String()
	makeStageSpecs := func(values []ResolvedOperationalStage, firstDue time.Time) []operations.ActivationStageSpec {
		result := make([]operations.ActivationStageSpec, 0, len(values))
		due := firstDue
		for _, stage := range values {
			attempts := stage.Retry.Attempts
			if attempts < 1 {
				attempts = 1
			}
			result = append(result, operations.ActivationStageSpec{Name: stage.Name, JobTemplate: stage.Template, Workflow: stage.Workflow,
				DueAt: due, TimeoutAt: due.Add(stage.Timeout), MaximumAttempts: attempts, Parameters: stage.Parameters,
				ProgressMessage: stage.ProgressMessages.Initial, RetryMessage: stage.ProgressMessages.Retry, WaitAfter: stage.WaitAfter,
				InitialDelay: stage.Retry.InitialDelay, Backoff: stage.Retry.Backoff, MaximumDelay: stage.Retry.MaximumDelay,
				TotalTimeout: stage.Retry.TotalTimeout, RetryableCodes: stage.Retry.RetryableCodes, FailureGuidance: mustMarshalOperationalGuidance(stage.FailureGuidance)})
			due = due.Add(stage.Timeout + stage.WaitAfter)
		}
		return result
	}
	stageSpecs := makeStageSpecs(stages, now)
	cleanupSpecs := makeStageSpecs(cleanup, now)
	resolvedPlan, err := json.Marshal(struct {
		Stages  []ResolvedOperationalStage `json:"stages"`
		Cleanup []ResolvedOperationalStage `json:"cleanup,omitempty"`
	}{Stages: stages, Cleanup: cleanup})
	if err != nil {
		return operations.ActivationRun{}, false, err
	}
	first := stages[0]
	jobID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("booking-activation-job\x00"+runID+"\x000\x001")).String()
	deliveryID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("booking-activation-delivery\x00"+jobID)).String()
	jobKey := "booking-activation:" + runID + ":0:1"
	command := operations.Command{Version: 1, JobID: jobID, Workflow: first.Workflow, Resource: slot.Resource, Kind: "preflight",
		StartsAt: now, EndsAt: now.Add(first.Timeout), BookingName: bookingName, PlanRevision: 1, IdempotencyKey: jobKey, Parameters: first.Parameters}
	body, err := json.Marshal(command)
	if err != nil {
		return operations.ActivationRun{}, false, err
	}
	request := operations.CreateActivationRequest{RunID: runID, BookingName: bookingName, User: booking.User, Resource: slot.Resource,
		Stream: streamName, Pipeline: s.Resources[slot.Resource].StreamOperations[streamName].ActivationPipeline, ManifestVersion: s.manifestVersion,
		IdempotencyKey: idempotencyKey, RequestedAt: now, ResolvedPlan: resolvedPlan, Stages: stageSpecs, CleanupStages: cleanupSpecs,
		FirstJob: operations.Job{ID: jobID, Resource: slot.Resource, Workflow: first.Workflow, Kind: "preflight", State: "reserved", DueAt: now,
			StartsAt: now, EndsAt: now.Add(first.Timeout), TriggeringBookingName: bookingName, ManifestVersion: s.manifestVersion,
			PlanRevision: 1, IdempotencyKey: jobKey, Payload: body},
		FirstDelivery: operations.Delivery{ID: deliveryID, JobID: jobID, Body: body}}
	return repository.CreateActivation(ctx, request)
}

func mustMarshalOperationalGuidance(value OperationalFailureGuidance) json.RawMessage {
	if value.Title == "" && value.Message == "" && len(value.Actions) == 0 {
		return nil
	}
	body, _ := json.Marshal(value)
	return body
}

func (s *Store) GetBookingActivation(ctx context.Context, runID string) (operations.ActivationRun, error) {
	s.RLock()
	repository, ok := s.repository.(BookingActivationRepository)
	s.RUnlock()
	if !ok {
		return operations.ActivationRun{}, errors.New("booking activation requires durable persistence")
	}
	return repository.GetActivation(ctx, runID)
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
			Usage: &OperationalUsageAttribution{Phase: "scheduled", PayerKind: "experiment_owner",
				PayerID: operationalCostOwner(slot.Resource, s.Resources[slot.Resource].Operations), Chargeable: true},
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
	CostOwner       string             `json:"cost_owner,omitempty" yaml:"cost_owner,omitempty"`
	BeforeBooking   []OperationalGuard `json:"before_booking,omitempty" yaml:"before_booking,omitempty"`
	AfterBooking    []OperationalGuard `json:"after_booking,omitempty" yaml:"after_booking,omitempty"`
}

func operationalCostOwner(resourceName string, profile OperationalProfile) string {
	if owner := strings.TrimSpace(profile.CostOwner); owner != "" {
		return owner
	}
	return resourceName
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
		if workflow.Kind != "" && workflow.Kind != "action" && workflow.Kind != "health_check" {
			messages = append(messages, "operational workflow "+name+" has invalid kind")
		}
	}
	for name, template := range m.OperationalJobTemplates {
		workflow, ok := m.OperationalWorkflows[template.Workflow]
		if !ok {
			messages = append(messages, "operational job template "+name+" references non-existent workflow: "+template.Workflow)
		} else if template.Timeout <= 0 || template.Timeout > workflow.MaximumDuration {
			messages = append(messages, "operational job template "+name+" timeout must be positive and within workflow maximum")
		}
		for parameter, kind := range template.AllowedOverrides {
			if !validOperationalParameterType(kind) {
				messages = append(messages, "operational job template "+name+" parameter "+parameter+" has invalid type")
			}
			if value, exists := template.Parameters[parameter]; exists && !validOperationalParameterValue(kind, value) {
				messages = append(messages, "operational job template "+name+" parameter "+parameter+" default has invalid "+kind+" value")
			}
		}
		if template.Retry.Attempts < 0 || template.Retry.Attempts > 20 {
			messages = append(messages, "operational job template "+name+" retry attempts must be between 0 and 20")
		}
		if template.Retry.Attempts > 1 && (template.Retry.InitialDelay < 0 || template.Retry.MaximumDelay < template.Retry.InitialDelay || template.Retry.TotalTimeout <= 0 || template.Retry.Backoff < 1) {
			messages = append(messages, "operational job template "+name+" has invalid retry timing")
		}
		for _, action := range template.FailureGuidance.Actions {
			if action.Type != "retry" && action.Type != "choose_another" && action.Type != "contact_support" {
				messages = append(messages, "operational job template "+name+" has invalid failure action: "+action.Type)
			}
			if action.Label == "" {
				messages = append(messages, "operational job template "+name+" has a failure action with no label")
			}
			if action.URL != "" {
				parsed, err := url.Parse(action.URL)
				if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
					messages = append(messages, "operational job template "+name+" failure action URL must use https")
				}
			}
		}
	}
	for name, pipeline := range m.OperationalPipelineTemplates {
		seen := make(map[string]bool)
		for _, stages := range [][]OperationalPipelineStage{pipeline.Stages, pipeline.Cleanup} {
			for _, stage := range stages {
				if stage.Name == "" || seen[stage.Name] {
					messages = append(messages, "operational pipeline "+name+" has missing or duplicate stage name: "+stage.Name)
				}
				seen[stage.Name] = true
				if _, ok := m.OperationalJobTemplates[stage.JobTemplate]; !ok {
					messages = append(messages, "operational pipeline "+name+" references non-existent job template: "+stage.JobTemplate)
				}
				if stage.WaitAfter < 0 {
					messages = append(messages, "operational pipeline "+name+" stage "+stage.Name+" has negative wait_after")
				}
			}
		}
		if len(pipeline.Stages) == 0 {
			messages = append(messages, "operational pipeline "+name+" must contain at least one stage")
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
		streamSet := make(map[string]bool)
		for _, stream := range resource.Streams {
			streamSet[stream] = true
		}
		for stream, binding := range resource.StreamOperations {
			if !streamSet[stream] {
				messages = append(messages, "resource "+resourceName+" configures operations for unlisted stream: "+stream)
			}
			pipeline, ok := m.OperationalPipelineTemplates[binding.ActivationPipeline]
			if !ok {
				messages = append(messages, "resource "+resourceName+" stream "+stream+" references non-existent activation pipeline: "+binding.ActivationPipeline)
				continue
			}
			allowed := make(map[string]string)
			for _, stages := range [][]OperationalPipelineStage{pipeline.Stages, pipeline.Cleanup} {
				for _, stage := range stages {
					for parameter, kind := range m.OperationalJobTemplates[stage.JobTemplate].AllowedOverrides {
						if previous, exists := allowed[parameter]; exists && previous != kind {
							messages = append(messages, "resource "+resourceName+" stream "+stream+" parameter "+parameter+" has conflicting types in its operational pipeline")
						}
						allowed[parameter] = kind
					}
				}
			}
			for parameter, value := range binding.Parameters {
				kind, ok := allowed[parameter]
				if !ok {
					messages = append(messages, "resource "+resourceName+" stream "+stream+" overrides unapproved parameter: "+parameter)
					continue
				}
				if (value.Value == "") == (value.From == "") {
					messages = append(messages, "resource "+resourceName+" stream "+stream+" parameter "+parameter+" must set exactly one of value or from")
					continue
				}
				resolved := value.Value
				if value.From != "" {
					const prefix = "resource.properties."
					if !strings.HasPrefix(value.From, prefix) {
						messages = append(messages, "resource "+resourceName+" stream "+stream+" parameter "+parameter+" has invalid binding source")
						continue
					}
					property := strings.TrimPrefix(value.From, prefix)
					var exists bool
					resolved, exists = resource.Properties[property]
					if !exists {
						messages = append(messages, "resource "+resourceName+" stream "+stream+" parameter "+parameter+" references missing property: "+property)
						continue
					}
				}
				if !validOperationalParameterValue(kind, resolved) {
					messages = append(messages, "resource "+resourceName+" stream "+stream+" parameter "+parameter+" has invalid "+kind+" value")
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

func validOperationalParameterType(kind string) bool {
	return kind == "string" || kind == "number" || kind == "boolean" || kind == "duration"
}

func validOperationalParameterValue(kind, value string) bool {
	switch kind {
	case "string":
		return value != ""
	case "number":
		_, err := strconv.ParseFloat(value, 64)
		return err == nil
	case "boolean":
		_, err := strconv.ParseBool(value)
		return err == nil
	case "duration":
		duration, err := time.ParseDuration(value)
		return err == nil && duration >= 0
	default:
		return false
	}
}
