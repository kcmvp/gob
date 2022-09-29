package scaffolds

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"strings"
	"testing"
)

func TestListStack(t *testing.T) {
	tests := []struct {
		name     string
		category string
		count    int
	}{
		{"setup", "setup", 3},
		{"gen", "gen", 2},
		{"run", "run", 5},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vs := ListStack(test.category)
			require.Equal(t, test.count, len(vs))
			for _, v := range vs {
				require.True(t, strings.HasPrefix(v.Command, fmt.Sprintf("gob %s", test.category)))
				if test.category == "run" {
					require.Equal(t, v.Module, "-")
				}
			}
		})
	}
	log.Println(tests)
}

func TestValidStack(t *testing.T) {
	tests := []struct {
		name     string
		category string
		args     []string
	}{
		{"setup", "setup", []string{"builder", "hook", "linter"}},
		{"gen", "gen", []string{"viper", "sql"}},
		{"run", "run", []string{"clean", "build", "test", "lint", "report"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := ValidStack(test.category)
			require.Equal(t, test.args, args)
		})
	}
	log.Println(tests)
}
