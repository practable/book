package interval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	items := []Description{a, b, c}

	err, msg := CheckDescriptions(items)

	assert.NoError(t, err)

	items = []Description{a, b, c, d}

	err, msg = CheckDescriptions(items)

	assert.Error(t, err)

	assert.Equal(t, msg, []string{"Duplicate description definition #3: a"})

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

	items := []Policy{a, b, c}

	err, msg := CheckPolicies(items)

	assert.NoError(t, err)

	items = []Policy{a, b, c, d}

	err, msg = CheckPolicies(items)

	assert.Error(t, err)

	assert.Equal(t, msg, []string{"Duplicate policy definition #3: a"})

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

	items := []Resource{a, b, c}

	err, msg := CheckResources(items)

	assert.NoError(t, err)

	items = []Resource{a, b, c, d}

	err, msg = CheckResources(items)

	assert.Error(t, err)

	assert.Equal(t, msg, []string{"Duplicate resource definition #3: a"})

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

	items := []Stream{a, b, c}

	err, msg := CheckStreams(items)

	assert.NoError(t, err)

	items = []Stream{a, b, c, d}

	err, msg = CheckStreams(items)

	assert.Error(t, err)

	assert.Equal(t, msg, []string{"Duplicate stream definition #3: a"})

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

	items := []Slot{a, b, c}

	err, msg := CheckSlots(items)

	assert.NoError(t, err)

	items = []Slot{a, b, c, d}

	err, msg = CheckSlots(items)

	assert.Error(t, err)

	assert.Equal(t, msg, []string{"Duplicate slot definition #3: a"})

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

	items := []UI{a, b, c}

	err, msg := CheckUIs(items)

	assert.NoError(t, err)

	items = []UI{a, b, c, d}

	err, msg = CheckUIs(items)

	assert.Error(t, err)

	assert.Equal(t, msg, []string{"Duplicate UI definition #3: a"})

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

	items := []UISet{a, b, c}

	err, msg := CheckUISets(items)

	assert.NoError(t, err)

	items = []UISet{a, b, c, d}

	err, msg = CheckUISets(items)

	assert.Error(t, err)

	assert.Equal(t, msg, []string{"Duplicate UISet definition #3: a"})

}
