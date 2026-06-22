package inspect

// Inspector provides CLI-accessible inspection of session artifacts.
type Inspector struct{}

// New creates an Inspector.
func New() *Inspector {
	return &Inspector{}
}

// PrintSession displays session details to the given writer.
func (ins *Inspector) PrintSession(path string) error {
	return nil
}
