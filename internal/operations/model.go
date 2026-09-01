package operations

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("operational job not found")
	ErrCallbackConflict  = errors.New("callback delivery ID was already used with different content")
	ErrInvalidTransition = errors.New("invalid operational job state transition")
)

type Job struct {
	ID                    string
	Resource              string
	Workflow              string
	Kind                  string
	State                 string
	DueAt                 time.Time
	StartsAt              time.Time
	EndsAt                time.Time
	BookingRowID          *int64
	TriggeringBookingName string
	ManifestVersion       int64
	PlanRevision          int64
	IdempotencyKey        string
	Payload               []byte
	Attempts              int
	LastError             string
}

type Delivery struct {
	ID       string
	JobID    string
	Body     []byte
	Attempts int
}

type Callback struct {
	DeliveryID string
	JobID      string
	State      string
	At         time.Time
	Code       string
	Error      string
}

type ScheduleOccurrence struct {
	Schedule        string
	OccurrenceAt    time.Time
	ManifestVersion int64
	State           string
	Slot            string
	Resource        string
	Workflow        string
	BookingName     string
	JobID           string
	Detail          string
}

type ScheduleReader interface {
	ListScheduleOccurrences(context.Context, time.Time, time.Time, string, int) ([]ScheduleOccurrence, error)
}

type Repository interface {
	CreateJob(context.Context, Job, Delivery) (Job, bool, error)
	ClaimDeliveries(context.Context, string, time.Time, time.Duration, int) ([]Delivery, error)
	CompleteDelivery(context.Context, string, string, bool, int, string, time.Time, time.Time) error
	ApplyCallback(context.Context, Callback, string) (Job, bool, error)
	GetJob(context.Context, string) (Job, error)
}

// ActivationTimeoutRepository is implemented by durable repositories that can
// recover activation stages whose runner callback deadline has elapsed.
type ActivationTimeoutRepository interface {
	SweepActivationTimeouts(context.Context, time.Time, int) (int, error)
}
