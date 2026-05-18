# PRD: Key Bar — единая система хоткеев

## Problem Statement

Подсказки по клавишам сейчас вшиты строками внутрь каждой панели и не имеют единой логики. Один и тот же режим показывает одни подсказки, а обрабатывает другие клавиши — нет гарантии что отображаемое совпадает с работающим. Левая (неактивная) панель не блокирует обработку своих клавиш в правой. Пользователь не может увидеть полный список доступных клавиш для текущего Navigation Level.

## Solution

Добавить **Key Bar** — постоянную полосу с рамкой внизу экрана на всю ширину. Key Bar показывает две строки: Local Keys текущего Navigation Level (обрезанные по ширине с `…`) и Global Keys (всегда одинаковые). Клавиша `h` разворачивает Key Bar вверх, показывая полный список Local Keys в две колонки — верхняя часть, разделитель, Global Keys. Панели уменьшаются на высоту Key Bar. Каждый Navigation Level декларирует свои клавиши через `key.Binding` из `bubbles/key`. Input-режимы отключают Global Keys через `key.Binding.SetEnabled(false)`.

## User Stories

1. As a developer, I want to always see the available keys for the current Navigation Level at the bottom of the screen, so that I do not need to memorise every key binding.
2. As a developer, I want Local Keys and Global Keys shown on separate lines in the Key Bar, so that I can tell which keys are context-specific and which are always available.
3. As a developer, I want the Key Bar to have a border matching the main panes, so that the UI feels consistent.
4. As a developer, I want Local Keys truncated with `…` when they do not fit the terminal width, so that the Key Bar stays at a fixed height regardless of how many bindings the current level has.
5. As a developer, I want to press `h` to expand the Key Bar and see all Local Keys, so that I can discover less common bindings without leaving the current context.
6. As a developer, I want the expanded Key Bar to show Local Keys in two columns above a separator and Global Keys below, so that the layout is consistent with the collapsed view.
7. As a developer, I want the main panes to shrink when the Key Bar expands, so that the total screen height is preserved.
8. As a developer typing a comment or reply, I want `q` to input the letter rather than quit, so that text entry is not interrupted by global shortcuts.
9. As a developer in an input mode, I want the Key Bar to show input-specific hints (Enter: send, Esc: cancel) instead of global keys, so that I know how to complete or cancel the input.
10. As a developer, I want each Navigation Level to declare its own key bindings in one place, so that display and behaviour never drift apart.
11. As a developer, I want left-pane (Context Pane) key bindings to have no effect when the right pane is active, so that navigation is unambiguous.

## Implementation Decisions

- **Key Bar** is a new bordered strip rendered below the two main panes using `lipgloss.JoinVertical`. It spans the full terminal width.
- **Collapsed state**: 2 content lines + lipgloss border = 4 lines total. Line 1: Local Keys truncated to terminal width with `…`. Line 2: Global Keys (always the same).
- **Expanded state**: Key Bar grows to fit all Local Keys in two columns, separator line, then Global Keys. Main panes receive `paneHeight()` which subtracts the current Key Bar height.
- **`paneHeight()` helper**: introduced to centralise height calculation, replacing the scattered `max(8, m.height)` calls in every render function.
- **`key.Binding`** from `charmbracelet/bubbles/key` is used for all binding declarations (see ADR 0009). Custom Key Bar renderer replaces `bubbles/help` to support the two-section layout and lipgloss border.
- **Per-level KeyMap structs**: `ProjectListKeys`, `SectionsKeys`, `EntityListKeys`, `MRDetailKeys`, `DiffViewKeys`, `FileDiffKeys`. Each implements a `LocalKeys() []key.Binding` method.
- **`GlobalKeys` struct** (singleton): `Quit`, `Back`, `ToggleKeyBar` bindings shared by all levels.
- **Input mode suppression**: entering any text input mode (`commentInput`, `replyInput`, `mrCommentInput`, `editInput`, `ModeProjectInput`, `FocusFilter`) calls `globals.Quit.SetEnabled(false)` and `globals.Back.SetEnabled(false)`. Key Bar line 2 switches to input-specific hints. Exiting input mode restores `SetEnabled(true)`.
- **`h` global key**: toggles `model.keyBarExpanded bool`. Both collapsed and expanded states are rendered by the same Key Bar component; only height changes.
- **Existing inline hints** (strings appended at the bottom of each render function) are removed and replaced by Key Bar.
- **Context Pane key isolation**: left-pane bindings are not registered in `updateKey` — the Context Pane is read-only by design (ADR 0007), so no change in logic is needed; the isolation is already structural.

## Testing Decisions

- Tests should verify observable model state and rendered output structure, not internal implementation details.
- **Key Bar renderer**: unit-test collapsed output (two lines, correct truncation with `…`), expanded output (two-column local keys + separator + global keys), and height calculation with various terminal widths.
- **`paneHeight()`**: test that pane height equals `m.height - keyBarHeight()` in both collapsed and expanded states.
- **Input mode suppression**: model transition tests — enter `commentInput`, verify `Quit.Enabled() == false`; exit, verify `Quit.Enabled() == true`.
- **KeyMap structs**: each KeyMap declares the expected set of bindings; test that `LocalKeys()` returns non-empty slices and all bindings have non-empty help text.
- **`h` toggle**: model transition test — press `h`, verify `keyBarExpanded == true`; press again, verify `false`.
- **Key isolation**: verify that keys handled only in the active right pane (e.g. `A` for approve) produce no state change when the model is in a mode where they are not applicable.

## Out of Scope

- Mouse interaction with the Key Bar (clicking a hint to trigger the action).
- Configurable key remapping by the user.
- Animated expansion of the Key Bar.
- Showing key bindings for the Context Pane (it is read-only, no actionable bindings).
- `bubbles/help` full-screen help mode (`?` toggle).

## Further Notes

The Key Bar replaces all existing inline hint strings appended at the bottom of render functions. After this slice, no render function should contain hardcoded hint strings — all key documentation lives in KeyMap structs.
