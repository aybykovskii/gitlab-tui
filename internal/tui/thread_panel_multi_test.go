package tui

import (
	"strings"
	"testing"

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
	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2

	updated, _ := model.Update(keyMsg("]"))
	model = updated.(Model)

	if model.threadPanelCursor != 1 {
		t.Fatalf("expected threadPanelCursor=1 after ']', got %d", model.threadPanelCursor)
	}

	view := model.renderFileDiffPane()
	if !strings.Contains(view, "bob") || !strings.Contains(view, "Second note") {
		t.Fatalf("expected Thread Panel to show d2 (bob/Second note), got:\n%s", view)
	}
}

// Cycle 2 — `[` decreases threadPanelCursor.
func TestThreadPanelCursorDecreasesWithOpenBracket(t *testing.T) {
	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2
	model.threadPanelCursor = 2

	updated, _ := model.Update(keyMsg("["))
	model = updated.(Model)

	if model.threadPanelCursor != 1 {
		t.Fatalf("expected threadPanelCursor=1 after '[', got %d", model.threadPanelCursor)
	}
}

// Cycle 3 — `[` clamps at 0, does not go negative.
func TestThreadPanelCursorClampsAtMin(t *testing.T) {
	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2
	model.threadPanelCursor = 0

	updated, _ := model.Update(keyMsg("["))
	model = updated.(Model)

	if model.threadPanelCursor != 0 {
		t.Fatalf("expected threadPanelCursor to stay at 0, got %d", model.threadPanelCursor)
	}
}

// Cycle 4 — `]` clamps at max (len-1).
func TestThreadPanelCursorClampsAtMax(t *testing.T) {
	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2
	model.threadPanelCursor = 2

	updated, _ := model.Update(keyMsg("]"))
	model = updated.(Model)

	if model.threadPanelCursor != 2 {
		t.Fatalf("expected threadPanelCursor to stay at 2, got %d", model.threadPanelCursor)
	}
}

// Cycle 5 — threadPanelCursor resets to 0 when diff cursor moves to another row.
func TestThreadPanelCursorResetsOnDiffCursorMove(t *testing.T) {
	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2
	model.threadPanelCursor = 2

	updated, _ := model.Update(keyMsg("k")) // move up
	model = updated.(Model)

	if model.threadPanelCursor != 0 {
		t.Fatalf("expected threadPanelCursor reset to 0 after moving cursor, got %d", model.threadPanelCursor)
	}
}

// Cycle 6 — Thread Panel header shows [N/M  [/]: switch] when there are >1 discussions.
func TestThreadPanelHeaderShowsCounterForMultipleDiscussions(t *testing.T) {
	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2

	view := model.renderFileDiffPane()

	if !strings.Contains(view, "[1/3") {
		t.Fatalf("expected counter [1/3 in Thread Panel header, got:\n%s", view)
	}
}

// Cycle 7 — Single thread: no counter shown in header.
func TestThreadPanelHeaderHidesCounterForSingleDiscussion(t *testing.T) {
	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, []mr.Discussion{testDiscussion}, nil)
	model.diffCursor = 2

	view := model.renderFileDiffPane()

	if strings.Contains(view, "[1/1") {
		t.Fatalf("expected no counter for single discussion, got:\n%s", view)
	}
}

// Cycle 8 — `r` uses threadPanelCursor to select the active discussion.
func TestFileDiffRKeyUsesThreadPanelCursorDiscussion(t *testing.T) {
	items := []mr.MergeRequest{{IID: 42, Title: "Test MR"}}
	model := makeFileDiffModel(items, testFile, threeDiscussions, nil)
	model.diffCursor = 2
	model.threadPanelCursor = 1 // second discussion = d2

	updated, _ := model.Update(keyMsg("r"))
	model = updated.(Model)

	if !model.replyInput {
		t.Fatal("expected replyInput true after 'r'")
	}

	if model.replyDiscussionID != "d2" {
		t.Fatalf("expected replyDiscussionID d2, got %q", model.replyDiscussionID)
	}
}
