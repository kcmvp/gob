package builder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

type Report struct {
	Methods      int
	Tests        int
	Coverage     Coverage
	LinterIssues *LinterIssue
}

type Coverage struct {
	Method string
	Line   string
}

// @todo rename to Linter
// @ add Linter version.
type LinterIssue struct {
	Files  int
	Issues int
	Detail map[string]int
}
type TestCase struct {
	Package string
	Test    string
	Action  string
	Output  string
	Elapsed float64
}

type testCase struct {
	Package string
	Test    string
	Action  string
	Output  string
	Elapsed float64
}

func reporting() {

}

func (report *Report) GenLinterReport() {

}

func (report *Report) GenTestReport() {

}

func CoverageReport(rawTestReport, methodCoverageReport string, quality *Report) {
	file, err := os.Open(rawTestReport)
	if err != nil {
		log.Fatalln(color.RedString("failed to open the file %v \n", rawTestReport))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	testSet := map[string]bool{}
	for scanner.Scan() {
		text := scanner.Text()
		c := testCase{}
		json.Unmarshal([]byte(text), &c)
		testSet[c.Test] = true
	}

	quality.Tests = len(testSet)

	mc, err := os.Open(methodCoverageReport)
	if err != nil {
		log.Fatalln(color.RedString("failed to open the file %v", methodCoverageReport))
	}
	defer mc.Close()
	testedMethod := 0
	scanner = bufio.NewScanner(mc)
	for scanner.Scan() {
		text := scanner.Text()
		items := strings.Fields(text)
		coverage, _ := strconv.ParseFloat(strings.TrimRight(items[2], "%"), 64)

		if strings.EqualFold(items[0], "total:") {
			quality.Coverage.Line = items[2]
		} else {
			quality.Methods++
			if coverage > 0 {
				testedMethod++
			}
		}
	}
	quality.Coverage.Method = fmt.Sprintf("%.2f%%", float64(testedMethod)*100/float64(quality.Methods))
}

func buildTestReport(rawData, methodData string) {
	file, err := os.Open(rawData)
	if err != nil {
		log.Fatalln(color.RedString("failed to open the file %v \n", rawData))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	quality := Report{}

	testSet := map[string]bool{}
	for scanner.Scan() {
		text := scanner.Text()
		c := TestCase{}
		json.Unmarshal([]byte(text), &c)
		testSet[c.Test] = true
	}
	quality.Tests = len(testSet)

	mc, err := os.Open(methodData)
	if err != nil {
		log.Fatalln(color.RedString("failed to open the file %v", methodData))
	}
	defer mc.Close()
	testedMethod := 0
	scanner = bufio.NewScanner(mc)
	for scanner.Scan() {
		text := scanner.Text()
		items := strings.Fields(text)
		coverage, _ := strconv.ParseFloat(strings.TrimRight(items[2], "%"), 64)

		if strings.EqualFold(items[0], "total:") {
			quality.Coverage.Line = items[2]
		} else {
			quality.Methods++
			if coverage > 0 {
				testedMethod++
			}
		}
	}
	quality.Coverage.Method = fmt.Sprintf("%.2f%%", float64(testedMethod)*100/float64(quality.Methods))
}
