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
		return errors.New("duplicate or missing name"), msg
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
		return errors.New("duplicate or missing name"), msg
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
		return errors.New("duplicate or missing name"), msg
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
		return errors.New("duplicate or missing name"), msg
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
		return errors.New("duplicate or missing name"), msg
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
		return errors.New("duplicate or missing name"), msg
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
		return errors.New("duplicate or missing name"), msg
	}

	return nil, []string{}

}

// CheckManifest checks for internal consistency, throwing an error
// if there are any unresolved references by name
func CheckManifest(m Manifest) (error, []string) {

	// check if any elements have duplicate or missing names

	err, msg := CheckDescriptions(m.Descriptions)

	if err != nil {
		return err, msg
	}

	err, msg = CheckPolicies(m.Policies)

	if err != nil {
		return err, msg
	}

	err, msg = CheckResources(m.Resources)

	if err != nil {
		return err, msg
	}

	err, msg = CheckStreams(m.Streams)

	if err != nil {
		return err, msg
	}
	err, msg = CheckSlots(m.Slots)

	if err != nil {
		return err, msg
	}
	err, msg = CheckUIs(m.UIs)

	if err != nil {
		return err, msg
	}
	err, msg = CheckUISets(m.UISets)

	if err != nil {
		return err, msg
	}

	// Make maps of all our entities
	dm := make(map[string]*Description)
	pm := make(map[string]*Policy)
	rm := make(map[string]*Resource)
	slm := make(map[string]*Slot)
	stm := make(map[string]*Stream)
	uim := make(map[string]*UI)
	usm := make(map[string]*UISet)

	for idx, d := range m.Descriptions {
		dm[d.Name] = &m.Descriptions[idx]
	}

	for idx, p := range m.Policies {
		pm[p.Name] = &m.Policies[idx]
	}

	for idx, r := range m.Resources {
		rm[r.Name] = &m.Resources[idx]
	}

	for idx, s := range m.Slots {
		slm[s.Name] = &m.Slots[idx]
	}

	for idx, s := range m.Streams {
		stm[s.Name] = &m.Streams[idx]
	}

	for idx, u := range m.UIs {
		uim[u.Name] = &m.UIs[idx]
	}

	for idx, u := range m.UISets {
		usm[u.Name] = &m.UISets[idx]
	}

	// Check that all references are present

	// Description -> N/A

	// Policy -> Description
	for k, v := range pm {
		if _, ok := dm[v.Description]; !ok {
			m := "Policy " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
	}

	// Resource ->  Description, Stream
	for k, v := range rm {
		if _, ok := dm[v.Description]; !ok {
			m := "Resource " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		for _, s := range v.Streams {
			if _, ok := stm[s]; !ok {
				m := "Resource " + k + " references non-existent stream: " + v.Description
				msg = append(msg, m)
			}
		}
	}

	// Slot -> Description, Policy, Resource, UISet
	for k, v := range slm {
		if _, ok := dm[v.Description]; !ok {
			m := "Slot " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		if _, ok := pm[v.Policy]; !ok {
			m := "Slot " + k + " references non-existent policy: " + v.Policy
			msg = append(msg, m)
		}
		if _, ok := rm[v.Resource]; !ok {
			m := "Slot " + k + " references non-existent resource: " + v.Resource
			msg = append(msg, m)
		}
		if _, ok := usm[v.UISet]; !ok {
			m := "Slot " + k + " references non-existent UISet: " + v.UISet
			msg = append(msg, m)
		}
	}

	// Stream -> N/A

	// UI -> Description, Stream

	for k, v := range uim {
		if _, ok := dm[v.Description]; !ok {
			m := "UI " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		// this check still applies, even though it relates in part to the templating process
		for _, s := range v.StreamsRequired {
			if _, ok := stm[s]; !ok {
				m := "UI " + k + " references non-existent stream: " + v.Description
				msg = append(msg, m)
			}
		}
	}

	// UISet -> UIs
	for k, v := range usm {
		for _, u := range v.UIs {
			if _, ok := uim[u]; !ok {
				m := "UISet " + k + " references non-existent UI: " + u
				msg = append(msg, m)
			}
		}
	}

	if len(msg) > 0 {
		return errors.New("missing reference(s)"), msg
	}

	return nil, []string{}

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
