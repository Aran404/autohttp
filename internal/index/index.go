package index

import (
	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

// Index builds an inverted index of every scalar value across trees.
type Index struct{}

// New creates an Index.
func New() *Index {
	return &Index{}
}

// Build constructs the value index from parsed trees.
func (idx *Index) Build(trees []*pb.ParsedTree) error {
	return nil
}
