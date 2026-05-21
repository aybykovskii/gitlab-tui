package tui

import (
	"errors"
	"strings"
	"testing"


	"github.com/stretchr/testify/assert"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func TestMRDetailStateHandlesDiscussionsStarted(t *testing.T) {
	t.Parallel()

	s := NewMRDetailState()
	s.Update(discussionsStartedMsg{})

	assert.True(t, s.discussionsLoading)

	assert.Equal(t, "", s.discussionsError)
}

func TestMRDetailStateHandlesDiscussionsFinished(t *testing.T) {
	t.Parallel()

	discussions := []mr.Discussion{{ID: "d1", Notes: []mr.Note{{Author: "alice", Body: "looks good"}}}}
	s := NewMRDetailState()
	s.discussionsLoading = true
	s.Update(discussionsFinishedMsg{iid: 5, discussions: discussions})

	assert.False(t, s.discussionsLoading)

	assert.Len(t, s.discussions[5], 1)
}

func TestMRDetailStateHandlesDiscussionsFinishedError(t *testing.T) {
	t.Parallel()

	s := NewMRDetailState()
	s.discussionsLoading = true
	s.Update(discussionsFinishedMsg{iid: 5, err: errors.New("network error")})

	assert.False(t, s.discussionsLoading)

	assert.NotEqual(t, "", s.discussionsError)

	assert.Len(t, s.discussions[5], 0)
}

func TestMRDetailStateHandlesFilesStarted(t *testing.T) {
	t.Parallel()

	s := NewMRDetailState()
	s.Update(filesStartedMsg{})

	assert.True(t, s.filesLoading)
}

func TestMRDetailStateHandlesFilesFinished(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{{Path: "main.go"}}
	s := NewMRDetailState()
	s.filesLoading = true
	s.Update(filesFinishedMsg{iid: 7, files: files})

	assert.False(t, s.filesLoading)

	assert.Len(t, s.changedFiles[7], 1)
}

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
			assert.Contains(t, view, tc.want)
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
		assert.Contains(t, view, want)
	}
}
