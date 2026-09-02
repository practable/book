package serve

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/practable/book/internal/store"
)

func TestOperationalActivityIncludesDeterministicScopedRelayToken(t *testing.T) {
	start := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	activity := store.Activity{
		BookingID: "booking-1", NotBefore: start, ExpiresAt: start.Add(time.Minute),
		Description: store.Description{Name: "test", Type: "experiment"},
		Streams: map[string]store.Stream{
			"video": {Audience: "https://relay.example", ConnectionType: "session", For: "video", Scopes: []string{"read"}, Topic: "resource-video", URL: "https://relay.example"},
		},
	}
	secret := []byte("0123456789abcdef0123456789abcdef")
	first, err := operationalActivityModel(activity, secret)
	if err != nil {
		t.Fatal(err)
	}
	second, err := operationalActivityModel(activity, secret)
	if err != nil {
		t.Fatal(err)
	}
	left, _ := json.Marshal(first)
	right, _ := json.Marshal(second)
	if string(left) != string(right) {
		t.Fatalf("activation response changed across retry\n%s\n%s", left, right)
	}
	if len(first.Streams) != 1 || first.Streams[0].Token == "" || *first.Streams[0].URL != "https://relay.example/session/resource-video" {
		t.Fatalf("unexpected stream response: %+v", first.Streams)
	}
	claims := &Permission{}
	parsed, err := (&jwt.Parser{SkipClaimsValidation: true}).ParseWithClaims(first.Streams[0].Token, claims, func(token *jwt.Token) (interface{}, error) { return secret, nil })
	if err != nil || !parsed.Valid {
		t.Fatalf("invalid relay token: %v", err)
	}
	if claims.BookingID != activity.BookingID || claims.Topic != "resource-video" || claims.Prefix != "session" || claims.Subject != "__operations__" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if !claims.IssuedAt.Time.Equal(start.Add(-time.Second)) || !claims.NotBefore.Time.Equal(start.Add(-time.Second)) || !claims.ExpiresAt.Time.Equal(activity.ExpiresAt) {
		t.Fatalf("unexpected claim times: %+v", claims.RegisteredClaims)
	}
}
