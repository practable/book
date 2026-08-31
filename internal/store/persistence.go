package store

import (
	"context"
	"errors"
	"time"

	"github.com/practable/book/internal/diary"
)

var (
	ErrBookingConflict    = errors.New("booking overlaps an existing booking")
	ErrBookingIDConflict  = errors.New("booking identifier is already in use")
	ErrMaxBookings        = errors.New("maximum current bookings reached")
	ErrMaxUsage           = errors.New("maximum usage reached")
	ErrPersistentNotFound = errors.New("booking not found")
)

// PersistentBooking is the durable representation needed to reconstruct the
// in-memory domain projection. Resource is resolved when the booking is made so
// a later manifest change cannot silently weaken resource exclusivity.
type PersistentBooking struct {
	Booking             Booking
	Resource            string
	ResourceConstrained bool
	Current             bool
	UsageCharge         time.Duration
}

type PersistentState struct {
	Bookings []PersistentBooking
	Groups   map[string][]string
}

type CreateBookingRequest struct {
	Booking             Booking
	Resource            string
	ResourceConstrained bool
	Now                 time.Time
	EnforceMaxBookings  bool
	MaxBookings         int64
	EnforceMaxUsage     bool
	MaxUsage            time.Duration
}

// BookingRepository is deliberately narrower than Store: policy/window domain
// decisions remain in Store, while cross-process state transitions and their
// invariants are owned by the repository transaction.
type BookingRepository interface {
	Load(context.Context) (PersistentState, error)
	CreateBooking(context.Context, CreateBookingRequest) (PersistentBooking, bool, error)
	CancelBooking(context.Context, string, time.Time, string, time.Duration) (PersistentBooking, error)
	StartBooking(context.Context, string, time.Time) (PersistentBooking, error)
	ExpireBookings(context.Context, time.Time) error
	ReplaceBookings(context.Context, []CreateBookingRequest) error
	ReplaceOldBookings(context.Context, []PersistentBooking) error
	GrantGroup(context.Context, string, string) error
	RevokeGroup(context.Context, string, string) error
	Close()
}

// WithRepository attaches authoritative persistence and immediately loads any
// state that can be projected with the currently installed manifest.
func (s *Store) WithRepository(repository BookingRepository) error {
	s.Lock()
	defer s.Unlock()
	s.repository = repository
	return s.recoverFromRepositoryLocked(context.Background())
}

func (s *Store) refreshFromRepositoryLocked(ctx context.Context) error {
	if s.repository == nil {
		return nil
	}
	state, err := s.repository.Load(ctx)
	if err != nil {
		return err
	}

	groups := make(map[string]map[string]bool)
	for user, values := range state.Groups {
		if groups[user] == nil {
			groups[user] = make(map[string]bool)
		}
		for _, group := range values {
			groups[user][group] = true
		}
	}

	availability := make(map[string]struct {
		available bool
		reason    string
	})
	for name, resource := range s.Resources {
		if resource.Diary != nil {
			available, reason := resource.Diary.IsAvailable()
			availability[name] = struct {
				available bool
				reason    string
			}{available, reason}
		}
	}
	s.Checker.Clean()
	s.Bookings = make(map[string]*Booking)
	s.OldBookings = make(map[string]*Booking)
	s.Users = make(map[string]*User)
	for user, values := range groups {
		u := NewUser()
		u.Groups = values
		s.Users[user] = u
	}
	for name, resource := range s.Resources {
		resource.Diary = diary.New(name)
		if status, ok := availability[name]; ok && !status.available {
			resource.Diary.SetUnavailable(status.reason)
		}
		s.Resources[name] = resource
	}

	for _, persisted := range state.Bookings {
		booking := persisted.Booking
		u := s.Users[booking.User]
		if u == nil {
			u = NewUser()
			s.Users[booking.User] = u
		}
		usage := time.Duration(0)
		if existing := u.Usage[booking.Policy]; existing != nil {
			usage = *existing
		}
		usage += persisted.UsageCharge
		u.Usage[booking.Policy] = &usage

		if persisted.Current {
			s.Bookings[booking.Name] = &booking
			u.Bookings[booking.Name] = &booking
			if persisted.ResourceConstrained {
				resource, ok := s.Resources[persisted.Resource]
				if len(s.Resources) == 0 {
					continue
				}
				if !ok || resource.Diary == nil {
					return errors.New("persisted booking " + booking.Name + " references resource " + persisted.Resource + " absent from the manifest")
				}
				if err := resource.Diary.Request(booking.When, booking.Name); err != nil {
					return errors.New("could not restore booking " + booking.Name + ": " + err.Error())
				}
			}
			if policy, ok := s.Policies[booking.Policy]; ok && policy.EnforceGracePeriod && !booking.Started {
				_ = s.Checker.Push(booking.When.Start.Add(policy.GracePeriod), booking.Name)
			}
		} else {
			s.OldBookings[booking.Name] = &booking
			u.OldBookings[booking.Name] = &booking
		}
	}
	return nil
}

func (s *Store) expireAndRefreshLocked(ctx context.Context) error {
	if s.repository == nil {
		return nil
	}
	if err := s.repository.ExpireBookings(ctx, s.now()); err != nil {
		return err
	}
	return s.recoverFromRepositoryLocked(ctx)
}

func (s *Store) recoverFromRepositoryLocked(ctx context.Context) error {
	if err := s.refreshFromRepositoryLocked(ctx); err != nil {
		return err
	}
	if s.repository == nil || len(s.Policies) == 0 {
		return nil
	}
	changed := false
	for _, booking := range s.Bookings {
		policy, ok := s.Policies[booking.Policy]
		if !ok || !policy.EnforceGracePeriod || booking.Started || !booking.When.Start.Add(policy.GracePeriod).Before(s.now()) {
			continue
		}
		candidate := *booking
		candidate.Cancelled = true
		candidate.CancelledAt = s.now()
		candidate.CancelledBy = "auto-grace-check"
		usage, err := calculateUsage(candidate, policy)
		if err != nil {
			return err
		}
		if _, err := s.repository.CancelBooking(ctx, booking.Name, candidate.CancelledAt, candidate.CancelledBy, usage); err != nil && !errors.Is(err, ErrPersistentNotFound) {
			return err
		}
		changed = true
	}
	if changed {
		return s.refreshFromRepositoryLocked(ctx)
	}
	return nil
}
