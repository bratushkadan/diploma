package template_test

import (
	"fmt"
	"testing"

	"github.com/bratushkadan/floral/pkg/template"
	"github.com/stretchr/testify/assert"
)

func TestReplaceAllPairs(t *testing.T) {
	replacerName, replacerPlace := "{{name}}", "{{place}}"
	name, place := "Dan", "Kindgdom"

	tpl := fmt.Sprintf(`
    Hello, %s. %s, you're always welcome at %s.
    `, replacerName, replacerName, replacerPlace)
	expectedPartial := fmt.Sprintf(`
    Hello, %s. %s, you're always welcome at %s.
    `, name, name, replacerPlace)
	expectedFull := fmt.Sprintf(`
    Hello, %s. %s, you're always welcome at %s.
    `, name, name, place)

	assert.Equal(t, tpl, template.ReplaceAllPairs(tpl))
	assert.Equal(t, tpl, template.ReplaceAllPairs(tpl, ""))
	assert.Equal(t, tpl, template.ReplaceAllPairs(tpl, replacerName))

	assert.Equal(t, expectedPartial, template.ReplaceAllPairs(tpl, replacerName, name))
	assert.Equal(t, expectedPartial, template.ReplaceAllPairs(tpl, replacerName, name, replacerPlace))

	assert.Equal(t, expectedFull, template.ReplaceAllPairs(tpl, replacerName, name, replacerPlace, place))
}
