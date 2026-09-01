package serve

import (
	"context"
	"testing"
	"time"

	"github.com/practable/book/internal/config"
	"github.com/practable/book/internal/operations"
	"github.com/practable/book/internal/serve/restapi/operations/admin"
)

type occurrenceRepository struct {
	items []operations.ScheduleOccurrence
}

func (*occurrenceRepository) CreateJob(context.Context, operations.Job, operations.Delivery) (operations.Job, bool, error) {
	panic("unused")
}
func (*occurrenceRepository) ClaimDeliveries(context.Context, string, time.Time, time.Duration, int) ([]operations.Delivery, error) {
	panic("unused")
}
func (*occurrenceRepository) CompleteDelivery(context.Context, string, string, bool, int, string, time.Time, time.Time) error {
	panic("unused")
}
func (*occurrenceRepository) ApplyCallback(context.Context, operations.Callback, string) (operations.Job, bool, error) {
	panic("unused")
}
func (*occurrenceRepository) GetJob(context.Context, string) (operations.Job, error) { panic("unused") }
func (r *occurrenceRepository) ListScheduleOccurrences(context.Context, time.Time, time.Time, string, int) ([]operations.ScheduleOccurrence, error) {
	return r.items, nil
}

func TestListOperationalOccurrencesRequiresAdminAndReturnsContext(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	repository := &occurrenceRepository{items: []operations.ScheduleOccurrence{{Schedule: "daily", OccurrenceAt: now,
		ManifestVersion: 4, State: "conflict", Slot: "slot-a", Resource: "resource-a", Workflow: "check", Detail: "busy"}}}
	handler := listOperationalOccurrencesHandler(config.ServerConfig{OperationsRepository: repository, Now: func() time.Time { return now }})
	unauthorized := handler(admin.NewListOperationalOccurrencesParams(), principalWithScopes("booking:maintenance"))
	if _, ok := unauthorized.(*admin.ListOperationalOccurrencesUnauthorized); !ok {
		t.Fatalf("unauthorized response = %T", unauthorized)
	}
	response := handler(admin.NewListOperationalOccurrencesParams(), principalWithScopes("booking:admin"))
	ok, valid := response.(*admin.ListOperationalOccurrencesOK)
	if !valid || len(ok.Payload) != 1 || *ok.Payload[0].State != "conflict" || *ok.Payload[0].Resource != "resource-a" {
		t.Fatalf("response = %#v", response)
	}
}
