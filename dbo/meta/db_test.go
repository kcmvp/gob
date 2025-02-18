package meta

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlatforms(t *testing.T) {
	assert.Len(t, Platforms(), 3)
}
