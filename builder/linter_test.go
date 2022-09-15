package builder

import (
	"github.com/kcmvp/gob/boot"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type LinterTestSuit struct {
	suite.Suite
	linter  *Linter
	builder *Builder
}

func (linter *LinterTestSuit) SetupSuite() {
	linter.linter = newLinter()
	linter.builder = NewBuilder()
}

func TestLinterSuit(t *testing.T) {
	suite.Run(t, new(LinterTestSuit))
}

func (linter *LinterTestSuit) TestHappyFlow() {
	require.NotNil(linter.T(), linter.linter)
	require.NotNil(linter.T(), linter.builder)
}

func (linter *LinterTestSuit) TestLintContextValue() {
	tests := []struct {
		name     string
		command  boot.Command
		ctxValue string
	}{
		{
			boot.Report.Name(),
			boot.Report,
			"run -v --out-format json ./... --fix false golangci-lint-v1-49-0",
		},
		{
			boot.Lint.Name(),
			boot.Lint,
			"run -v --out-format json ./... --new-from-rev HEAD~ golangci-lint-v1-49-0",
		},
	}
	for _, test := range tests {
		linter.T().Run(test.name, func(t *testing.T) {
			lintAction(linter.builder, test.command)
			require.Equal(t, test.ctxValue, boot.ExecCtx(test.command))
		})
	}
}
