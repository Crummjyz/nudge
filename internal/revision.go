package internal

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
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
