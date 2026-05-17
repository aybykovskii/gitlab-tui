# PRD: Go Port of gitlab-tui

## Problem Statement

The current `gitlab-tui` implementation is tied to the React-like Ink ecosystem. That makes the code feel less direct than desired for a terminal-first GitLab review tool and encourages UI-bound state machines that are harder to extend and test. The user wants a simpler, more idiomatic Go codebase while preserving the useful parts of the current product: keyboard-driven MR review, two-pane context, GitLab REST integration, and eventually server-side review workflows.

## Solution

Build a parallel Go implementation inside this repository under the temporary binary name `gitlab-tui-go`. The Go version will use Bubble Tea, Bubbles, and Lipgloss for a keyboard and mouse-friendly TUI inspired by JiraTUI, and will follow Go project structure and maintenance practices inspired by jira-cli.

The first MVP is read-only MR review: resolve a GitLab project, load opened merge requests, browse the MR list, inspect MR details, and view a side-by-side diff. The current TypeScript/Ink implementation remains as the behavior reference until the Go version reaches parity.

## User Stories

1. As a developer, I want to run `gitlab-tui-go`, so that I can review GitLab merge requests from the terminal.
2. As a developer, I want the Go implementation to live alongside the TypeScript implementation, so that I can compare behavior during the port.
3. As a developer, I want the old TypeScript implementation to remain usable during the port, so that migration work does not block existing usage.
4. As a developer, I want a non-React TUI stack, so that the codebase feels simpler and more terminal-native.
5. As a developer, I want a two-pane layout, so that I can keep navigation context visible while reading MR details or diffs.
6. As a developer, I want keyboard navigation, so that I can operate the tool efficiently without leaving the keyboard.
7. As a developer, I want basic mouse support, so that I can click to focus/select items and use the wheel to scroll.
8. As a developer, I want mouse support to stay simple in the MVP, so that the port does not get blocked by complex drag, resize, or context menu behavior.
9. As a developer, I want `gitlab-tui-go init`, so that I can bootstrap a Go-specific config file.
10. As a developer, I want `gitlab-tui-go version`, so that I can verify which binary version I am running.
11. As a developer, I want the default `gitlab-tui-go` command to launch the TUI, so that normal usage is short and direct.
12. As a developer, I want config stored as YAML, so that I can read and edit it manually.
13. As a developer, I want config to use XDG paths, so that it follows Unix conventions.
14. As a developer, I want `GITLAB_TUI_CONFIG_FILE` to override the config path, so that I can test or switch configs easily.
15. As a developer, I want the config to support multiple accounts, so that I can later use several GitLab hosts/accounts without a schema migration.
16. As a developer, I want the MVP to use one default account, so that the first implementation stays simple.
17. As a developer, I want GitLab tokens read from environment variables, so that secrets are not written to config files.
18. As a developer, I want each account to declare its `token_env`, so that multi-account support works without storing secrets.
19. As a developer working inside a git repository, I want the project to be inferred from git remote, so that launching the tool usually requires no manual project selection.
20. As a developer outside a recognized git repository, I want to choose from recent projects, so that I can quickly return to projects I have reviewed before.
21. As a developer, I want a manual project input fallback, so that I can open any project even if detection and recent projects fail.
22. As a developer, I want recent projects stored in config with account, path, and last-used time, so that project selection is account-aware and sorted by usage.
23. As a developer, I want the MR list to show opened merge requests, so that the MVP focuses on the most common review workflow.
24. As a developer, I want a local text filter for the MR list, so that I can quickly find a relevant MR without extra API calls.
25. As a developer, I want selected MR details shown in the right pane, so that I can inspect author, branches, state, pipeline/approval summary, and description before opening a diff.
26. As a developer, I want Enter or mouse click to open the selected MR/diff path, so that keyboard and mouse navigation both work.
27. As a developer, I want a side-by-side diff view, so that the Go MVP matches the current review mental model.
28. As a developer, I want the diff view to be read-only in the MVP, so that the first release can validate navigation and rendering before write flows.
29. As a developer, I want Esc/back navigation from diff to detail, so that I can move through the review flow without restarting the app.
30. As a developer, I want data to load when entering a screen, so that the displayed state is fresh enough for active use.
31. As a developer, I want a manual refresh key, so that I can update MR data on demand.
32. As a developer, I do not want background polling in the MVP, so that the app avoids unnecessary API traffic and state complexity.
33. As a maintainer, I want the GitLab API layer to use the official GitLab Go client only, so that API access is consistent and supported.
34. As a maintainer, I want raw HTTP and alternate GitLab clients avoided, so that the API layer does not fragment.
35. As a maintainer, I want Bubble Tea models to avoid direct GitLab client access, so that UI state and application logic remain separated.
36. As a maintainer, I want MR data shaping outside the TUI layer, so that behavior can be tested without terminal rendering.
37. As a maintainer, I want diff parsing and projection outside the TUI layer, so that side-by-side rendering is fed by stable, tested data structures.
38. As a maintainer, I want a fake-data TUI shell as the first tracer bullet, so that layout and input behavior can be validated before API integration.
39. As a maintainer, I want Go unit tests for config loading, git remote detection, MR mapping, and diff parsing, so that core behavior survives refactors.
40. As a maintainer, I want Bubble Tea model transition tests, so that keyboard and mouse navigation behavior is verified without brittle terminal snapshots.
41. As a maintainer, I want Makefile targets for Go build, test, and lint, so that local development has a clear workflow.
42. As a maintainer, I want golangci-lint configured early, so that Go code quality stays consistent from the start.
43. As a maintainer, I want GoReleaser deferred, so that release automation does not distract from the MVP.
44. As a future contributor, I want the Go package layout to separate app, TUI, GitLab API, config, git detection, MR domain, diff domain, and IDE integration, so that I can find the right place for changes.
45. As a future contributor, I want ADRs to document the port decisions, so that the move away from Ink and GitBeaker is understandable later.
46. As a user of the current tool, I want write actions excluded from the first MVP, so that read-only review becomes stable before comments, drafts, approve, merge, create, or edit flows are ported.
47. As a user, I eventually want review write flows restored, so that the Go version can replace the TypeScript version after feature parity.
48. As a user, I want behavior to stay close to the existing tool where appropriate, so that the port feels like an evolution rather than a different product.

## Implementation Decisions

- The Go implementation will be built in parallel inside the existing repository.
- The temporary binary name is `gitlab-tui-go`; the final name can become `gitlab-tui` after feature parity.
- The TUI stack is Bubble Tea, Bubbles, and Lipgloss.
- JiraTUI is a UX reference, especially for keyboard and mouse-friendly interaction.
- jira-cli is a Go structure and maintenance reference.
- The first MVP is read-only MR review.
- The first tracer bullet is a fake-data two-pane TUI shell with MR list, MR detail, and fake side-by-side diff.
- The Go package layout will be organized around app orchestration, TUI, GitLab API, config, git detection, MR domain, diff domain, and IDE integration.
- The TUI layer must not call the GitLab client directly.
- Application use-cases and service interfaces sit outside Bubble Tea models.
- MR data shaping and diff parsing/projection are deep modules with small testable interfaces.
- GitLab API access uses only the official `gitlab.com/gitlab-org/api/client-go` library.
- The REST-over-GraphQL product direction remains in place for the Go port.
- The old Ink ADR is superseded by the Go/Bubble Tea port ADR.
- The old GitBeaker-specific REST ADR is superseded by the official GitLab Go client ADR.
- The old React-style Navigation Stack ADR is superseded by a Bubble Tea-native two-pane navigation decision.
- The Go config is new and does not reuse the TypeScript config format.
- Config is YAML and follows XDG config path conventions.
- `GITLAB_TUI_CONFIG_FILE` overrides config path resolution.
- Config supports multiple accounts but MVP uses one default account.
- Account tokens are provided by environment variables named in `account.token_env`.
- Secrets are not stored in YAML config.
- Project resolution order is `--project` override, git remote, recent projects list, first GitLab projects, then manual input.
- Positional project paths are not part of the CLI grammar; explicit project selection uses `--project <path>`.
- Optional CLI section aliases (`mr`, `issue`, `pipeline`) can deep-link into project sections; `gitlab-tui mr` inside a git repository opens the current project's merge requests, and `gitlab-tui mr 123` opens a specific merge request.
- Recent projects are stored in config as objects with account, path, and last-used timestamp.
- MR list MVP loads opened merge requests only.
- MR list filtering is local text filtering over already loaded MRs.
- Data refresh is load-on-enter plus explicit manual refresh.
- There is no background polling in MVP.
- Side-by-side diff is the only MVP diff mode.
- Mouse support in MVP is limited to click-to-focus/select and wheel scroll.
- CLI MVP includes default TUI launch, `init`, and `version`.
- Tooling MVP includes Go module setup, Makefile targets, and golangci-lint.
- GoReleaser and full release automation are deferred.

## Testing Decisions

Good tests should verify external behavior and stable contracts, not incidental implementation details. Pure modules should be tested through their public interfaces. Bubble Tea model tests should assert state transitions and emitted commands for user inputs rather than comparing full terminal render snapshots.

Modules to test in the MVP:

- Config path resolution, YAML load/save, account selection, token env lookup, and recent project updates.
- Git remote project detection and fallback behavior.
- MR domain mapping from GitLab client types into app-owned MR models.
- MR list filtering behavior.
- Diff parsing and side-by-side projection behavior.
- Bubble Tea model transitions for focus movement, list selection, filtering, diff open/back, click selection, and wheel scroll.
- Application use-case behavior using fake services.

Prior art in the current codebase includes unit tests for config manager, git detector, deep link parsing, stack building, diff parser/position logic, MR filters/mappers/title parsing, review/session behavior, and UI-adjacent screen tests. The Go port should keep the same preference for fast, isolated tests and avoid relying on live GitLab API calls in normal test runs.

## Out of Scope

- Replacing the TypeScript binary immediately.
- Deleting the TypeScript/Ink implementation before Go feature parity.
- Draft comments, instant comments, thread replies, resolve/unresolve, approve, merge, MR create, and MR edit in the first MVP.
- Unified diff mode in the first MVP.
- Drag resizing, context menus, hover-heavy interactions, and advanced mouse UX in the first MVP.
- Background polling or live auto-refresh.
- GraphQL API integration.
- Raw HTTP fallback around the official GitLab Go client.
- Keychain integration for token storage.
- Full account/project picker UX beyond default account, git remote detection, recent projects, and manual input.
- GoReleaser and full release packaging automation.
- Golden snapshot testing of full terminal output as the primary test strategy.

## Further Notes

The port should preserve useful product decisions from the current implementation while deliberately changing the implementation model. The goal is not a line-by-line rewrite; it is a Go-native TUI with clearer module boundaries.

The first fake-data tracer bullet is important because it validates the new architecture before API integration creates additional coupling. Once the shell proves the UX and boundaries, the next slices can connect config, project detection, GitLab API loading, MR detail data, and real diff rendering.

Published ADRs relevant to this PRD:

- ADR 0005: Parallel Go Port with Bubble Tea
- ADR 0006: Official GitLab Go Client for API Access
- ADR 0007: Two-Pane Bubble Tea Navigation
- ADR 0008: Strict TUI and Domain Separation in the Go Port
