package warning

import (
	sitter "github.com/smacker/go-tree-sitter"
)

type Warning struct {
	Path       string
	Content    string
	Start, End sitter.Point
}
