package operations

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/practable/book/internal/webhook"
)

const maxResponseBody = 4096

type Dispatcher struct {
	Repository  Repository
	Endpoint    string
	Secret      []byte
	Client      *http.Client
	Owner       string
	Now         func() time.Time
	Lease       time.Duration
	BatchSize   int
	MaxAttempts int
}

// Run continuously drains durable deliveries until the context is cancelled.
func (d *Dispatcher) Run(ctx context.Context, pollEvery time.Duration) {
	if pollEvery <= 0 {
		pollEvery = time.Second
	}
	ticker := time.NewTicker(pollEvery)
	defer ticker.Stop()
	for {
		_, _ = d.DispatchOnce(ctx) // retry state remains durable in the repository
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// DispatchOnce claims and attempts one batch. The database lease permits many
// book instances to run this safely; a delivery ID remains stable across retries.
func (d *Dispatcher) DispatchOnce(ctx context.Context) (int, error) {
	if d.Repository == nil || d.Endpoint == "" || len(d.Secret) != 32 || d.Owner == "" {
		return 0, fmt.Errorf("dispatcher is not fully configured")
	}
	now := time.Now().UTC()
	if d.Now != nil {
		now = d.Now().UTC()
	}
	lease := d.Lease
	if lease <= 0 {
		lease = 30 * time.Second
	}
	batch := d.BatchSize
	if batch <= 0 {
		batch = 10
	}
	maxAttempts := d.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 12
	}
	client := d.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	deliveries, err := d.Repository.ClaimDeliveries(ctx, d.Owner, now, lease, batch)
	if err != nil {
		return 0, err
	}
	for _, delivery := range deliveries {
		headers, signErr := webhook.Sign(d.Secret, webhook.DirectionCommand, delivery.ID, now, delivery.Body)
		if signErr != nil {
			return 0, signErr
		}
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, d.Endpoint, bytes.NewReader(delivery.Body))
		if reqErr != nil {
			return 0, reqErr
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(webhook.HeaderTimestamp, headers.Timestamp)
		req.Header.Set(webhook.HeaderDeliveryID, headers.DeliveryID)
		req.Header.Set(webhook.HeaderDirection, headers.Direction)
		req.Header.Set(webhook.HeaderSignature, headers.Signature)

		status, message, delivered := 0, "", false
		resp, sendErr := client.Do(req)
		if sendErr != nil {
			message = sendErr.Error()
		} else {
			status = resp.StatusCode
			body, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
			_ = resp.Body.Close()
			message = string(body)
			delivered = status >= 200 && status < 300
		}
		next := now.Add(retryDelay(delivery.Attempts))
		if delivery.Attempts >= maxAttempts {
			next = now
		}
		if err := d.Repository.CompleteDelivery(ctx, delivery.ID, d.Owner, delivered, status, message, now, next); err != nil {
			return len(deliveries), err
		}
	}
	return len(deliveries), nil
}

func retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	shift := attempt - 1
	if shift > 9 {
		shift = 9
	}
	delay := time.Second << shift
	if delay > 5*time.Minute {
		return 5 * time.Minute
	}
	return delay
}
