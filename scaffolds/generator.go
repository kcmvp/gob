package scaffolds

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

	"github.com/spf13/viper"

	"github.com/samber/lo"

	"github.com/fatih/color"

	"github.com/kcmvp/gob/boot"
)

const (
	AppCfgName  = "application"
	genFlagName = "stack"
)

//go:embed staging
var stackDir embed.FS

var generate boot.Action = func(session *boot.Session, project boot.Project, command boot.Command) error {
	stack := getStack(session.GetFlagString(command, genFlagName))
	register := []string{}
	return visit(stack, project, register)
}

func visit(stack Stack, project boot.Project, register []string) error {
	if stack.Register {
		register = lo.Reverse(append(register, stack.Name))
	}
	err := scaffold(stack.Name, stack.Module, project, register)
	for len(stack.DependsOn) > 0 {
		t := getStack(stack.DependsOn)
		return visit(t, project, register)
	}
	return err //nolint
}

func scaffold(stack string, module string, project boot.Project, register []string) error {
	err := fs.WalkDir(stackDir, ".", func(path string, d fs.DirEntry, err error) error {
		if !strings.HasPrefix(d.Name(), stack) {
			return err
		}
		template, _ := fs.ReadFile(stackDir, path)
		switch d.Name() {
		case fmt.Sprintf("%s.tmpl", stack):
			genCode(string(template), stack, register, project)
		case fmt.Sprintf("%s.yml", stack):
			applicationYml(string(template), project)
		default:
			//
		}
		return err
	})
	if err != nil {
		return err //nolint
	}
	dependencies := strings.Split(module, ",")
	for _, dependency := range dependencies {
		// @todo need to optimize when the module exists then no need to run this command
		log.Printf("Adding dependency %s\n", dependency)
		cmd := exec.Command("go", "get", dependency)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(color.YellowString("Failed to get module: %s, please setup it manually", dependency))
			log.Println(color.YellowString(string(out)))
		}
	}
	return err //nolint
}

func ymlVariable(project boot.Project) map[string]interface{} {
	env := map[string]interface{}{}
	segs := strings.Split(project.Mod().Module.Mod.String(), "/")
	env["Module"] = segs[len(segs)-1]
	return env
}

func applicationYml(yml string, project boot.Project) {
	processed := boot.GenerateString(yml, ymlVariable(project))
	v1 := viper.New()
	v1.SetConfigType("yml")
	v1.ReadConfig(bytes.NewBuffer([]byte(processed))) //nolint:errcheck
	v2 := viper.New()
	v2.SetConfigName(AppCfgName)
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
			log.Println(color.YellowString("Failed to update %s.yml", AppCfgName))
		}
	}
	v1.WriteConfigAs(filepath.Join(project.RootDir(), fmt.Sprintf("%s.yml", AppCfgName))) //nolint
}

func genCode(text string, stack string, vars []string, project boot.Project) {
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
	err := boot.GenerateFile(text, filepath.Join(dir, fmt.Sprintf("%s.go", stack)), vars, stack == "boot")
	if err != nil {
		log.Println(color.YellowString("Failed to generate code:%s", err.Error()))
	}
}

func codeVariable(stack string, vars []string, project boot.Project) []string {
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
