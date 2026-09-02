package operations

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/practable/book/internal/webhook"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type dispatchRepo struct {
	delivery  Delivery
	completed bool
	success   bool
	status    int
	now       time.Time
	next      time.Time
}

func (r *dispatchRepo) CreateJob(context.Context, Job, Delivery) (Job, bool, error) { panic("unused") }
func (r *dispatchRepo) ClaimDeliveries(context.Context, string, time.Time, time.Duration, int) ([]Delivery, error) {
	return []Delivery{r.delivery}, nil
}
func (r *dispatchRepo) CompleteDelivery(_ context.Context, _ string, _ string, success bool, status int, _ string, now, next time.Time) error {
	r.completed, r.success, r.status = true, success, status
	r.now, r.next = now, next
	return nil
}
func (r *dispatchRepo) ApplyCallback(context.Context, Callback, string) (Job, bool, error) {
	panic("unused")
}
func (r *dispatchRepo) GetJob(context.Context, string) (Job, error) { panic("unused") }

func TestDispatcherSignsExactBody(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	secret := []byte("01234567890123456789012345678901")
	body := []byte(`{"job_id":"job-1"}`)
	repo := &dispatchRepo{delivery: Delivery{ID: "delivery-1", JobID: "job-1", Body: body, Attempts: 1}}
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got, _ := io.ReadAll(req.Body)
		err := webhook.Verify(secret, webhook.DirectionCommand, webhook.Headers{
			Timestamp: req.Header.Get(webhook.HeaderTimestamp), DeliveryID: req.Header.Get(webhook.HeaderDeliveryID),
			Direction: req.Header.Get(webhook.HeaderDirection), Signature: req.Header.Get(webhook.HeaderSignature),
		}, now, time.Minute, got)
		if err != nil {
			t.Fatalf("signature: %v", err)
		}
		if string(got) != string(body) {
			t.Fatalf("body changed: %s", got)
		}
		return &http.Response{StatusCode: http.StatusAccepted, Body: io.NopCloser(strings.NewReader("accepted")), Header: make(http.Header)}, nil
	})}
	d := Dispatcher{Repository: repo, Endpoint: "https://runner.example/jobs", Secret: secret, Client: client, Owner: "book-1", Now: func() time.Time { return now }}
	if n, err := d.DispatchOnce(context.Background()); err != nil || n != 1 {
		t.Fatalf("DispatchOnce = %d, %v", n, err)
	}
	if !repo.completed || !repo.success || repo.status != http.StatusAccepted {
		t.Fatalf("completion = %#v", repo)
	}
}

func TestRetryDelayIsCapped(t *testing.T) {
	if got := retryDelay(99); got != 5*time.Minute {
		t.Fatalf("retryDelay = %s", got)
	}
}

func TestDispatcherSchedulesFailedDeliveryAfterCurrentTime(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	repo := &dispatchRepo{delivery: Delivery{ID: "delivery-1", JobID: "job-1", Body: []byte(`{}`), Attempts: 1}}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader("unavailable")), Header: make(http.Header)}, nil
	})}
	dispatcher := Dispatcher{Repository: repo, Endpoint: "https://runner.example/jobs", Secret: []byte("01234567890123456789012345678901"), Client: client, Owner: "book-1", Now: func() time.Time { return now }}

	if _, err := dispatcher.DispatchOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if repo.success || !repo.now.Equal(now) || !repo.next.Equal(now.Add(time.Second)) {
		t.Fatalf("failed delivery completion has now=%s next=%s success=%v", repo.now, repo.next, repo.success)
	}
}
