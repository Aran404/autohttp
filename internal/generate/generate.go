package generate

import pb "github.com/autohttp/autohttp/gen/autohttp/v1"

// Generator emits standalone Go or Python scripts from an execution graph.
type Generator struct{}

// New creates a Generator.
func New() *Generator {
	return &Generator{}
}

// Generate emits a script for the given graph and target language.
func (g *Generator) Generate(graph *pb.ExecutionGraph, target string) ([]byte, error) {
	return nil, nil
}
