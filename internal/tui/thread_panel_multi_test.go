package tui

import (
	"strings"
	"testing"


	"github.com/stretchr/testify/assert"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

var threeDiscussions = []mr.Discussion{
	{
		ID:       "d1",
		Notes:    []mr.Note{{Author: "alice", Body: "First note"}},
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 3},
	},
	{
		ID:       "d2",
		Notes:    []mr.Note{{Author: "bob", Body: "Second note"}},
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 3},
	},
	{
		ID:       "d3",
		Notes:    []mr.Note{{Author: "carol", Body: "Third note"}},
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 3},
	},
}

// Cycle 1 — `]` advances threadPanelCursor on a line with multiple discussions.
func TestThreadPanelCursorAdvancesWithCloseBracket(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2

	updated, _ := model.Update(keyMsg("]"))
	model = updated.(Model)

	assert.Equal(t, 1, model.threadPanelCursor)

	view := model.renderFileDiffPane()
	if !strings.Contains(view, "bob") || !strings.Contains(view, "Second note") {
		t.Fatalf("expected Thread Panel to show d2 (bob/Second note), got:\n%s", view)
	}
}

// Cycle 2 — `[` decreases threadPanelCursor.
func TestThreadPanelCursorDecreasesWithOpenBracket(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2
	model.threadPanelCursor = 2

	updated, _ := model.Update(keyMsg("["))
	model = updated.(Model)

	assert.Equal(t, 1, model.threadPanelCursor)
}

// Cycle 3 — `[` clamps at 0, does not go negative.
func TestThreadPanelCursorClampsAtMin(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2
	model.threadPanelCursor = 0

	updated, _ := model.Update(keyMsg("["))
	model = updated.(Model)

	assert.Equal(t, 0, model.threadPanelCursor)
}

// Cycle 4 — `]` clamps at max (len-1).
func TestThreadPanelCursorClampsAtMax(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2
	model.threadPanelCursor = 2

	updated, _ := model.Update(keyMsg("]"))
	model = updated.(Model)

	assert.Equal(t, 2, model.threadPanelCursor)
}

// Cycle 5 — threadPanelCursor resets to 0 when diff cursor moves to another row.
func TestThreadPanelCursorResetsOnDiffCursorMove(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2
	model.threadPanelCursor = 2

	updated, _ := model.Update(keyMsg("k")) // move up
	model = updated.(Model)

	assert.Equal(t, 0, model.threadPanelCursor)
}

// Cycle 6 — Thread Panel header shows [N/M  [/]: switch] when there are >1 discussions.
func TestThreadPanelHeaderShowsCounterForMultipleDiscussions(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2

	view := model.renderFileDiffPane()

	assert.Contains(t, view, "[1/3")
}

// Cycle 7 — Single thread: no counter shown in header.
func TestThreadPanelHeaderHidesCounterForSingleDiscussion(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, []mr.Discussion{testDiscussion}, nil)
	model.diffCursor = 2

	view := model.renderFileDiffPane()

	assert.NotContains(t, view, "[1/1")
}

// Cycle 8 — `r` uses threadPanelCursor to select the active discussion.
func TestFileDiffRKeyUsesThreadPanelCursorDiscussion(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2
	model.threadPanelCursor = 1 // second discussion = d2

	updated, _ := model.Update(keyMsg("r"))
	model = updated.(Model)

	assert.True(t, model.replyInput)

	assert.Equal(t, "d2", model.replyDiscussionID)
}
