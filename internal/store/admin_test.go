package store

import (
	"testing"
	"time"

	"github.com/practable/book/internal/interval"
	"github.com/stretchr/testify/require"
)

func TestQueryBookingsFiltersCurrentRecordsByResourceAndRange(t *testing.T) {
	s, now := newCalendarTestStore(t)
	booking, err := s.MakeCalendarBooking("calendar-user", CalendarSelector{Policy: "p-a"}, interval.Interval{Start: now.Add(70 * time.Minute), End: now.Add(80 * time.Minute)}, "admin-query-request")
	require.NoError(t, err)
	from, to := now.Add(time.Hour), now.Add(2*time.Hour)
	records, err := s.QueryBookings(BookingQuery{Resource: "r-a", State: "current", From: &from, To: &to})
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, booking.Name, records[0].Booking.Name)
	require.Equal(t, "r-a", records[0].Resource)
	require.Equal(t, "current", records[0].Collection)

	records, err = s.QueryBookings(BookingQuery{Resource: "r-missing"})
	require.NoError(t, err)
	require.Empty(t, records)
}

func TestMaintenanceBookingCanTargetResource(t *testing.T) {
	s, now := newCalendarTestStore(t)
	require.NoError(t, s.SetResourceIsAvailable("r-a", false, "repair"))
	booking, err := s.MakeMaintenanceBookingForResource("r-a", "technician", interval.Interval{Start: now.Add(time.Hour), End: now.Add(70 * time.Minute)})
	require.NoError(t, err)
	require.True(t, booking.Maintenance)
	require.Equal(t, "r-a", s.Slots[booking.Slot].Resource)
	_, err = s.MakeMaintenanceBookingForResource("r-a", "technician", booking.When)
	require.Error(t, err)
}

func TestFilteredUsageClipsToReportWindow(t *testing.T) {
	s, now := newCalendarTestStore(t)
	booking, err := s.MakeCalendarBooking("calendar-user", CalendarSelector{Policy: "p-a"}, interval.Interval{Start: now.Add(time.Hour), End: now.Add(2 * time.Hour)}, "usage-report-request")
	require.NoError(t, err)
	s.Bookings[booking.Name].Started = true
	s.Bookings[booking.Name].StartedAt = now.Add(time.Hour).Format(time.RFC3339Nano)
	s.SetNow(func() time.Time { return now.Add(90 * time.Minute) })
	from, to := now.Add(70*time.Minute), now.Add(80*time.Minute)
	summary := s.GetFilteredUsageSummary(UsageQuery{Resource: "r-a", From: &from, To: &to})
	require.Equal(t, 10*time.Minute, summary.ActualUsage)
	require.EqualValues(t, 1, summary.StartedBookings)
}
