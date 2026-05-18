# ADR 0009: Declarative key bindings with bubbles/key

## Status

Accepted

## Context

The TUI needs a Key Bar — a persistent bottom strip showing Local Keys for the active Navigation Level and Global Keys available everywhere. Key Bar must expand on `h` to show the full key list, and input modes must suppress certain Global Keys (e.g. `q` during comment entry).

The existing approach handles keys imperatively: every render function appends hint strings inline, and `updateKey` dispatches via `msg.String() == "q"` comparisons. This creates two separate sources of truth — the display hints and the actual key logic — which drift apart as modes grow.

## Decision

Use `github.com/charmbracelet/bubbles/key` (`key.Binding`) for all key binding declarations. Each Navigation Level gets its own KeyMap struct (e.g. `ProjectListKeys`, `MRDetailKeys`). A single `GlobalKeys` struct holds `Quit`, `Back`, and `ToggleKeyBar` bindings shared across all levels.

Key dispatch in `updateKey` switches from string comparison to `key.Matches(msg, k.Up)`. Input modes call `k.SetEnabled(false)` on Global Keys for the duration of text entry. The Key Bar renderer reads `LocalKeys()` and `GlobalKeys` to build both the collapsed (truncated) and expanded (two-column) views.

A custom Key Bar renderer is written instead of using `bubbles/help`, because `bubbles/help` renders in its own style and does not support the two-section (local + separator + global) layout with a lipgloss border matching the rest of the UI.

## Consequences

- `charmbracelet/bubbles` is added as a direct dependency (currently indirect via Bubble Tea).
- Key declarations and key display are co-located in KeyMap structs — no drift between what is shown and what works.
- `key.Binding.SetEnabled(false)` is the single mechanism for input-mode suppression, replacing ad-hoc checks scattered through `updateKey`.
- The alternative — a custom `KeyBinding{Key, Desc string}` struct — was rejected because it lacks the built-in `Enabled()` toggle and would require re-implementing the multiple-alias (`q`/`ctrl+c`) pattern that `key.WithKeys` already provides.
