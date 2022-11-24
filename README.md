# nudge

Spot when implementations change, but docs don't.

Running `nudge` with no arguments will check the current directory against `HEAD~`.
Or, list files/directories to check, and specify a commit/range with `-d`.

#### Check a PR with GitHub Actions

```yaml
- name: Nudge
  run: |
    cargo install doc-nudge
    nudge -d ${{ github.event.pull_request.base.sha }}..${{ github.event.pull_request.head.sha }}
```
