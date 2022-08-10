//go:build gbt

package main

import (
	"bufio"
	"github.com/kcmvp/gbt/builder"
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
	//_, file, _, ok := runtime.Caller(0)
	//fmt.Printf(filepath.Dir(file))
	//os.Exit(1)
	// run test for all the modules

	project := builder.NewProject().Clean().Scan(refs)
	os.Exit(1)

}

func checkIfError(err error) {
	if err == nil {
		return
	} else {
		log.Fatalf("runs into error %v", err)
	}
}
