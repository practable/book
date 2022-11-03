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
			msg = append(msg, "Unnamed description #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "Duplicate description definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("Issues found"), msg
	}

	return nil, []string{}

}

func CheckPolicies(items []Policy) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "Unnamed policy #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "Duplicate policy definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("Issues found"), msg
	}

	return nil, []string{}

}

func CheckResources(items []Resource) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "Unnamed resource #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "Duplicate resource definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("Issues found"), msg
	}

	return nil, []string{}

}

func CheckSlots(items []Slot) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "Unnamed slot #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "Duplicate slot definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("Issues found"), msg
	}

	return nil, []string{}

}

func CheckStreams(items []Stream) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "Unnamed stream #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "Duplicate stream definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("Issues found"), msg
	}

	return nil, []string{}

}

func CheckUIs(items []UI) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "Unnamed UI #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "Duplicate UI definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("Issues found"), msg
	}

	return nil, []string{}

}

func CheckUISets(items []UISet) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "Unnamed UISet #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "Duplicate UISet definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("Issues found"), msg
	}

	return nil, []string{}

}

// CheckManifest checks for internal consistency, throwing an error
// if there are any unresolved references by name
func CheckManifest(m Manifest) (error, []string) {
	// check if any elements have duplicate names

	return errors.New("not implemented"), []string{}

}

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
