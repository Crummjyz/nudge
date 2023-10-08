package internal

import (
	"log"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
)

func Diff(r *git.Repository, revisionRange RevisionRange) []diff.FilePatch {
	fromHash, err := r.ResolveRevision(revisionRange.From)
	toHash, err := r.ResolveRevision(revisionRange.To)

	from, err := r.CommitObject(*fromHash)
	to, err := r.CommitObject(*toHash)

	patch, err := from.Patch(to)
	if err != nil {
		log.Fatal(err)
	}

	return patch.FilePatches()
}

func Collect(chunks []diff.Chunk) (file []byte, changes [][2]int) {
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
