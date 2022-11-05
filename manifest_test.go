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
		Descriptions: map[string]Description{
			"d-p-a": Description{
				Name:  "policy-a",
				Type:  "policy",
				Short: "a",
			},
			"d-p-b": Description{
				Name:  "policy-b",
				Type:  "policy",
				Short: "b",
			},
			"d-r-a": Description{
				Name:  "resource-a",
				Type:  "resource",
				Short: "a",
			},
			"d-r-b": Description{
				Name:  "resource-b",
				Type:  "resource",
				Short: "b",
			},
			"d-sl-a": Description{
				Name:  "slot-a",
				Type:  "slot",
				Short: "a",
			},
			"d-sl-b": Description{
				Name:  "slot-b",
				Type:  "slot",
				Short: "b",
			},
			"d-ui-a": Description{
				Name:  "ui-a",
				Type:  "ui",
				Short: "a",
			},
			"d-ui-b": Description{
				Name:  "ui-b",
				Type:  "ui",
				Short: "b",
			},
		},
		Policies: map[string]Policy{
			"p-a": Policy{
				Description: "d-p-a",
				Slots:       []string{"sl-a"},
			},
			"p-b": Policy{
				BookAhead:          time.Duration(2 * time.Hour),
				Description:        "d-p-b",
				EnforceBookAhead:   true,
				EnforceMaxBookings: true,
				EnforceMinDuration: true,
				EnforceMaxDuration: true,
				EnforceMaxUsage:    true,
				MaxUsage:           time.Duration(30 * time.Minute),
				MaxBookings:        2,
				MaxDuration:        time.Duration(10 * time.Minute),
				MinDuration:        time.Duration(5 * time.Minute),
				Slots:              []string{"sl-b"},
			},
		},
		Resources: map[string]Resource{
			"r-a": Resource{
				Description: "d-r-a",
				Streams:     []string{"st-a", "st-b"},
				TopicStub:   "aaaa00",
			},
			"r-b": Resource{
				Description: "d-r-b",
				Streams:     []string{"st-a", "st-b"},
				TopicStub:   "bbbb00",
			},
		},
		Slots: map[string]Slot{
			"sl-a": Slot{
				Description: "d-sl-a",
				Policy:      "p-a",
				Resource:    "r-a",
				UISet:       "us-a",
				Window:      "w-a",
			},
			"sl-b": Slot{
				Description: "d-sl-b",
				Policy:      "p-b",
				Resource:    "r-b",
				UISet:       "us-b",
				Window:      "w-b",
			},
		},
		Streams: map[string]Stream{
			"st-a": Stream{
				Audience:       "a",
				ConnectionType: "a",
				For:            "a",
				Scopes:         []string{"r", "w"},
				Topic:          "a",
				URL:            "a",
			},
			"st-b": Stream{
				Audience:       "b",
				ConnectionType: "b",
				For:            "b",
				Scopes:         []string{"r", "w"},
				Topic:          "b",
				URL:            "b",
			},
		},
		UIs: map[string]UI{
			"ui-a": UI{
				Description:     "d-ui-a",
				StreamsRequired: []string{"st-a", "st-b"},
				URL:             "a",
			},
			"ui-b": UI{
				Description:     "d-ui-b",
				StreamsRequired: []string{"st-a", "st-b"},
				URL:             "b",
			},
		},
		UISets: map[string]UISet{
			"us-a": UISet{
				UIs: []string{"ui-a"},
			},
			"us-b": UISet{
				UIs: []string{"ui-a", "ui-b"},
			},
		},
		Windows: map[string]Window{
			"w-a": Window{
				Allowed: []interval.Interval{
					interval.Interval{
						Start: time.Date(2022, 11, 4, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			"w-b": Window{
				Allowed: []interval.Interval{
					interval.Interval{
						Start: time.Date(2022, 11, 4, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2022, 11, 6, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	},
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

	m.UISets["us-b"].UIs[1] = "ui-c" //ui-c does not exist

	err, msg := CheckManifest(m)

	assert.Error(t, err)
	assert.Equal(t, []string{"ui_set us-b references non-existent ui: ui-c"}, msg)

	//fix manifest for other tests
	m.UISets["us-b"].UIs[1] = "ui-b"

	err, _ = CheckManifest(m)
	assert.NoError(t, err)
}

func TestCheckManifestCatchMissingResource(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	testManifest.Manifest.Resources["r-c"] = testManifest.Manifest.Resources["r-b"]
	delete(testManifest.Manifest.Resources, "r-b")

	err, msg := CheckManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"slot sl-b references non-existent resource: r-b"}, msg)

	// fix manifest
	testManifest.Manifest.Resources["r-b"] = testManifest.Manifest.Resources["r-c"]
	delete(testManifest.Manifest.Resources, "r-c")

	err, _ = CheckManifest(testManifest.Manifest)
	assert.NoError(t, err)
}

func TestCheckManifestCatchMissingDescriptions(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	dsla := testManifest.Manifest.Descriptions["d-sl-a"]
	delete(testManifest.Manifest.Descriptions, "d-sl-a")

	err, msg := CheckManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"slot sl-a references non-existent description: d-sl-a"}, msg)

	//fix manifest for other tests
	testManifest.Manifest.Descriptions["d-sl-a"] = dsla
	err, _ = CheckManifest(testManifest.Manifest)
	assert.NoError(t, err)

}

func TestCheckManifestCatchMissingStream(t *testing.T) {

	testManifest.Lock()
	defer testManifest.Unlock()

	u := testManifest.Manifest.UIs["ui-b"]
	s := u.StreamsRequired
	u.StreamsRequired = []string{"st-c"}
	testManifest.Manifest.UIs["ui-b"] = u

	err, msg := CheckManifest(testManifest.Manifest)

	assert.Error(t, err)
	assert.Equal(t, []string{"ui ui-b references non-existent stream: st-c"}, msg)

	//fix manifest for other tests
	u.StreamsRequired = s
	testManifest.Manifest.UIs["ui-b"] = u
	err, _ = CheckManifest(testManifest.Manifest)
	assert.NoError(t, err)

}

//TODO increase test coverage to include all the checks we do on manifest
