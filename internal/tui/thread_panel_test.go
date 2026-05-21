package tui

import (
	"strings"
	"testing"


	"github.com/stretchr/testify/assert"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

// makeFileDiffModel builds a ModeFileDiff model for the first item in items.
func makeFileDiffModel(items []mr.MergeRequest, file mr.ChangedFile, discussions []mr.Discussion, drafts []mr.DraftComment) Model {
	model := NewModel(items)
	model.mode = ModeFileDiff
	model.threadPanelVisible = true
	model.selectedFile = 0
	model.diffCursor = 0
	model.width = 100
	model.height = 30
	iid := items[0].IID

	model.changedFiles[iid] = []mr.ChangedFile{file}
	if discussions != nil {
		model.discussions[iid] = discussions
	}

	if drafts != nil {
		model.drafts[iid] = drafts
	}

	return model
}

var testFile = mr.ChangedFile{
	Path: "main.go",
	Diff: []mr.DiffRow{
		{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"},
		{OldLine: 2, NewLine: 2, OldText: "b", NewText: "b"},
		{OldLine: 3, NewLine: 3, OldText: "c", NewText: "c"},
	},
}

var testDiscussion = mr.Discussion{
	ID:       "d1",
	Resolved: false,
	Notes:    []mr.Note{{Author: "alice", Body: "Nice change!"}},
	Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 3},
}

// Cycle 1 — tracer bullet: Thread Panel appears when cursor is on a Discussion line.
func TestThreadPanelShowsDiscussionAtCursorLine(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, []mr.Discussion{testDiscussion}, nil)
	model.diffCursor = 2 // row index 2 → NewLine 3

	view := model.renderFileDiffPane()

	assert.Contains(t, view, "alice")

	assert.Contains(t, view, "Nice change!")
}

// Cycle 2 — no panel when cursor is on a line without a thread.
func TestThreadPanelAbsentOnNonCommentedLine(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, []mr.Discussion{testDiscussion}, nil)
	model.diffCursor = 0 // row 0 → NewLine 1, no discussion

	view := model.renderFileDiffPane()

	assert.NotContains(t, view, "alice")
}

// Cycle 3 — `t` hides Thread Panel; gutter marker ○ remains visible.
func TestToggleTHidesThreadPanelButKeepsGutterMarker(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, []mr.Discussion{testDiscussion}, nil)
	model.diffCursor = 2

	updated, _ := model.Update(keyMsg("t"))
	model = updated.(Model)

	view := model.renderFileDiffPane()

	assert.NotContains(t, view, "alice")

	assert.Contains(t, view, "○")
}

// Cycle 4 — `t` twice restores the Thread Panel.
func TestToggleTTwiceRestoresThreadPanel(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, []mr.Discussion{testDiscussion}, nil)
	model.diffCursor = 2

	updated, _ := model.Update(keyMsg("t"))
	updatedAgain, _ := updated.(Model).Update(keyMsg("t"))
	model = updatedAgain.(Model)

	view := model.renderFileDiffPane()

	assert.Contains(t, view, "alice")
}

// Cycle 5 — Resolved discussion renders with resolved indicator.
func TestResolvedDiscussionShowsResolvedIndicator(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	resolved := mr.Discussion{
		ID:       "d2",
		Resolved: true,
		Notes:    []mr.Note{{Author: "bob", Body: "LGTM"}},
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 3},
	}
	model := makeFileDiffModel(items, testFile, []mr.Discussion{resolved}, nil)
	model.diffCursor = 2

	view := model.renderFileDiffPane()

	if !strings.Contains(view, "✅") && !strings.Contains(view, "resolved") {
		t.Fatalf("expected resolved indicator in Thread Panel, got: %q", view)
	}
}

// Cycle 6 — Draft comment shows draft indicator in Thread Panel.
func TestDraftCommentShowsDraftIndicatorInThreadPanel(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	draft := mr.DraftComment{
		LocalID:  "local-1",
		Body:     "WIP note",
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 3},
	}
	model := makeFileDiffModel(items, testFile, nil, []mr.DraftComment{draft})
	model.diffCursor = 2

	view := model.renderFileDiffPane()

	if !strings.Contains(view, "📝") && !strings.Contains(view, "Draft") {
		t.Fatalf("expected draft indicator in Thread Panel, got: %q", view)
	}

	assert.Contains(t, view, "WIP note")
}

// Cycle 7 — `r` key opens reply input when cursor is on a Discussion line.
func TestFileDiffRKeyOpensReplyInputAtCursorDiscussion(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, []mr.Discussion{testDiscussion}, nil)
	model.diffCursor = 2

	updated, _ := model.Update(keyMsg("r"))
	model = updated.(Model)

	assert.True(t, model.replyInput)

	assert.Equal(t, "d1", model.replyDiscussionID)
}

// Cycle 8 — `x` key toggles resolved on the discussion at cursor.
func TestFileDiffXKeyTogglesResolveAtCursorRow(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, []mr.Discussion{testDiscussion}, nil)
	model.resolveDiscussion = nil // no async fn — model resolves locally
	model.diffCursor = 2

	updated, _ := model.Update(keyMsg("x"))
	model = updated.(Model)

	if len(model.discussions[42]) == 0 || !model.discussions[42][0].Resolved {
		t.Fatal("expected discussion to be marked resolved after 'x'")
	}
}
