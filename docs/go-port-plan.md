# Go Port Plan

## Goal

Port `gitlab-tui` to Go to remove the React/Ink dependency and build a simpler, extensible TUI codebase.

## References

- JiraTUI: UX reference for keyboard and mouse-friendly TUI navigation.
- jira-cli: Go project structure and maintenance practices reference.

## Decisions

- Port in parallel inside this repository.
- Temporary binary name: `gitlab-tui-go`.
- TUI stack: Bubble Tea, Bubbles, Lipgloss.
- GitLab API client: official `gitlab.com/gitlab-org/api/client-go` only.
- First MVP: read-only MR review.
- MVP flow: project selection → opened MR list → MR detail → side-by-side diff view.
- Mouse MVP: click-to-focus/select and wheel scroll only.
- Data refresh: load on screen entry, manual refresh with `r`; no polling.
- CLI MVP: default command launches TUI; `init` creates config; `version` prints version.
- Config: new Go YAML config using XDG paths and `GITLAB_TUI_CONFIG_FILE` override.
- Secrets: tokens are read from env vars only; config stores `account.token_env`.
- Project resolution: git remote → recent projects list → manual input.
- Recent projects are stored in `config.yaml` as `{account, path, last_used_at}`.
- Tests: unit tests for pure logic plus Bubble Tea model transition tests.
- Tooling: `go.mod`, `Makefile`, `.golangci.yml`; GoReleaser later.

## Go package layout

```text
cmd/gitlab-tui-go/
internal/app/
internal/tui/
internal/gitlab/
internal/config/
internal/git/
internal/mr/
internal/diff/
internal/ide/
```

## First tracer bullet

Build a Go project scaffold and a Bubble Tea two-pane shell backed by fake MR data:

- left pane: opened MR list with local text filter
- right pane: MR detail summary
- Enter/click opens fake side-by-side diff in the right pane
- Esc/back returns from diff to detail
- wheel scroll works in focused scrollable pane

GitLab API integration comes after the fake-data shell validates the UX and package boundaries.
