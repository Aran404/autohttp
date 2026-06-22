package analyze

import (
	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

// Result holds analysis output.
type Result struct {
	Dependencies []*pb.DependencyCandidate
	Noise        []*pb.NoiseCandidate
	Dynamic      []*pb.DynamicCandidate
	Operations   []*pb.LogicalOperationCandidate
}

// Analyzer performs deterministic dependency analysis on a recorded session.
type Analyzer struct{}

// New creates an Analyzer.
func New() *Analyzer {
	return &Analyzer{}
}

// Analyze runs deterministic analysis and returns candidates.
func (a *Analyzer) Analyze(session *pb.RecordedSession, trees []*pb.ParsedTree) (*Result, error) {
	return nil, nil
}
