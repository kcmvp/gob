package builder

import (
	"github.com/kcmvp/gob/cmd/gob/project"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestSupported(t *testing.T) {
	assert.Len(t, supported, 5)
	plugins := lo.Filter(supported, func(item project.Plugin, _ int) bool {
		return len(item.Url) > 0
		return true
	})
	assert.Len(t, plugins, 2)
	for _, plugin := range plugins {
		assert.True(t, strings.Count(plugin.Url, "@") == 1)
	}

}
