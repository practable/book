package interval

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/timdrysdale/interval/interval"
)

// Note that complex types and slices are shallow copied so changes are visible
// to other tests. Since tests may eventually run in parallel, add a mutex
// All tests must restore any changes they make to the manifest
// Note :- the mutex might have been an over-reaction to a confusing
// test result .... but it's in there now.
type MutexManifest struct {
	*sync.Mutex
	Manifest Manifest
}

var testManifest = MutexManifest{
	&sync.Mutex{},
	Manifest{
		Descriptions: []Description{
			Description{
				Name:  "d-p-a",
				Type:  "policy",
				Short: "a",
			},
			Description{
				Name:  "d-p-b",
				Type:  "policy",
				Short: "b",
			},
			Description{
				Name:  "d-r-a",
				Type:  "resource",
				Short: "a",
			},
			Description{
				Name:  "d-r-b",
				Type:  "resource",
				Short: "b",
			},
			Description{
				Name:  "d-sl-a",
				Type:  "slot",
				Short: "a",
			},
			Description{
				Name:  "d-sl-b",
				Type:  "slot",
				Short: "b",
			},
			Description{
				Name:  "d-ui-a",
				Type:  "ui",
				Short: "a",
			},
			Description{
				Name:  "d-ui-b",
				Type:  "ui",
				Short: "b",
			},
		},
		Policies: []Policy{
			Policy{
				Name:        "p-a",
				Description: "d-p-a",
			},
			Policy{
				Name:        "p-b",
				Description: "d-p-b",
			},
		},
		Resources: []Resource{
			Resource{
				Name:        "r-a",
				Description: "d-r-a",
				Streams:     []string{"st-a", "st-b"},
			},
			Resource{
				Name:        "r-b",
				Description: "d-r-b",
				Streams:     []string{"st-a", "st-b"},
			},
		},
		Slots: []Slot{
			Slot{
				Name:        "sl-a",
				Description: "d-sl-a",
				Policy:      "p-a",
				Resource:    "r-a",
				UISet:       "us-a",
				Window:      "w-a",
			},
			Slot{
				Name:        "sl-b",
				Description: "d-sl-b",
				Policy:      "p-b",
				Resource:    "r-b",
				UISet:       "us-b",
				Window:      "w-b",
			},
		},
		Streams: []Stream{
			Stream{
				Name:           "st-a",
				Audience:       "a",
				ConnectionType: "a",
				For:            "a",
				Scopes:         []string{"r", "w"},
				Topic:          "a",
				URL:            "a",
			},
			Stream{
				Name:           "st-b",
				Audience:       "b",
				ConnectionType: "b",
				For:            "b",
				Scopes:         []string{"r", "w"},
				Topic:          "b",
				URL:            "b",
			},
		},
		UIs: []UI{
			UI{
				Name:            "ui-a",
				Description:     "d-ui-a",
				StreamsRequired: []string{"st-a", "st-b"},
				URL:             "a",
			},
			UI{
				Name:            "ui-b",
				Description:     "d-ui-b",
				StreamsRequired: []string{"st-a", "st-b"},
				URL:             "b",
			},
		},
		UISets: []UISet{
			UISet{
				Name: "us-a",
				UIs:  []string{"ui-a"},
			},
			UISet{
				Name: "us-b",
				UIs:  []string{"ui-a", "ui-b"},
			},
		},
		Windows: []Window{
			Window{
				Name: "w-a",
				Allowed: []interval.Interval{
					interval.Interval{
						Start: time.Now(),
						End:   time.Now().Add(time.Hour),
					},
				},
			},
			Window{
				Name: "w-b",
				Allowed: []interval.Interval{
					interval.Interval{
						Start: time.Now(),
						End:   time.Now().Add(time.Hour),
					},
				},
			},
		},
	},
}

func TestCheckDescriptions(t *testing.T) {

	a := Description{
		Name:  "a",
		Type:  "test",
		Short: "a",
	}

	b := Description{
		Name:  "b",
		Type:  "test",
		Short: "b",
	}

	c := Description{
		Name:  "c",
		Type:  "test",
		Short: "c",
	}

	d := Description{
		Name:  "a",
		Type:  "test",
		Short: "duplicate a",
	}

	e := Description{}

	items := []Description{a, b, c}

	err, msg := CheckDescriptions(items)

	assert.NoError(t, err)

	if err != nil {
		t.Log(msg)
	}

	items = []Description{a, b, c, d}

	err, msg = CheckDescriptions(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"duplicate description definition #3: a"}, msg)

	items = []Description{a, e, b, c}

	err, msg = CheckDescriptions(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"unnamed description #1"}, msg)

}

func TestCheckPolicies(t *testing.T) {

	a := Policy{
		Name:        "a",
		Description: "d-p-a",
	}

	b := Policy{
		Name:        "b",
		Description: "d-p-b",
	}

	c := Policy{
		Name:        "c",
		Description: "d-p-c",
	}

	d := Policy{
		Name:        "a",
		Description: "d-p-a",
	}

	e := Policy{}

	items := []Policy{a, b, c}

	err, msg := CheckPolicies(items)

	assert.NoError(t, err)

	if err != nil {
		t.Log(msg)
	}

	items = []Policy{a, b, c, d}

	err, msg = CheckPolicies(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"duplicate policy definition #3: a"}, msg)

	items = []Policy{a, e, b, c}

	err, msg = CheckPolicies(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"unnamed policy #1"}, msg)

}

func TestCheckResources(t *testing.T) {

	a := Resource{
		Name:        "a",
		Description: "d-r-a",
		Streams:     []string{"a", "b"},
	}

	b := Resource{
		Name:        "b",
		Description: "d-r-b",
		Streams:     []string{"a", "b"},
	}

	c := Resource{
		Name:        "c",
		Description: "d-r-c",
		Streams:     []string{"a", "b"},
	}

	d := Resource{
		Name:        "a",
		Description: "d-r-a",
		Streams:     []string{"a", "b"},
	}

	e := Resource{}

	items := []Resource{a, b, c}

	err, msg := CheckResources(items)

	assert.NoError(t, err)

	if err != nil {
		t.Log(msg)
	}

	items = []Resource{a, b, c, d}

	err, msg = CheckResources(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"duplicate resource definition #3: a"}, msg)

	items = []Resource{a, e, b, c}

	err, msg = CheckResources(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"unnamed resource #1"}, msg)

}

func TestCheckSlots(t *testing.T) {

	a := Slot{
		Name:        "a",
		Description: "d-sl-a",
		Policy:      "p-a",
		Resource:    "r-a",
		UISet:       "us-a",
		Window:      "w-a",
	}

	b := Slot{
		Name:        "b",
		Description: "d-sl-b",
		Policy:      "p-b",
		Resource:    "r-b",
		UISet:       "us-b",
		Window:      "w-b",
	}

	c := Slot{
		Name:        "c",
		Description: "d-sl-c",
		Policy:      "p-c",
		Resource:    "r-c",
		UISet:       "us-c",
		Window:      "w-c",
	}

	d := Slot{
		Name:        "a",
		Description: "d-sl-a",
		Policy:      "p-a",
		Resource:    "r-a",
		UISet:       "us-a",
		Window:      "w-a",
	}

	e := Slot{}

	items := []Slot{a, b, c}

	err, msg := CheckSlots(items)

	assert.NoError(t, err)

	if err != nil {
		t.Log(msg)
	}

	items = []Slot{a, b, c, d}

	err, msg = CheckSlots(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"duplicate slot definition #3: a"}, msg)

	items = []Slot{a, e, b, c}

	err, msg = CheckSlots(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"unnamed slot #1"}, msg)

}

func TestCheckStreams(t *testing.T) {

	a := Stream{
		Name:           "a",
		Audience:       "a",
		ConnectionType: "a",
		For:            "a",
		Scopes:         []string{"r", "w"},
		Topic:          "a",
		URL:            "a",
	}

	b := Stream{
		Name:           "b",
		Audience:       "a",
		ConnectionType: "a",
		For:            "a",
		Scopes:         []string{"r", "w"},
		Topic:          "a",
		URL:            "a",
	}
	c := Stream{
		Name:           "c",
		Audience:       "a",
		ConnectionType: "a",
		For:            "a",
		Scopes:         []string{"r", "w"},
		Topic:          "a",
		URL:            "a",
	}

	d := Stream{
		Name:           "a",
		Audience:       "a",
		ConnectionType: "a",
		For:            "a",
		Scopes:         []string{"r", "w"},
		Topic:          "a",
		URL:            "a",
	}

	e := Stream{}

	items := []Stream{a, b, c}

	err, msg := CheckStreams(items)

	assert.NoError(t, err)

	if err != nil {
		t.Log(msg)
	}

	items = []Stream{a, b, c, d}

	err, msg = CheckStreams(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"duplicate stream definition #3: a"}, msg)

	items = []Stream{a, e, b, c}

	err, msg = CheckStreams(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"unnamed stream #1"}, msg)

}

func TestCheckUIs(t *testing.T) {

	a := UI{
		Name: "a",
		URL:  "a",
	}

	b := UI{
		Name: "b",
		URL:  "a",
	}

	c := UI{
		Name: "c",
		URL:  "a",
	}

	d := UI{
		Name: "a",
		URL:  "a",
	}

	e := UI{}

	items := []UI{a, b, c}

	err, msg := CheckUIs(items)

	assert.NoError(t, err)

	if err != nil {
		t.Log(msg)
	}

	items = []UI{a, b, c, d}

	err, msg = CheckUIs(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, msg, []string{"duplicate ui definition #3: a"})

	items = []UI{a, e, b, c}

	err, msg = CheckUIs(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, msg, []string{"unnamed ui #1"})
}

func TestCheckUISets(t *testing.T) {

	a := UISet{
		Name: "a",
	}

	b := UISet{
		Name: "b",
	}

	c := UISet{
		Name: "c",
	}

	d := UISet{
		Name: "a",
	}

	e := UISet{}

	items := []UISet{a, b, c}

	err, msg := CheckUISets(items)

	assert.NoError(t, err)

	if err != nil {
		t.Log(msg)
	}

	items = []UISet{a, b, c, d}

	err, msg = CheckUISets(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"duplicate ui_set definition #3: a"}, msg)

	items = []UISet{a, e, b, c}

	err, msg = CheckUISets(items)

	assert.Error(t, err)
	assert.Equal(t, "duplicate or missing name", err.Error())
	assert.Equal(t, []string{"unnamed ui_set #1"}, msg)
}

func TestCheckOKManifest(t *testing.T) {

	err, msg := CheckManifest(testManifest.Manifest)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, msg)
}

func TestCheckManifestCatchMissingUI(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()
	m := testManifest.Manifest

	us := m.UISets
	us[1].UIs[1] = "ui-c" //does not exist
	m.UISets = us

	err, msg := CheckManifest(m)

	assert.Error(t, err)
	assert.Equal(t, []string{"ui_set us-b references non-existent ui: ui-c"}, msg)

	us[1].UIs[1] = "ui-b" //fix manifest for other tests
	m.UISets = us
	err, _ = CheckManifest(m)
	assert.NoError(t, err)
}

func TestCheckManifestCatchMissingResource(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	r := testManifest.Manifest.Resources
	r[1].Name = "r-c" //makes r-b not exist
	testManifest.Manifest.Resources = r

	err, msg := CheckManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"slot sl-b references non-existent resource: r-b"}, msg)

	r[1].Name = "r-b" //fix manifest for other tests
	testManifest.Manifest.Resources = r
	err, _ = CheckManifest(testManifest.Manifest)
	assert.NoError(t, err)
}

func TestCheckManifestCatchMissingDescriptions(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	d := testManifest.Manifest.Descriptions
	d[4].Name = "foo" //makes d-sl-a not exist
	testManifest.Manifest.Descriptions = d

	err, msg := CheckManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"slot sl-a references non-existent description: d-sl-a"}, msg)

	d[4].Name = "d-sl-a" //fix manifest for other tests
	testManifest.Manifest.Descriptions = d
	err, _ = CheckManifest(testManifest.Manifest)
	assert.NoError(t, err)

}

func TestCheckManifestCatchMissingStream(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	s := testManifest.Manifest.UIs[1].StreamsRequired
	testManifest.Manifest.UIs[1].StreamsRequired = []string{"st-c"} // st-c not exist

	err, msg := CheckManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"ui ui-b references non-existent stream: st-c"}, msg)

	testManifest.Manifest.UIs[1].StreamsRequired = s //fix manifest for other tests

	err, _ = CheckManifest(testManifest.Manifest)
	assert.NoError(t, err)

}

//TODO add extra checks if this by-name reference approach works out
