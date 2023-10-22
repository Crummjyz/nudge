package internal

import (
	"log"

	"github.com/crummjyz/nudge/v2/internal/find"
	"github.com/crummjyz/nudge/v2/internal/warning"
	"github.com/go-git/go-git/v5"
)

func Nudge(r *git.Repository, revisions RevisionRange, ignoreHeaders bool) (warnings []warning.Warning) {
	patches, err := filePatches(r, revisions)
	if err != nil {
		log.Fatal(err)
	}

	for _, patch := range patches {
		content, changes := collectChunks(patch.Chunks())
		comments, err := find.FindComments(content, changes)
		if err != nil {
			log.Fatal(err)
		}

		if len(comments) == 0 {
			continue
		}

		_, file := patch.Files()
		path := file.Path()

		warnings = append(warnings, warning.Warning{
			Path:    path,
			Content: comments[0].Content(content) + "\n",
			Start:   comments[0].StartPoint(),
			End:     comments[0].EndPoint(),
		})

		for _, comment := range comments[1:] {
			block := &warnings[len(warnings)-1]

			if comment.StartPoint().Row == block.End.Row+1 {
				block.End = comment.EndPoint()
				block.Content += comment.Content(content) + "\n"
			} else {
				warnings = append(warnings, warning.Warning{
					Path:    path,
					Content: comment.Content(content) + "\n",
					Start:   comment.StartPoint(),
					End:     comment.EndPoint(),
				})
			}
		}

		if ignoreHeaders {
			for len(warnings) > 0 && warnings[0].Start.Row == 0 {
				warnings = warnings[1:]
			}
		}
	}
	return
}
