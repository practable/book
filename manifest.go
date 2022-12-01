package interval

import (
	"errors"
	"time"

	"github.com/timdrysdale/interval/diary"
	"github.com/timdrysdale/interval/filter"
)

// Manifest represents all the available equipment and how to access it
// Slots are the primary entities, so reference checking starts with them
type Manifest struct {
	Descriptions map[string]Description `json:"descriptions" yaml:"descriptions"`
	Policies     map[string]Policy      `json:"policies" yaml:"policies"`
	Resources    map[string]Resource    `json:"resources" yaml:"resources"`
	Slots        map[string]Slot        `json:"slots" yaml:"slots"`
	Streams      map[string]Stream      `json:"streams" yaml:"streams"`
	UIs          map[string]UI          `json:"uis" yaml:"uis"`
	UISets       map[string]UISet       `json:"ui_sets" yaml:"ui_sets"`
	Windows      map[string]Window      `json:"windows" yaml:"windows"`
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

	for k, w := range m.Windows {

		f := filter.New()
		err = f.SetAllowed(w.Allowed)
		if err != nil {
			return errors.New("failed to create allowed intervals for window " + k + ":" + err.Error())
		}
		err := f.SetDenied(w.Denied)
		if err != nil {
			return errors.New("failed to create denied intervals for window " + k + ":" + err.Error())
		}

		fm[k] = f
	}

	// we're going to do the replacement now, goodbye old manifest data.
	s.Lock()
	defer s.Unlock()

	s.Filters = fm

	// Make new maps for our new entities
	s.Descriptions = m.Descriptions
	s.Policies = m.Policies
	s.Resources = m.Resources
	s.Slots = m.Slots
	s.Streams = m.Streams
	s.UISets = m.UISets
	s.Windows = m.Windows

	status := "Loaded at " + s.Now().Format(time.RFC3339)

	// SlotMap is used for checking if slots are listed in policy
	for k, v := range s.Policies {
		v.SlotMap = make(map[string]bool)
		for _, k := range v.Slots {
			v.SlotMap[k] = true
		}
		s.Policies[k] = v
	}

	for k := range s.Resources {
		r := s.Resources[k]
		r.Diary = diary.New(k)
		s.Resources[k] = r
		// default to available because unavailable kit is the exception
		s.Resources[k].Diary.SetAvailable(status)
	}

	// populate UIs with descriptions now to save doing it repetively later
	s.UIs = make(map[string]UIDescribed)

	for k, v := range m.UIs {

		d, err := s.GetDescription(v.Description)

		if err != nil {
			return err
		}

		uid := UIDescribed{
			Description:          d,
			DescriptionReference: v.Description,
			URL:                  m.UIs[k].URL,
			StreamsRequired:      m.UIs[k].StreamsRequired,
		}
		s.UIs[k] = uid
	}

	return nil

}

// ExportManifest returns the manifest from the store
func (s *Store) ExportManifest() Manifest {

	uis := make(map[string]UI)

	// Manifest only has the name of the description in the UI
	for k, v := range s.UIs {
		uis[k] = UI{
			Description:     v.DescriptionReference,
			URL:             v.URL,
			StreamsRequired: v.StreamsRequired,
		}
	}

	// Resources have diary pointers which we should nullify by omission for security and readability
	rm := make(map[string]Resource)
	for k, v := range s.Resources {
		rm[k] = Resource{
			ConfigURL:   v.ConfigURL,
			Description: v.Description,
			Streams:     v.Streams,
			TopicStub:   v.TopicStub,
		}
	}

	return Manifest{
		Descriptions: s.Descriptions,
		Policies:     s.Policies,
		Resources:    rm,
		Slots:        s.Slots,
		Streams:      s.Streams,
		UIs:          uis,
		UISets:       s.UISets,
		Windows:      s.Windows,
	}
}

func CheckDescriptions(items map[string]Description) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.Name == "" {
			msg = append(msg, "missing name field in description "+k)
		}
		if item.Type == "" {
			msg = append(msg, "missing type field in description "+k)
		}
		if item.Short == "" {
			msg = append(msg, "missing short field in description "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckPolicies(items map[string]Policy) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.Description == "" {
			msg = append(msg, "missing description field in policy "+k)
		}
		if item.Slots == nil {
			msg = append(msg, "missing slots field in policy "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckResources(items map[string]Resource) (error, []string) {

	msg := []string{}

	for k, item := range items {
		// ConfigURL is optional
		if item.Description == "" {
			msg = append(msg, "missing description field in resource "+k)
		}
		if item.Streams == nil {
			msg = append(msg, "missing streams field in resource "+k)
		}
		if item.TopicStub == "" {
			msg = append(msg, "missing topic_stub field in resource "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckSlots(items map[string]Slot) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.Description == "" {
			msg = append(msg, "missing description field in slot "+k)
		}
		if item.Policy == "" {
			msg = append(msg, "missing policy field in slot "+k)
		}
		if item.Resource == "" {
			msg = append(msg, "missing resource field in slot "+k)
		}
		if item.UISet == "" {
			msg = append(msg, "missing ui_set field in slot "+k)
		}
		if item.Window == "" {
			msg = append(msg, "missing window field in slot "+k)
		}

	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckStreams(items map[string]Stream) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.Audience == "" {
			msg = append(msg, "missing audience field in stream "+k)
		}
		if item.ConnectionType == "" {
			msg = append(msg, "missing ct field in stream "+k)
		}
		if item.For == "" {
			msg = append(msg, "missing for field in stream "+k)
		}
		if item.Scopes == nil {
			msg = append(msg, "missing scopes field in stream "+k)
		}
		if item.Topic == "" {
			msg = append(msg, "missing topic field in stream "+k)
		}
		if item.URL == "" {
			msg = append(msg, "missing url field in stream "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckUIs(items map[string]UI) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.URL == "" {
			msg = append(msg, "missing url field in ui "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckUISets(items map[string]UISet) (error, []string) {

	msg := []string{}

	for k, item := range items {
		if item.UIs == nil {
			msg = append(msg, "missing uis field in ui_set "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	return nil, []string{}

}

func CheckWindows(items map[string]Window) (error, []string) {

	msg := []string{}

	for k, item := range items {
		// a window has to have at least one allowed period to be valid
		// a slot should be deleted rather than have a window with no allowed periods
		if item.Allowed == nil {
			msg = append(msg, "missing allowed field in window "+k)
		}
	}

	if len(msg) > 0 {
		return errors.New("missing field"), msg
	}

	// we can get errors making filters, so check that

	for k, w := range items {

		f := filter.New()
		err := f.SetAllowed(w.Allowed)
		if err != nil {
			msg = append(msg, "failed to create allowed intervals for window "+k+": "+err.Error())
		}
		err = f.SetDenied(w.Denied)
		if err != nil {
			msg = append(msg, "failed to create denied intervals for window "+k+": "+err.Error())
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

	// Check that all references are present

	// Description -> N/A

	// Policy -> Description, Slots
	for k, v := range m.Policies {
		if _, ok := m.Descriptions[v.Description]; !ok {
			m := "policy " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		for _, s := range v.Slots {
			if _, ok := m.Slots[s]; !ok {
				m := "policy " + k + " references non-existent slot: " + s
				msg = append(msg, m)
			}
		}
	}

	// Resource ->  Description, Stream
	for k, v := range m.Resources {
		if _, ok := m.Descriptions[v.Description]; !ok {
			m := "resource " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		for _, s := range v.Streams {
			if _, ok := m.Streams[s]; !ok {
				m := "resource " + k + " references non-existent stream: " + s
				msg = append(msg, m)
			}
		}
	}

	// Slot -> Description, Policy, Resource, UISet, Window
	for k, v := range m.Slots {
		if _, ok := m.Descriptions[v.Description]; !ok {
			m := "slot " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		if _, ok := m.Policies[v.Policy]; !ok {
			m := "slot " + k + " references non-existent policy: " + v.Policy
			msg = append(msg, m)
		}
		if _, ok := m.Resources[v.Resource]; !ok {
			m := "slot " + k + " references non-existent resource: " + v.Resource
			msg = append(msg, m)
		}
		if _, ok := m.UISets[v.UISet]; !ok {
			m := "slot " + k + " references non-existent ui_set: " + v.UISet
			msg = append(msg, m)
		}
		if _, ok := m.Windows[v.Window]; !ok {
			m := "slot " + k + " references non-existent window: " + v.Window
			msg = append(msg, m)
		}
	}

	// Stream -> N/A

	// UI -> Description, StreamsRequired

	for k, v := range m.UIs {
		if _, ok := m.Descriptions[v.Description]; !ok {
			m := "ui " + k + " references non-existent description: " + v.Description
			msg = append(msg, m)
		}
		// this check still applies, even though it relates in part to the templating process
		for _, s := range v.StreamsRequired {
			if _, ok := m.Streams[s]; !ok {
				m := "ui " + k + " references non-existent stream: " + s
				msg = append(msg, m)
			}
		}
	}

	// UISet -> UIs
	for k, v := range m.UISets {
		for _, u := range v.UIs {
			if _, ok := m.UIs[u]; !ok {
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
