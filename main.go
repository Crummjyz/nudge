package main

import (
	"log"

	"github.com/alexflint/go-arg"
	"github.com/go-git/go-git/v5"

	"fantail.dev/nudge/v2/internal"
	"fantail.dev/nudge/v2/internal/warning"
)

type args struct {
	Path          string                 `arg:"positional" default:"." help:"path to git repository"`
	WarningFormat warning.WarningFormat  `arg:"--format" default:"default" help:"warning format"`
	RevisionRange internal.RevisionRange `arg:"--revisions" default:"HEAD~" help:"revision range"`
	IgnoreHeaders bool                   `arg:"--ignore-headers" default:"false" help:"ignore file header comments"`
}

func (args) Description() string {
	return "Spot when implementations change, but docs don't."
}

func (args) Epilogue() string {
	return `In an ideal world, docs don't depend on implementations.
In the real world, they do.`
}

func main() {
	var args args
	arg.MustParse(&args)

	r, err := git.PlainOpen(args.Path)
	if err != nil {
		log.Fatal(err)
	}

	warnings := internal.Nudge(r, args.RevisionRange, args.IgnoreHeaders)
	for _, warning := range warnings {
		warning.PrintWarning(args.WarningFormat)
	}
}
