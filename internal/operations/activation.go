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
	RetryMessage    string
	WaitAfter       time.Duration
	InitialDelay    time.Duration
	Backoff         float64
	MaximumDelay    time.Duration
	TotalTimeout    time.Duration
	RetryableCodes  []string
	FailureGuidance json.RawMessage
	Kind            string
}

type CreateActivationRequest struct {
	RunID            string
	BookingName      string
	User             string
	Resource         string
	Stream           string
	Pipeline         string
	ManifestVersion  int64
	IdempotencyKey   string
	RequestedAt      time.Time
	ResolvedPlan     json.RawMessage
	Stages           []ActivationStageSpec
	CleanupStages    []ActivationStageSpec
	RecoveryStages   []ActivationStageSpec
	RecoveryAttempts int
	FirstJob         Job
	FirstDelivery    Delivery
}

type ActivationStage struct {
	Index           int
	Phase           string
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
	RetryMessage    string
	WaitAfter       time.Duration
	InitialDelay    time.Duration
	Backoff         float64
	MaximumDelay    time.Duration
	TotalTimeout    time.Duration
	RetryableCodes  []string
	FailureGuidance json.RawMessage
}

type ActivationRun struct {
	ID                      string
	BookingName             string
	User                    string
	Resource                string
	Stream                  string
	Pipeline                string
	ManifestVersion         int64
	IdempotencyKey          string
	State                   string
	CleanupState            string
	CurrentStage            int
	RecoveryAttempt         int
	MaximumRecoveryAttempts int
	ProgressMessage         string
	FailureCode             string
	FailureMessage          string
	FailureGuidance         json.RawMessage
	StartedAt               time.Time
	UpdatedAt               time.Time
	CompletedAt             *time.Time
	Stages                  []ActivationStage
	RecoveryStages          []ActivationStage
	CleanupStages           []ActivationStage
}
