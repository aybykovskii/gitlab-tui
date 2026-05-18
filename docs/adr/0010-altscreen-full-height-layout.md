# ADR 0010: AltScreen full-height layout

## Status

Accepted

## Context

The TUI was launched without `tea.WithAltScreen()`, rendering inline in the terminal scroll buffer. This caused two problems: resize events produced re-render artifacts in the scroll history above the application, and layout elements like the Key Bar and Thread Panel could not be anchored to fixed positions (bottom of the screen) because the terminal height was not exclusively owned by the application.

## Decision

Add `tea.WithAltScreen()` to `tea.NewProgram(...)`. The application now occupies the full terminal height in an isolated alternate screen buffer. On exit, the terminal is restored to its previous state.

## Consequences

- Resize artifacts in scroll history are eliminated; `tea.WindowSizeMsg` continues to handle dynamic resizing correctly.
- The Key Bar is always anchored to the bottom of the screen.
- The Thread Panel can occupy a predictable portion of the available height without competing with scroll history.
- The trade-off: after exit, the terminal history shows no TUI output (the alternate buffer is discarded). Users who relied on scrolling back to see previous TUI renders lose that ability. This is the standard behaviour of full-screen terminal applications (vim, htop, lazygit) and is the expected UX for this class of tool.
- The alternative — keeping inline rendering and working around layout anchoring with cursor positioning — was rejected because it cannot reliably anchor elements to the bottom across all terminal emulators.
