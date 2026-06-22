package tree

import (
	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

// Parser converts request/response artifacts into typed trees.
type Parser struct{}

// New creates a Parser.
func New() *Parser {
	return &Parser{}
}

// ParseSession converts every exchange in the session into parsed trees.
func (p *Parser) ParseSession(session *pb.RecordedSession) ([]*pb.ParsedTree, error) {
	return nil, nil
}
