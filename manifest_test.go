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

	assert.Equal(t, msg, []string{"Duplicate named member #3: a"})

}
