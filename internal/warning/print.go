package warning

import "fmt"

func (w *Warning) PrintWarning(format WarningFormat) {
	switch format {
	case WarningFormat(DefaultWarningFormat):
		fmt.Printf(
			"%s:%d:%d: comment unchanged: \n%s\n\n",
			w.Path,
			w.Start.Row+1,
			w.Start.Column+1,
			w.Content,
		)
	case WarningFormat(GitHubWarningFormat):
		fmt.Printf(
			"::warning file=%s,line=%d,col=%d,endLine=%d,endCol=%d::%s\n",
			w.Path,
			w.Start.Row+1,
			w.Start.Column+1,
			w.End.Row+1,
			w.End.Column+1,
			"comment unchanged",
		)
	}
}
