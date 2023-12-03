package tool

type Tool struct {
	Name string `json:"name"`
	Cmd  string `json:"cmd"`
	Url  string `json:"url"`
	Ver  string `json:"ver"`
}

func Tools() []Tool {
	return []Tool{}
}

func (tool Tool) Init() error {
	// install command to ~/.gob/tool.name/$version/tool.name
	return nil
}

func (tool Tool) Version(arg string) string {
	return ""
}
func (tool Tool) Path() string {
	return ""
}
