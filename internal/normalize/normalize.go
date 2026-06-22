package normalize

import (
	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

// Normalizer converts raw recorder data into a canonical RecordedSession.
type Normalizer struct{}

// New creates a Normalizer.
func New() *Normalizer {
	return &Normalizer{}
}

// Normalize converts raw data into a canonical session.
func (n *Normalizer) Normalize(raw interface{}) (*pb.RecordedSession, error) {
	return nil, nil
}
