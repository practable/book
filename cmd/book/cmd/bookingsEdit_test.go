package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/practable/book/internal/client/models"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestBookingEditFileYAMLRoundTrip(t *testing.T) {
	name, policy, slot, user, original := "replacement", "policy", "slot", "user", "old-booking"
	revision := int64(42)
	model := &models.BookingEdit{OriginalName: &original, Revision: &revision,
		Booking: &models.Booking{Name: &name, Policy: &policy, Slot: &slot, User: &user,
			When: &models.Interval{Start: strfmt.DateTime(time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)), End: strfmt.DateTime(time.Date(2026, 9, 1, 10, 30, 0, 0, time.UTC))}}}
	file, err := editFileFromModel(model)
	require.NoError(t, err)
	document, err := yaml.Marshal(file)
	require.NoError(t, err)
	require.Contains(t, string(document), "original_name:")
	require.Contains(t, string(document), "revision:")
	require.NotContains(t, strings.ToLower(string(document)), "originalname:")
	var restored bookingEditFile
	require.NoError(t, yaml.Unmarshal(document, &restored))
	got, err := restored.toModel()
	require.NoError(t, err)
	require.Equal(t, original, *got.OriginalName)
	require.Equal(t, revision, *got.Revision)
	require.Equal(t, name, *got.Booking.Name)
	require.Equal(t, model.Booking.When.Start.String(), got.Booking.When.Start.String())
}
