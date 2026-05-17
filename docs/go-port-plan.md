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
- MVP flow: project selection → project sections → opened MR list → MR detail → changed files → side-by-side diff view.
- Mouse MVP: click-to-select in the active right pane and wheel scroll only; the left context pane is read-only.
- Data refresh: load on screen entry, manual refresh with `r`; no polling.
- CLI MVP: default command launches TUI; `init` creates config; `version` prints version.
- Config: new Go YAML config using XDG paths and `GITLAB_TUI_CONFIG_FILE` override.
- Secrets: tokens are read from env vars only; config stores `account.token_env`.
- Project resolution: `--project` override → git remote → recent projects list → first GitLab projects → manual input.
- Positional project paths are not part of the CLI grammar; explicit project selection uses `--project <path>`.
- Section resolution: optional section aliases (`mr`, `issue`, `pipeline`) open the matching project section directly; entity IDs follow the section alias, e.g. `gitlab-tui mr 123`.
- The left pane is a read-only context pane; `Esc`/back moves to the previous navigation level instead of focusing the left pane.
- When a project is resolved but no section is specified, the two-pane UI opens with the project list on the left and section selection on the right, with the resolved project highlighted.
- Project List contents are recent projects followed by the first 10 GitLab projects, with duplicates removed. Any project the user enters from Project List or opens through CLI is remembered immediately and moves to the top of recents.
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
- MR detail has three tabs: Summary, Discussions, Files.
- Discussions are visible both in the MR Discussions tab and inline in Diff View.
- MR diff uses Changed Files as the left context pane and Diff View as the active right pane
- Esc/back returns from diff to detail
- wheel scroll works in focused scrollable pane

GitLab API integration comes after the fake-data shell validates the UX and package boundaries.
