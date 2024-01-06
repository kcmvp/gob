package cmd

import (
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
	testDir string
}

func (suite *ExecTestSuite) SetupSuite() {
	s, _ := os.Open(filepath.Join(internal.CurProject().Root(), "gob.yaml"))
	os.MkdirAll(suite.testDir, os.ModePerm)
	t, _ := os.Create(filepath.Join(suite.testDir, "gob.yaml"))
	io.Copy(t, s)
	s.Close()
	t.Close()
	internal.CurProject().LoadSettings()
}

func (suite *ExecTestSuite) TearDownSuite() {
	os.RemoveAll(suite.testDir)
}

func TestExecSuite(t *testing.T) {
	_, file := internal.TestCallee()
	suite.Run(t, &ExecTestSuite{
		testDir: filepath.Join(internal.CurProject().Target(), file),
	})
}

func (suite *ExecTestSuite) TestActions() {
	assert.Equal(suite.T(), 3, len(execValidArgs()))
	assert.True(suite.T(), lo.Every(execValidArgs(), []string{internal.CommitMsgCmd, internal.PreCommitCmd, internal.PreCommitCmd}))
}

func (suite *ExecTestSuite) TestCmdArgs() {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"no match", []string{lo.RandomString(10, lo.LettersCharset)}, true},
		{"first match", []string{internal.CommitMsgCmd, lo.RandomString(10, lo.LettersCharset)}, false},
		{"second match", []string{lo.RandomString(10, lo.LettersCharset), "msghook"}, true},
		{"more than 3", []string{internal.CommitMsgCmd, lo.RandomString(10, lo.AlphanumericCharset),
			lo.RandomString(10, lo.AlphanumericCharset),
			lo.RandomString(10, lo.AlphanumericCharset)},
			true,
		},
	}
	for _, test := range tests {
		suite.T().Run(test.name, func(t *testing.T) {
			err := execCmd.Args(execCmd, test.args)
			assert.True(t, test.wantErr == (err != nil))
		})
	}
}

func TestValidateCommitMsg(t *testing.T) {
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
		t.Run(test.name, func(t *testing.T) {
			err := validateCommitMsg(execCmd, test.args...)
			assert.True(t, test.wantErr == (err != nil))
		})
	}
}
