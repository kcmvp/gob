package builder

import (
	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type ProjectSuit struct {
	suite.Suite
	project *Project
	repo    *git.Repository
}

func TestProjectSuit(t *testing.T) {
	suite.Run(t, new(ProjectSuit))
}

func (suit *ProjectSuit) SetupTest() {
	suit.project = NewProject()
	os.Chdir(suit.project.ModuleDir())
	if repo, err := git.PlainOpen(suit.project.ModuleDir()); err != nil {
		log.Fatalf("failed to read git repository %+v", err)
	} else {
		suit.repo = repo
	}
}

func (suit *ProjectSuit) BeforeTest(suiteName, testName string) {
	if strings.Contains(testName, "CommitHook") {
		suit.project.scanChanged = true
		for _, f := range []string{"linter.go", "project.go"} {
			gof, err := os.OpenFile(filepath.Join(suit.project.ModuleDir(), "builder", f), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalf("failed open the file %s", f)
			}
			defer gof.Close()
			if _, err = gof.WriteString("//"); err != nil {
				log.Fatalf("failed write the file %s, %+v", f, err)
			}
		}
	}
}

func (suit *ProjectSuit) TestScanCommitHook() {
	suit.project.Clean().Scan()
	require.True(suit.T(), suit.project.Quality().LinterIssues.Files > 0)
}
