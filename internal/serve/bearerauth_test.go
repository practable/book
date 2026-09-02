package serve

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
	lit "github.com/practable/book/internal/login"
	"github.com/practable/book/internal/store"
	"github.com/stretchr/testify/require"
)

func principalWithScopes(scopes ...string) interface{} {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &lit.Token{Scopes: scopes})
}

func TestMaintenanceScopeCancelsOnlyMaintenanceBooking(t *testing.T) {
	claims := &lit.Token{Scopes: []string{"booking:maintenance"}}
	require.True(t, cancelScopeAllowsBooking(false, claims, store.Booking{Maintenance: true}))
	require.False(t, cancelScopeAllowsBooking(false, claims, store.Booking{}))
	require.True(t, cancelScopeAllowsBooking(false, &lit.Token{Scopes: []string{"booking:user"}}, store.Booking{}))
	require.True(t, cancelScopeAllowsBooking(true, claims, store.Booking{}))
}

func TestTechnicianScopesAreNarrowAndUsable(t *testing.T) {
	_, err := isMaintenanceOrAdmin(principalWithScopes("booking:maintenance"))
	require.NoError(t, err)
	_, err = isOverrideOrAdmin(principalWithScopes("booking:booking-override"))
	require.NoError(t, err)
	admin, _, err := isActivityCaller(principalWithScopes("booking:maintenance"))
	require.NoError(t, err)
	require.False(t, admin)
	_, err = isAdmin(principalWithScopes("booking:maintenance"))
	require.Error(t, err)
}
