package store

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/practable/book/internal/interval"
)

// BookingQuery is the bounded, read-only projection used by operational tools.
type BookingQuery struct {
	Resource string
	Slot     string
	Policy   string
	User     string
	State    string
	From     *time.Time
	To       *time.Time
	Limit    int
}

type BookingRecord struct {
	Booking     Booking
	Resource    string
	Collection  string
	ActualUsage time.Duration
}

type BookingEvent struct {
	BookingName string
	Type        string
	OccurredAt  time.Time
	Actor       string
}

type BookingEventReader interface {
	ListBookingEvents(context.Context, string) ([]BookingEvent, error)
}

type UsageQuery struct {
	Resource string
	Slot     string
	Policy   string
	User     string
	From     *time.Time
	To       *time.Time
}

type OperationalUsageReader interface {
	GetOperationalUsageSummary(context.Context, UsageQuery) (time.Duration, time.Duration, time.Duration, int64, error)
}

type OperationalAlert struct {
	ID              int64
	Resource        string
	Stream          string
	Code            string
	Message         string
	JobID           string
	ManifestVersion int64
	Status          string
	Occurrences     int64
	FirstSeen       time.Time
	LastSeen        time.Time
	AcknowledgedAt  *time.Time
	AcknowledgedBy  string
	ResolvedAt      *time.Time
	ResolvedBy      string
}

type OperationalStreamHealth struct {
	Resource        string
	Stream          string
	Status          string
	Code            string
	Message         string
	JobID           string
	ManifestVersion int64
	CheckedAt       time.Time
}

type ResourceHold struct {
	Resource  string
	Reason    string
	HeldSince time.Time
	HeldBy    string
}

type ResourceHoldRepository interface {
	ListResourceHolds(context.Context) ([]ResourceHold, error)
	SetResourceAvailabilityBy(context.Context, string, bool, string, string, int64) error
}

type OperationalAlertRepository interface {
	ListOperationalAlerts(context.Context, string, int) ([]OperationalAlert, error)
	ListOperationalStreamHealth(context.Context) ([]OperationalStreamHealth, error)
	SetOperationalAlertStatus(context.Context, int64, string, string, time.Time) (OperationalAlert, error)
}

func (s *Store) ListOperationalAlerts(ctx context.Context, status string, limit int) ([]OperationalAlert, error) {
	s.RLock()
	repository, ok := s.repository.(OperationalAlertRepository)
	s.RUnlock()
	if !ok {
		return nil, errors.New("operational alerts require durable persistence")
	}
	return repository.ListOperationalAlerts(ctx, status, limit)
}

func (s *Store) ListOperationalStreamHealth(ctx context.Context) ([]OperationalStreamHealth, error) {
	s.RLock()
	repository, ok := s.repository.(OperationalAlertRepository)
	s.RUnlock()
	if !ok {
		return nil, errors.New("operational health requires durable persistence")
	}
	return repository.ListOperationalStreamHealth(ctx)
}

func (s *Store) SetOperationalAlertStatus(ctx context.Context, id int64, status, actor string) (OperationalAlert, error) {
	s.RLock()
	repository, ok := s.repository.(OperationalAlertRepository)
	now := s.now()
	s.RUnlock()
	if !ok {
		return OperationalAlert{}, errors.New("operational alerts require durable persistence")
	}
	return repository.SetOperationalAlertStatus(ctx, id, status, actor, now)
}

func (s *Store) ListResourceHolds(ctx context.Context) ([]ResourceHold, error) {
	s.RLock()
	repository, ok := s.repository.(ResourceHoldRepository)
	s.RUnlock()
	if !ok {
		return nil, errors.New("resource hold audit requires durable persistence")
	}
	return repository.ListResourceHolds(ctx)
}

func (s *Store) GetFilteredUsageSummary(query UsageQuery) UsageSummary {
	result, _ := s.GetFilteredUsageSummaryPersistent(query)
	return result
}

// GetFilteredUsageSummaryPersistent returns database-backed operational usage
// errors to callers that must not silently publish an incomplete report.
func (s *Store) GetFilteredUsageSummaryPersistent(query UsageQuery) (UsageSummary, error) {
	s.Lock()
	defer s.Unlock()
	now := s.now()
	result := UsageSummary{}
	for _, bookings := range []map[string]*Booking{s.Bookings, s.OldBookings} {
		for _, booking := range bookings {
			slot := s.Slots[booking.Slot]
			if query.Resource != "" && query.Resource != slot.Resource || query.Slot != "" && query.Slot != booking.Slot || query.Policy != "" && query.Policy != booking.Policy || query.User != "" && query.User != booking.User || booking.StartedAt == "" {
				continue
			}
			started, err := time.Parse(time.RFC3339Nano, booking.StartedAt)
			if err != nil {
				continue
			}
			ended := booking.When.End
			if booking.Cancelled && !booking.CancelledAt.IsZero() && booking.CancelledAt.Before(ended) {
				ended = booking.CancelledAt
			}
			if now.Before(ended) {
				ended = now
			}
			if query.From != nil && started.Before(*query.From) {
				started = *query.From
			}
			if query.To != nil && ended.After(*query.To) {
				ended = *query.To
			}
			if !ended.After(started) {
				continue
			}
			result.StartedBookings++
			if booking.Cancelled || !booking.When.End.After(now) {
				result.CompletedBookings++
			}
			result.ActualUsage += ended.Sub(started)
		}
	}
	if reader, ok := s.repository.(OperationalUsageReader); ok {
		preparation, cleanup, scheduled, jobs, err := reader.GetOperationalUsageSummary(context.Background(), query)
		if err != nil {
			return UsageSummary{}, err
		}
		result.PreparationUsage, result.CleanupUsage, result.ScheduledUsage, result.OperationalJobs = preparation, cleanup, scheduled, jobs
	}
	return result, nil
}

func bookingMatchesQuery(booking *Booking, resource, collection string, query BookingQuery) bool {
	if query.Resource != "" && query.Resource != resource || query.Slot != "" && query.Slot != booking.Slot || query.Policy != "" && query.Policy != booking.Policy || query.User != "" && query.User != booking.User {
		return false
	}
	switch query.State {
	case "", "all":
	case "current":
		if collection != "current" {
			return false
		}
	case "history":
		if collection != "history" {
			return false
		}
	case "started":
		if !booking.Started || booking.Cancelled {
			return false
		}
	case "cancelled":
		if !booking.Cancelled {
			return false
		}
	case "unfulfilled":
		if !booking.Unfulfilled {
			return false
		}
	default:
		return false
	}
	if query.From != nil && !booking.When.End.After(*query.From) {
		return false
	}
	if query.To != nil && !booking.When.Start.Before(*query.To) {
		return false
	}
	return true
}

func (s *Store) QueryBookings(query BookingQuery) ([]BookingRecord, error) {
	s.Lock()
	defer s.Unlock()
	if err := s.expireAndRefreshLocked(context.Background()); err != nil {
		return nil, err
	}
	if query.Limit <= 0 || query.Limit > 1000 {
		query.Limit = 200
	}
	if query.State != "" && query.State != "all" && query.State != "current" && query.State != "history" && query.State != "started" && query.State != "cancelled" && query.State != "unfulfilled" {
		return nil, errors.New("invalid booking state filter")
	}
	now := s.now()
	result := make([]BookingRecord, 0)
	for collection, bookings := range map[string]map[string]*Booking{"current": s.Bookings, "history": s.OldBookings} {
		for _, booking := range bookings {
			slot, ok := s.Slots[booking.Slot]
			resource := ""
			if ok {
				resource = slot.Resource
			}
			if bookingMatchesQuery(booking, resource, collection, query) {
				result = append(result, BookingRecord{Booking: *booking, Resource: resource, Collection: collection, ActualUsage: booking.ActualUsage(now)})
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Booking.When.Start.Equal(result[j].Booking.When.Start) {
			return result[i].Booking.Name < result[j].Booking.Name
		}
		return result[i].Booking.When.Start.Before(result[j].Booking.When.Start)
	})
	if len(result) > query.Limit {
		result = result[:query.Limit]
	}
	return result, nil
}

func (s *Store) GetBookingEvents(name string) ([]BookingEvent, error) {
	s.Lock()
	repository := s.repository
	s.Unlock()
	reader, ok := repository.(BookingEventReader)
	if !ok {
		return nil, errors.New("booking audit history requires persistent storage")
	}
	return reader.ListBookingEvents(context.Background(), name)
}

// MakeMaintenanceBookingForResource lets the store choose a stable slot for a
// physical resource, avoiding slot knowledge in maintenance dashboards.
func (s *Store) MakeMaintenanceBookingForResource(resource, operator string, when interval.Interval) (Booking, error) {
	s.Lock()
	slots := make([]string, 0)
	for name, slot := range s.Slots {
		if slot.Resource == resource {
			slots = append(slots, name)
		}
	}
	s.Unlock()
	sort.Strings(slots)
	if len(slots) == 0 {
		return Booking{}, errors.New("resource " + resource + " not found or has no slot")
	}
	return s.MakeMaintenanceBooking(slots[0], operator, when)
}
