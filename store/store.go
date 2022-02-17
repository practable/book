// package store holds bookings with arbitrary durations
package store

import (
	avl "interval/internal/trees/avltree"

	"interval/internal/interval"
)

type Resource struct {
	bookings *avl.Tree
}

func New() *Store {

	return &Store{
		bookings: avl.NewWith(interval.Comparator),
	}

}
