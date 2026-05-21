package tui

import (
	"strings"
	"testing"


	"github.com/stretchr/testify/assert"
	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func TestIssueDetailStateHandlesDiscussionsFinished(t *testing.T) {
	t.Parallel()

	discussions := []issue.Discussion{{ID: "d1", Notes: []mr.Note{{Author: "alice", Body: "good"}}}}
	s := NewIssueDetailState()
	s.Update(issueDiscussionsFinishedMsg{iid: 9, discussions: discussions})

	assert.Len(t, s.discussions[9], 1)
}

func TestIssueDetailStateViewRendersActiveTabs(t *testing.T) {
	t.Parallel()

	item := issue.Issue{IID: 82, Title: "Issue component", Author: "alice", Description: "Issue body"}
	cases := []struct {
		name string
		tab  DetailTab
		want string
	}{
		{name: "summary", tab: TabSummary, want: "[>Summary<] [Discussions]"},
		{name: "discussions", tab: TabDiscussions, want: "[Summary] [>Discussions<]"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			state := NewIssueDetailState()
			state.activeTab = tc.tab

			view := state.View(LayoutState{Width: 80, Height: 20}, item)
			assert.Contains(t, view, tc.want)
		})
	}
}

func TestIssueDetailStateViewRendersViewportContent(t *testing.T) {
	t.Parallel()

	state := NewIssueDetailState()
	item := issue.Issue{IID: 82, Title: "Viewport issue", Author: "alice", Description: "Issue body"}

	view := state.View(LayoutState{Width: 80, Height: 20}, item)
	if !strings.Contains(view, "Viewport issue") || !strings.Contains(view, "Issue body") {
		t.Fatalf("expected issue viewport content in view:\n%s", view)
	}
}

func TestIssueDetailStateViewRendersDiscussions(t *testing.T) {
	t.Parallel()

	state := NewIssueDetailState()
	state.activeTab = TabDiscussions
	state.discussions[82] = []issue.Discussion{{ID: "d1", Notes: []mr.Note{{Author: "alice", Body: "Needs detail"}}}}
	item := issue.Issue{IID: 82, Title: "Discuss issue"}

	view := state.View(LayoutState{Width: 80, Height: 20}, item)
	assert.Contains(t, view, "Needs detail")
}
