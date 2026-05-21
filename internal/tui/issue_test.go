package tui

import (
	"strings"
	"testing"


	"github.com/stretchr/testify/assert"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func TestEnterOnIssuesSectionOpensIssueList(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project"})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.Equal(t, ModeEntityList, model.mode)

	assert.Equal(t, SectionIssues, model.section)

	assert.Contains(t, model.View(), "Issues [opened]")
}

func TestIssuesEntityListRendersTwoLineRows(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeEntityList
	model.issueItems = []issue.Issue{{
		IID:          80,
		Title:        "Issues Entity List",
		Author:       "Alice",
		Labels:       []string{"frontend", "tui", "extra"},
		CommentCount: 3,
	}, {
		IID:          81,
		Title:        "No comments",
		Author:       "Bob",
		Labels:       []string{"backend"},
		CommentCount: 0,
	}}

	view := model.renderEntityListPane()

	for _, want := range []string{"Issues [opened]", "#80 Issues Entity List", "Alice · [frontend] [tui] · 💬 3", "#81 No comments", "Bob · [backend]"} {
		assert.Contains(t, view, want)
	}

	assert.NotContains(t, view, "extra")

	assert.NotContains(t, view, "💬 0")
}

func TestIssueStateFilterCyclesAndUpdatesTitle(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{
		Path:       "group/project",
		Section:    SectionIssues,
		LoadIssues: func(state string, search string) ([]issue.Issue, error) { return nil, nil },
	})
	model.mode = ModeEntityList

	for _, want := range []string{"closed", "", "opened"} {
		updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

		model = updated.(Model)
		assert.Equal(t, want, model.issueState)

		assert.NotNil(t, cmd)
	}

	assert.Contains(t, model.renderEntityListPane(), "Issues [opened]")
}

func TestIssueDetailSummaryRendersMetadata(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeEntityList
	model.issueItems = []issue.Issue{{
		IID:          81,
		Title:        "Issue Detail",
		Author:       "Alice",
		Assignees:    []string{"Bob"},
		State:        "opened",
		Labels:       []string{"frontend", "tui"},
		DueDate:      "2026-05-20",
		Milestone:    "v1",
		Weight:       5,
		CommentCount: 7,
		Description:  "Detail description",
	}}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	view := model.renderRight()

	for _, want := range []string{"#81 Issue Detail", "[>Summary<] [Discussions]", "👤 Alice · assigned: Bob", "🟢 opened · 💬 7", "🏷️ [frontend] [tui]", "📅 Due: 2026-05-20 · 🏁 v1", "⚖️ Weight: 5", "Detail description"} {
		assert.Contains(t, view, want)
	}
}

func TestIssueDetailHidesUnsetWeight(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeDetail
	model.issueItems = []issue.Issue{{IID: 82, Title: "No Estimate", State: "closed"}}

	view := model.renderRight()
	assert.NotContains(t, view, "Weight")

	assert.Contains(t, view, "🔴 closed")
}

func TestIssueDetailTabsStayWithinSummaryAndDiscussions(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeDetail
	model.issueItems = []issue.Issue{{IID: 81, Title: "Issue Detail"}}

	for _, want := range []DetailTab{TabDiscussions, TabSummary, TabDiscussions} {
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})

		model = updated.(Model)
		assert.Equal(t, want, model.IssueDetailState.activeTab)
	}
}

func TestIssueDetailKeyBarUsesIssueKeys(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeDetail
	model.issueItems = []issue.Issue{{IID: 81, Title: "Issue Detail"}}

	view := model.renderKeyBar()
	if !strings.Contains(view, "Tab next tab") || strings.Contains(view, "approve") || strings.Contains(view, "merge") {
		t.Fatalf("expected issue detail local keys, got %q", view)
	}
}

func TestIssueEditOpenAssignAndLabelsActions(t *testing.T) {
	t.Parallel()

	edited := false
	assigned := false
	openedURL := ""
	model := NewModelWithProject(nil, ProjectOptions{
		Path:    "group/project",
		Section: SectionIssues,
		Issues:  []issue.Issue{{IID: 84, Title: "Old", Description: "Desc", State: "opened", WebURL: "https://gitlab.example/issue/84"}},
		EditIssue: func(iid int, title, description string) error {
			edited = iid == 84 && title == "New" && description == "Desc"
			return nil
		},
		AssignSelfIssue: func(iid int) error {
			assigned = iid == 84
			return nil
		},
		OpenURL: func(url string) error {
			openedURL = url
			return nil
		},
	})
	model.mode = ModeDetail

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})

	model = updated.(Model)
	assert.True(t, model.editInput)

	model.BeginWithValue("New")
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.NotNil(t, cmd)

	updated, _ = model.Update(cmd())

	model = updated.(Model)
	if !edited || model.issueItems[0].Title != "New" {
		t.Fatalf("expected edit to update title, edited=%t issue=%+v", edited, model.issueItems[0])
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	model = updated.(Model)

	assert.NotNil(t, cmd)

	updated, _ = model.Update(cmd())

	model = updated.(Model)
	if !assigned || len(model.issueItems[0].Assignees) != 1 || model.issueItems[0].Assignees[0] != "me" {
		t.Fatalf("expected assign self to update assignees, assigned=%t issue=%+v", assigned, model.issueItems[0])
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})

	model = updated.(Model)
	assert.Equal(t, ModeLabelSelect, model.mode)

	model.mode = ModeDetail
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	model = updated.(Model)

	assert.NotNil(t, cmd)

	cmd()

	assert.Equal(t, "https://gitlab.example/issue/84", openedURL)
}

func TestIssueCloseReopenActionUsesStateAndUpdatesModel(t *testing.T) {
	t.Parallel()

	closed := false
	reopened := false
	model := NewModelWithProject(nil, ProjectOptions{
		Path:    "group/project",
		Section: SectionIssues,
		Issues:  []issue.Issue{{IID: 83, Title: "Close me", State: "opened"}},
		CloseIssue: func(iid int) error {
			closed = iid == 83
			return nil
		},
		ReopenIssue: func(iid int) error {
			reopened = iid == 83
			return nil
		},
	})
	model.mode = ModeDetail

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	model = updated.(Model)

	assert.NotNil(t, cmd)

	updated, _ = model.Update(cmd())

	model = updated.(Model)
	if !closed || reopened || model.issueItems[0].State != "closed" {
		t.Fatalf("expected close to update state, closed=%t reopened=%t issue=%+v", closed, reopened, model.issueItems[0])
	}

	assert.Contains(t, model.renderRight(), "🔴 closed")

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	model = updated.(Model)

	assert.NotNil(t, cmd)

	updated, _ = model.Update(cmd())

	model = updated.(Model)
	if !reopened || model.issueItems[0].State != "opened" {
		t.Fatalf("expected reopen to update state, reopened=%t issue=%+v", reopened, model.issueItems[0])
	}
}

func TestIssueDiscussionsTabRendersCommentsAndReplyInput(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeDetail
	model.IssueDetailState.activeTab = TabDiscussions
	model.issueItems = []issue.Issue{{IID: 82, Title: "Issue Discussions"}}
	model.IssueDetailState.discussions = map[int][]issue.Discussion{82: {{ID: "d1", Notes: []mr.Note{{Author: "Alice", Body: "Needs work"}}}}}

	view := model.renderRight()
	assert.Contains(t, view, "Needs work")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})

	model = updated.(Model)
	if !model.replyInput || model.replyDiscussionID != "d1" {
		t.Fatalf("expected reply input for d1, got input=%t discussion=%q", model.replyInput, model.replyDiscussionID)
	}
}

func TestIssueGeneralCommentInputCallsPostIssueComment(t *testing.T) {
	t.Parallel()

	called := false
	model := NewModelWithProject(nil, ProjectOptions{
		Path:    "group/project",
		Section: SectionIssues,
		Issues:  []issue.Issue{{IID: 82, Title: "Issue Discussions"}},
		PostIssueComment: func(iid int, body string) error {
			called = true
			if iid != 82 || body != "General comment" {
				t.Fatalf("unexpected comment iid/body: %d %q", iid, body)
			}
			return nil
		},
	})
	model.mode = ModeDetail

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("General comment")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.False(t, model.issueCommentInput)

	assert.NotNil(t, cmd)

	cmd()

	assert.True(t, called)
}

func TestIssueDiscussionsIgnoreResolveAndDraftKeys(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeDetail
	model.IssueDetailState.activeTab = TabDiscussions
	model.issueItems = []issue.Issue{{IID: 82, Title: "Issue Discussions"}}
	model.IssueDetailState.discussions = map[int][]issue.Discussion{82: {{ID: "d1", Notes: []mr.Note{{Author: "Alice", Body: "Needs work"}}}}}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	model = updated.(Model)

	if cmd != nil || model.IssueDetailState.discussions[82][0].Resolved {
		t.Fatalf("expected x to be ignored, cmd=%v discussion=%+v", cmd, model.IssueDetailState.discussions[82][0])
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})

	model = updated.(Model)
	assert.False(t, model.replyInput)
}

func TestIssueFilterNarrowsByTitleAndAuthor(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeEntityList
	model.issueItems = []issue.Issue{{Title: "Render issues", Author: "Alice"}, {Title: "Other", Author: "Bob"}}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("alice")})
	model = updated.(Model)

	items := model.filteredIssues()
	if len(items) != 1 || items[0].Author != "Alice" {
		t.Fatalf("unexpected filtered issues: %+v", items)
	}
}
