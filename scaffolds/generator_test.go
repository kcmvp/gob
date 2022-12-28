package scaffolds

import (
	"bufio"
	"github.com/kcmvp/gob/boot"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type GeneratorTestSuite struct {
	suite.Suite
	project *Project
}

func (b *GeneratorTestSuite) SetupSuite() {

	pwd, _ := os.Getwd()
	b.project = NewProject(pwd)
}

func (b *GeneratorTestSuite) SetupTest() {
	os.RemoveAll(filepath.Join(b.project.RootDir(), "infra"))
}

func TestGeneratorSuit(t *testing.T) {
	suite.Run(t, new(GeneratorTestSuite))
}

func (b *GeneratorTestSuite) TestGenerateConfig() {
	session := boot.NewSession()
	session.BindFlag(boot.Generate, "stack", "config")
	session.Run(b.project, boot.Generate)
	v := viper.New()
	v.SetConfigType("yml")
	v.SetConfigName(AppCfgName)
	v.AddConfigPath(b.project.RootDir())
	v.ReadInConfig()
	require.Equal(b.T(), "gob", v.GetString("application.name"))
	_, err := os.Stat(filepath.Join(b.project.RootDir(), "infra", "boot.go"))
	require.NoError(b.T(), err)
	boot, err := os.Open(filepath.Join(b.project.RootDir(), "infra", "boot.go"))
	require.NoError(b.T(), err)
	defer boot.Close() //nolint
	registerReg := regexp.MustCompile(`Register\((.*?)\)`)
	scanner := bufio.NewScanner(boot)
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if registerReg.MatchString(line) && !found {
			found = strings.Contains(line, "config")
		}
	}
	require.True(b.T(), found)
}

func (b *GeneratorTestSuite) TestGenerateDatabase() {
	session := boot.NewSession()
	session.BindFlag(boot.Generate, "stack", "database")
	session.Run(b.project, boot.Generate)
	v := viper.New()
	v.SetConfigType("yml")
	v.SetConfigName(AppCfgName)
	v.AddConfigPath(b.project.RootDir())
	v.ReadInConfig()
	require.Equal(b.T(), "gob", v.GetString("application.name"))
	_, err := os.Stat(filepath.Join(b.project.RootDir(), "infra", "boot.go"))
	require.NoError(b.T(), err)
	boot, err := os.Open(filepath.Join(b.project.RootDir(), "infra", "boot.go"))
	require.NoError(b.T(), err)
	defer boot.Close() //nolint
	registerReg := regexp.MustCompile(`Register\((.*?)\)`)
	scanner := bufio.NewScanner(boot)
	databaseLine := 0
	configLine := 0
	database := false
	config := false
	num := 0
	for scanner.Scan() {
		num++
		line := scanner.Text()
		if registerReg.MatchString(line) {

			if strings.Contains(line, "database") {
				databaseLine = num
				database = true
			}
			if strings.Contains(line, "config") {
				configLine = num
				config = true
			}
		}
	}
	require.True(b.T(), database)
	require.True(b.T(), config)
	require.True(b.T(), configLine < databaseLine)
}

func (b *GeneratorTestSuite) TestGenerateConfigDatabase() {
	session := boot.NewSession()
	session.BindFlag(boot.Generate, "stack", "config")
	session.Run(b.project, boot.Generate)

	session = boot.NewSession()
	session.BindFlag(boot.Generate, "stack", "database")
	session.Run(b.project, boot.Generate)

	v := viper.New()
	v.SetConfigType("yml")
	v.SetConfigName(AppCfgName)
	v.AddConfigPath(b.project.RootDir())
	v.ReadInConfig()
	require.Equal(b.T(), "gob", v.GetString("application.name"))
	_, err := os.Stat(filepath.Join(b.project.RootDir(), "infra", "boot.go"))
	require.NoError(b.T(), err)
	boot, err := os.Open(filepath.Join(b.project.RootDir(), "infra", "boot.go"))
	require.NoError(b.T(), err)
	defer boot.Close() //nolint
	registerReg := regexp.MustCompile(`Register\((.*?)\)`)
	scanner := bufio.NewScanner(boot)
	databaseLine := 0
	configLine := 0
	database := false
	config := false
	num := 0
	for scanner.Scan() {
		num++
		line := scanner.Text()
		if registerReg.MatchString(line) {

			if strings.Contains(line, "database") {
				databaseLine = num
				database = true
			}
			if strings.Contains(line, "config") {
				configLine = num
				config = true
			}
		}
	}
	require.True(b.T(), database)
	require.True(b.T(), config)
	require.True(b.T(), configLine < databaseLine)
}
