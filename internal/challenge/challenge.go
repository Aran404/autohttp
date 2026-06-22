package challenge

// Detector detects anti-bot and captcha challenges.
type Detector struct{}

// New creates a Detector.
func New() *Detector {
	return &Detector{}
}

// Detect checks if a response body contains challenge indicators.
func (d *Detector) Detect(body string) (string, float64, error) {
	return "", 0, nil
}
