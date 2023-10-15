package warning

import (
	"fmt"
	"strings"
)

type WarningFormat int

const (
	// A human-readable format.
	DefaultWarningFormat int = iota
	// A format compatible with GitHub Actions.
	GitHubWarningFormat
)

func (f *WarningFormat) UnmarshalText(b []byte) error {
	s := string(b)
	switch strings.ToLower(s) {
	case "default":
		*f = WarningFormat(DefaultWarningFormat)
	case "github":
		*f = WarningFormat(GitHubWarningFormat)
	default:
		return fmt.Errorf("invalid warning format: %s", s)
	}
	return nil
}
