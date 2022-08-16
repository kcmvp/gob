package githook

import (
	"strings"
)

var hooks = []string{"pre-commit", "commit-msg", "pre-push"}

func Hooks() map[string]string {
	m := map[string]string{}
	for _, hook := range hooks {
		m[hook] = strings.Replace(hook, "-", "_", 1)
	}
	return m
}

func Validate(msg string) {

}
