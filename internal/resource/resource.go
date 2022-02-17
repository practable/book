// package resource holds non-overlapping bookings with arbitrary durations
package resource

import (
	avl "interval/internal/trees/avltree"

	"interval/internal/interval"

	"github.com/google/uuid"
)

type Resource struct {
	bookings *avl.Tree
}

func New() *Resource {
	return &Resource{
		bookings: avl.NewWith(interval.Comparator),
	}
}

func (r *Resource) Request(when interval.Interval) (uuid.UUID, error) {

	u := uuid.New()

	_, err := r.bookings.Put(when, u)

	return u, err

}
