package interval

import (
	"errors"

	"github.com/timdrysdale/interval/diary"
	"github.com/timdrysdale/interval/interval"
)

// Export and reload bookings....

// To reload bookings, need to clear all the diaries, then insert
// bookings.

func (s *Store) CheckBooking(b Booking) (error, []string) {

	msg := []string{}

	if b.Name == "" {
		msg = append(msg, "missing name")
	}

	if b.Policy == "" {
		msg = append(msg, b.Name+" missing policy")
	}
	if b.Slot == "" {
		msg = append(msg, b.Name+" missing slot")
	}
	if b.User == "" {
		msg = append(msg, b.Name+" missing user")
	}
	if (b.When == interval.Interval{
		Start: interval.ZeroTime,
		End:   interval.ZeroTime,
	}) {
		msg = append(msg, "missing when")
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	if _, ok := s.Policies[b.Policy]; !ok {
		msg = append(msg, b.Name+" policy "+b.Policy+" not found")
	}
	if _, ok := s.Slots[b.Slot]; !ok {
		msg = append(msg, b.Name+" slot "+b.Slot+" not found")
	}
	if _, ok := s.Users[b.User]; !ok {
		msg = append(msg, b.Name+" user "+b.User+" not found")
	}

	if len(msg) > 0 {
		return errors.New("missing references"), msg
	}

	return nil, []string{}
}

func (s *Store) ExportBookings() map[string]Booking {

	s.Lock()
	defer s.Unlock()

	bm := make(map[string]Booking)

	for k, v := range s.Bookings {
		bm[k] = *v
	}

	return bm
}

// ReplaceBookings will replace all bookings with a new set
// each booking must be valid for the manifest, i.e. all
// references to other entities must be valid.
// Note that the manifest should be set first
func (s *Store) ReplaceBookings(bm map[string]Booking) (error, []string) {
	s.Lock()
	defer s.Unlock()

	// Check bookings are individually sane given our manifest
	msg := []string{}

	for _, v := range bm {
		err, ms := s.CheckBooking(v)
		if err != nil {
			for _, m := range ms {
				msg = append(msg, m)
			}
		}
	}

	if len(msg) > 0 {
		return errors.New("malformed booking"), msg
	}

	// bookings are ok, so clean house.
	// we want to refund our users, so go through each booking and cancel

	for k, v := range s.Bookings {
		err := s.CancelBooking(*v)
		if err != nil {
			msg = append(msg,
				"could not refund user "+
					v.User+" "+HumaniseDuration(v.When.End.Sub(v.When.Start))+
					" for replaced booking "+k+" on policy "+v.Policy)
		}
	}
	// can't delete bookings as we iterate over map, so just create a fresh map
	s.Bookings = make(map[string]*Booking)

	for k := range s.Resources {
		r := s.Resources[k]
		r.Diary = diary.New(k)
		s.Resources[k] = r
	}

	// Now make the bookings, respecting policy and usage
	for _, v := range bm {
		_, err := s.MakeBookingWithName(v.Policy, v.Slot, v.User, v.When, v.Name)

		if err != nil {
			msg = append(msg, "booking "+v.Name+" failed because "+err.Error())
		}

		// s.Bookings is updated by MakeBookingWithID so we mustn't update it ourselves
	}

	return nil, []string{}
}

func (s *Store) ExportOldBookings() map[string]Booking {
	s.Lock()
	defer s.Unlock()

	bm := make(map[string]Booking)

	for k, v := range s.OldBookings {
		bm[k] = *v
	}

	return bm
}

func (s *Store) ReplaceOldBookings(bm map[string]Booking) (error, []string) {
	s.Lock()
	defer s.Unlock()

	// Check bookings are individually sane given our manifest
	msg := []string{}

	for _, v := range bm {
		err, ms := s.CheckBooking(v)
		if err != nil {
			for _, m := range ms {
				msg = append(msg, m)
			}
		}
	}

	if len(msg) > 0 {
		return errors.New("malformed booking"), msg
	}

	// bookings are ok, so clean house.

	// no need to handle any diaries or cancellations, because these are old bookings
	s.OldBookings = make(map[string]*Booking)

	// Map the bookings
	for k, v := range bm {

		ob := v //make local copy so we can get a pointer detached from the loop variable

		s.OldBookings[k] = &ob

	}

	return nil, []string{}

}

// ExportUsers returns a map of users, listing the names of bookings, old bookings, policies and
// their usage to date by policy name
func (s *Store) ExportUsers() map[string]UserExternal {

	s.Lock()
	defer s.Unlock()

	um := make(map[string]UserExternal)

	for k, v := range s.Users {

		bs := []string{}
		obs := []string{}
		ps := []string{}
		ds := make(map[string]string)

		for k := range v.Bookings {
			bs = append(bs, k)
		}

		for k := range v.OldBookings {
			obs = append(obs, k)
		}
		for k := range v.Policies {
			ps = append(ps, k)
		}
		for k, v := range v.Usage {
			ds[k] = HumaniseDuration(*v)
		}

		um[k] = UserExternal{
			Bookings:    bs,
			OldBookings: obs,
			Policies:    ps,
			Usage:       ds,
		}
	}

	return um
}

// Replace Users is not implemented because it would allow
// the consistency of the store to be broken (e.g. which users
// were associated with which bookings). As for usage, the
// ReplaceBookings method already handles adjustments to usage
// automatically so there is no need to edit users.
// If a user needs more usage allowance, then they need a new policy,
// rather than an adjustment to their old usage value.
func (s *Store) ReplaceUsers(u map[string]UserExternal) (error, []string) {
	return errors.New("not implemented"), []string{}
}
