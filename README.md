# Nudge
_Spot when implementations change, but docs don't._

Nudge suggests documentation that might need review based on changes to
implementations. It does this by looking at your commit history, parsing the
files with changes, and flagging unchanged comments on changed code. Nudge
integrates nicely with CI pipelines and IDEs to make it easy to keep
documentation up to date.

## Usage

Run `nudge` in your repository to check unstaged changes. Alternatively, specify
a revision range like `--revisions=main..branch`. Ignore file header-comments
with `--ignore-headers`. To use Nudge with GitHub Actions, pass the
`--format=github` option, which will annotate suggestions in the GitHub file
viewer.

Nudge can work on any language that has a
[Tree-sitter](https://tree-sitter.github.io/tree-sitter/) grammar. Currently C,
Go, and Rust are implemented.

## Visual Studio Code Extension

When a supported file is saved, the extension checks all changes in your
workspace and adds info squiggles under relevant documentation. Enable/disable
linting with the 'Toggle Nudge' command.
