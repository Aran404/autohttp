package index

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/autohttp/autohttp/internal/tree"
)

type Reference struct {
	Tree     *tree.Tree
	NodePath string
	Node     *tree.Node
}

type Index struct {
	mu      sync.RWMutex
	exact   map[string][]Reference
	decoded map[string][]Reference
	shape   map[string][]Reference
}

func New() *Index {
	return &Index{
		exact:   make(map[string][]Reference),
		decoded: make(map[string][]Reference),
		shape:   make(map[string][]Reference),
	}
}

func (idx *Index) Add(t *tree.Tree) {
	if t == nil || t.Root == nil {
		return
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	walkNode("", t.Root, idx, t)
}

func (idx *Index) Lookup(value string) []Reference {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if refs, ok := idx.exact[value]; ok {
		return refs
	}
	return nil
}

func walkNode(parentPath string, n *tree.Node, idx *Index, t *tree.Tree) {
	if n == nil {
		return
	}

	fullPath := parentPath
	if n.Path != "" {
		if fullPath == "" {
			fullPath = n.Path
		} else {
			fullPath = parentPath + "." + n.Path
		}
	}

	switch n.Type {
	case "string", "number", "boolean":
		raw := n.RawValue
		if raw == "" && n.Value != nil {
			raw = fmt.Sprintf("%v", n.Value)
		}
		if raw == "" {
			break
		}

		ref := Reference{Tree: t, NodePath: fullPath, Node: n}

		idx.exact[raw] = append(idx.exact[raw], ref)

		norm := strings.ToLower(strings.TrimSpace(raw))
		if norm != raw {
			idx.decoded[norm] = append(idx.decoded[norm], ref)
		}

		shape := n.Shape
		if shape == "" {
			shape = detectShape(raw)
		}
		if shape != "" {
			idx.shape[shape] = append(idx.shape[shape], ref)
		}
	}

	for _, child := range n.Children {
		walkNode(fullPath, child, idx, t)
	}
}

var (
	jwtRE    = regexp.MustCompile(`^[A-Za-z0-9\-_=]+\.[A-Za-z0-9\-_=]+\.[A-Za-z0-9\-_=]+$`)
	uuidRE   = regexp.MustCompile(`^(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	tsNanoRE = regexp.MustCompile(`^\d{16,19}$`)
)

func detectShape(value string) string {
	if jwtRE.MatchString(value) {
		return "jwt"
	}
	if uuidRE.MatchString(value) {
		return "uuid"
	}
	if isTimestamp(value) {
		return "timestamp"
	}
	return ""
}

func isTimestamp(value string) bool {
	if tsNanoRE.MatchString(value) {
		return true
	}

	if len(value) != 10 && len(value) != 13 {
		return false
	}
	for _, c := range value {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func (idx *Index) LookupDecoded(value string) []Reference {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	norm := strings.ToLower(strings.TrimSpace(value))
	if norm == value {
		if refs, ok := idx.decoded[norm]; ok {
			return refs
		}
		return nil
	}

	exact := idx.exact[norm]
	decoded := idx.decoded[norm]

	if len(exact) > 0 || len(decoded) > 0 {
		result := make([]Reference, 0, len(exact)+len(decoded))
		result = append(result, exact...)
		result = append(result, decoded...)
		return result
	}
	return nil
}

func (idx *Index) LookupByShape(shape string) []Reference {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if refs, ok := idx.shape[shape]; ok {
		return refs
	}
	return nil
}
