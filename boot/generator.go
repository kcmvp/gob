package boot

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/modfile"

	"github.com/spf13/viper"

	"github.com/samber/lo"

	"github.com/fatih/color"
)

const (
	AppCfgName  = "application"
	genFlagName = "stack"
)

//go:embed staging
var stackDir embed.FS

var generate Action = func(session *Session, project *Project, command Command) error {
	stack := getStack(session.GetFlagString(command, genFlagName))
	var register []string
	return visit(stack, project, register)
}

func visit(stack Stack, project *Project, register []string) error {
	if stack.Register {
		register = lo.Reverse(append(register, stack.Name))
	}
	err := scaffold(stack, project, register)
	for len(stack.DependsOn) > 0 {
		t := getStack(stack.DependsOn)
		return visit(t, project, register)
	}
	return err //nolint
}

func scaffold(stack Stack, project *Project, register []string) error {
	err := fs.WalkDir(stackDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !strings.HasPrefix(d.Name(), stack.Name) {
			return err
		}
		template, _ := fs.ReadFile(stackDir, path)
		switch {
		case strings.HasSuffix(d.Name(), ".tmpl"):
			genCode(string(template), stack.Name, register, project)
		case strings.HasSuffix(d.Name(), ".yml"):
			genYml(string(template), stack, project, strings.HasSuffix(d.Name(), "_test.yml"))
		default:
			//
		}
		return err
	})
	setupDependencies(stack, *project)
	return err //nolint
}

func setupDependencies(stack Stack, project Project) {
	dependencies := strings.Split(stack.Module, ",")
	dependencies = append(dependencies, strings.Split(stack.TestModule, ",")...)
	requires := lo.Map(project.Mod().Require, func(t *modfile.Require, i int) string {
		return t.Mod.Path
	})
	lo.ForEach(dependencies, func(dep string, _ int) {
		if !lo.Contains(requires, dep) {
			log.Printf("Adding dependency %s\n", dep)
			cmd := exec.Command("go", "get", dep)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Println(color.YellowString("Failed to get module: %s, please setup it manually", dep))
				log.Println(color.YellowString(string(out)))
			}
		} else {
			log.Println(color.YellowString("Module: %s exists ", dep))
		}
	})
}

func ymlVariable(stack Stack, project *Project, test bool) map[string]interface{} {
	env := map[string]interface{}{}
	segs := strings.Split(project.Mod().Module.Mod.String(), "/")
	env["Module"] = segs[len(segs)-1]
	if test {
		for k, v := range stack.TestEnvVariables {
			env[k] = v
		}
	}
	return env
}

func genYml(yml string, stack Stack, project *Project, test bool) {
	vars := ymlVariable(stack, project, test)
	processed := GenerateString(yml, vars)
	v1 := viper.New()
	v1.SetConfigType("yml")
	v1.ReadConfig(bytes.NewBuffer([]byte(processed))) //nolint:errcheck
	v2 := viper.New()
	fileName := AppCfgName
	if test {
		fileName = fmt.Sprintf("%s_test", AppCfgName)
	}
	v2.SetConfigName(fileName)
	v2.SetConfigType("yml")
	v2.AddConfigPath(project.RootDir())
	if err := v2.ReadInConfig(); err != nil {
		var t0 viper.ConfigFileNotFoundError
		if ok := errors.Is(err, t0); !ok {
			log.Fatalln(color.RedString("Failed to read configuration %s", err.Error()))
		}
	} else {
		cfg := v2.ConfigFileUsed()
		f, err := os.Open(cfg)
		if err != nil {
			log.Fatalf("Failed to read %s", cfg)
		}
		err = v1.MergeConfig(f)
		if err != nil {
			log.Println(color.YellowString("Failed to update %s.yml", fileName))
		}
	}
	v1.WriteConfigAs(filepath.Join(project.RootDir(), fmt.Sprintf("%s.yml", fileName))) //nolint
}

func genCode(text string, stack string, vars []string, project *Project) {
	lines := strings.Split(text, "\n")
	pkgDecl, found := lo.Find(lines, func(line string) bool {
		return len(line) > 0 && strings.HasPrefix(line, "package ")
	})
	dir, _ := os.Getwd()
	if found {
		dir = filepath.Join(project.RootDir(), strings.Fields(pkgDecl)[1])
		os.Mkdir(dir, os.ModePerm) //nolint
	}
	vars = codeVariable(stack, vars, project)
	err := GenerateFile(text, filepath.Join(dir, fmt.Sprintf("%s.go", stack)), vars, stack == "boot")
	if err != nil {
		log.Println(color.YellowString("Failed to generate code:%s", err.Error()))
	}
}

func codeVariable(stack string, vars []string, project *Project) []string {
	if strings.Compare(stack, "boot") != 0 {
		return vars
	}
	boot, err := os.Open(filepath.Join(project.RootDir(), "infra", "boot.go"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return vars
		}
		log.Fatalln(color.YellowString("Failed to open the file:%s", err.Error()))
	}
	bootReg := regexp.MustCompile(`func\s+Boot\(\s*\)\s+\{.*`)
	registerReg := regexp.MustCompile(`Register\((.*?)\)`)
	scanner := bufio.NewScanner(boot)
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		found = found || bootReg.MatchString(line)
		if found {
			if registerReg.MatchString(line) {
				rs := registerReg.FindStringSubmatch(line)
				_, ok := lo.Find(vars, func(v string) bool {
					return v == rs[1]
				})
				if !ok {
					vars = append(vars, rs[1])
				}
			}
		}
	}
	boot.Close()
	if err = scanner.Err(); err != nil {
		log.Fatalln(color.YellowString("Failed to open the file:%s", err.Error()))
	}
	return vars
}
