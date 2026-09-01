package store

import (
	"testing"
	"time"

	"github.com/practable/book/internal/interval"
)

func operationalManifest() Manifest {
	day := func(hour int) time.Time { return time.Date(2026, 9, 7, hour, 0, 0, 0, time.UTC) }
	return Manifest{
		OperationalWorkflows: map[string]OperationalWorkflow{
			"fill":   {Description: "Fill tank", ExpectedDuration: 10 * time.Minute, MaximumDuration: 15 * time.Minute},
			"drain":  {Description: "Drain tank", ExpectedDuration: 10 * time.Minute, MaximumDuration: 15 * time.Minute},
			"settle": {Description: "Thermally settle", ExpectedDuration: 30 * time.Minute, MaximumDuration: 30 * time.Minute},
		},
		Windows: map[string]Window{"working": {Allowed: []interval.Interval{{Start: day(9), End: day(17)}}}},
		Resources: map[string]Resource{
			"tank": {Operations: OperationalProfile{OperatingWindow: "working",
				BeforeBooking: []OperationalGuard{{Workflow: "fill", Duration: 10 * time.Minute, Applies: OperationalOutsideOperatingWindow, Reclaimable: true}},
				AfterBooking:  []OperationalGuard{{Workflow: "drain", Duration: 10 * time.Minute, Applies: OperationalOutsideOperatingWindow, Reclaimable: true}}}},
			"fridge": {Operations: OperationalProfile{AfterBooking: []OperationalGuard{{Workflow: "settle", Duration: 30 * time.Minute, Applies: OperationalAlways}}}},
		},
	}
}

func TestOperationalPlanOmitsOutOfHoursWorkDuringOperatingWindow(t *testing.T) {
	m := operationalManifest()
	booking := interval.Interval{Start: time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 7, 11, 0, 0, 0, time.UTC)}
	plan, err := PlanOperationalGuards(m, "tank", booking, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 0 {
		t.Fatalf("plan = %#v", plan)
	}
}

func TestOperationalPlanCreatesOutOfHoursSetupAndTeardown(t *testing.T) {
	m := operationalManifest()
	booking := interval.Interval{Start: time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)}
	plan, err := PlanOperationalGuards(m, "tank", booking, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 2 {
		t.Fatalf("plan = %#v", plan)
	}
	if plan[0].Kind != OperationalBeforeBooking || !plan[0].When.Start.Equal(booking.Start.Add(-10*time.Minute)) {
		t.Fatalf("setup = %#v", plan[0])
	}
	if plan[1].Kind != OperationalAfterBooking || !plan[1].When.End.Equal(booking.End.Add(10*time.Minute)) {
		t.Fatalf("teardown = %#v", plan[1])
	}
}

func TestReclaimableTeardownIsSuppressedByFollowingBooking(t *testing.T) {
	m := operationalManifest()
	booking := interval.Interval{Start: time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)}
	next := interval.Interval{Start: booking.End.Add(5 * time.Minute), End: booking.End.Add(time.Hour)}
	plan, err := PlanOperationalGuards(m, "tank", booking, nil, &next)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 1 || plan[0].Kind != OperationalBeforeBooking {
		t.Fatalf("plan = %#v", plan)
	}
}

func TestMandatorySettlingIsNeverReclaimed(t *testing.T) {
	m := operationalManifest()
	booking := interval.Interval{Start: time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 7, 11, 0, 0, 0, time.UTC)}
	next := interval.Interval{Start: booking.End.Add(5 * time.Minute), End: booking.End.Add(time.Hour)}
	plan, err := PlanOperationalGuards(m, "fridge", booking, nil, &next)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 1 || plan[0].Workflow != "settle" {
		t.Fatalf("plan = %#v", plan)
	}
}

func TestMultipleOperationalRulesAreSequencedInManifestOrder(t *testing.T) {
	m := operationalManifest()
	m.OperationalWorkflows["power"] = OperationalWorkflow{Description: "Power equipment", ExpectedDuration: 5 * time.Minute, MaximumDuration: 5 * time.Minute}
	r := m.Resources["fridge"]
	r.Operations.BeforeBooking = []OperationalGuard{
		{Workflow: "power", Duration: 5 * time.Minute, Applies: OperationalAlways},
		{Workflow: "fill", Duration: 10 * time.Minute, Applies: OperationalAlways},
	}
	r.Operations.AfterBooking = []OperationalGuard{
		{Workflow: "drain", Duration: 10 * time.Minute, Applies: OperationalAlways},
		{Workflow: "power", Duration: 5 * time.Minute, Applies: OperationalAlways},
	}
	m.Resources["fridge"] = r
	booking := interval.Interval{Start: time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 7, 11, 0, 0, 0, time.UTC)}
	plan, err := PlanOperationalGuards(m, "fridge", booking, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 4 {
		t.Fatalf("plan = %#v", plan)
	}
	if plan[0].Workflow != "power" || plan[1].Workflow != "fill" || !plan[0].When.End.Equal(plan[1].When.Start) || !plan[1].When.End.Equal(booking.Start) {
		t.Fatalf("setup order = %#v", plan[:2])
	}
	if plan[2].Workflow != "drain" || plan[3].Workflow != "power" || !plan[2].When.End.Equal(plan[3].When.Start) {
		t.Fatalf("teardown order = %#v", plan[2:])
	}
}

func TestOperationalManifestRejectsUnknownAndOverlongWorkflow(t *testing.T) {
	m := operationalManifest()
	r := m.Resources["tank"]
	r.Operations.AfterBooking = append(r.Operations.AfterBooking,
		OperationalGuard{Workflow: "fill", Duration: time.Hour, Applies: OperationalAlways},
		OperationalGuard{Workflow: "unknown", Duration: time.Minute, Applies: OperationalAlways})
	m.Resources["tank"] = r
	if messages := validateOperationalManifest(m); len(messages) != 2 {
		t.Fatalf("messages = %#v", messages)
	}
}

func TestOperationalSchedulesUseCivilTimeAndSemesterBounds(t *testing.T) {
	m := operationalManifest()
	m.Slots = map[string]Slot{"tank-slot": {}}
	m.OperationalSchedules = map[string]OperationalSchedule{
		"weekday-fill": {Slot: "tank-slot", Workflow: "fill", Duration: 10 * time.Minute, Conflict: OperationalConflictRequire,
			Recurrence: OperationalRecurrence{Timezone: "Europe/London", StartDate: "2026-03-27", EndDate: "2026-03-31",
				Weekdays: []string{"fri", "mon", "tue"}, Time: "09:00", Exceptions: []string{"2026-03-31"}}},
	}
	occurrences, err := MaterializeOperationalSchedules(m, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(occurrences) != 2 {
		t.Fatalf("occurrences = %#v", occurrences)
	}
	if got := occurrences[0].When.Start.Hour(); got != 9 {
		t.Fatalf("pre-DST UTC hour = %d", got)
	}
	if got := occurrences[1].When.Start.Hour(); got != 8 {
		t.Fatalf("post-DST UTC hour = %d", got)
	}
	if occurrences[1].When.End.Sub(occurrences[1].When.Start) != 10*time.Minute {
		t.Fatalf("duration = %s", occurrences[1].When.End.Sub(occurrences[1].When.Start))
	}
}

func TestScheduledOperationRequiresExperimentCostOwner(t *testing.T) {
	m := operationalManifest()
	m.Slots = map[string]Slot{"tank-slot": {Resource: "tank"}}
	m.OperationalSchedules = map[string]OperationalSchedule{"daily-qc": {
		Slot: "tank-slot", Workflow: "fill", Duration: time.Minute, Conflict: OperationalConflictSkip,
		Recurrence: OperationalRecurrence{Timezone: "UTC", StartDate: "2026-09-01", EndDate: "2026-09-01", Weekdays: []string{"tue"}, Time: "09:00"},
	}}
	messages := validateOperationalManifest(m)
	if len(messages) != 1 || messages[0] != "operational schedule daily-qc resource tank has no operations cost_owner" {
		t.Fatalf("messages = %#v", messages)
	}
	resource := m.Resources["tank"]
	resource.Operations.CostOwner = "physics-teaching-lab"
	m.Resources["tank"] = resource
	if messages := validateOperationalManifest(m); len(messages) != 0 {
		t.Fatalf("messages after owner = %#v", messages)
	}
}

func TestReusableStreamPipelineResolvesTypedResourceBinding(t *testing.T) {
	m := operationalManifest()
	m.OperationalWorkflows["enable-video"] = OperationalWorkflow{Description: "Enable video", Kind: "action", ExpectedDuration: time.Second, MaximumDuration: 4 * time.Second}
	m.OperationalWorkflows["check-video"] = OperationalWorkflow{Description: "Check video", Kind: "health_check", ExpectedDuration: time.Second, MaximumDuration: 2 * time.Second}
	m.OperationalJobTemplates = map[string]OperationalJobTemplate{
		"enable": {Workflow: "enable-video", Timeout: 4 * time.Second},
		"check": {Workflow: "check-video", Timeout: 2 * time.Second, Parameters: map[string]string{"minimum_fps": "20"},
			AllowedOverrides: map[string]string{"minimum_fps": "number"},
			Retry:            OperationalRetryPolicy{Attempts: 3, InitialDelay: 500 * time.Millisecond, Backoff: 1.5, MaximumDelay: time.Second, TotalTimeout: 6 * time.Second}},
	}
	m.OperationalPipelineTemplates = map[string]OperationalPipelineTemplate{"video": {Stages: []OperationalPipelineStage{
		{Name: "enable", JobTemplate: "enable"}, {Name: "verify", JobTemplate: "check"},
	}}}
	resource := m.Resources["tank"]
	resource.Streams = []string{"video"}
	resource.Properties = map[string]string{"camera_minimum_fps": "18"}
	resource.StreamOperations = map[string]OperationalStreamBinding{"video": {ActivationPipeline: "video", Parameters: map[string]OperationalParameterBinding{
		"minimum_fps": {From: "resource.properties.camera_minimum_fps"},
	}}}
	m.Resources["tank"] = resource
	if messages := validateOperationalManifest(m); len(messages) != 0 {
		t.Fatalf("messages = %#v", messages)
	}
	stages, _, err := ResolveOperationalPipeline(m, "tank", "video")
	if err != nil {
		t.Fatal(err)
	}
	if len(stages) != 2 || stages[1].Parameters["minimum_fps"] != "18" || stages[1].Retry.Attempts != 3 {
		t.Fatalf("stages = %#v", stages)
	}
}

func TestReusableStreamPipelineResolvesCleanupOnlyBinding(t *testing.T) {
	m := operationalManifest()
	m.OperationalWorkflows["disable-video"] = OperationalWorkflow{Description: "Disable video", Kind: "action", ExpectedDuration: time.Second, MaximumDuration: 4 * time.Second}
	m.OperationalJobTemplates = map[string]OperationalJobTemplate{
		"enable":  {Workflow: "fill", Timeout: time.Minute},
		"disable": {Workflow: "disable-video", Timeout: 4 * time.Second, AllowedOverrides: map[string]string{"stream": "string"}},
	}
	m.OperationalPipelineTemplates = map[string]OperationalPipelineTemplate{"video": {
		Stages:  []OperationalPipelineStage{{Name: "enable", JobTemplate: "enable"}},
		Cleanup: []OperationalPipelineStage{{Name: "disable", JobTemplate: "disable"}},
	}}
	resource := m.Resources["tank"]
	resource.Streams = []string{"video"}
	resource.Properties = map[string]string{"video_stream": "tank-camera"}
	resource.StreamOperations = map[string]OperationalStreamBinding{"video": {ActivationPipeline: "video", Parameters: map[string]OperationalParameterBinding{
		"stream": {From: "resource.properties.video_stream"},
	}}}
	m.Resources["tank"] = resource
	if messages := validateOperationalManifest(m); len(messages) != 0 {
		t.Fatalf("messages = %#v", messages)
	}
	_, cleanup, err := ResolveOperationalPipeline(m, "tank", "video")
	if err != nil {
		t.Fatal(err)
	}
	if len(cleanup) != 1 || cleanup[0].Parameters["stream"] != "tank-camera" {
		t.Fatalf("cleanup = %#v", cleanup)
	}
}

func TestOperationalPipelineRejectsUnapprovedOrMistypedBinding(t *testing.T) {
	m := operationalManifest()
	m.OperationalWorkflows["check-video"] = OperationalWorkflow{Description: "Check video", Kind: "health_check", ExpectedDuration: time.Second, MaximumDuration: 2 * time.Second}
	m.OperationalJobTemplates = map[string]OperationalJobTemplate{"check": {Workflow: "check-video", Timeout: time.Second, AllowedOverrides: map[string]string{"minimum_fps": "number"}}}
	m.OperationalPipelineTemplates = map[string]OperationalPipelineTemplate{"video": {Stages: []OperationalPipelineStage{{Name: "verify", JobTemplate: "check"}}}}
	resource := m.Resources["tank"]
	resource.Streams = []string{"video"}
	resource.StreamOperations = map[string]OperationalStreamBinding{"video": {ActivationPipeline: "video", Parameters: map[string]OperationalParameterBinding{
		"minimum_fps": {Value: "fast"}, "command": {Value: "arbitrary"},
	}}}
	m.Resources["tank"] = resource
	if messages := validateOperationalManifest(m); len(messages) != 2 {
		t.Fatalf("messages = %#v", messages)
	}
}
