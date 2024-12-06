package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDSMap(t *testing.T) {
	assert.Len(t, dsMap(), 1)
}
