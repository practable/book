package store

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/practable/book/internal/interval"
)

// CalendarSelector identifies a fungible policy pool, optionally narrowed to
// one resource or a set of structured resource properties.
type CalendarSelector struct {
	Policy     string            `json:"policy" yaml:"policy"`
	Resource   string            `json:"resource,omitempty" yaml:"resource,omitempty"`
	Properties map[string]string `json:"properties,omitempty" yaml:"properties,omitempty"`
}

type CalendarResource struct {
	Name       string            `json:"name" yaml:"name"`
	Class      string            `json:"class,omitempty" yaml:"class,omitempty"`
	Properties map[string]string `json:"properties,omitempty" yaml:"properties,omitempty"`
}

type CalendarCatalogueItem struct {
	Policy              string             `json:"policy" yaml:"policy"`
	Description         Description        `json:"description" yaml:"description"`
	RecommendedDuration time.Duration      `json:"recommended_duration" yaml:"recommended_duration"`
	BookingIncrement    time.Duration      `json:"booking_increment" yaml:"booking_increment"`
	Resources           []CalendarResource `json:"resources" yaml:"resources"`
}

type CalendarAvailabilityBand struct {
	Start             time.Time `json:"start" yaml:"start"`
	End               time.Time `json:"end" yaml:"end"`
	MatchingResources int64     `json:"matching_resources" yaml:"matching_resources"`
	Bookable          bool      `json:"bookable" yaml:"bookable"`
	Reason            string    `json:"reason,omitempty" yaml:"reason,omitempty"`
	OperatingMode     string    `json:"operating_mode" yaml:"operating_mode"`
}

type CalendarPreview struct {
	When               interval.Interval `json:"when" yaml:"when"`
	Selector           CalendarSelector  `json:"selector" yaml:"selector"`
	MatchingResources  []string          `json:"matching_resources" yaml:"matching_resources"`
	Bookable           bool              `json:"bookable" yaml:"bookable"`
	Reasons            []string          `json:"reasons,omitempty" yaml:"reasons,omitempty"`
	UsageAfter         time.Duration     `json:"usage_after" yaml:"usage_after"`
	UsageLimit         time.Duration     `json:"usage_limit,omitempty" yaml:"usage_limit,omitempty"`
	ManifestVersion    int64             `json:"manifest_version" yaml:"manifest_version"`
	OperationalEffects []string          `json:"operational_effects" yaml:"operational_effects"`
}

func resourceMatches(resourceName string, resource Resource, selector CalendarSelector) bool {
	if selector.Resource != "" && selector.Resource != resourceName {
		return false
	}
	for key, value := range selector.Properties {
		if resource.Properties[key] != value {
			return false
		}
	}
	return true
}

func (s *Store) calendarSlotsLocked(selector CalendarSelector) ([]string, error) {
	policy, ok := s.Policies[selector.Policy]
	if !ok {
		return nil, fmt.Errorf("policy %s not found", selector.Policy)
	}
	result := make([]string, 0, len(policy.Slots))
	seenResources := make(map[string]bool)
	for _, slotName := range policy.Slots {
		slot, ok := s.Slots[slotName]
		if !ok {
			continue
		}
		resource, ok := s.Resources[slot.Resource]
		if !ok || seenResources[slot.Resource] || !resourceMatches(slot.Resource, resource, selector) {
			continue
		}
		seenResources[slot.Resource] = true
		result = append(result, slotName)
	}
	sort.Strings(result)
	if len(result) == 0 {
		return nil, errors.New("selector matches no resources")
	}
	return result, nil
}

func intervalContained(available []interval.Interval, requested interval.Interval) bool {
	for _, candidate := range available {
		if !requested.Start.Before(candidate.Start) && !requested.End.After(candidate.End) {
			return true
		}
	}
	return false
}

func (s *Store) matchingCalendarResourcesLocked(selector CalendarSelector, when interval.Interval) ([]string, []string, error) {
	slots, err := s.calendarSlotsLocked(selector)
	if err != nil {
		return nil, nil, err
	}
	resources := make([]string, 0, len(slots))
	matchingSlots := make([]string, 0, len(slots))
	for _, slotName := range slots {
		available, err := s.getAvailability(slotName)
		if err != nil || (!intervalContained(available, when) && !s.bookableByReclaimingOperationsLocked(slotName, when)) {
			continue
		}
		slot := s.Slots[slotName]
		resources = append(resources, slot.Resource)
		matchingSlots = append(matchingSlots, slotName)
	}
	return resources, matchingSlots, nil
}

func (s *Store) bookableByReclaimingOperationsLocked(slotName string, when interval.Interval) bool {
	slot, ok := s.Slots[slotName]
	if !ok {
		return false
	}
	window, ok := s.Filters[slot.Window]
	if !ok || !window.Allowed(when) {
		return false
	}
	available, _, err := s.getSlotIsAvailable(slotName)
	if err != nil || !available {
		return false
	}
	for _, existing := range s.Bookings {
		if existing.Cancelled || interval.Comparator(existing.When, when) != 0 {
			continue
		}
		existingSlot, ok := s.Slots[existing.Slot]
		if !ok || existingSlot.Resource != slot.Resource {
			continue
		}
		if existing.Policy != "__operations_reclaimable__" {
			return false
		}
	}
	return true
}

// GetCalendarCatalogue returns one entry per policy in a group. Policies are
// the existing manifest boundary for a fungible set of slots, so this adds no
// parallel experiment-class registry.
func (s *Store) GetCalendarCatalogue(groupName string) ([]CalendarCatalogueItem, error) {
	s.Lock()
	defer s.Unlock()
	if err := s.expireAndRefreshLocked(context.Background()); err != nil {
		return nil, err
	}
	group, ok := s.Groups[groupName]
	if !ok {
		return nil, fmt.Errorf("group %s not found", groupName)
	}
	items := make([]CalendarCatalogueItem, 0, len(group.Policies))
	for _, policyName := range group.Policies {
		policy, ok := s.Policies[policyName]
		if !ok {
			continue
		}
		description := s.Descriptions[policy.Description]
		item := CalendarCatalogueItem{Policy: policyName, Description: description, BookingIncrement: time.Minute}
		for _, guideName := range policy.DisplayGuides {
			if guide, ok := s.DisplayGuides[guideName]; ok && (item.RecommendedDuration == 0 || guide.Duration < item.RecommendedDuration) {
				item.RecommendedDuration = guide.Duration
			}
		}
		seen := make(map[string]bool)
		for _, slotName := range policy.Slots {
			slot, ok := s.Slots[slotName]
			if !ok || seen[slot.Resource] {
				continue
			}
			resource, ok := s.Resources[slot.Resource]
			if !ok {
				continue
			}
			seen[slot.Resource] = true
			item.Resources = append(item.Resources, CalendarResource{Name: slot.Resource, Class: resource.Class, Properties: resource.Properties})
		}
		sort.Slice(item.Resources, func(i, j int) bool { return item.Resources[i].Name < item.Resources[j].Name })
		items = append(items, item)
	}
	return items, nil
}

func (s *Store) GetCalendarAvailability(selector CalendarSelector, from, to time.Time, duration, resolution time.Duration) ([]CalendarAvailabilityBand, error) {
	s.Lock()
	defer s.Unlock()
	if err := s.expireAndRefreshLocked(context.Background()); err != nil {
		return nil, err
	}
	if !to.After(from) || duration <= 0 || resolution <= 0 {
		return nil, errors.New("invalid calendar range, duration, or resolution")
	}
	if to.Sub(from) > 62*24*time.Hour || resolution < time.Minute {
		return nil, errors.New("calendar query exceeds range or resolution limit")
	}
	result := make([]CalendarAvailabilityBand, 0, int(to.Sub(from)/resolution)+1)
	for start := from.UTC(); start.Before(to); start = start.Add(resolution) {
		end := start.Add(duration)
		resources, _, err := s.matchingCalendarResourcesLocked(selector, interval.Interval{Start: start, End: end})
		if err != nil {
			return nil, err
		}
		band := CalendarAvailabilityBand{Start: start, End: start.Add(resolution), MatchingResources: int64(len(resources)), Bookable: len(resources) > 0, OperatingMode: "normal"}
		if !band.Bookable {
			band.Reason = "unavailable"
		}
		result = append(result, band)
	}
	return result, nil
}

func (s *Store) PreviewCalendarBooking(user string, selector CalendarSelector, when interval.Interval) (CalendarPreview, error) {
	s.Lock()
	defer s.Unlock()
	if err := s.expireAndRefreshLocked(context.Background()); err != nil {
		return CalendarPreview{}, err
	}
	preview := CalendarPreview{When: when, Selector: selector, ManifestVersion: s.manifestVersion, OperationalEffects: []string{}}
	policy, ok := s.Policies[selector.Policy]
	if !ok {
		return preview, fmt.Errorf("policy %s not found", selector.Policy)
	}
	if !when.End.After(when.Start) {
		preview.Reasons = append(preview.Reasons, "invalid_interval")
		return preview, nil
	}
	userState, ok := s.Users[user]
	if !ok {
		preview.Reasons = append(preview.Reasons, "user_not_found")
		return preview, nil
	}
	usage := time.Duration(0)
	if tracker := userState.Usage[selector.Policy]; tracker != nil {
		usage = *tracker
	}
	preview.UsageAfter = usage + when.End.Sub(when.Start)
	if policy.EnforceMaxUsage {
		preview.UsageLimit = policy.MaxUsage
		if preview.UsageAfter > policy.MaxUsage {
			preview.Reasons = append(preview.Reasons, "usage_limit")
		}
	}
	if policy.EnforceMinDuration && when.End.Sub(when.Start) < policy.MinDuration {
		preview.Reasons = append(preview.Reasons, "minimum_duration")
	}
	if policy.EnforceMaxDuration && when.End.Sub(when.Start) > policy.MaxDuration {
		preview.Reasons = append(preview.Reasons, "maximum_duration")
	}
	if policy.EnforceMaxBookings {
		current := int64(0)
		for _, booking := range userState.Bookings {
			if booking.Policy == selector.Policy {
				current++
			}
		}
		if current >= policy.MaxBookings {
			preview.Reasons = append(preview.Reasons, "booking_count_limit")
		}
	}
	now := s.now()
	earliestStart := now
	if policy.EnforceAllowStartInPast {
		earliestStart = earliestStart.Add(-policy.AllowStartInPastWithin)
	}
	if when.Start.Before(earliestStart) {
		preview.Reasons = append(preview.Reasons, "starts_in_past")
	}
	if policy.EnforceBookAhead && when.End.After(now.Add(policy.BookAhead)) {
		preview.Reasons = append(preview.Reasons, "book_ahead_limit")
	}
	if policy.EnforceStartsWithin && when.Start.After(now.Add(policy.StartsWithin)) {
		preview.Reasons = append(preview.Reasons, "starts_within_limit")
	}
	resources, _, err := s.matchingCalendarResourcesLocked(selector, when)
	if err != nil {
		return preview, err
	}
	preview.MatchingResources = resources
	if len(resources) == 0 {
		preview.Reasons = append(preview.Reasons, "unavailable")
	}
	if s.Locked {
		preview.Reasons = append(preview.Reasons, "booking_creation_paused")
	}
	preview.Bookable = len(preview.Reasons) == 0
	return preview, nil
}

// MakeCalendarBooking transactionally assigns one matching slot. The caller's
// idempotency key becomes a stable opaque booking identifier scoped to the user.
func (s *Store) MakeCalendarBooking(user string, selector CalendarSelector, when interval.Interval, idempotencyKey string) (Booking, error) {
	s.Lock()
	defer s.Unlock()
	if strings.TrimSpace(idempotencyKey) == "" {
		return Booking{}, errors.New("idempotency key is required")
	}
	if s.Locked {
		return Booking{}, errors.New("booking creation paused: " + s.Message)
	}
	if err := s.expireAndRefreshLocked(context.Background()); err != nil {
		return Booking{}, err
	}
	name := uuid.NewSHA1(uuid.NameSpaceURL, []byte(user+"\x00"+idempotencyKey)).String()
	if existing := s.Bookings[name]; existing != nil {
		if existing.User == user && existing.Policy == selector.Policy && existing.When == when {
			return *existing, nil
		}
		return Booking{}, ErrBookingIDConflict
	}
	_, slots, err := s.matchingCalendarResourcesLocked(selector, when)
	if err != nil {
		return Booking{}, err
	}
	if len(slots) == 0 {
		return Booking{}, ErrBookingConflict
	}
	var lastErr error
	for _, slotName := range slots {
		booking, err := s.makeBookingWithName(slotName, user, when, name, true)
		if err == nil {
			return booking, nil
		}
		lastErr = err
		if errors.Is(err, ErrMaxBookings) || errors.Is(err, ErrMaxUsage) {
			break
		}
	}
	if lastErr == nil {
		lastErr = ErrBookingConflict
	}
	return Booking{}, lastErr
}

// RescheduleCalendarBooking moves or resizes an unstarted booking using an
// optimistic revision. The replacement is revalidated against the current
// manifest, policy usage, and database overlap constraints.
func (s *Store) RescheduleCalendarBooking(user, name string, revision int64, selector CalendarSelector, when interval.Interval) (EditableBooking, error) {
	current, err := s.GetBookingForEdit(name)
	if err != nil {
		return EditableBooking{}, err
	}
	if current.Booking.User != user {
		return EditableBooking{}, errors.New("booking is not owned by this user")
	}
	if current.Revision != revision {
		return EditableBooking{}, ErrBookingRevision
	}
	s.Lock()
	slots, err := s.calendarSlotsLocked(selector)
	s.Unlock()
	if err != nil {
		return EditableBooking{}, err
	}
	var lastErr error
	for _, slotName := range slots {
		replacement := current.Booking
		replacement.Policy = selector.Policy
		replacement.Slot = slotName
		replacement.When = when
		updated, err := s.replaceBooking(EditableBooking{OriginalName: name, Revision: revision, Booking: replacement}, "user:"+user)
		if err == nil {
			return updated, nil
		}
		lastErr = err
		if errors.Is(err, ErrBookingRevision) || errors.Is(err, ErrBookingStarted) || errors.Is(err, ErrMaxBookings) || errors.Is(err, ErrMaxUsage) {
			break
		}
	}
	if lastErr == nil {
		lastErr = ErrBookingConflict
	}
	return EditableBooking{}, lastErr
}
