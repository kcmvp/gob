package boot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/thedevsaddam/gojsonq/v2"
)

const (
	allOver = "AllOver"
	pkg     = "pkg"
	file    = "file"
)

const reportJSON = "report.json"

type Issue struct {
	Total   int            `json:"Total,omitempty"`
	TypeMap map[string]int `json:"Linter,omitempty"`
}

type Metrics struct {
	Coverage string
	Issues   *Issue `json:"Issues,omitempty"`
}

type BuildReport struct {
	Metrics
	Pkgs []*PkgReport `json:"Packages,omitempty"`
}

// PkgReport package dimension.
type PkgReport struct {
	Name string
	Metrics
	Files []*FileReport
}

// FileReport file dimension.
type FileReport struct {
	Name string
	Metrics
}

func (report *BuildReport) Save(dir string, session *Session) error {
	data, err := json.MarshalIndent(report, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal project report:%w", err)
	}
	err = os.WriteFile(filepath.Join(dir, session.Specified(reportJSON)), data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to save project report:%w", err)
	}
	return nil
}

var reportAction Action = func(session *Session, project *Project, command Command) error {
	report := BuildReport{}
	data, err := os.ReadFile(filepath.Join(project.TargetDir(), session.Specified(reportJSON)))
	if err != nil {
		return fmt.Errorf("failed to create project report:%w", err)
	}
	err = json.Unmarshal(data, &report)
	if err != nil {
		return fmt.Errorf("failed to create project report:%w", err)
	}
	jq := gojsonq.New().File(filepath.Join(project.TargetDir(), session.Specified(LintJSONReport))).From(IssueNode)

	module := project.Mod().Module.Mod.Path
	jq.Select("FromLinter as Lint", "Pos.Filename as File")
	v := jq.Get().([]interface{})
	t3s := lo.Map(v, func(t interface{}, _ int) lo.Tuple3[string, string, string] {
		vt := t.(map[string]interface{})
		lint := vt["Lint"].(string)
		file := vt["File"].(string)
		items := strings.Split(file, string(os.PathSeparator))
		pkgName := module
		if len(items) == 2 {
			pkgName = filepath.Join(pkgName, items[0])
		}
		return lo.Tuple3[string, string, string]{
			A: lint,
			B: pkgName,
			C: filepath.Join(module, file),
		}
	})

	report.Issues = issueMap(allOver, t3s)[allOver]
	if report.Issues.Total == 0 {
		return nil
	}
	// package level report
	pkgIssuesMap := issueMap(pkg, t3s)
	fileIssuesMap := issueMap(file, t3s)
	for pkgName, pkgIssue := range pkgIssuesMap {
		pkgName := pkgName
		pkg, ok := lo.Find(report.Pkgs, func(t *PkgReport) bool {
			return pkgName == t.Name
		})
		if !ok {
			pkg = &PkgReport{
				Name: pkgName,
				Metrics: Metrics{
					Coverage: "-",
				},
			}
			report.Pkgs = append(report.Pkgs, pkg)
		}
		pkg.Issues = pkgIssue

		// file level report
		pkgFileIssues := lo.PickBy(fileIssuesMap, func(file string, issue *Issue) bool {
			return filepath.Dir(file) == pkg.Name
		})
		for fileName, fileIssue := range pkgFileIssues {
			fileName := fileName
			fileReport, ok := lo.Find(pkg.Files, func(fileReport *FileReport) bool {
				return fileName == fileReport.Name
			})
			if !ok {
				fileReport = &FileReport{
					Name: fileName,
				}
				fileReport.Coverage = "-"
				pkg.Files = append(pkg.Files, fileReport)
			}
			fileReport.Issues = fileIssue
		}
	}
	err = report.Save(project.TargetDir(), session)
	return err
}

func issueMap(levelName string, data []lo.Tuple3[string, string, string]) map[string]*Issue {
	var levels map[string][]lo.Tuple3[string, string, string]
	switch levelName {
	case pkg:
		levels = lo.GroupBy(data, func(t lo.Tuple3[string, string, string]) string {
			return t.B
		})
	case file:
		levels = lo.GroupBy(data, func(t lo.Tuple3[string, string, string]) string {
			return t.C
		})
	default:
		levels = map[string][]lo.Tuple3[string, string, string]{
			levelName: data,
		}
	}
	levelMap := map[string]*Issue{}
	for l, v := range levels {
		issue := &Issue{
			Total:   len(v),
			TypeMap: map[string]int{},
		}
		byIssueType := lo.GroupBy(v, func(t lo.Tuple3[string, string, string]) string {
			return t.A
		})
		for t, items := range byIssueType {
			issue.TypeMap[t] = len(items)
		}
		levelMap[l] = issue
	}
	return levelMap
}
