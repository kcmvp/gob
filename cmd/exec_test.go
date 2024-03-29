package cmd

import (
	"fmt"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type ExecTestSuite struct {
	suite.Suite
}

func (suite *ExecTestSuite) BeforeTest(_, testName string) {
	s, _ := os.Open(filepath.Join(internal.CurProject().Root(), "testdata", "gob.yaml"))
	root := filepath.Join(internal.CurProject().Root(), "target", fmt.Sprintf("cmd_exec_test_%s", testName))
	os.MkdirAll(root, os.ModePerm)
	t, _ := os.Create(filepath.Join(root, "gob.yaml"))
	io.Copy(t, s)
	s.Close()
	t.Close()
}

func (suite *ExecTestSuite) TearDownSuite() {
	TearDownSuite("cmd_exec_test_")
}

func TestExecSuite(t *testing.T) {
	suite.Run(t, &ExecTestSuite{})
}

func (suite *ExecTestSuite) TestActions() {
	assert.Equal(suite.T(), 3, len(execValidArgs()))
	assert.True(suite.T(), lo.Every(execValidArgs(), []string{internal.CommitMsg, internal.PreCommit, internal.PreCommit}))
}

func (suite *ExecTestSuite) TestCmdArgs() {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"no match", []string{lo.RandomString(10, lo.LettersCharset)}, true},
		{"first match", []string{internal.CommitMsg, lo.RandomString(10, lo.LettersCharset)}, false},
		{"second match", []string{lo.RandomString(10, lo.LettersCharset), "msghook"}, true},
		{"more than 3", []string{internal.CommitMsg, lo.RandomString(10, lo.AlphanumericCharset),
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
			err := do(internal.Execution{CmdKey: internal.CommitMsg}, nil, test.args...)
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
			cmdKey: internal.PrePush,
			msg:    fmt.Sprintf("delete %s", pushDeleteHash),
			result: true,
		},
		{
			name:   "invalid",
			cmdKey: internal.PrePush,
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
			rs := pushDelete(internal.PrePush)
			assert.Equal(t, test.result, rs)
		})
	}
}

func (suite *ExecTestSuite) TestDo() {
	tests := []struct {
		name      string
		execution internal.Execution
		wantErr   bool
	}{

		{
			name: "pre-push",
			execution: internal.Execution{
				CmdKey:  internal.PrePush,
				Actions: []string{"build"},
			},
			wantErr: false,
		},
		{
			name: "pre-commit",
			execution: internal.Execution{
				CmdKey:  internal.PreCommit,
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
