//go:build gbt

package main

import (
	"bufio"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kcmvp/gbt/script"
	"log"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	refs := strings.Fields(line)
	// do nothing for push delete
	if strings.Contains(refs[0], "delete") {
		os.Exit(0)
	}
	cqc := script.NewCQC(0.35, 0.85)
	rep, err := git.PlainOpen(cqc.RootDir())
	checkIfError(err)
	w, err := rep.Worktree()
	checkIfError(err)
	status, err := w.Status()
	checkIfError(err)
	// check uncommitted changes by git status --porcelain
	if len(status) > 0 {
		fmt.Println("uncommitted files found, please commit them first")
		fmt.Printf("%+v", status)
		os.Exit(1)
	}

	cqc.Clean().Test()
	if cqc.Error() != nil {
		fmt.Printf("test failed %v \n", cqc.Error())
		os.Exit(1)
	}

	// check coverage
	l, err := rep.CommitObject(plumbing.NewHash(refs[1]))
	r, err := rep.CommitObject(plumbing.NewHash(refs[3]))
	p, err := r.Patch(l)
	checkIfError(err)
	for _, patch := range p.FilePatches() {
		f, t := patch.Files()
		fmt.Printf(f.Path())
		fmt.Printf(t.Path())
	}

	os.Exit(1)

}

func checkIfError(err error) {
	if err == nil {
		return
	} else {
		log.Fatalf("runs into error %v", err)
	}
}
