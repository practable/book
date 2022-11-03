package interval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var OKManifest = Manifest{
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
			Name:        "ui-a",
			Description: "d-ui-a",
		},
		UI{
			Name:        "ui-b",
			Description: "d-ui-b",
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

	err, msg := CheckManifest(OKManifest)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, msg)
}

func TestCheckManifestCatchMissingUI(t *testing.T) {

	m := OKManifest

	us := m.UISets
	us[1].UIs[1] = "ui-c" //does not exist
	m.UISets = us

	err, msg := CheckManifest(m)

	assert.Error(t, err)
	assert.Equal(t, []string{"UISet us-b references non-existent UI: ui-c"}, msg)
}
