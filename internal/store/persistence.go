package store

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/practable/book/internal/diary"
)

var (
	ErrBookingConflict    = errors.New("booking overlaps an existing booking")
	ErrBookingIDConflict  = errors.New("booking identifier is already in use")
	ErrMaxBookings        = errors.New("maximum current bookings reached")
	ErrMaxUsage           = errors.New("maximum usage reached")
	ErrStaleManifest      = errors.New("manifest changed while handling the request")
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
	Manifest *PersistentManifest
}

// PersistentManifest identifies the one immutable manifest version currently
// active in the database. Manifest contains only the source configuration; all
// derived filters, diaries, and lookup maps are rebuilt by Store.
type PersistentManifest struct {
	Manifest    Manifest
	Version     int64
	Checksum    string
	ActivatedAt time.Time
}

type CreateBookingRequest struct {
	Booking             Booking
	Resource            string
	ResourceConstrained bool
	Now                 time.Time
	ManifestVersion     int64
	EnforceMaxBookings  bool
	MaxBookings         int64
	EnforceMaxUsage     bool
	MaxUsage            time.Duration
}

// ManifestValidator runs while the repository holds the exclusive maintenance
// lock and before a candidate becomes active. It must not call the repository.
type ManifestValidator func(PersistentState) error

// BookingRepository is deliberately narrower than Store: policy/window domain
// decisions remain in Store, while cross-process state transitions and their
// invariants are owned by the repository transaction.
type BookingRepository interface {
	Load(context.Context) (PersistentState, error)
	ActiveManifestVersion(context.Context) (int64, error)
	CreateBooking(context.Context, CreateBookingRequest) (PersistentBooking, bool, error)
	CancelBooking(context.Context, string, time.Time, string, time.Duration, int64) (PersistentBooking, error)
	StartBooking(context.Context, string, time.Time, int64) (PersistentBooking, error)
	ExpireBookings(context.Context, time.Time) error
	ReplaceBookings(context.Context, []CreateBookingRequest, int64) error
	ReplaceOldBookings(context.Context, []PersistentBooking, int64) error
	ReplaceManifest(context.Context, Manifest, ManifestValidator) (PersistentManifest, error)
	GrantGroup(context.Context, string, string, int64) error
	RevokeGroup(context.Context, string, string, int64) error
	Close()
}

// RefreshManifest reloads authoritative state when another service instance
// has activated a newer manifest. It is also useful to make an administrative
// update visible immediately instead of waiting for the background poll.
func (s *Store) RefreshManifest(ctx context.Context) error {
	s.Lock()
	defer s.Unlock()
	if s.repository == nil {
		return nil
	}
	version, err := s.repository.ActiveManifestVersion(ctx)
	if err != nil {
		return err
	}
	if version == s.manifestVersion {
		return nil
	}
	return s.refreshFromRepositoryLocked(ctx)
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
	if state.Manifest != nil && state.Manifest.Version != s.manifestVersion {
		if err := s.applyManifestLocked(state.Manifest.Manifest); err != nil {
			return fmt.Errorf("apply persistent manifest version %d: %w", state.Manifest.Version, err)
		}
		s.manifestVersion = state.Manifest.Version
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

// validateManifestState replays persisted state into an isolated Store. It is
// intentionally free of repository calls so PostgreSQL can invoke it while
// holding the manifest activation transaction and maintenance lock.
func validateManifestState(manifest Manifest, state PersistentState, now func() time.Time) error {
	candidate := New().WithNow(now)
	if err := candidate.applyManifestLocked(manifest); err != nil {
		return err
	}
	for user, groups := range state.Groups {
		u := NewUser()
		for _, group := range groups {
			u.Groups[group] = true
		}
		candidate.Users[user] = u
	}
	current := make([]PersistentBooking, 0)
	for _, persisted := range state.Bookings {
		if persisted.Current {
			current = append(current, persisted)
			continue
		}
		u := candidate.Users[persisted.Booking.User]
		if u == nil {
			u = NewUser()
			candidate.Users[persisted.Booking.User] = u
		}
		usage := time.Duration(0)
		if existing := u.Usage[persisted.Booking.Policy]; existing != nil {
			usage = *existing
		}
		usage += persisted.UsageCharge
		u.Usage[persisted.Booking.Policy] = &usage
		booking := persisted.Booking
		candidate.OldBookings[booking.Name] = &booking
		u.OldBookings[booking.Name] = &booking
	}
	sort.Slice(current, func(i, j int) bool {
		left, right := current[i].Booking, current[j].Booking
		if left.When.Start.Equal(right.When.Start) {
			return left.Name < right.Name
		}
		return left.When.Start.Before(right.When.Start)
	})
	// Started bookings have already passed admission-time checks. Revalidate
	// their durable references, resource exclusivity, windows, and duration,
	// then include them in aggregate usage before replaying future bookings.
	for _, persisted := range current {
		if !persisted.Booking.Started {
			continue
		}
		booking := persisted.Booking
		slot, ok := candidate.Slots[booking.Slot]
		if !ok || slot.Policy != booking.Policy {
			return fmt.Errorf("persisted booking %s is incompatible: slot or policy changed", booking.Name)
		}
		policy, ok := candidate.Policies[booking.Policy]
		if !ok {
			return fmt.Errorf("persisted booking %s is incompatible: policy %s is absent", booking.Name, booking.Policy)
		}
		if slot.Resource != persisted.Resource {
			return fmt.Errorf("persisted booking %s is incompatible: resource changed from %s to %s", booking.Name, persisted.Resource, slot.Resource)
		}
		if persisted.ResourceConstrained == policy.EnforceUnlimitedUsers {
			return fmt.Errorf("persisted booking %s is incompatible: resource constraint policy changed", booking.Name)
		}
		window, ok := candidate.Filters[slot.Window]
		if !ok || !window.Allowed(booking.When) {
			return fmt.Errorf("persisted booking %s is outside its slot window", booking.Name)
		}
		duration := booking.When.End.Sub(booking.When.Start)
		if policy.EnforceMinDuration && duration < policy.MinDuration {
			return fmt.Errorf("persisted booking %s is shorter than the new minimum duration", booking.Name)
		}
		if policy.EnforceMaxDuration && duration > policy.MaxDuration {
			return fmt.Errorf("persisted booking %s is longer than the new maximum duration", booking.Name)
		}
		resource, ok := candidate.Resources[persisted.Resource]
		if !ok {
			return fmt.Errorf("persisted booking %s references absent resource %s", booking.Name, persisted.Resource)
		}
		if persisted.ResourceConstrained {
			if err := resource.Diary.Request(booking.When, booking.Name); err != nil {
				return fmt.Errorf("persisted booking %s conflicts under the candidate manifest: %w", booking.Name, err)
			}
		}
		u := candidate.Users[booking.User]
		if u == nil {
			u = NewUser()
			candidate.Users[booking.User] = u
		}
		usage := time.Duration(0)
		if existing := u.Usage[booking.Policy]; existing != nil {
			usage = *existing
		}
		usage += persisted.UsageCharge
		u.Usage[booking.Policy] = &usage
		copy := booking
		candidate.Bookings[booking.Name] = &copy
		u.Bookings[booking.Name] = &copy
	}
	for _, persisted := range current {
		booking := persisted.Booking
		if booking.Started {
			continue
		}
		slot, ok := candidate.Slots[booking.Slot]
		if !ok || slot.Policy != booking.Policy {
			return fmt.Errorf("persisted booking %s is incompatible: slot or policy changed", booking.Name)
		}
		policy := candidate.Policies[booking.Policy]
		if slot.Resource != persisted.Resource {
			return fmt.Errorf("persisted booking %s is incompatible: resource changed from %s to %s", booking.Name, persisted.Resource, slot.Resource)
		}
		if persisted.ResourceConstrained == policy.EnforceUnlimitedUsers {
			return fmt.Errorf("persisted booking %s is incompatible: resource constraint policy changed", booking.Name)
		}
		if _, err := candidate.makeBookingWithName(booking.Slot, booking.User, booking.When, booking.Name, false); err != nil {
			return fmt.Errorf("persisted booking %s is incompatible: %w", booking.Name, err)
		}
	}
	for userName, user := range candidate.Users {
		counts := make(map[string]int64)
		for _, booking := range user.Bookings {
			if !booking.When.End.Before(now()) {
				counts[booking.Policy]++
			}
		}
		for policyName, count := range counts {
			policy, ok := candidate.Policies[policyName]
			if ok && policy.EnforceMaxBookings && count > policy.MaxBookings {
				return fmt.Errorf("user %s has %d bookings under policy %s, exceeding the new limit of %d", userName, count, policyName, policy.MaxBookings)
			}
		}
		for policyName, usage := range user.Usage {
			policy, ok := candidate.Policies[policyName]
			if ok && policy.EnforceMaxUsage && *usage > policy.MaxUsage {
				return fmt.Errorf("user %s usage under policy %s exceeds the new limit", userName, policyName)
			}
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
		if _, err := s.repository.CancelBooking(ctx, booking.Name, candidate.CancelledAt, candidate.CancelledBy, usage, s.manifestVersion); err != nil && !errors.Is(err, ErrPersistentNotFound) {
			return err
		}
		changed = true
	}
	if changed {
		return s.refreshFromRepositoryLocked(ctx)
	}
	return nil
}
