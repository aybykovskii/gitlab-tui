package tui

import (
	"strings"
	"testing"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func TestDiffViewStateViewRendersColoredDiffLines(t *testing.T) {
	t.Parallel()

	state := NewDiffViewState()
	state.diffFiles = []mr.ChangedFile{{Path: "main.go", Diff: []mr.DiffRow{
		{OldLine: 1, NewLine: 1, OldText: "same", NewText: "same"},
		{OldLine: 0, NewLine: 2, NewText: "added"},
		{OldLine: 3, NewLine: 0, OldText: "removed"},
	}}}

	view := state.View(LayoutState{Width: 100, Height: 20})
	for _, want := range []string{"\x1b[38;5;240m", "\x1b[38;5;2m", "\x1b[38;5;1m", "+ added", "- removed", "same"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected diff view to contain %q, got:\n%q", want, view)
		}
	}
}

func TestDiffViewStateViewShowsAndHidesThreadPanel(t *testing.T) {
	t.Parallel()

	state := NewDiffViewState()
	state.diffFiles = []mr.ChangedFile{{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "same", NewText: "same"}}}}
	state.diffDiscussions = []mr.Discussion{{ID: "d1", Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 1}, Notes: []mr.Note{{Author: "alice", Body: "Needs work"}}}}

	shown := state.View(LayoutState{Width: 100, Height: 20})
	if !strings.Contains(shown, "Discussion") || !strings.Contains(shown, "Needs work") {
		t.Fatalf("expected thread panel shown, got:\n%s", shown)
	}

	state.threadPanelVisible = false
	hidden := state.View(LayoutState{Width: 100, Height: 20})
	if strings.Contains(hidden, "Needs work") {
		t.Fatalf("expected thread panel hidden, got:\n%s", hidden)
	}
}

func TestDiffViewStateViewRendersMultiDiscussionCounter(t *testing.T) {
	t.Parallel()

	state := NewDiffViewState()
	state.diffFiles = []mr.ChangedFile{{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "same", NewText: "same"}}}}
	state.threadPanelCursor = 1
	state.diffDiscussions = []mr.Discussion{
		{ID: "d1", Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 1}, Notes: []mr.Note{{Author: "alice", Body: "First"}}},
		{ID: "d2", Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 1}, Notes: []mr.Note{{Author: "bob", Body: "Second"}}},
		{ID: "d3", Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 1}, Notes: []mr.Note{{Author: "carol", Body: "Third"}}},
	}

	view := state.View(LayoutState{Width: 100, Height: 20})
	if !strings.Contains(view, "[2/3") || !strings.Contains(view, "Second") {
		t.Fatalf("expected multi-discussion counter and active discussion, got:\n%s", view)
	}
}
