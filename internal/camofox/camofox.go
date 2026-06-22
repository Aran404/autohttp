package camofox

// Manager controls an external camofox-browser Node.js process.
type Manager struct{}

// New creates a Manager with the given config path.
func New(configPath string) *Manager {
	return &Manager{}
}
