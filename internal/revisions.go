package internal

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/storage/memory"
)

type RevisionRange struct {
	From, To plumbing.Revision
}

func (n *RevisionRange) UnmarshalText(b []byte) error {
	s := string(b)
	pos := strings.Index(s, "..")
	if pos == -1 {
		n.From = plumbing.Revision(s)
		return nil
	}

	n.From = plumbing.Revision(s[:pos])
	n.To = plumbing.Revision(s[pos+2:])
	return nil
}

// filePatches returns a list of file patches between two revisions.
func filePatches(r *git.Repository, revisions RevisionRange) ([]diff.FilePatch, error) {
	if revisions.To == "" {
		return worktreeFilePatches(r, revisions.From)
	}

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
			length := len(content)
			file = append(file, content...)
			changes = append(changes, [2]int{pos, pos + length})
			pos += length
		case diff.Delete:
			changes = append(changes, [2]int{pos, pos})
		}
	}
	return
}

// worktreeFilePatches returns a list of file patches between a revision and the worktree.
func worktreeFilePatches(r *git.Repository, fromRevision plumbing.Revision) ([]diff.FilePatch, error) {
	// Hacky workaround to diff workingtree.

	// 1. Copy the repo and worktree to an in-memory repository.
	// 3. Commit these changes to the in-memory repository.
	// 4. Diff this commit to the from commit.

	// Setup in-memory repository.
	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	memFilesystem := memfs.New()
	memRepository, err := git.Clone(memory.NewStorage(), memFilesystem, &git.CloneOptions{
		URL:      w.Filesystem.Root(),
		Progress: nil,
	})
	if err != nil {
		return nil, err
	}
	memWorktree, err := memRepository.Worktree()
	if err != nil {
		return nil, err
	}

	// Copy worktree to in-memory repository.
	status, err := w.Status()
	if err != nil {
		return nil, err
	}

	for path, s := range status {
		switch s.Worktree {
		case git.Modified, git.Added:
			f, _ := w.Filesystem.Open(path)
			b, _ := io.ReadAll(f)

			memFile, _ := memFilesystem.Create(path)
			memFile.Write(b)
			memFile.Close()

			_, err := memWorktree.Add(path)
			if err != nil {
				return nil, err
			}
		}
	}

	// Commit changes to in-memory repository.
	toHash, err := memWorktree.Commit("nudge: working tree", &git.CommitOptions{})
	if err != nil {
		return nil, err
	}

	// Diff this commit to the from commit.
	fromHash, err := r.ResolveRevision("HEAD")
	if err != nil {
		return nil, err
	}

	from, err := r.CommitObject(*fromHash)
	to, err := memRepository.CommitObject(toHash)
	if err != nil {
		return nil, err
	}

	patch, err := from.Patch(to)
	if err != nil {
		return nil, err
	}

	return patch.FilePatches(), nil
}
