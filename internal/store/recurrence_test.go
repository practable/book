package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestWeeklyRecurrenceUsesCivilTimeAcrossDST(t *testing.T) {
	occurrences, err := materializeWeekly(WeeklyRecurrence{
		Timezone: "Europe/London", StartDate: "2026-03-23", EndDate: "2026-04-03",
		Weekdays: []string{"Monday", "Friday"}, StartTime: "09:00", EndTime: "17:00",
	})
	require.NoError(t, err)
	require.Len(t, occurrences, 4)
	require.Equal(t, "2026-03-23T09:00:00Z", occurrences[0].Start.Format(time.RFC3339))
	require.Equal(t, "2026-03-30T08:00:00Z", occurrences[2].Start.Format(time.RFC3339))
	for _, occurrence := range occurrences {
		require.Equal(t, 8*time.Hour, occurrence.End.Sub(occurrence.Start))
	}
}

func TestWeeklyRecurrenceSupportsSemesterExceptionsAndOvernightWindows(t *testing.T) {
	occurrences, err := materializeWeekly(WeeklyRecurrence{
		Timezone: "Europe/London", StartDate: "2026-09-01", EndDate: "2026-09-15",
		Weekdays: []string{"tue"}, StartTime: "22:00", EndTime: "02:00", Exceptions: []string{"2026-09-08"},
	})
	require.NoError(t, err)
	require.Len(t, occurrences, 2)
	require.Equal(t, 4*time.Hour, occurrences[0].End.Sub(occurrences[0].Start))
	require.Equal(t, "2026-09-15", occurrences[1].Start.In(time.FixedZone("BST", 3600)).Format("2006-01-02"))
}

func TestInvalidRecurringWindowIsRejectedWithoutReplacingManifest(t *testing.T) {
	s, _ := newCalendarTestStore(t)
	encoded, err := yaml.Marshal(s.ExportManifest())
	require.NoError(t, err)
	var manifest Manifest
	require.NoError(t, yaml.Unmarshal(encoded, &manifest))
	window := manifest.Windows["w-a"]
	window.RecurringAllowed = []WeeklyRecurrence{{Timezone: "Not/AZone", StartDate: "2026-01-01", EndDate: "2026-02-01", Weekdays: []string{"mon"}, StartTime: "09:00", EndTime: "17:00"}}
	manifest.Windows["w-a"] = window
	require.Error(t, s.ReplaceManifest(manifest))
	require.Empty(t, s.ExportManifest().Windows["w-a"].RecurringAllowed)
}
