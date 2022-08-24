package infra

type Report struct {
	Tests    int
	Packages map[string]string
}
