package record

import (
	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

// Recorder captures browser network activity into a RecordedSession.
type Recorder struct{}

// New creates a new Recorder.
func New() *Recorder {
	return &Recorder{}
}

// Capture starts recording and returns the recorded session.
func (r *Recorder) Capture() (*pb.RecordedSession, error) {
	return nil, nil
}
