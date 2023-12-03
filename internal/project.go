package internal

import (
	"bufio"
	"github.com/fatih/color"
	"log"
	"os/exec"
	"strings"
)

var (
	Yellow *color.Color
	Red    *color.Color
	Blue   *color.Color
)

func init() {
	Yellow = color.New(color.FgYellow)
	Red = color.New(color.FgRed)
}

type Project struct {
	root   string
	module string
	deps   []string
}

func NewProject() *Project {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}:{{.Path}}")
	output, err := cmd.Output()
	if err != nil || len(string(output)) == 0 {
		log.Fatal(Red.Sprintf("Error executing command:", err))
	}

	item := strings.Split(strings.TrimSpace(string(output)), ":")
	project := &Project{
		root:   item[0],
		module: item[1],
	}
	cmd = exec.Command("go", "list", "-f", "{{if not .Standard}}{{.ImportPath}}{{end}}", "-deps")
	output, err = cmd.Output()
	if err != nil {
		log.Fatal(Red.Sprintf("Error executing command:", err))
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var deps []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			deps = append(deps, line)
		}
	}
	project.deps = deps
	return project
}

func (project *Project) Root() string {
	return project.root
}

func (project *Project) Module() string {
	return project.module
}
