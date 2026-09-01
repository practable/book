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
