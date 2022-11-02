package pool

import (
	"errors"
	"sync"
	"time"
)

var ErrorNoneAvailable = errors.New("none available")

// NewPool returns a pointer to a newly initialised, empty Pool with the given name
func NewPool(name string) *Pool {

	pool := &Pool{
		&sync.RWMutex{},
		*NewDescription(name),
		make(map[string]*Activity),
		make(map[string]int64),
		make(map[string]int64),
		60,
		7200,
		func() int64 { return time.Now().Unix() },
	}

	return pool
}

// getTime is an internal test function to check on time mocking
func (p *Pool) getTime() time.Time {
	return p.Now()
}

// WithNow sets the function which reports the current time (useful in testing)
func (p *Pool) WithNow(now func() time.Time) *Pool {
	p.Lock()
	defer p.Unlock()
	p.Now = now
	return p
}

// WithDescription adds a description to the Pool
func (p *Pool) WithDescription(d shared.Description) *Pool {
	p.Lock()
	defer p.Unlock()
	p.Description = d
	return p
}

// WithID sets the ID for the Pool
func (p *Pool) WithID(id string) *Pool {
	p.Lock()
	defer p.Unlock()
	p.ID = id
	return p
}

// GetID returns the ID of the Pool
func (p *Pool) GetID() string {
	p.Lock()
	defer p.Unlock()
	return p.ID
}

// AddActivity adds a single Activity to a pool
func (p *Pool) AddActivity(activity *Activity) error {

	p.RemoveStaleEntries()

	if activity == nil {
		return errors.New("nil pointer to activity")
	}

	p.Lock()
	defer p.Unlock()

	a := p.Activities
	a[activity.Name] = activity
	p.Activities = a

	return nil

}

// DeleteActivity removes a single Activity from the Pool
func (p *Pool) DeleteActivity(activity *Activity) {
	p.Lock()
	defer p.Unlock()
	act := p.Activities
	delete(act, activity.Name)
	p.Activities = act
}

// GetActivityIDs returns an array containing the names of all activities in the Pool
func (p *Pool) GetActivityNames() []string {

	p.RemoveStaleEntries()

	p.RLock()
	defer p.RUnlock()

	names := []string{}

	for k := range p.Activities {
		names = append(names, k)
	}

	return names

}

// GetEquipmentByID returns a pointer to the Activity of the given ID
func (s *Store) GetEquipmentByName(name string) (*Activity, error) {

	s.RLock()
	defer s.RUnlock()
	a := s.Equipments[id]
	if a == nil {
		return a, errors.New("not found")
	}
	return a, nil
}

// EquipmentExists checks whether an activity of the given ID exists in the pool (returns true if exists)
func (s *Store) EquipmentExists(name string) bool {

	s.RLock()
	defer s.RUnlock()

	_, ok := s.Activities[name]
	return ok
}

// RequestEquipment returns the name, and booking ID for the first free
// Equipment in the pool. The order in which Equipment is used
// is not defined.
// Throws an error if no free activities.
// Most of the time, this should find free equipment, because requests are rate-limited
// by the slot booking system and should only make requests when equipment is reasonably
// expected to be free, so we don't have to worry about the overhead of supplying lots of
// "none available" messages like we did in the previous version that did not have advance booking
func (s *Store) BookEquipment(pool string, when Interval) (string, string, error) {

	p.Lock()
	defer p.Unlock()

	if p, ok := s.Pools[pool]; ok {

		for k := range p.Equipments {
			if q, ok := s.Equipments[k]; ok {
				if q.Available.Allowed(when) { //check available
					uuid, err := q.Bookings.Request(when)
					if err != nil {
						continue //no booking possible, try another
					} else {
						// got booking
						return k, uuid, nil
					}
				}
			}
		}
	}

	return nil, nil, ErrorNoneAvailable

}
