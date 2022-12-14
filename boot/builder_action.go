package boot

import (
	"bufio"
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

	"github.com/PuerkitoBio/goquery"
	"github.com/samber/lo"

	"github.com/fatih/color"
)

const (
	testCoverOut  = "cover.out"
	testCoverHTML = "cover.html"
)

//go:embed template/*.tmpl
var templateDir embed.FS

var createDirAction Action = func(session *Session, project *Project, command Command) error {
	log.Println("Creating project directories")
	var dir string
	switch command.Name() {
	case SetupHook.Name(), SetupBuilder.Name():
		dir = project.ScriptDir()
	case Lint.Name(), Test.Name(), Build.Name(), Report.Name(),
		PreCommit.Name(), CommitMsg.Name(), PrePush.Name():
		dir = project.TargetDir()
	case SetupGitFlow.Name():
		dir = ".github/workflows"
	}
	if len(dir) < 1 {
		return nil
	}
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		err = fmt.Errorf("failed to create dir:%w", err)
	}
	return err
}

var cleanAction Action = func(session *Session, project *Project, command Command) error {
	log.Println("Cleaning project")
	flags := lo.FilterMap(command.ValidFlags(), func(flag string, i int) (string, bool) {
		return flag, session.GetFlagBool(command, flag) && flag != "delete"
	})
	args := append([]string{"clean"}, flags...)
	log.Printf("Flags: %s\n", strings.Join(flags, ","))
	output, err := exec.Command("go", args...).CombinedOutput()
	if err != nil {
		log.Println(color.RedString(string(output)))
		return err //nolint:wrapcheck
	}
	delAll := session.GetFlagBool(command, "delete")
	session.SaveCtxValue(command, fmt.Sprintf("%s %s delete=%v", "go", strings.Join(args, " "), delAll))
	err = filepath.WalkDir(project.TargetDir(), func(path string, d fs.DirEntry, err error) error {
		// @todo revisit the logic is correct or not
		if err == nil && !d.IsDir() && (delAll || strings.HasSuffix(d.Name(), session.ID())) {
			err = os.Remove(path)
			if err != nil {
				log.Println(color.YellowString("failed to delete %s:%s", path, err.Error()))
			}
		} else if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err //nolint:wrapcheck
	})
	if err != nil {
		return fmt.Errorf("failed to delete %s :%w", project.TargetDir(), err)
	}
	return err //nolint:wrapcheck
}

var commitMsgAction Action = func(session *Session, project *Project, command Command) error {
	log.Println("Validate commit message")
	input, _ := os.ReadFile(os.Args[1])
	return validateCommitMsg(string(input), string(project.Option().MsgPattern)) //nolint:wrapcheck
}

var lintAction Action = func(session *Session, project *Project, command Command) error {
	log.Println("Running linters against source code")
	return newLinter().scan(session, project, command) //nolint:wrapcheck
}

var testAction Action = func(session *Session, project *Project, command Command) error {
	err := os.Chdir(project.RootDir())
	log.Println("Running unit test")
	if err != nil {
		return fmt.Errorf("failed to change directory:%w", err)
	}
	params := []string{"test", "-v", "-coverpkg", "./...", "-coverprofile", filepath.Join(project.TargetDir(), session.Specified(testCoverOut))}
	// selective test scope in commit-msg hook, default are all packages
	scope := []string{"./..."}
	// @todo add test for this configuration
	selectiveTest := command == CommitMsg && !project.Config().GetBool(fmt.Sprintf("%s.testall", command.Hook()))
	if selectiveTest {
		changes, _ := changeSet(project)
		paths := lo.FilterMap(changes, func(c string, _ int) (string, bool) {
			p := strings.Split(c, string(os.PathSeparator))[0]
			_, err := os.Stat(filepath.Join(project.root, p))
			if strings.HasSuffix(p, ".go") || err != nil {
				return "", false
			}
			return fmt.Sprintf(".%s%s%s...", string(os.PathSeparator), p, string(os.PathSeparator)), true
		})
		paths = lo.Uniq[string](paths)
		if len(paths) > 0 {
			log.Println(color.GreenString("Selective tests in: %s", strings.Join(paths, " ")))
			log.Println(color.GreenString("Set 'gob.commit-msg.testall: true' to change the behavior"))
			scope = paths
		}
	}
	params = append(params, scope...)
	testCmd := exec.Command("go", params...)
	stdout, err := testCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("test failed:%w", err)
	}
	err = testCmd.Start()
	if err != nil {
		return fmt.Errorf("test failed:%w", err)
	}
	scanner := bufio.NewScanner(stdout)
	// ok  	github.com/kcmvp/gob/project	0.155s	coverage: 16.9% of statements
	pkr := regexp.MustCompile(`\sok\s+\S+\s+\S+s\s+coverage:\s+\S+% of statements`)
	// ?   	github.com/kcmvp/gob/infra	[no test files]
	ntr := regexp.MustCompile(`\s+\S+\s+\[no test files\]`)
	pkgCoverage := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.Contains(line, "[build failed]"):
			line = color.RedString(line)
			err = errors.New("build failed") //nolint
		case strings.Contains(line, "--- FAIL:"), strings.Contains(line, "FAIL"):
			line = color.RedString(line)
			err = errors.New("test failure") //nolint
		case pkr.MatchString(line):
			line = color.GreenString(line)
			find := strings.Fields(pkr.FindString(line))
			pkgCoverage[find[1]] = find[4]
		case ntr.MatchString(line):
			line = color.YellowString(line)
			find := strings.Fields(ntr.FindString(line))
			pkgCoverage[find[0]] = "-"
		}
		log.Println(line)
	}
	if err != nil {
		return fmt.Errorf("test failed:%w", err)
	}
	// run 'go tool cover -func ./targetDir/coverage.data' to get project level coverage
	params = []string{"tool", "cover", "-func", filepath.Join(project.TargetDir(), session.Specified(testCoverOut))}
	out, err := exec.Command("go", params...).CombinedOutput()
	if err != nil {
		log.Println(color.RedString("failed to get coverage report:%s", err.Error()))
		return fmt.Errorf("failed to get coverage report:%w", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")
	totalRep := regexp.MustCompile(`total:\s+\(statements\)\s+\S+`)
	report := BuildReport{}
	for _, line := range lines {
		if totalRep.MatchString(line) {
			report.Coverage = strings.Fields(line)[2]
		}
	}
	//  go tool cover -html ./targetDir/coverage.data to get detail coverage html report
	htmlReport := filepath.Join(project.TargetDir(), session.Specified(testCoverHTML))
	params = []string{"tool", "cover", "-html", filepath.Join(project.TargetDir(), session.Specified(testCoverOut)), "-o", htmlReport}
	out, err = exec.Command("go", params...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s:%w", string(out), err)
	}
	// generate file coverage report
	report.Pkgs = lo.MapToSlice(pkgCoverage, func(k string, v string) *PkgReport {
		return &PkgReport{
			Name: k,
			Metrics: Metrics{
				Coverage: v,
			},
			Files: []*FileReport{},
		}
	})
	reader, err := os.Open(htmlReport)
	if err != nil {
		return fmt.Errorf("failed to open coverage report:%w", err)
	}
	doc, _ := goquery.NewDocumentFromReader(reader)
	doc.Find("#files option").Each(func(i int, s *goquery.Selection) {
		fc := strings.Fields(s.Text())
		if len(fc) == 2 {
			file := &FileReport{
				Name: fc[0],
				Metrics: Metrics{
					Coverage: strings.ReplaceAll(strings.ReplaceAll(fc[1], "(", ""), ")", ""),
				},
			}
			for _, pkg := range report.Pkgs {
				if pkg.Coverage != "-" && strings.Contains(file.Name, pkg.Name) {
					pkg.Files = append(pkg.Files, file)
				}
			}
		}
	})

	err = report.Save(project.TargetDir(), session)
	if err != nil {
		return err //nolint
	}
	log.Printf("coverage report is generated at %s \n", filepath.Join(project.TargetDir(), testCoverHTML))
	return err //nolint:wrapcheck
}

var buildAction Action = func(session *Session, builder *Project, command Command) error {
	var targetFiles []string
	if len(targetFiles) == 0 {
		targetFiles = append(targetFiles, "main.go")
	}
	log.Println("build project ......")
	err := filepath.WalkDir(builder.RootDir(), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		for _, t := range targetFiles {
			if strings.EqualFold(d.Name(), t) {
				// @todo optimize need to read application.yml to determine the binary name
				if output, err := exec.Command("go", "build", "-o", builder.TargetDir(), path).CombinedOutput(); err != nil { //nolint
					log.Println(string(output))
					return err //nolint
				}
			}
		}
		return nil
	})
	if err != nil {
		err = fmt.Errorf("failed to build proejct:%w", err)
	}
	return err
}
