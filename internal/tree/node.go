package tree

import "strings"

type Node struct {
	Path        string
	Type        string
	RawValue    string
	Value       any
	Children    []*Node
	Entropy     float64
	Source      string
	SourceField string
	Shape       string
}

type Tree struct {
	Root       *Node
	ExchangeID string
}

func (t *Tree) Find(path string) *Node {
	if t == nil || t.Root == nil {
		return nil
	}
	parts := splitPath(path)
	return findNode(t.Root, parts, 0)
}

func findNode(n *Node, parts []string, depth int) *Node {
	if n == nil {
		return nil
	}
	if depth == len(parts) {
		return n
	}
	for _, child := range n.Children {
		if child.Path == parts[depth] {
			return findNode(child, parts, depth+1)
		}
	}
	return nil
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, ".")
}
