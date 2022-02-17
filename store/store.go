// package store holds bookings with arbitrary durations
package store

import (
	"errors"
	"time"

	"interval/internal/resource"

	"interval/internal/interval"

	"github.com/google/uuid"
)

var errNotFound = errors.New("resource not found")

type Interval struct {
	Start time.Time
	End   time.Time
}

type Store struct {
	Resources map[uuid.UUID]*resource.Resource
}

type Booking struct {
	When Interval
	ID   uuid.UUID
}

func New() *Store {
	return &Store{
		Resources: make(map[uuid.UUID]*resource.Resource),
	}
}

func (s *Store) Add() uuid.UUID {

	u := uuid.New()

	r := resource.New()

	s.Resources[u] = r

	return u
}

func (s *Store) ClearBeforeAll(t time.Time) {

	for _, r := range s.Resources {
		r.ClearBefore(t)
	}
}

func (s *Store) Request(rID uuid.UUID, when Interval) (uuid.NullUUID, error) {

	nu := uuid.NullUUID{}

	if r, ok := s.Resources[rID]; ok {

		u, err := r.Request(interval.Interval{
			Start: when.Start,
			End:   when.End,
		})

		if err != nil {
			return nu, err
		}

		nu.UUID = u
		nu.Valid = true
		return nu, nil

	}

	return nu, errNotFound
}

func (s *Store) Cancel(rID uuid.UUID, bID uuid.UUID) error {

	if r, ok := s.Resources[rID]; ok {

		return r.Delete(bID)

	}

	return errNotFound

}

func (s *Store) GetBookings(rID uuid.UUID) ([]Booking, error) {

	bookings := []Booking{}

	if r, ok := s.Resources[rID]; ok {

		bb, err := r.GetBookings()

		if err != nil {
			return bookings, err
		}

		for _, b := range bb {
			bookings = append(bookings,
				Booking{
					When: Interval{
						Start: b.When.Start,
						End:   b.When.End,
					},
					ID: b.ID,
				})
		}

		return bookings, nil

	}

	return bookings, errNotFound

}
