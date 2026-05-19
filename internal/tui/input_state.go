package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type InputState struct {
	commentInput      bool
	commentInstant    bool
	commentError      string
	mrCommentInput    bool
	mrCommentError    string
	issueCommentInput bool
	issueCommentError string
	editInput         bool
	editField         string
	editTitle         string
	replyInput        bool
	replyDraft        bool
	replyDiscussionID string
	textInput         textinput.Model
}

func NewInputState() InputState {
	ti := textinput.New()

	return InputState{textInput: ti}
}

func (s InputState) Active() bool {
	return s.commentInput || s.mrCommentInput || s.issueCommentInput || s.editInput || s.replyInput
}

func (s InputState) Value() string {
	return s.textInput.Value()
}

func (s *InputState) Begin() {
	s.textInput.SetValue("")
	s.textInput.Focus()
}

func (s *InputState) BeginWithValue(value string) {
	s.textInput.SetValue(value)
	s.textInput.CursorEnd()
	s.textInput.Focus()
}

func (s *InputState) Reset() {
	s.textInput.SetValue("")
	s.textInput.Blur()
}

func (s *InputState) Append(v string) {
	s.textInput.SetValue(s.textInput.Value() + v)
}

func (s *InputState) Backspace() {
	v := []rune(s.textInput.Value())
	if len(v) > 0 {
		s.textInput.SetValue(string(v[:len(v)-1]))
	}
}

func (s *InputState) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.textInput, cmd = s.textInput.Update(msg)

	return cmd
}
