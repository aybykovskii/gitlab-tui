package tui

import (
	"testing"


	"github.com/stretchr/testify/assert"
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

		assert.True(t, state.Active())
	}
}

func TestInputStateValueIsEmptyAfterBegin(t *testing.T) {
	t.Parallel()

	state := NewInputState()
	state.BeginWithValue("some prior value")
	state.Begin()

	assert.Equal(t, "", state.Value())
}

func TestInputStateBeginWithValueSetsValue(t *testing.T) {
	t.Parallel()

	state := NewInputState()
	state.BeginWithValue("Draft: My feature")

	assert.Equal(t, "Draft: My feature", state.Value())
}

func TestInputStateUpdateRoutesCharacterKeysToTextInput(t *testing.T) {
	t.Parallel()

	state := NewInputState()
	state.Begin()

	state.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	state.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	assert.Equal(t, "hi", state.Value())
}

func TestInputStateResetClearsValue(t *testing.T) {
	t.Parallel()

	state := NewInputState()
	state.BeginWithValue("typed text")
	state.Reset()

	assert.Equal(t, "", state.Value())
}
