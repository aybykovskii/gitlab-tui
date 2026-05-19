package tui

import (
	"strings"
	"testing"

	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func TestEntityListStateViewRendersMergeRequests(t *testing.T) {
	t.Parallel()

	state := NewEntityListState([]mr.MergeRequest{{IID: 1, Title: "First MR", Author: "alice", SourceBranch: "feat", TargetBranch: "main"}}, nil)
	state.projectPath = "group/project"

	view := state.View(LayoutState{Width: 80, Height: 20}, EntityListViewData{Section: SectionMergeRequests})
	for _, want := range []string{"Project: group/project", "Merge Requests", "!1 First MR", "alice feat → main"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected MR entity list to contain %q, got:\n%s", want, view)
		}
	}
}

func TestEntityListStateViewRendersIssues(t *testing.T) {
	t.Parallel()

	state := NewEntityListState(nil, []issue.Issue{{IID: 2, Title: "First issue", Author: "bob", Labels: []string{"bug"}, CommentCount: 3}})
	state.projectPath = "group/project"

	view := state.View(LayoutState{Width: 80, Height: 20}, EntityListViewData{Section: SectionIssues, IssueStateLabel: "opened"})
	for _, want := range []string{"Project: group/project", "Issues [opened]", "#2 First issue", "bob · [bug] · 💬 3"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected issue entity list to contain %q, got:\n%s", want, view)
		}
	}
}

func TestEntityListStateViewUsesEncapsulatedFilterAndSelection(t *testing.T) {
	t.Parallel()

	state := NewEntityListState([]mr.MergeRequest{{IID: 1, Title: "Hidden MR"}, {IID: 2, Title: "Visible MR"}}, nil)
	state.query = "visible"

	view := state.View(LayoutState{Width: 80, Height: 20}, EntityListViewData{Section: SectionMergeRequests})
	if !strings.Contains(view, ">") || !strings.Contains(view, "!2 Visible MR") || strings.Contains(view, "Hidden MR") {
		t.Fatalf("expected filtered selected MR list, got:\n%s", view)
	}
}
