package internal

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
)

type RevisionRange struct {
	From, To plumbing.Revision
}

func (n *RevisionRange) UnmarshalText(b []byte) error {
	s := string(b)
	pos := strings.Index(s, "..")
	if pos == -1 {
		return fmt.Errorf("invalid revision range: %s", s)
	}

	n.From = plumbing.Revision(s[:pos])
	n.To = plumbing.Revision(s[pos+2:])
	return nil
}

// filePatches returns a list of file patches between two revisions.
func filePatches(r *git.Repository, revisions RevisionRange) ([]diff.FilePatch, error) {
	hashFrom, err := r.ResolveRevision(revisions.From)
	HashTo, err := r.ResolveRevision(revisions.To)
	if err != nil {
		return nil, fmt.Errorf("invalid revision range: %s..%s", revisions.From, revisions.To)
	}

	from, err := r.CommitObject(*hashFrom)
	to, err := r.CommitObject(*HashTo)
	if err != nil {
		return nil, err
	}

	patch, err := from.Patch(to)
	if err != nil {
		return nil, err
	}

	return patch.FilePatches(), nil
}

// collectChunks collects diff chunks into the (to) file and a list of byte ranges that changed.
func collectChunks(chunks []diff.Chunk) (file []byte, changes [][2]int) {
	pos := 0
	for _, chunk := range chunks {
		content := chunk.Content()
		mode := chunk.Type()

		switch mode {
		case diff.Equal:
			pos += len(content)
			file = append(file, content...)
		case diff.Add:
			pos += len(content)
			file = append(file, content...)
			changes = append(changes, [2]int{pos, pos + len(content)})
		case diff.Delete:
			changes = append(changes, [2]int{pos, pos})
		}
	}
	return
}
