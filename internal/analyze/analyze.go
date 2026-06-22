package analyze

import (
	"sort"

	"github.com/autohttp/autohttp/internal/index"
	"github.com/autohttp/autohttp/internal/tree"
	"github.com/autohttp/autohttp/session"
)

func Analyze(sess *session.Session, trees []*tree.Tree) *Result {
	if sess == nil || len(trees) == 0 {
		return &Result{}
	}

	idx := index.New()
	exchangeOrder := buildExchangeOrder(sess)
	sorted := sortTreesByOrder(trees, exchangeOrder)

	for _, t := range sorted {
		idx.Add(t)
	}

	result := &Result{}
	seenDeps := make(map[string]bool)

	for _, t := range sorted {
		treePos := exchangeOrder[t.ExchangeID]
		walkAnalyze(t.Root, "", t, idx, exchangeOrder, treePos, result, seenDeps)
	}

	classifyFields(sorted, result)
	return result
}

func buildExchangeOrder(sess *session.Session) map[string]int {
	order := make(map[string]int, len(sess.Exchanges))
	for i, ex := range sess.Exchanges {
		order[ex.ID] = i
	}
	return order
}

func sortTreesByOrder(trees []*tree.Tree, order map[string]int) []*tree.Tree {
	sorted := make([]*tree.Tree, len(trees))
	copy(sorted, trees)
	sort.SliceStable(sorted, func(i, j int) bool {
		return order[sorted[i].ExchangeID] < order[sorted[j].ExchangeID]
	})
	return sorted
}

func walkAnalyze(n *tree.Node, parentPath string, t *tree.Tree, idx *index.Index, order map[string]int, treePos int, result *Result, seenDeps map[string]bool) {
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

	if n.Type == "string" || n.Type == "number" || n.Type == "boolean" {
		raw := n.RawValue
		if raw == "" {
			goto recurse
		}

		refs := idx.Lookup(raw)
		for _, ref := range refs {
			if ref.Tree.ExchangeID == t.ExchangeID {
				continue
			}
			refPos, ok := order[ref.Tree.ExchangeID]
			if !ok || refPos >= treePos {
				continue
			}

			depKey := ref.Tree.ExchangeID + ":" + ref.NodePath + ":" + t.ExchangeID + ":" + fullPath
			if seenDeps[depKey] {
				continue
			}
			seenDeps[depKey] = true

			result.Dependencies = append(result.Dependencies, &Dependency{
				From:       ref.Tree.ExchangeID,
				To:         t.ExchangeID,
				Value:      raw,
				Path:       ref.NodePath,
				TargetPath: fullPath,
				Confidence: 1.0,
				Reason:     "exact match",
			})
		}
	}

recurse:
	for _, child := range n.Children {
		walkAnalyze(child, fullPath, t, idx, order, treePos, result, seenDeps)
	}
}

func classifyFields(trees []*tree.Tree, result *Result) {
	fieldEntries := make(map[string][]entry)
	seen := make(map[string]bool)

	for _, t := range trees {
		walkCollect(t.Root, "", fieldEntries, seen)
	}

	for path, entries := range fieldEntries {
		isDynamic := false
		firstValue := entries[0].value

		for _, e := range entries {
			if isDynamicShape(e.shape) {
				isDynamic = true
				break
			}
			if e.value != firstValue {
				isDynamic = true
				break
			}
		}

		if isDynamic {
			result.Dynamic = append(result.Dynamic, path)
		} else {
			result.Static = append(result.Static, path)
		}
	}
}

func walkCollect(n *tree.Node, parentPath string, fieldEntries map[string][]entry, seen map[string]bool) {
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

	if n.Type == "string" || n.Type == "number" || n.Type == "boolean" {
		if fullPath != "" && !seen[fullPath+"="+n.RawValue+"@"+n.Source] {
			seen[fullPath+"="+n.RawValue+"@"+n.Source] = true
			fieldEntries[fullPath] = append(fieldEntries[fullPath], entry{
				value: n.RawValue,
				shape: n.Shape,
			})
		}
	}

	for _, child := range n.Children {
		walkCollect(child, fullPath, fieldEntries, seen)
	}
}

type entry struct {
	value string
	shape string
}

func isDynamicShape(shape string) bool {
	return shape == "jwt" || shape == "uuid" || shape == "timestamp"
}
