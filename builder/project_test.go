//go:build integrated

package builder

import (
	"github.com/stretchr/testify/suite"
	"log"
	"os"
	"testing"
)

type ProjectSuit struct {
	suite.Suite
	project *project
}

func TestProjectSuit(t *testing.T) {
	suite.Run(t, new(ProjectSuit))
}

func (suite *ProjectSuit) SetupTest() {
	suite.project = newProject(".")
	os.Chdir(suite.project.ModuleDir())
}

func CheckIfError(err error) {
	if err == nil {
		return
	}
	log.Fatalf("runs into error %+v", err)
}

func (suite *ProjectSuit) TestCoverage() {
	suite.project.clean()
	suite.project.test()
}
