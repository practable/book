package operations

import "time"

// Command is the stable, pre-approved message sent to the job runner. Params
// are validated by the planner; they are data, never executable commands.
type Command struct {
	Version        int               `json:"version"`
	JobID          string            `json:"job_id"`
	Workflow       string            `json:"workflow"`
	Resource       string            `json:"resource"`
	Kind           string            `json:"kind"`
	StartsAt       time.Time         `json:"starts_at"`
	EndsAt         time.Time         `json:"ends_at"`
	BookingName    string            `json:"booking_name,omitempty"`
	PlanRevision   int64             `json:"plan_revision"`
	IdempotencyKey string            `json:"idempotency_key"`
	Parameters     map[string]string `json:"parameters,omitempty"`
}

type CallbackPayload struct {
	Version int       `json:"version"`
	JobID   string    `json:"job_id"`
	State   string    `json:"state"`
	At      time.Time `json:"at"`
	Code    string    `json:"code,omitempty"`
	Error   string    `json:"error,omitempty"`
}
