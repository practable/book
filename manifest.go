package interval

import (
	"errors"
	"strconv"
	"time"

	"github.com/timdrysdale/interval/diary"
	"github.com/timdrysdale/interval/filter"
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
	Descriptions []Description `json:"descriptions" yaml:"descriptions"`
	Policies     []Policy      `json:"policies" yaml:"policies"`
	Resources    []Resource    `json:"resources" yaml:"resources"`
	Slots        []Slot        `json:"slots" yaml:"slots"`
	Streams      []Stream      `json:"streams" yaml:"streams"`
	UIs          []UI          `json:"uis" yaml:"uis"`
	UISets       []UISet       `json:"ui_sets" yaml:"ui_sets"`
	Windows      []Window      `json:"windows" yaml:"windows"`
}

// ReplaceManifest overwrites the existing manifest with a new one i.e. does not retain existing elements from any previous manifests
// but it does retain non-Manifest elements such as bookings.
func (s *Store) ReplaceManifest(m Manifest) error {

	err, _ := CheckManifest(m)

	if err != nil {
		return err //user can call CheckDescriptions some other way if they want the manifest error details
	}

	// we can get errors making filters, so do that before doing anything destructive
	// even though we checked it with CheckManifest, we have to handle the errors
	fm := make(map[string]*filter.Filter)
	wm := make(map[string]*Window)

	for idx, w := range m.Windows {

		f := filter.New()
		err = f.SetAllowed(w.Allowed)
		if err != nil {
			return errors.New("failed to create allowed intervals for window #" + strconv.Itoa(idx) + " (" + w.Name + "):" + err.Error())
		}
		err := f.SetDenied(w.Denied)
		if err != nil {
			return errors.New("failed to create denied intervals for window #" + strconv.Itoa(idx) + " (" + w.Name + "):" + err.Error())
		}

		fm[w.Name] = f
		wm[w.Name] = &m.Windows[idx]
	}

	// we're going to do the replacement now, goodbye old manifest data.
	s.Lock()
	defer s.Unlock()

	s.Filters = fm
	s.Windows = wm

	// Make new maps for our new entities
	s.Descriptions = make(map[string]*Description)
	s.Policies = make(map[string]*Policy)
	s.Resources = make(map[string]*Resource)
	s.Slots = make(map[string]*Slot)
	s.Streams = make(map[string]*Stream)
	s.UIs = make(map[string]*UI)
	s.UISets = make(map[string]*UISet)

	for idx, d := range m.Descriptions {
		s.Descriptions[d.Name] = &m.Descriptions[idx]
	}

	for idx, p := range m.Policies {
		s.Policies[p.Name] = &m.Policies[idx]
	}

	status := "Loaded at " + time.Now().Format(time.RFC3339)
	for idx, r := range m.Resources {
		m.Resources[idx].Diary = diary.New(r.Name)
		// default to available because unavailable kit is the exception
		m.Resources[idx].Diary.SetAvailable(status)
		s.Resources[r.Name] = &m.Resources[idx]
	}

	for idx, sl := range m.Slots {
		s.Slots[sl.Name] = &m.Slots[idx]
	}

	for idx, st := range m.Streams {
		s.Streams[st.Name] = &m.Streams[idx]
	}

	for idx, u := range m.UIs {
		s.UIs[u.Name] = &m.UIs[idx]
	}

	for idx, u := range m.UISets {
		s.UISets[u.Name] = &m.UISets[idx]
	}

	return nil

}

func CheckDescriptions(items []Description) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "unnamed description #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "duplicate description definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("duplicate or missing name"), msg
	}

	for idx, item := range items {
		if item.Type == "" {
			msg = append(msg, "missing type field in description #"+strconv.Itoa(idx)+": "+item.Name)
		}
		if item.Short == "" {
			msg = append(msg, "missing short field in description #"+strconv.Itoa(idx)+": "+item.Name)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckPolicies(items []Policy) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "unnamed policy #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "duplicate policy definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("duplicate or missing name"), msg
	}

	for idx, item := range items {
		if item.Description == "" {
			msg = append(msg, "missing description field in policy #"+strconv.Itoa(idx)+": "+item.Name)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckResources(items []Resource) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "unnamed resource #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "duplicate resource definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("duplicate or missing name"), msg
	}

	for idx, item := range items {
		// ConfigURL is optional
		if item.Description == "" {
			msg = append(msg, "missing description field in resource #"+strconv.Itoa(idx)+": "+item.Name)
		}
		if item.Streams == nil {
			msg = append(msg, "missing streams field in resource #"+strconv.Itoa(idx)+": "+item.Name)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckSlots(items []Slot) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "unnamed slot #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "duplicate slot definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("duplicate or missing name"), msg
	}

	for idx, item := range items {
		if item.Description == "" {
			msg = append(msg, "missing description field in slot #"+strconv.Itoa(idx)+": "+item.Name)
		}
		if item.Policy == "" {
			msg = append(msg, "missing policy field in slot #"+strconv.Itoa(idx)+": "+item.Name)
		}
		if item.Resource == "" {
			msg = append(msg, "missing resource field in slot #"+strconv.Itoa(idx)+": "+item.Name)
		}
		if item.UISet == "" {
			msg = append(msg, "missing ui_set field in slot #"+strconv.Itoa(idx)+": "+item.Name)
		}
		if item.Window == "" {
			msg = append(msg, "missing window field in slot #"+strconv.Itoa(idx)+": "+item.Name)
		}

	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckStreams(items []Stream) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "unnamed stream #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "duplicate stream definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("duplicate or missing name"), msg
	}

	for idx, item := range items {
		if item.Audience == "" {
			msg = append(msg, "missing audience field in stream #"+strconv.Itoa(idx)+": "+item.Name)
		}
		if item.ConnectionType == "" {
			msg = append(msg, "missing ct field in stream #"+strconv.Itoa(idx)+": "+item.Name)
		}
		if item.For == "" {
			msg = append(msg, "missing for field in stream #"+strconv.Itoa(idx)+": "+item.Name)
		}
		if item.Scopes == nil {
			msg = append(msg, "missing scopes field in stream #"+strconv.Itoa(idx)+": "+item.Name)
		}
		if item.Topic == "" {
			msg = append(msg, "missing topic field in stream #"+strconv.Itoa(idx)+": "+item.Name)
		}
		if item.URL == "" {
			msg = append(msg, "missing url field in stream #"+strconv.Itoa(idx)+": "+item.Name)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckUIs(items []UI) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			// lowercase capitalisation of ui to match json, yaml
			msg = append(msg, "unnamed ui #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "duplicate ui definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("duplicate or missing name"), msg
	}
	for idx, item := range items {
		if item.URL == "" {
			msg = append(msg, "missing url field in ui #"+strconv.Itoa(idx)+": "+item.Name)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckUISets(items []UISet) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			// lowercase capitalisation of ui_set to match json, yaml
			msg = append(msg, "unnamed ui_set #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "duplicate ui_set definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("duplicate or missing name"), msg
	}

	return nil, []string{}

}

func CheckWindows(items []Window) (error, []string) {

	msg := []string{}

	n := make(map[string]bool)

	for idx, item := range items {

		if item.Name == "" {
			msg = append(msg, "unnamed window #"+strconv.Itoa(idx))
		}

		if _, ok := n[item.Name]; ok {
			msg = append(msg, "duplicate window definition #"+strconv.Itoa(idx)+": "+item.Name)
		} else {
			n[item.Name] = true
		}
	}

	if len(msg) > 0 {
		return errors.New("duplicate or missing name"), msg
	}

	for idx, item := range items {
		// a window has to have at least one allowed period to be valid
		// a slot should be deleted rather than have a window with no allowed periods
		if item.Allowed == nil {
			msg = append(msg, "missing allowed field in window #"+strconv.Itoa(idx)+": "+item.Name)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	// we can get errors making filters, so check that

	for idx, w := range items {

		f := filter.New()
		err := f.SetAllowed(w.Allowed)
		if err != nil {
			msg = append(msg, "failed to create allowed intervals for window #"+strconv.Itoa(idx)+" ("+w.Name+"):"+err.Error())
		}
		err = f.SetDenied(w.Denied)
		if err != nil {
			msg = append(msg, "failed to create denied intervals for window #"+strconv.Itoa(idx)+" ("+w.Name+"):"+err.Error())
		}

	}

	if len(msg) > 0 {
		return errors.New("failed creating filter"), msg
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

	err, msg = CheckWindows(m.Windows)

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
	wm := make(map[string]*Window)

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

	for idx, w := range m.Windows {
		wm[w.Name] = &m.Windows[idx]
	}

	// Check that all references are present

	// Description -> N/A

	// Policy -> Description
	for k, v := range pm {
		if _, ok := dm[v.Description]; !ok {
			m := "policy " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
	}

	// Resource ->  Description, Stream
	for k, v := range rm {
		if _, ok := dm[v.Description]; !ok {
			m := "resource " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		for _, s := range v.Streams {
			if _, ok := stm[s]; !ok {
				m := "resource " + k + " references non-existent stream: " + s
				msg = append(msg, m)
			}
		}
	}

	// Slot -> Description, Policy, Resource, UISet, Window
	for k, v := range slm {
		if _, ok := dm[v.Description]; !ok {
			m := "slot " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		if _, ok := pm[v.Policy]; !ok {
			m := "slot " + k + " references non-existent policy: " + v.Policy
			msg = append(msg, m)
		}
		if _, ok := rm[v.Resource]; !ok {
			m := "slot " + k + " references non-existent resource: " + v.Resource
			msg = append(msg, m)
		}
		if _, ok := usm[v.UISet]; !ok {
			m := "slot " + k + " references non-existent ui_set: " + v.UISet
			msg = append(msg, m)
		}
		if _, ok := wm[v.Window]; !ok {
			m := "slot " + k + " references non-existent window: " + v.Window
			msg = append(msg, m)
		}
	}

	// Stream -> N/A

	// UI -> Description, StreamsRequired

	for k, v := range uim {
		if _, ok := dm[v.Description]; !ok {
			m := "ui " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		// this check still applies, even though it relates in part to the templating process
		for _, s := range v.StreamsRequired {
			if _, ok := stm[s]; !ok {
				m := "ui " + k + " references non-existent stream: " + s
				msg = append(msg, m)
			}
		}
	}

	// UISet -> UIs
	for k, v := range usm {
		for _, u := range v.UIs {
			if _, ok := uim[u]; !ok {
				m := "ui_set " + k + " references non-existent ui: " + u
				msg = append(msg, m)
			}
		}
	}

	if len(msg) > 0 {
		return errors.New("missing reference(s)"), msg
	}

	return nil, []string{}

}
