package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/go-git/go-git/v5"

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

		for _, comment := range comments {
			_, file := patch.Files()
			path := file.Path()
			switch args.WarningFormat {
			case WarningFormat(DefaultWarningFormat):
				fmt.Printf("%s:%d:%d: comment unchanged: %s\n", path, comment.StartPoint().Row+1, comment.StartPoint().Column+1, comment.Content(content))
			case WarningFormat(GitHubWarningFormat):
				fmt.Printf("::warning file=%s,line=%d,col=%d::%s\n", path, comment.StartPoint().Row+1, comment.StartPoint().Column+1, "comment unchanged")
			}
		}
	}
}
