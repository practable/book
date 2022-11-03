package interval

import (
	"errors"
	"strconv"
)

// Named represents a reduced form of our other structs that have a name
// field, so we can access that field from an interface
// and reduce code repetition checking for duplicate names
type Named struct {
	Name string
}

// Manifest represents all the available equipment and how to access it
// Slots are the primary entities, so reference checking starts with them
type Manifest struct {
	Descriptions []Description
	Policies     []Policy
	Resources    []Resource
	Slots        []Slot
	Streams      []Stream
	UIs          []UI
	UISets       []UISet
}

/*
// Populate assumes an empty store i.e. does not retain existing elements
// from any previous manifests, but it does retain bookings.
func (s *Store) Populate(m Manifest) {
	for _, d := range m.Descriptions {
		s.Descriptions[d.Name] = d
	}
}
*/

func CheckDescriptions(items []Description) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "Unnamed member #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "Duplicate named member #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("Issues found"), msg
	}

	return nil, []string{}

}

/*
// CheckManifest checks for internal consistency, throwing an error
// if there are any unresolved references by name
func CheckManifest(m Manifest) (error, []string) {
	// check if any elements have duplicate names
	msg := []string{}

	dn := make(map[string]bool)

	for idx, d := range m.Descriptions {

		if d.Name == "" {
			msg = append(msg, "Unnamed description #" + +strconv.Itoa(idx))
		}

		if val, ok := dn[d.Name]; ok {
			msg = append(msg, "Duplicate description #" + strconv.Itoa(idx) + " " + d.Name)
		} else {
			dn[d.Name] = true
		}
	}

	pn := make(map[string]bool)

	for idx, p := range m.Polices {

		if p.Name == "" {
			msg.append("Unnamed policy #" + +strconv.Itoa(idx))
		}

		if val, ok := pn[p.Name]; ok {
			msg.append("Duplicate policy #" + strconv.Itoa(idx) + " " + p.Name)
		} else {
			pn[p.Name] = true
		}
	}

	for _, s := range m.Slots {

	}
}*/

/*

// CheckSlot checks for internal consistency of a slot
func CheckSlot(s Slot) (error, []string) {
	msg := []string{}

	if s.Name == "" {
		msg.append("Unnamed slot")
	}

	if len(msg) > 0 {
		return errors.New("Error"), msg
	}

	return nil, []string{}
}
*/
