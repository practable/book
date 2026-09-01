package webhook

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseSecretRequiresExactly256Bits(t *testing.T) {
	hexSecret := strings.Repeat("ab", 32)
	secret, err := ParseSecret(hexSecret)
	require.NoError(t, err)
	require.Len(t, secret, 32)
	_, err = ParseSecret("550e8400-e29b-41d4" + "-a716-446655440000")
	require.Error(t, err)
}

func TestSignAndVerifyBindsDirectionDeliveryAndBody(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	body := []byte(`{"job_id":"job-1"}`)
	headers, err := Sign(secret, DirectionCommand, "delivery-1", now, body)
	require.NoError(t, err)
	require.NoError(t, Verify(secret, DirectionCommand, headers, now.Add(time.Minute), 5*time.Minute, body))
	require.ErrorIs(t, Verify(secret, DirectionCallback, headers, now, 5*time.Minute, body), ErrInvalidSignature)
	require.ErrorIs(t, Verify(secret, DirectionCommand, headers, now, 5*time.Minute, []byte(`{"job_id":"job-2"}`)), ErrInvalidSignature)
	headers.DeliveryID = "delivery-2"
	require.ErrorIs(t, Verify(secret, DirectionCommand, headers, now, 5*time.Minute, body), ErrInvalidSignature)
}

func TestVerifyRejectsOldAndFutureSignatures(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	for _, signedAt := range []time.Time{now.Add(-6 * time.Minute), now.Add(6 * time.Minute)} {
		headers, err := Sign(secret, DirectionCallback, "callback-1", signedAt, nil)
		require.NoError(t, err)
		require.True(t, errors.Is(Verify(secret, DirectionCallback, headers, now, 5*time.Minute, nil), ErrExpiredSignature))
	}
}
