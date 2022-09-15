package builder

type PkgReport struct {
	Name     string
	Coverage string
	Issue    *Issue `json:"issue,omitempty"`
	Files    []FileReport
}

type FileReport struct {
	Name     string
	Coverage string
	Issue    *Issue `json:"issue,omitempty"`
}

type Issue struct {
	Count    int            `json:"count,omitempty"`
	Category map[string]int `json:"category,omitempty"`
}

type Report struct {
	Issues *Issue `json:"issue,omitempty"`
	Pkgs   []*PkgReport
}
