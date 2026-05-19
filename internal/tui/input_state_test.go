package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInputStateIsInactiveByDefault(t *testing.T) {
	t.Parallel()

	state := NewInputState()

	if state.Active() {
		t.Fatal("expected new InputState to be inactive")
	}
}

func TestInputStateIsActiveWhenAnyFlagSet(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"comment", "mrComment", "issueComment", "edit", "reply"} {
		state := NewInputState()
		switch name {
		case "comment":
			state.commentInput = true
		case "mrComment":
			state.mrCommentInput = true
		case "issueComment":
			state.issueCommentInput = true
		case "edit":
			state.editInput = true
		case "reply":
			state.replyInput = true
		}

		if !state.Active() {
			t.Fatalf("expected Active()=true when %s is set", name)
		}
	}
}

func TestInputStateValueIsEmptyAfterBegin(t *testing.T) {
	t.Parallel()

	state := NewInputState()
	state.BeginWithValue("some prior value")
	state.Begin()

	if state.Value() != "" {
		t.Fatalf("expected empty value after Begin(), got %q", state.Value())
	}
}

func TestInputStateBeginWithValueSetsValue(t *testing.T) {
	t.Parallel()

	state := NewInputState()
	state.BeginWithValue("Draft: My feature")

	if state.Value() != "Draft: My feature" {
		t.Fatalf("expected %q, got %q", "Draft: My feature", state.Value())
	}
}

func TestInputStateUpdateRoutesCharacterKeysToTextInput(t *testing.T) {
	t.Parallel()

	state := NewInputState()
	state.Begin()

	state.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	state.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	if state.Value() != "hi" {
		t.Fatalf("expected value %q after typing, got %q", "hi", state.Value())
	}
}

func TestInputStateResetClearsValue(t *testing.T) {
	t.Parallel()

	state := NewInputState()
	state.BeginWithValue("typed text")
	state.Reset()

	if state.Value() != "" {
		t.Fatalf("expected empty value after Reset(), got %q", state.Value())
	}
}
