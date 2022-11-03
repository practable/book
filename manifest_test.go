package interval

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note that slices are shallow copied so changes are visible
// to other tests. Since tests may eventually run in parallel, add a mutex
// All tests must restore any changes they make to the manifest
type MutexManifest struct {
	*sync.Mutex
	Manifest Manifest
}

var testManifest = MutexManifest{
	&sync.Mutex{},
	Manifest{
		Descriptions: []Description{
			Description{Name: "d-p-a"},
			Description{Name: "d-p-b"},
			Description{Name: "d-r-a"},
			Description{Name: "d-r-b"},
			Description{Name: "d-sl-a"},
			Description{Name: "d-sl-b"},
			Description{Name: "d-ui-a"},
			Description{Name: "d-ui-b"},
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
			},
			Resource{
				Name:        "r-b",
				Description: "d-r-b",
			},
		},
		Slots: []Slot{
			Slot{
				Name:        "sl-a",
				Description: "d-sl-a",
				Policy:      "p-a",
				Resource:    "r-a",
				UISet:       "us-a",
			},
			Slot{
				Name:        "sl-b",
				Description: "d-sl-b",
				Policy:      "p-b",
				Resource:    "r-b",
				UISet:       "us-b",
			},
		},
		Streams: []Stream{
			Stream{Name: "st-a"},
			Stream{Name: "st-b"},
		},
		UIs: []UI{
			UI{
				Name:            "ui-a",
				Description:     "d-ui-a",
				StreamsRequired: []string{"st-a", "st-b"},
			},
			UI{
				Name:            "ui-b",
				Description:     "d-ui-b",
				StreamsRequired: []string{"st-a", "st-b"},
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
	},
}

func TestCheckDescriptions(t *testing.T) {

	a := Description{
		Name: "a",
	}

	b := Description{
		Name: "b",
	}

	c := Description{
		Name: "c",
	}

	d := Description{
		Name: "a",
	}

	e := Description{}

	items := []Description{a, b, c}

	err, msg := CheckDescriptions(items)

	assert.NoError(t, err)

	items = []Description{a, b, c, d}

	err, msg = CheckDescriptions(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Duplicate description definition #3: a"})

	items = []Description{a, e, b, c}

	err, msg = CheckDescriptions(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Unnamed description #1"})

}

func TestCheckPolicies(t *testing.T) {

	a := Policy{
		Name: "a",
	}

	b := Policy{
		Name: "b",
	}

	c := Policy{
		Name: "c",
	}

	d := Policy{
		Name: "a",
	}

	e := Policy{}

	items := []Policy{a, b, c}

	err, msg := CheckPolicies(items)

	assert.NoError(t, err)

	items = []Policy{a, b, c, d}

	err, msg = CheckPolicies(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Duplicate policy definition #3: a"})

	items = []Policy{a, e, b, c}

	err, msg = CheckPolicies(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Unnamed policy #1"})

}

func TestCheckResources(t *testing.T) {

	a := Resource{
		Name: "a",
	}

	b := Resource{
		Name: "b",
	}

	c := Resource{
		Name: "c",
	}

	d := Resource{
		Name: "a",
	}

	e := Resource{}

	items := []Resource{a, b, c}

	err, msg := CheckResources(items)

	assert.NoError(t, err)

	items = []Resource{a, b, c, d}

	err, msg = CheckResources(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Duplicate resource definition #3: a"})

	items = []Resource{a, e, b, c}

	err, msg = CheckResources(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Unnamed resource #1"})

}

func TestCheckStreams(t *testing.T) {

	a := Stream{
		Name: "a",
	}

	b := Stream{
		Name: "b",
	}

	c := Stream{
		Name: "c",
	}

	d := Stream{
		Name: "a",
	}

	e := Stream{}

	items := []Stream{a, b, c}

	err, msg := CheckStreams(items)

	assert.NoError(t, err)

	items = []Stream{a, b, c, d}

	err, msg = CheckStreams(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Duplicate stream definition #3: a"})

	items = []Stream{a, e, b, c}

	err, msg = CheckStreams(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Unnamed stream #1"})

}

func TestCheckSlots(t *testing.T) {

	a := Slot{
		Name: "a",
	}

	b := Slot{
		Name: "b",
	}

	c := Slot{
		Name: "c",
	}

	d := Slot{
		Name: "a",
	}

	e := Slot{}

	items := []Slot{a, b, c}

	err, msg := CheckSlots(items)

	assert.NoError(t, err)

	items = []Slot{a, b, c, d}

	err, msg = CheckSlots(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Duplicate slot definition #3: a"})

	items = []Slot{a, e, b, c}

	err, msg = CheckSlots(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Unnamed slot #1"})

}

func TestCheckUIs(t *testing.T) {

	a := UI{
		Name: "a",
	}

	b := UI{
		Name: "b",
	}

	c := UI{
		Name: "c",
	}

	d := UI{
		Name: "a",
	}

	e := UI{}

	items := []UI{a, b, c}

	err, msg := CheckUIs(items)

	assert.NoError(t, err)

	items = []UI{a, b, c, d}

	err, msg = CheckUIs(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Duplicate UI definition #3: a"})

	items = []UI{a, e, b, c}

	err, msg = CheckUIs(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Unnamed UI #1"})
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

	items = []UISet{a, b, c, d}

	err, msg = CheckUISets(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Duplicate UISet definition #3: a"})

	items = []UISet{a, e, b, c}

	err, msg = CheckUISets(items)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate or missing name")
	assert.Equal(t, msg, []string{"Unnamed UISet #1"})
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
	assert.Equal(t, []string{"UISet us-b references non-existent UI: ui-c"}, msg)

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
	assert.Equal(t, []string{"Slot sl-b references non-existent resource: r-b"}, msg)

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
	assert.Equal(t, []string{"Slot sl-a references non-existent description: d-sl-a"}, msg)

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
	assert.Equal(t, []string{"UI ui-b references non-existent stream: st-c"}, msg)

	testManifest.Manifest.UIs[1].StreamsRequired = s //fix manifest for other tests

	err, _ = CheckManifest(testManifest.Manifest)
	assert.NoError(t, err)

}

//TODO add extra checks if this by-name reference approach works out
