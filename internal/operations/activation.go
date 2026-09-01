package operations

import (
	"encoding/json"
	"errors"
	"time"
)

var ErrActivationConflict = errors.New("booking activation conflicts with an existing activation")

type ActivationStageSpec struct {
	Name            string
	JobTemplate     string
	Workflow        string
	DueAt           time.Time
	TimeoutAt       time.Time
	MaximumAttempts int
	Parameters      map[string]string
	ProgressMessage string
}

type CreateActivationRequest struct {
	RunID           string
	BookingName     string
	User            string
	Resource        string
	Stream          string
	Pipeline        string
	ManifestVersion int64
	IdempotencyKey  string
	RequestedAt     time.Time
	ResolvedPlan    json.RawMessage
	Stages          []ActivationStageSpec
	FirstJob        Job
	FirstDelivery   Delivery
}

type ActivationStage struct {
	Index           int
	Name            string
	JobTemplate     string
	Workflow        string
	State           string
	Attempt         int
	MaximumAttempts int
	DueAt           time.Time
	TimeoutAt       time.Time
	Parameters      map[string]string
	ProgressMessage string
	LastErrorCode   string
	LastError       string
	JobID           string
}

type ActivationRun struct {
	ID              string
	BookingName     string
	User            string
	Resource        string
	Stream          string
	Pipeline        string
	ManifestVersion int64
	IdempotencyKey  string
	State           string
	CurrentStage    int
	ProgressMessage string
	FailureCode     string
	FailureMessage  string
	FailureGuidance json.RawMessage
	StartedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     *time.Time
	Stages          []ActivationStage
}
