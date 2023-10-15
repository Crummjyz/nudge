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
	pending := []*sitter.Node{}

	ended := false   // the comment has ended
	changed := false // the comment has changed

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)

		overlap := overlappingChanges(child, changes)

		switch child.Type() {
		case "line_comment":
			if ended {
				pending = nil
				ended = false
			}

			if len(overlap) > 0 {
				changed = true
				continue
			}

			pending = append(pending, child)
		default:
			ended = true

			if changed {
				pending = nil
				changed = false
			}

			if len(overlap) > 0 {
				comments = append(comments, pending...)
				pending = nil

				comments = append(comments, recurse(child, overlap)...)
			}

			// Language Specific Rules

			// We only want to match direct siblings, but property wrappers and attributes
			// complicate this, so we special case them. This may lead to some suboptimal code paths
			// in the above logic.

			if child.Type() != "attribute_item" {
				break
			}

			pending = nil
		}
	}
	return
}

func overlappingChanges(node *sitter.Node, changes [][2]int) (overlapping [][2]int) {
	for _, change := range changes {
		if overlap(change, [2]int{int(node.StartByte()), int(node.EndByte())}) {
			overlapping = append(overlapping, change)
		}
	}
	return
}

// overlap reports whether the (ends-inclusive) ranges a and b overlap at all.
func overlap(a, b [2]int) bool {
	return (b[0] >= a[0] && b[0] <= a[1]) || (b[1] >= a[0] && b[1] <= a[1]) || (a[0] >= b[0] && a[0] <= b[1]) || (a[1] >= b[0] && a[1] <= b[1])
}
