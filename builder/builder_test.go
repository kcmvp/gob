package builder_test

import (
	"github.com/kcmvp/gos/builder"
	"testing"
)

func TestBuilder(t *testing.T) {
	b := builder.NewBuilder()
	b.Run(builder.Test)
}
