package store

import (
	"testing"
	"time"

	"github.com/practable/book/internal/interval"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func newCalendarTestStore(t *testing.T) (*Store, time.Time) {
	t.Helper()
	now := time.Date(2022, 11, 4, 9, 0, 0, 0, time.UTC)
	var manifest Manifest
	require.NoError(t, yaml.Unmarshal(manifestYAML, &manifest))
	resource := manifest.Resources["r-a"]
	resource.Class = "spinner"
	resource.Properties = map[string]string{"disc": "68g", "motor": "brushless"}
	manifest.Resources["r-a"] = resource
	s := New().WithNow(func() time.Time { return now })
	require.NoError(t, s.ReplaceManifest(manifest))
	require.NoError(t, s.AddGroupForUser("calendar-user", "g-a"))
	return s, now
}

func TestCalendarCatalogueUsesExistingPolicyAndGuideModel(t *testing.T) {
	s, _ := newCalendarTestStore(t)
	items, err := s.GetCalendarCatalogue("g-a")
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "p-a", items[0].Policy)
	require.Equal(t, time.Minute, items[0].BookingIncrement)
	require.Len(t, items[0].Resources, 1)
	require.Equal(t, "spinner", items[0].Resources[0].Class)
	require.Equal(t, "68g", items[0].Resources[0].Properties["disc"])
}

func TestCalendarAvailabilityAndPreviewRespectExactIntervals(t *testing.T) {
	s, now := newCalendarTestStore(t)
	selector := CalendarSelector{Policy: "p-a", Properties: map[string]string{"disc": "68g"}}
	bands, err := s.GetCalendarAvailability(selector, now.Add(time.Hour), now.Add(80*time.Minute), 10*time.Minute, 5*time.Minute)
	require.NoError(t, err)
	require.Len(t, bands, 4)
	require.True(t, bands[0].Bookable)
	require.EqualValues(t, 1, bands[0].MatchingResources)

	when := interval.Interval{Start: now.Add(70 * time.Minute), End: now.Add(80 * time.Minute)}
	preview, err := s.PreviewCalendarBooking("calendar-user", selector, when)
	require.NoError(t, err)
	require.True(t, preview.Bookable)
	require.Equal(t, []string{"r-a"}, preview.MatchingResources)
	require.Equal(t, 10*time.Minute, preview.UsageAfter)
}

func TestCalendarCreationIsIdempotentAndConsumesAvailability(t *testing.T) {
	s, now := newCalendarTestStore(t)
	selector := CalendarSelector{Policy: "p-a"}
	when := interval.Interval{Start: now.Add(70 * time.Minute), End: now.Add(80 * time.Minute)}
	first, err := s.MakeCalendarBooking("calendar-user", selector, when, "request-123")
	require.NoError(t, err)
	second, err := s.MakeCalendarBooking("calendar-user", selector, when, "request-123")
	require.NoError(t, err)
	require.Equal(t, first.Name, second.Name)

	bands, err := s.GetCalendarAvailability(selector, when.Start, when.End, 10*time.Minute, 5*time.Minute)
	require.NoError(t, err)
	require.False(t, bands[0].Bookable)
}
