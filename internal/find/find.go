package find

import (
	"context"
	"errors"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/rust"
)

// FindComments finds all comments that apply to changed code, but have not themselves changed.
func FindComments(file []byte, changes [][2]int) ([]*sitter.Node, error) {
	langs := []*sitter.Language{rust.GetLanguage(), golang.GetLanguage(), c.GetLanguage()}

	var n *sitter.Node
	var err error
	for _, lang := range langs {
		n, err = sitter.ParseCtx(context.Background(), file, lang)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, errors.New("nudge: not a supported language")
	}

	return recurse(n, changes), nil
}

func recurse(node *sitter.Node, changes [][2]int) (comments []*sitter.Node) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)

		subchanges := [][2]int{}
		for _, change := range changes {
			if overlap(change, [2]int{int(child.StartByte()), int(child.EndByte())}) {
				subchanges = append(subchanges, change)
			}
		}

		if len(subchanges) > 0 {
			comments = append(comments, recurse(child, subchanges)...)
		} else if child.Type() == "line_comment" {
			comments = append(comments, child)
		}
	}

	return
}

// overlap reports whether the (ends-inclusive) ranges a and b overlap at all.
func overlap(a, b [2]int) bool {
	return a[0] <= b[0] && b[0] <= a[1] || a[0] <= b[1] && b[1] <= a[1]
}
