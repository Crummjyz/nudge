# nudge

Spot when implementations change, but docs don't.

Running `nudge` with no arguments will check the current directory against `HEAD~`.
Or, list files/directories to check, and specify a commit/range with `-d`.

Currently works with Swift, Rust, and Go source files, but could support any language with a
[Tree Sitter parser](https://tree-sitter.github.io/tree-sitter/#parsers).

#### Check a PR with GitHub Actions

```yaml
- name: Nudge
  run: |
    cargo install --git https://github.com/Crummjyz/nudge
    nudge -d ${{ github.event.pull_request.base.sha }}..${{ github.event.pull_request.head.sha }}
```
