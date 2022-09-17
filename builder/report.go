package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kcmvp/gob/boot"
)

const reportJSON = "report.json"

type PkgReport struct {
	Name     string
	Coverage string
	Issues   *Issue `json:"Issues,omitempty"`
	Files    []*FileReport
}

type FileReport struct {
	Name     string
	Coverage string
	Issues   *Issue `json:"Issues,omitempty"`
}

type Issue struct {
	Total   int            `json:"Total,omitempty"`
	TypeMap map[string]int `json:"Linter,omitempty"`
}

type Report struct {
	Issues *Issue       `json:"Issues,omitempty"`
	Pkgs   []*PkgReport `json:"Packages,omitempty"`
}

func (report *Report) Save(dir string, session *boot.Session) error {
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
