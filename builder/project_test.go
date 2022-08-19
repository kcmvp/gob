//go:build integrated

package builder

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"log"
	"os"
	"strings"
	"testing"
)

type ProjectSuit struct {
	suite.Suite
	project *project
}

func TestProjectSuit(t *testing.T) {
	suite.Run(t, new(ProjectSuit))
}

func (suit *ProjectSuit) SetupTest() {
	suit.project = newProject(hook.DefaultHookCfg())
	os.Chdir(suit.project.ModuleDir())
}

func (suit *ProjectSuit) BeforeTest(suiteName, testName string) {
	if strings.Contains(testName, "CommitHook") {
		//suit.project.scanChanged = true
		//for _, f := range []string{"linter.go", "project.go"} {
		//	gof, err := os.OpenFile(filepath.Join(suit.project.ModuleDir(), "builder", f), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		//	if err != nil {
		//		log.Fatalf("failed open the file %s", f)
		//	}
		//	if _, err = gof.WriteString("//"); err != nil {
		//		log.Fatalf("failed write the file %s, %+v", f, err)
		//	}
		//	gof.Close()
		//}
	}
}

func CheckIfError(err error) {
	if err == nil {
		return
	}
	log.Fatalf("runs into error %+v", err)
}

func (suit *ProjectSuit) TestCoverage() {
	suit.project.Clean().Test()
}

func (suit *ProjectSuit) TestScanCommitHook() {
	suit.project.Clean().Scan()
	require.True(suit.T(), suit.project.Quality().LinterIssues.Files > 0)
}
