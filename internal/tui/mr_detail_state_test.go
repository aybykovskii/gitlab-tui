package tui

import (
	"strings"
	"testing"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func TestMRDetailStateViewRendersActiveTabs(t *testing.T) {
	t.Parallel()

	item := mr.MergeRequest{IID: 42, Title: "Extract component", Author: "alice", State: "opened"}
	cases := []struct {
		name string
		tab  DetailTab
		want string
	}{
		{name: "summary", tab: TabSummary, want: "[>Summary<] [Discussions] [Files] [Review]"},
		{name: "discussions", tab: TabDiscussions, want: "[Summary] [>Discussions<] [Files] [Review]"},
		{name: "files", tab: TabFiles, want: "[Summary] [Discussions] [>Files<] [Review]"},
		{name: "review", tab: TabReview, want: "[Summary] [Discussions] [Files] [>Review<]"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			state := NewMRDetailState()
			state.activeTab = tc.tab

			view := state.View(LayoutState{Width: 80, Height: 20}, item)
			if !strings.Contains(view, tc.want) {
				t.Fatalf("expected active tab %q in view:\n%s", tc.want, view)
			}
		})
	}
}

func TestMRDetailStateViewRendersViewportContent(t *testing.T) {
	t.Parallel()

	state := NewMRDetailState()
	item := mr.MergeRequest{IID: 42, Title: "Viewport MR", Author: "alice", Description: "Long body"}

	view := state.View(LayoutState{Width: 80, Height: 20}, item)
	if !strings.Contains(view, "Viewport MR") || !strings.Contains(view, "Long body") {
		t.Fatalf("expected viewport content in view, got:\n%s", view)
	}
}

func TestMRDetailStateViewShowsDraftMarkers(t *testing.T) {
	t.Parallel()

	state := NewMRDetailState()
	state.activeTab = TabReview
	state.drafts[42] = []mr.DraftComment{{Body: "Check naming", Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 7}}}
	item := mr.MergeRequest{IID: 42, Title: "Draft MR"}

	view := state.View(LayoutState{Width: 80, Height: 20}, item)
	for _, want := range []string{"[>Review (1)<]", "> main.go:7 Check naming"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected draft marker %q in view:\n%s", want, view)
		}
	}
}
