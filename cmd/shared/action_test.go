package shared

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLatestTag(t *testing.T) {
	// git@github.com:golangci/golangci-lint.git
	//ver, err := LatestVersion("https://github.com/golangci/golangci-lint.git", "v1.55.*")
	ver, err := LatestVersion("github.com/golangci/golangci-lint/cmd/golangci-lint", "v1.55.*")
	assert.NoError(t, err)
	assert.Equal(t, "v1.55.2", ver)
}
