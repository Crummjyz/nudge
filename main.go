package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/rust"
)

type WarningFormat int

const (
	DefaultWarningFormat int = iota
	GitHubWarningFormat
)

func (n *WarningFormat) UnmarshalText(b []byte) error {
	s := string(b)
	switch strings.ToLower(s) {
	case "default":
		*n = WarningFormat(DefaultWarningFormat)
	case "github":
		*n = WarningFormat(GitHubWarningFormat)
	default:
		return fmt.Errorf("invalid warning format: %s", s)
	}
	return nil
}

func main() {
	var args struct {
		Path          string        `arg:"positional" default:"." help:"path to git repository"`
		WarningFormat WarningFormat `arg:"--format" default:"default" help:"warning format"`
	}
	arg.MustParse(&args)

	r, err := git.PlainOpen(args.Path)
	if err != nil {
		log.Fatal(err)
	}

	patch := patch(r)[0]

	content, changes := collect(patch.Chunks())
	comments := sit(content, changes)

	for _, comment := range comments {
		_, file := patch.Files()
		path := file.Path()
		switch args.WarningFormat {
		case WarningFormat(DefaultWarningFormat):
			fmt.Printf("%s:%d:%d: unchanged: %s\n", path, comment.StartPoint().Row+1, comment.StartPoint().Column+1, comment.Content(content))
		case WarningFormat(GitHubWarningFormat):
			fmt.Printf("::warning file=%s,line=%d,col=%d::%s\n", path, comment.StartPoint().Row+1, comment.StartPoint().Column+1, "comment unchanged")
		}
	}
}

func patch(r *git.Repository) []diff.FilePatch {
	h, err := r.Head()

	a, err := r.CommitObject(h.Hash())
	b, err := a.Parent(0)

	patch, err := a.Patch(b)

	if err != nil {
		log.Fatal(err)
	}

	return patch.FilePatches()
}

func collect(chunks []diff.Chunk) (f []byte, changes [][2]int) {
	pos := 0
	for _, chunk := range chunks {
		content := chunk.Content()
		mode := chunk.Type()

		switch mode {
		case diff.Equal:
			pos += len(content)
			f = append(f, content...)
		case diff.Add:
			pos += len(content)
			f = append(f, content...)
			changes = append(changes, [2]int{pos, pos + len(content)})
		case diff.Delete:
			changes = append(changes, [2]int{pos, pos})
		}
	}

	return
}

func sit(file []byte, changes [][2]int) []*sitter.Node {
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
		log.Fatal(err)
	}

	comments := recurse(n, changes)
	return comments
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

func overlap(a, b [2]int) bool {
	return a[0] <= b[0] && b[0] <= a[1] || a[0] <= b[1] && b[1] <= a[1]
}
