# PRD: MR Detail — навигация, визуальное оформление, лейблы

## Problem Statement

MR Detail View в текущем виде имеет три проблемы. Первая: `↑/↓` в активной правой панели переключают выбор в левой панели (списке MR) — нарушение принципа изоляции активной панели. Вторая: вся информация об MR — сплошной текст без визуальных акцентов, сложно быстро сканировать ключевые поля. Третья: лейблы отсутствуют полностью — ни отображения, ни редактирования, хотя это один из основных инструментов классификации MR в GitLab.

## Solution

Исправить навигацию: `↑/↓` в MR Detail скроллят правую панель, а не левую. Добавить emoji-индикаторы для ключевых полей Summary (автор, ветки, состояние, pipeline, approvals, лейблы) с возможностью отключить или переопределить маппинг через конфиг. Добавить отображение Labels как цветных «таблеток» с фоном из GitLab. Добавить Label Selector — интерактивный список для добавления/удаления лейблов. Добавить Draft-статус с возможностью переключения. Добавить Reviewers и Assignees в Summary.

## User Stories

1. As a reviewer, I want `↑/↓` in MR Detail to scroll the MR description and metadata, so that navigating long descriptions does not accidentally change the selected MR.
2. As a reviewer, I want emoji icons next to MR metadata fields, so that I can scan author, branch, state, and pipeline at a glance without reading labels.
3. As a reviewer, I want to disable emoji in the config, so that the TUI works cleanly in terminals or workflows where emoji are unwanted.
4. As a reviewer, I want to override individual emoji in the config, so that I can use my preferred symbols without replacing the whole set.
5. As a reviewer, I want merge request labels displayed as colored pills in MR Detail Summary, so that I can see the classification at a glance.
6. As a reviewer, I want label colors to match their colors in GitLab, so that the TUI is visually consistent with the web interface.
7. As a reviewer, I want project labels loaded once when the project opens, so that label colors are immediately available in every MR without extra requests.
8. As a reviewer, I want to press `l` in MR Detail Summary to open the Label Selector, so that I can add or remove labels without leaving the TUI.
9. As a reviewer, I want the Label Selector to show all project labels with their colors and current selection state, so that I can see what is available and what is already applied.
10. As a reviewer, I want to toggle a label with `Space` and save with `Enter`, so that label editing is fast and keyboard-driven.
11. As a reviewer, I want to cancel label editing with `Esc` without saving changes, so that accidental `l` presses do not modify the MR.
12. As a reviewer, I want to see whether a merge request is a Draft directly in its title, so that I know its readiness without reading the description.
13. As a reviewer, I want to press `d` to toggle Draft status on the selected merge request, so that I can mark it ready-for-review without opening GitLab web.
14. As a reviewer, I want to see the MR's Reviewers and Assignees in Summary, so that I know who is responsible for the review without opening GitLab web.
15. As a reviewer, I want state emoji to differ between opened, merged, and closed, so that MR state is immediately visible without reading the word.

## Implementation Decisions

- **Navigation fix**: in `ModeDetail`, `↑/↓` and `j/k` increment/decrement `m.rightTop` (scroll right pane), not `m.selected` (left pane selection). Left pane selection only changes in `ModeEntityList`.
- **Emoji config**: new `EmojiConfig` struct in `config` with `Enabled bool` and `Map EmojiMap`. `EmojiMap` has one string field per icon slot: `Author`, `Branch`, `State`, `Pipeline`, `Approvals`, `Labels`, `Reviewers`, `Assignees`, `Draft`. On load, absent map fields fall back to built-in defaults. `enabled: false` replaces all emoji with empty strings.
- **State emoji**: `opened` → `🟢`, `merged` → `🟣`, `closed` → `🔴`. Draft prefix shown before the title: `📝 Draft: Fix login bug`.
- **Label model**: new `Label` struct `{Name string, Color string}` in `internal/mr`. `MergeRequest` gains `Labels []string`, `Draft bool`, `Reviewers []string`, `Assignees []string`.
- **Label loading**: new `ListProjectLabelsFunc func() ([]Label, error)` loaded once via `LoadProject` alongside MR list. Cached as `projectLabels []mr.Label` in TUI model.
- **Label rendering**: each label rendered as a lipgloss pill with `.Background(lipgloss.Color(hex))` and contrasting foreground (white or black based on luminance). Rendered on a `🏷️ ` prefixed line in Summary.
- **Label Selector mode**: new `ModeLabelSelect`. Right pane shows all project labels with current selection state (`●` selected, `○` not). `↑/↓` navigate list, `Space` toggles, `Enter` calls `UpdateMRLabels` API, `Esc` cancels. Opened with `l` from `ModeDetail` on `TabSummary`.
- **Draft toggle**: `d` in `ModeDetail` calls `ToggleDraftMR func(iid int) error` which uses `EditMergeRequest` with draft title prefix. `MergeRequest.Draft` field updated optimistically.
- **GitLab client**: add `ListProjectLabels() ([]mr.Label, error)` and `UpdateMRLabels(iid int, labels []string) error` to client. Both use existing `client-go` endpoints.
- **Summary layout** (top to bottom):
  ```
  📝 !42 Draft: Fix login bug      (title line, Draft prefix if applicable)
  [>Summary<] [Discussions] [Files]

  👤 John Doe @johndoe  ·  👥 Alice, Bob  ·  assigned: Carol
  🌿 feature/fix → main
  🟢 opened  ·  ⚙️ ✓ success  ·  ✅ 2/3
  🏷️  [bug] [frontend]

  <description>
  ```

## Testing Decisions

- Tests should verify model state and rendered output through public interfaces, not internal field assignments.
- **Navigation fix**: model transition test — in `ModeDetail`, send `↑`/`↓` key, assert `m.rightTop` changes and `m.selected` does not.
- **Emoji config**: unit-test `EmojiConfig` merge — absent fields use defaults, present fields override, `enabled: false` produces empty strings for all slots.
- **Label rendering**: unit-test pill renderer with known hex colors; assert correct lipgloss style applied; assert foreground contrast selection (white vs black).
- **Label Selector**: model transition tests — open with `l`, navigate with `↑/↓`, toggle with `Space`, save with `Enter` (assert `UpdateMRLabels` called with correct label list), cancel with `Esc` (assert no API call).
- **Draft toggle**: model transition test — send `d`, assert `ToggleDraftMR` called, assert `MergeRequest.Draft` flipped.
- **GitLab client**: unit-test `ListProjectLabels` and `UpdateMRLabels` with fake `client-go` interfaces.

## Out of Scope

- Filtering the Entity List by label.
- Creating or deleting project-level labels from the TUI.
- Showing labels in the Entity List (left pane).
- Multi-line description scrolling with `PgUp`/`PgDn` (can be added later).
- Showing label colors in the Label Selector as background pills (text with `●` marker is sufficient for MVP).

## Further Notes

Label foreground contrast (white vs black) should be computed from the label's hex background luminance using the W3C relative luminance formula, not hardcoded. This ensures readability across the full range of GitLab label colors.
