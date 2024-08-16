package command

import (
	"fmt"
	"github.com/kcmvp/gob/cmd/gob/artifact"
	"github.com/kcmvp/gob/utils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type ExecTestSuite struct {
	suite.Suite
}

func (suite *ExecTestSuite) BeforeTest(_, testName string) {
	os.Chdir(artifact.CurProject().Root())
	s, _ := os.Open(filepath.Join(artifact.CurProject().Root(), "cmd", "gbc", "testdata", "gob.yaml"))
	_, method := utils.TestCaller()
	root := filepath.Join(artifact.CurProject().Root(), "target", strings.ReplaceAll(method, "_BeforeTest", fmt.Sprintf("_%s", testName)))
	os.MkdirAll(root, os.ModePerm)
	t, _ := os.Create(filepath.Join(root, "gob.yaml"))
	io.Copy(t, s)
	t.Close()
	s.Close()
}

func (suite *ExecTestSuite) TearDownSuite() {
	_, method := utils.TestCaller()
	TearDownSuite(strings.TrimRight(method, "TearDownSuite"))
}

func TestExecSuite(t *testing.T) {
	suite.Run(t, &ExecTestSuite{})
}

func (suite *ExecTestSuite) TestActions() {
	assert.Equal(suite.T(), 3, len(execValidArgs()))
	assert.True(suite.T(), lo.Every(execValidArgs(), []string{artifact.CommitMsg, artifact.PreCommit, artifact.PreCommit}))
}

func (suite *ExecTestSuite) TestCmdArgs() {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"no match", []string{lo.RandomString(10, lo.LettersCharset)}, true},
		{"first match", []string{artifact.CommitMsg, lo.RandomString(10, lo.LettersCharset)}, false},
		{"second match", []string{lo.RandomString(10, lo.LettersCharset), "msghook"}, true},
		{"more than 3", []string{artifact.CommitMsg, lo.RandomString(10, lo.AlphanumericCharset),
			lo.RandomString(10, lo.AlphanumericCharset),
			lo.RandomString(10, lo.AlphanumericCharset)},
			true,
		},
	}
	for _, test := range tests {
		err := execCmd.Args(execCmd, test.args)
		assert.True(suite.T(), test.wantErr == (err != nil))
	}
}

func (suite *ExecTestSuite) TestValidateCommitMsg() {
	f, _ := os.CreateTemp("", "commit")
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	f.WriteString("#123: just for testing")
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no msg", []string{lo.RandomString(10, lo.LettersCharset)}, true},
		{"random msg", []string{lo.RandomString(10, lo.LettersCharset), lo.RandomString(10, lo.LettersCharset)}, true},
		{"valid msg", []string{lo.RandomString(10, lo.LettersCharset), f.Name(), "^#[0-9]+:\\s*.{10,}$"}, false},
	}
	for _, test := range tests {
		suite.T().Run(test.name, func(t *testing.T) {
			err := do(artifact.Execution{CmdKey: artifact.CommitMsg}, nil, test.args...)
			assert.True(t, test.wantErr == (err != nil))
		})
	}
}

func (suite *ExecTestSuite) TestPushDelete() {
	tests := []struct {
		name   string
		cmdKey string
		msg    string
		result bool
	}{
		{
			name:   "valid",
			cmdKey: artifact.PrePush,
			msg:    fmt.Sprintf("delete %s", pushDeleteHash),
			result: true,
		},
		{
			name:   "invalid",
			cmdKey: artifact.PrePush,
			msg:    pushDeleteHash,
			result: false,
		},
		{
			name:   "invalid",
			cmdKey: "abc",
			msg:    pushDeleteHash,
			result: false,
		},
	}
	for _, test := range tests {
		suite.T().Run(test.name, func(t *testing.T) {
			// Create a pipe
			reader, writer, err := os.Pipe()
			assert.NoError(suite.T(), err)
			defer reader.Close()
			// read from pipe
			os.Stdin = reader
			go func() {
				defer writer.Close()
				io.WriteString(writer, test.msg)
			}()
			rs := pushDelete(artifact.PrePush)
			assert.Equal(t, test.result, rs)
		})
	}
}

func (suite *ExecTestSuite) TestDo() {
	tests := []struct {
		name      string
		execution artifact.Execution
		wantErr   bool
	}{

		{
			name: "pre-push",
			execution: artifact.Execution{
				CmdKey:  artifact.PrePush,
				Actions: []string{"build"},
			},
			wantErr: false,
		},
		{
			name: "pre-commit",
			execution: artifact.Execution{
				CmdKey:  artifact.PreCommit,
				Actions: []string{"lint"},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		suite.T().Run(test.name, func(t *testing.T) {
			err := do(test.execution, execCmd)
			assert.True(suite.T(), test.wantErr == (err != nil))
		})
	}
}

func (suite *ExecTestSuite) TestRuneE() {
	err := execCmd.RunE(execCmd, []string{"pre-commit"})
	assert.NoError(suite.T(), err)
}
