# Nudge
_Spot when implementations change, but docs don't._

The `nudge` tool suggests documentation that might need review, based on changes to implementations.

### Options

Without any options, `nudge` checks the current repository's working tree against the latest commit.
Alternatively, specify a revision range like `--revisions=main..branch`. Ignore file header comments
with `--ignore-headers`. To use Nudge in GitHub Actions, pass the `--format=github` option. This
annotates found docs in the GitHub file viewer.

Nudge can work with any language that has a [Tree
Sitter](https://tree-sitter.github.io/tree-sitter/) grammar. Currently the following are
implemented:
- C
- Go
- Rust

### Visual Studio Code Extension

Folders in your open workspace are checked when a supported file is saved, and info squiggles are
added to relevant docs. Enable/disable linting with the 'Toggle Nudge' command.
