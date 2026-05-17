# Task

Port `gitlab-tui` to Go in parallel with the existing TypeScript implementation.

## Goals
- Build a simpler, extensible Go TUI using Bubble Tea/Bubbles/Lipgloss.
- Keep the TypeScript/Ink app as behavior reference until Go feature parity.
- Deliver the first MVP as read-only MR review: project selection → opened MR list → MR detail → side-by-side diff.
- Track the work through GitHub issues under PRD #28.

## Checklist
- [x] Capture Go port PRD and publish it to GitHub issue #28.
- [x] Break PRD into vertical GitHub issues #29-#36.
- [x] Start #29: scaffold `gitlab-tui-go` binary and Go tooling.
- [x] Start #30: add YAML config init and loading.
- [x] Start #31: build fake-data two-pane Bubble Tea shell.
- [x] Continue dependent MVP issue #32: project resolution via git remote, recents, and input.
- [x] Continue dependent MVP issue #33: load opened merge requests from GitLab.
- [x] Continue dependent MVP issue #34: local MR filtering and manual refresh.
- [x] Continue dependent MVP issue #35: render real read-only side-by-side diff.
- [ ] Continue dependent MVP issue #36: validate read-only MVP against TS reference.

## Notes
- ADRs updated/created: #0005 Go/Bubble Tea port, #0006 official GitLab Go client, #0007 two-pane Bubble Tea navigation, #0008 strict TUI/domain separation.
- `CONTEXT.md` was removed as outdated.
- Iteration 2: started #29 by adding Go scaffold: `go.mod`, `cmd/gitlab-tui-go`, `internal/*` package layout, `Makefile`, `.golangci.yml`, version command, and `internal/app` tests. `make go-check` passes locally; `golangci-lint` is skipped because it is not installed in the environment.
- Iteration 3: completed #30 with YAML config model/load/save/init, XDG and `GITLAB_TUI_CONFIG_FILE` path resolution, env-based account tokens, `gitlab-tui-go init`, and config/app tests. `make go-check` passes.
- Commit created for completed #29/#30 with `Closes #29` and `Closes #30`.
- Iteration 4: completed #31 with Bubble Tea fake-data two-pane TUI shell, fake MR list/detail/diff, keyboard navigation, filter input, click selection/focus, wheel scroll, and model tests. `make go-check` passes.
- Commit created for completed #31 with `Closes #31`.
- Iteration 5: completed #32 with git remote parsing, project resolver, recent project config helpers, manual project input/recent project picker in TUI, and tests. `make go-check` passes.
- Commit created for completed #32 with `Closes #32`.
- Iteration 6: completed #33 with official GitLab Go client integration, opened MR loading, pagination, API-to-domain MR mapping, app startup real-data load when project is resolved, and tests. `make go-check` passes.
- Commit created for completed #33 with `Closes #33`.
- Iteration 7: completed #34 by wiring manual `r` refresh through an app-provided callback, preserving local MR filter behavior, adding loading/error refresh state, and model tests. `make go-check` passes.
- Commit created for completed #34 with `Closes #34`.
- Iteration 8: completed #35 with official GitLab diff loading, unified diff parser/projection into side-by-side rows, lazy TUI diff loading on Enter/click, loading/error states, and tests. `make go-check` passes.
- Commit created for completed #35 with `Closes #35`.
- Published issue links:
  - PRD: https://github.com/aybykovskii/gitlab-tui/issues/28
  - #29 scaffold/tooling: https://github.com/aybykovskii/gitlab-tui/issues/29
  - #30 config: https://github.com/aybykovskii/gitlab-tui/issues/30
  - #31 fake TUI shell: https://github.com/aybykovskii/gitlab-tui/issues/31
  - #32 project resolution: https://github.com/aybykovskii/gitlab-tui/issues/32
  - #33 GitLab MR loading: https://github.com/aybykovskii/gitlab-tui/issues/33
  - #34 filtering/refresh: https://github.com/aybykovskii/gitlab-tui/issues/34
  - #35 real side-by-side diff: https://github.com/aybykovskii/gitlab-tui/issues/35
  - #36 MVP validation: https://github.com/aybykovskii/gitlab-tui/issues/36
