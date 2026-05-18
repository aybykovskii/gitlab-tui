# PRD: Issues Section — список и детальный просмотр

## Problem Statement

Issues Section в текущей реализации помечена как `available: false` — пункт в меню есть, но при выборе ничего не происходит. Пользователь не может просматривать, фильтровать или управлять GitLab Issues из TUI. Это разрывает рабочий процесс: для работы с задачами приходится переключаться в браузер.

## Solution

Реализовать Issues Section как полноценный Navigation Level: Entity List с двустрочным отображением каждой задачи, текстовый фильтр и переключение state (`s`), Issue Detail с двумя табами (Summary + Discussions), и набор действий — close/reopen, edit, label selector, assign self, comment, open in browser. Переиспользовать существующие компоненты: Label Selector, Discussion Tab, emoji-конфиг, Key Bar.

## User Stories

1. As a developer, I want to open the Issues section from the Sections Navigation Level, so that I can browse project tasks without leaving the TUI.
2. As a developer, I want each issue shown as two lines in the Entity List (IID + title, then author + labels + comment count), so that I can scan issues quickly.
3. As a developer, I want to filter issues by typing `/`, so that I can narrow the list by title, author, or label.
4. As a developer, I want to press `s` to cycle through opened / closed / all states, so that I can quickly switch between active and resolved issues.
5. As a developer, I want the current state filter shown in the Entity List header, so that I always know which issues I am looking at.
6. As a developer, I want to open an issue and see a Summary tab with all metadata, so that I can read the full context without opening a browser.
7. As a developer, I want the Summary tab to show author, assignees, state, labels, milestone, due date, and optionally weight, so that all relevant metadata is visible at a glance.
8. As a developer, I want weight shown only when it is set, so that the Summary is not cluttered for projects that do not use weight.
9. As a developer, I want a Discussions tab showing all comments on the issue, so that I can read the conversation history.
10. As a developer, I want to press `r` on a Discussion to reply, so that I can continue conversations without leaving the TUI.
11. As a developer, I want to press `c` to close or reopen the issue, so that I can update its state from the TUI.
12. As a developer, I want to press `e` to edit the issue title and description, so that I can correct mistakes or update scope.
13. As a developer, I want to press `l` to open the Label Selector for the issue, so that I can add or remove labels.
14. As a developer, I want to press `a` to assign or unassign myself on the issue, so that I can claim or release ownership without opening a browser.
15. As a developer, I want to press `m` to add a general comment, so that I can leave feedback without entering the Discussions tab.
16. As a developer, I want to press `o` to open the issue in a browser, so that I can access features not available in the TUI.
17. As a developer, I want issue IIDs displayed with `#` prefix (e.g. `#42`), so that the notation is consistent with GitLab conventions.
18. As a developer, I want state displayed with the same emoji as MR state (`🟢` opened, `🔴` closed), so that state is immediately recognisable across sections.

## Implementation Decisions

- **`internal/issue` package**: new package analogous to `internal/mr`. Types: `Issue` (IID, Title, Author, AuthorUsername, State, Labels, Assignees, Description, WebURL, CommentCount, Milestone, DueDate, Weight, Confidential), `Discussion` (reuse `mr.Discussion` or alias).
- **`internal/gitlab` additions**: `ListProjectIssues(state string, search string) ([]issue.Issue, error)` with `per_page: 50`; `CloseIssue(iid int) error`; `ReopenIssue(iid int) error`; `EditIssue(iid int, title, description string) error`; `UpdateIssueLabels(iid int, labels []string) error`; `AssignSelfIssue(iid int) error` / `UnassignSelfIssue(iid int) error`; `AddIssueComment(iid int, body string) error`; `ListIssueDiscussions(iid int) ([]issue.Discussion, error)`.
- **TUI — Section activation**: `SectionIssues` set to `available: true`. Entity List in issues mode loads `[]issue.Issue` instead of `[]mr.MergeRequest`. A generalized data container or separate issue-specific func types in `ProjectOptions`.
- **Entity List — two rows per issue**: row 1 `#IID Title`, row 2 `Author · [label1] [label2] · 💬 N`. Labels truncated if more than 2. Comment count shown only if > 0.
- **State filter**: `issueState string` in model (`"opened"` default, cycles to `"closed"`, then `""` for all on `s`). State shown in Entity List header: `Issues [opened]`. Filter change reloads from API.
- **Text filter**: `/` filters locally on the loaded list (title + author), same as MR filter.
- **Issue Detail**: `ModeDetail` with `TabSummary` and `TabDiscussions`. No `TabFiles` and no `TabReview`. `Tab` cycles between two tabs only.
- **Summary layout**: `#42 Title` → tabs → `👤 Author · assigned: Alice` → `🟢 opened · 💬 3` → `🏷️ [labels]` → `📅 Due: date · 🏁 Milestone` → (optional) `⚖️ Weight: N` → blank → description.
- **Discussions tab**: reuses existing Discussion rendering. Discussions for issues have no diff positions and no resolved state. `r` replies; no `x` (resolve), no `d` (draft). Reply calls `AddIssueComment` or a reply-to-discussion endpoint.
- **Actions**: `c` calls `CloseIssue` or `ReopenIssue` based on current state; `e` reuses edit input flow; `l` opens Label Selector (same `ModeLabelSelect` as MR); `a` calls `AssignSelfIssue`/`UnassignSelfIssue` (toggle based on whether current user is in assignees); `m` opens comment input; `o` opens WebURL.
- **Label Selector**: reused without modification — same `ModeLabelSelect`, `projectLabels` cache.
- **Emoji config**: reuses existing `EmojiConfig`; issues use same state emoji slots (`State`, `Labels`, `Assignees`).
- **Key Bar**: Issue Detail KeyMap struct declares its own Local Keys matching the action set above.

## Testing Decisions

Good tests verify observable behavior through public interfaces — model state after messages, rendered output — not internal field assignments.

- **`internal/issue`**: unit-test `Issue` struct field mapping from `client-go` types.
- **`internal/gitlab`**: unit-test `ListProjectIssues` (state param passed correctly, mapping), `CloseIssue`, `ReopenIssue`, `AddIssueComment` with fake interfaces.
- **Entity List**: render test — two-line format per issue, `#` prefix, label truncation at 2, comment count hidden when 0.
- **State filter**: `s` key cycles `issueState` through `opened → closed → "" → opened`; header label updates accordingly.
- **Text filter**: filter reduces visible list; cleared filter restores full list.
- **Issue Detail tabs**: `Tab` cycles between Summary and Discussions only (not Files, not Review).
- **Actions**: model transition tests — `c` calls correct func based on state; `l` opens `ModeLabelSelect`; `a` toggles assignee; `m` opens comment input.
- **Discussions**: `r` on a Discussion opens reply input; `Enter` calls `AddIssueComment`; no `x` key handler.
- Prior art: existing `model_test.go` and `client_test.go` patterns.

## Out of Scope

- Creating new issues from the TUI.
- Assigning other users (requires user search).
- Milestone management from the TUI.
- Issue boards or kanban view.
- Confidential issue creation or marking.
- Emoji reactions on comments.
- Extended filter panel (labels, milestone, assignee filter) — text + state covers MVP.

## Further Notes

Issues reuse all visual and interaction components built for MR Detail: Label Selector, Discussion Tab rendering, emoji config, Key Bar KeyMap pattern. The primary new work is `internal/issue` types, GitLab client methods, and wiring the Issues section into the existing Navigation Level Stack.
