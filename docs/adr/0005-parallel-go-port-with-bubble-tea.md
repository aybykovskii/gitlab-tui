# ADR 0005: Parallel Go Port with Bubble Tea

**Status:** Accepted

We will port the TUI to Go in parallel inside this repository, keeping the current TypeScript/Ink implementation as a behavioral reference until feature parity. The Go TUI will use Bubble Tea, Bubbles, and Lipgloss to avoid the React-like ecosystem while retaining keyboard and mouse-driven terminal UX inspired by JiraTUI; Go project structure and release practices will follow idiomatic patterns similar to jira-cli.

## Consequences

ADR-0001 is superseded. Existing GitLab REST and Draft Notes decisions remain valid unless separately revisited.
