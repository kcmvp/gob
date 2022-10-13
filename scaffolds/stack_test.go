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
		{"setup", "setup", []string{"builder", "githook", "linter"}},
		{"gen", "gen", []string{"config", "database"}},
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

func TestGetStack(t *testing.T) {

	tests := []struct {
		Name        string
		Module      string
		Description string
		DependsOn   string
	}{
		{
			"config",
			"github.com/spf13/viper",
			"Generate viper configuration for project.",
			"boot",
		},
		{
			"database",
			"-",
			"Generate sql data source based on configuration.",
			"config",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			st := getStack(test.Name)
			require.Equal(t, test.Module, st.Module)
			require.Equal(t, test.Description, st.Description)
			require.Equal(t, test.DependsOn, st.DependsOn)
		})
	}

}
