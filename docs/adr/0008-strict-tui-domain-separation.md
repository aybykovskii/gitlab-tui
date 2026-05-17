# ADR 0008: Strict TUI and Domain Separation in the Go Port

**Status:** Accepted

The Go port will keep Bubble Tea models focused on view state, input handling, and rendering commands. GitLab API access, application use-cases, MR data shaping, and diff parsing/projection live outside TUI packages so the code remains testable and extensible instead of recreating a large UI-bound state machine.
