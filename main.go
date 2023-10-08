package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/go-git/go-git/v5"
	sitter "github.com/smacker/go-tree-sitter"

	"fantail.dev/nudge/v2/internal"
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
		Path          string                 `arg:"positional" default:"." help:"path to git repository"`
		WarningFormat WarningFormat          `arg:"--format" default:"default" help:"warning format"`
		RevisionRange internal.RevisionRange `arg:"--revision" default:"HEAD..HEAD~" help:"revision range"`
	}
	arg.MustParse(&args)

	r, err := git.PlainOpen(args.Path)
	if err != nil {
		log.Fatal(err)
	}

	patches := internal.Diff(r, args.RevisionRange)
	for _, patch := range patches {
		content, changes := internal.Collect(patch.Chunks())
		comments := internal.Sit(content, changes)

		blocks := []struct {
			start, end sitter.Point
			content    string
		}{}

		blocks = append(blocks, struct {
			start, end sitter.Point
			content    string
		}{
			start:   comments[0].StartPoint(),
			end:     comments[0].EndPoint(),
			content: comments[0].Content(content) + "\n",
		})

		for _, comment := range comments[1:] {
			block := &blocks[len(blocks)-1]

			if comment.StartPoint().Row == block.end.Row+1 {
				block.end = comment.EndPoint()
				block.content += comment.Content(content) + "\n"
			} else {
				blocks = append(blocks, struct {
					start, end sitter.Point
					content    string
				}{
					start:   comment.StartPoint(),
					end:     comment.EndPoint(),
					content: comment.Content(content),
				})
			}
		}

		for _, block := range blocks {
			_, file := patch.Files()
			path := file.Path()
			switch args.WarningFormat {
			case WarningFormat(DefaultWarningFormat):
				fmt.Printf(
					"%s:%d:%d: comment unchanged: \n%s\n",
					path,
					block.start.Row+1,
					block.start.Column+1,
					block.content,
				)
			case WarningFormat(GitHubWarningFormat):
				fmt.Printf(
					"::warning file=%s,line=%d,col=%d,endLine=%d,endCol=%d::%s\n",
					path,
					block.start.Row+1,
					block.start.Column+1,
					block.end.Row+1,
					block.end.Column+1,
					"comment unchanged",
				)
			}
		}
	}
}
