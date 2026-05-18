# PRD: Diff View — side-by-side, Thread Panel, Review Tab, AltScreen

## Problem Statement

Diff View в текущем виде имеет несколько критических проблем. Инлайновые комментарии вшиты между строками диффа, разрушая side-by-side читаемость. Отсутствует визуальное различие состояний Discussion (open, resolved, draft). Нет способа просмотреть накопленные Draft Comments перед сабмитом ревью. Приложение рендерится инлайн в терминале без AltScreen, из-за чего ресайз оставляет артефакты и Key Bar невозможно зафиксировать снизу.

## Solution

Перевести приложение на AltScreen (полная высота терминала, без артефактов при ресайзе). Убрать инлайновый рендеринг комментариев из строк диффа — вместо этого показывать полный тред в **Thread Panel** внизу правой панели при наведении курсора на закомментированную строку. Добавить `t` для скрытия/показа Thread Panel с сохранением гаттер-маркеров. Поддержать переключение между несколькими Discussion на одной строке (`[`/`]`). Добавить **Review Tab** — новый Entity Tab с просмотром накопленных Draft Comments, полем summary и действиями submit/discard.

## User Stories

1. As a reviewer, I want the TUI to occupy the full terminal height, so that the layout is stable and Key Bar is always anchored at the bottom.
2. As a reviewer, I want terminal resize to not produce render artifacts above the application, so that my terminal history stays clean.
3. As a reviewer, I want diff lines rendered without inline comment text between them, so that I can read the side-by-side diff without interruption.
4. As a reviewer, I want a gutter marker on lines that have Discussions or Draft Comments, so that I can see at a glance which lines have feedback without expanding threads.
5. As a reviewer, I want the Thread Panel to appear at the bottom of the right pane when my cursor is on a commented line, so that I can read the full Discussion in context.
6. As a reviewer, I want the Thread Panel to disappear when I move the cursor away from a commented line, so that it does not consume screen space unnecessarily.
7. As a reviewer, I want to press `t` to toggle Thread Panel visibility, so that I can read the diff without thread interruptions while still seeing gutter markers.
8. As a reviewer, I want the Thread Panel to show the full Discussion thread including all Notes, so that I can read the entire conversation.
9. As a reviewer, I want the Thread Panel to visually distinguish open, resolved, and draft states, so that I can understand the status of each thread at a glance.
10. As a reviewer, I want to press `[` and `]` to switch between multiple Discussions on the same line, so that I can read all threads at that position.
11. As a reviewer, I want a counter like `[1/3]` in the Thread Panel when multiple Discussions exist on a line, so that I know how many threads there are and which one I am viewing.
12. As a reviewer, I want emoji gutter markers (`📝` for draft, `💬` for discussion) when emoji is enabled, so that thread types are immediately distinguishable.
13. As a reviewer with emoji disabled, I want text gutter markers (`●` for draft, `○` for discussion) matching the TypeScript reference implementation, so that the diff is readable in any terminal.
14. As a reviewer, I want a Review Tab in MR Detail that lists all my Draft Comments with file and line context, so that I can review everything before submitting.
15. As a reviewer, I want to navigate to a Draft Comment from the Review Tab by pressing `Enter`, so that I can re-read the code context before confirming.
16. As a reviewer, I want pressing `Esc` in Diff View (when opened from Review Tab) to return me to the Review Tab, so that I can continue reviewing other drafts.
17. As a reviewer, I want an optional summary input field in the Review Tab, so that I can write a general review comment alongside my inline drafts.
18. As a reviewer, I want to press `p` in the Review Tab to submit all Draft Comments (and summary if written), so that I can publish my review in one action.
19. As a reviewer, I want to press `D` in the Review Tab to discard all Draft Comments, so that I can abandon a review that is no longer relevant.
20. As a reviewer, I want a draft count indicator visible in MR Detail (e.g. `📝 3 drafts`), so that I always know how many Draft Comments I have accumulated.

## Implementation Decisions

- **AltScreen**: add `tea.WithAltScreen()` to `tea.NewProgram`. See ADR 0010.
- **Inline comment removal**: `renderFileDiffPane` no longer appends Discussion or Draft Comment lines after each diff row. All thread content moves to Thread Panel.
- **Gutter markers**: two fixed-width columns left of each diff row. Column 1: `📝`/`●` if line has draft, `·` if line is in draft range, ` ` otherwise. Column 2: `💬`/`○` if line has Discussion, ` ` otherwise. Emoji vs text controlled by `EmojiConfig.Enabled`.
- **Thread Panel**: rendered at the bottom of the right pane when cursor is on a line with Discussion or Draft Comment. Separator line (`───`) divides diff content from Thread Panel. Height is dynamic — grows with thread content. `threadPanelVisible bool` in model, toggled by `t`. When hidden, Thread Panel is not rendered; gutter markers remain.
- **Thread Panel layout**: header line with state icon + author + `[N/M]` counter if multiple threads. Notes listed with author prefix. Draft Comments shown as `📝 [draft] body`. Resolved Discussions dimmed. Open Discussions normal color.
- **Multi-thread navigation**: `threadPanelCursor int` in model. `[` decrements, `]` increments, clamped to thread count for current line.
- **Reply/resolve from Thread Panel**: existing `r` (reply), `d` (draft reply), `x` (resolve/unresolve) keys operate on the thread currently shown in Thread Panel.
- **Review Tab**: new `TabReview` added to the Entity Tab sequence after `TabFiles`. `Tab` cycles through all four tabs.
- **Review Tab layout**: header with draft count, scrollable list of Draft Comments (file + line + body preview), separator, optional summary input field, submit/discard hints.
- **Review Tab navigation**: `↑/↓` moves cursor through Draft Comment list; `Enter` saves scroll position and opens `ModeFileDiff` at the comment's file and line; `Esc` in Diff View (when `returnToReview bool` is set) returns to Review Tab.
- **Submit from Review Tab**: `p` calls `SubmitDraftsFunc` (existing) + `PostMRCommentFunc` if summary is non-empty.
- **Draft count indicator**: shown in tab bar as `[Review (3)]` when drafts exist; zero drafts shows `[Review]`.

## Testing Decisions

Good tests verify observable behavior through public interfaces — model state transitions, rendered output structure — not internal field names or rendering implementation details.

- **AltScreen**: verify `tea.WithAltScreen()` is present in program options (integration-level check).
- **Thread Panel visibility**: model transition tests — cursor moves to commented line: Thread Panel content present in rendered output. Cursor moves away: Thread Panel absent. `t` toggles `threadPanelVisible`; gutter markers still present when hidden.
- **Multi-thread switching**: send `[`/`]`, assert `threadPanelCursor` changes; assert rendered Thread Panel shows correct thread author/body.
- **Gutter markers**: render test with `EmojiConfig.Enabled = true` and `false`; assert correct marker characters for draft-line, discussion-line, empty line.
- **Inline discussion removal**: render a diff with discussions; assert no discussion text appears between diff rows.
- **Review Tab**: model transition tests — `Tab` cycles to `TabReview`; draft list shows correct count and previews; `Enter` on draft opens correct file/line in `ModeFileDiff`; `Esc` returns to `TabReview`; `p` calls `SubmitDraftsFunc` with correct args.
- **Draft counter in tab bar**: render MR Detail with 0, 1, 3 drafts; assert tab label reflects count correctly.
- Prior art: existing `model_test.go` for message sequence and render tests.

## Out of Scope

- Syntax highlighting of diff content (separate PRD).
- Collapsing/expanding individual diff hunks.
- Scrolling within the Thread Panel separately from the diff (dynamic height covers this).
- Inline editing of existing Draft Comments from Review Tab (delete and re-add is sufficient for MVP).
- Review type selection (Approve / Request Changes / Comment) on submit — existing separate `A` key for approve covers the approval flow.

## Further Notes

The Thread Panel replaces the previous inline rendering entirely. After this slice, `renderFileDiffPane` must not contain any inline Discussion or Draft Comment text between diff rows — all thread content is exclusively in Thread Panel. This is a hard invariant enforced by the model tests.

Gutter marker fallbacks (`●`, `○`, `·`) match the TypeScript reference implementation to preserve visual familiarity for existing users.
