package work

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gob/project"
	"github.com/samber/mo"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	projects = map[string]*project.Project{}
)

// init initialize workspace
func init() {
	rs := mo.TupleToResult(exec.Command("go", "list", "-m", "-f", "{{.Dir}}_:_{{.Path}}").CombinedOutput())
	if rs.IsError() {
		log.Fatal(color.RedString("please execute command in workspace or project root directory"))
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(rs.MustGet()))
	for scanner.Scan() {
		text := scanner.Text()
		items := strings.Split(strings.TrimSpace(text), "_:_")
		fmt.Printf("--> %s \n", items[0])
		fmt.Printf("--> %s \n", items[1])
		//fmt.Println(len(projects))
		projects[items[0]] = project.NewProject(items[0], items[1])
	}
	fmt.Println(len(projects))
	//paths := lo.Map(lo.Keys(projects), func(key string, _ int) []string {
	//	return strings.FieldsFunc(key, func(r rune) bool {
	//		return r == os.PathSeparator
	//	})
	//})
	//uPath := paths[0]
	//lo.ForEach(paths, func(path []string, _ int) {
	//	uPath = lo.Intersect(uPath, path)
	//})
	//root := strings.Join(uPath, string(os.PathSeparator))
	//lo.ForEach(lo.Values(projects), func(project *project.Project, _ int) {
	//	project.SetWorkSpace(root)
	//})
	//// it's a multiple modules project, need to sort by dependency
	//if len(projects) > 1 {
	//	//
	//}
}

func Project() *project.Project {
	currentDir, _ := os.Getwd()
	if p, ok := projects[currentDir]; ok {
		return p
	}
	panic("Please execute command in the project root directory")
}
