package browsers

import "github.com/autohttp/autohttp/session"

type BrowserConfig struct {
}

type Browser interface {
	Navigate(url string, cfg ...BrowserConfig) error
	Capture() ([]*session.Exchange, error)
	Close() error
}
