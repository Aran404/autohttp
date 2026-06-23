package camofox

import (
	"github.com/autohttp/autohttp/internal/browsers"
	"github.com/autohttp/autohttp/session"
)

// Manager controls an external camofox-browser Node.js process.
type Manager struct{}

// New creates a Manager with the given config path.
func New(configPath string) browsers.Browser {
	return &Manager{}
}

func (m *Manager) Navigate(url string, cfg ...browsers.BrowserConfig) error {
	return nil
}
func (m *Manager) Capture() ([]*session.Exchange, error) {
	return nil, nil
}
func (m *Manager) Close() error {
	return nil
}
