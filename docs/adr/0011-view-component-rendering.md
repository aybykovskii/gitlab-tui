# ADR-0011 View Component Rendering

## Status

Accepted

## Context

The TUI `Model` struct accumulated 118 fields covering every mode, tab, scroll offset, and async state in the application. The `View()` method dispatched to free render functions (`renderRight()`, `renderFiles()`, `renderDiscussions()`) that built UI strings directly via `fmt.Sprintf` and manual switch/case blocks. This made it hard to reason about which state belonged to which screen and produced a growing file of render functions with no clear ownership.

The team explored adopting alternative TUI frameworks (`tview`, `gocui`) for their widget primitives, but these have incompatible event loops and would require discarding BubbleTea, Lipgloss, and all existing async message infrastructure.

## Decision

Decompose `Model` into typed **View Components** — sub-structs that own a logical slice of UI state and render it via `View(layout LayoutState) string`. The eight components are: `layout`, `projectPicker`, `entityList`, `mrDetail`, `issueDetail`, `diffView`, `labelSelector`, `input`.

When a component embeds a `bubbles` component (e.g. `viewport.Model` for scrollable panes), it also implements `Update(msg tea.Msg) tea.Cmd` to delegate message handling. The top-level `Model.Update()` routes messages to the relevant component's `Update()`. Business logic never moves into component methods.

Use `bubbles/viewport` inside `mrDetail`, `issueDetail`, and `diffView` to replace manual `scrollOffset` tracking.

Migrate incrementally: `TabsComponent` and `MRDetailState` first, remaining components in subsequent PRs.

## Alternatives considered

**Full `tea.Model` sub-components** — each component implements `Init/Update/View` as a complete nested model. Rejected because BubbleTea has no built-in event routing; every `WindowSizeMsg` and async domain message (`filesFinishedMsg`, `projectFinishedMsg`) would need manual forwarding at each level, producing more boilerplate than the current architecture.

**Keep render functions, add no structure** — minimal change, but leaves 118 flat fields and no clear ownership boundary as the app grows. Rejected because the decomposition also serves as the authoritative grouping for Update logic.

## Consequences

- `View()` in `view.go` becomes a composition of component `View()` calls, readable as a layout description.
- `Model` fields shrink from a flat list to eight named sub-structs; field access changes from `m.activeTab` to `m.mrDetail.ActiveTab`.
- Components that embed `viewport.Model` require an `Update()` method, making them slightly heavier than pure view structs — accepted as the cost of removing manual scroll tracking.
- `bubbles/viewport` scroll behaviour replaces manual `scrollOffset` arithmetic across three components.
