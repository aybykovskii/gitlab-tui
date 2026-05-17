# ADR 0007: Two-Pane Bubble Tea Navigation

**Status:** Accepted

The Go MVP will keep the product idea of a two-pane TUI through a Navigation Level Stack: the right pane always shows the active Navigation Level, and the left pane shows the parent/previous Navigation Level as context. The hierarchy is tool info → project selection → project sections → section entities → entity detail → optional entity tabs. The implementation will be Bubble Tea-native with focusable panes, keyboard navigation, and mouse selection/clicks rather than React-style context-based screen components.
