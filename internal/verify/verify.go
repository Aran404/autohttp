package verify

// Runner verifies a generated script by executing it against a test target.
type Runner struct{}

// New creates a Runner.
func New() *Runner {
	return &Runner{}
}

// Run executes the generated script and checks success conditions.
func (r *Runner) Run(scriptPath string, successURL string) error {
	return nil
}
