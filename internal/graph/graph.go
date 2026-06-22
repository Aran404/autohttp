package graph

import (
	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

// Engine converts analysis results into an executable graph.
type Engine struct{}

// New creates a graph Engine.
func New() *Engine {
	return &Engine{}
}

// Build constructs an execution graph from analysis results.
func (e *Engine) Build(analysis *pb.AnalysisResult) (*pb.ExecutionGraph, error) {
	return nil, nil
}
