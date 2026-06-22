package analyze

type Dependency struct {
	From       string
	To         string
	Value      string
	Path       string
	TargetPath string
	Confidence float64
	Reason     string
}

type Result struct {
	Dependencies []*Dependency
	Dynamic      []string
	Static       []string
}
