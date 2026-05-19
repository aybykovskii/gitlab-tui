package tui

import (
	"strings"
	"testing"

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

	if model.mode != ModeEntityList {
		t.Fatalf("expected ModeEntityList after entering Issues section, got %v", model.mode)
	}

	if model.section != SectionIssues {
		t.Fatalf("expected Issues section, got %q", model.section)
	}

	if !strings.Contains(model.View(), "Issues [opened]") {
		t.Fatalf("expected issue list view, got %q", model.View())
	}
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
		if !strings.Contains(view, want) {
			t.Fatalf("expected issue list to contain %q, got %q", want, view)
		}
	}

	if strings.Contains(view, "extra") {
		t.Fatalf("expected labels to be truncated, got %q", view)
	}

	if strings.Contains(view, "💬 0") {
		t.Fatalf("expected zero comment count to be hidden, got %q", view)
	}
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
		if model.issueState != want {
			t.Fatalf("expected state %q, got %q", want, model.issueState)
		}

		if cmd == nil {
			t.Fatalf("expected state %q to reload issues", want)
		}
	}

	if !strings.Contains(model.renderEntityListPane(), "Issues [opened]") {
		t.Fatalf("expected title to show opened state, got %q", model.renderEntityListPane())
	}
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
		if !strings.Contains(view, want) {
			t.Fatalf("expected issue detail to contain %q, got %q", want, view)
		}
	}
}

func TestIssueDetailHidesUnsetWeight(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeDetail
	model.issueItems = []issue.Issue{{IID: 82, Title: "No Estimate", State: "closed"}}

	view := model.renderRight()
	if strings.Contains(view, "Weight") {
		t.Fatalf("expected unset weight to be hidden, got %q", view)
	}

	if !strings.Contains(view, "🔴 closed") {
		t.Fatalf("expected closed state emoji, got %q", view)
	}
}

func TestIssueDetailTabsStayWithinSummaryAndDiscussions(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeDetail
	model.issueItems = []issue.Issue{{IID: 81, Title: "Issue Detail"}}

	for _, want := range []DetailTab{TabDiscussions, TabSummary, TabDiscussions} {
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})

		model = updated.(Model)
		if model.IssueDetailState.activeTab != want {
			t.Fatalf("expected tab %v, got %v", want, model.IssueDetailState.activeTab)
		}
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
	if !model.editInput {
		t.Fatal("expected e to open issue edit input")
	}

	model.BeginWithValue("New")
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected edit command")
	}

	updated, _ = model.Update(cmd())

	model = updated.(Model)
	if !edited || model.issueItems[0].Title != "New" {
		t.Fatalf("expected edit to update title, edited=%t issue=%+v", edited, model.issueItems[0])
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected assign command")
	}

	updated, _ = model.Update(cmd())

	model = updated.(Model)
	if !assigned || len(model.issueItems[0].Assignees) != 1 || model.issueItems[0].Assignees[0] != "me" {
		t.Fatalf("expected assign self to update assignees, assigned=%t issue=%+v", assigned, model.issueItems[0])
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})

	model = updated.(Model)
	if model.mode != ModeLabelSelect {
		t.Fatalf("expected label selector mode, got %v", model.mode)
	}

	model.mode = ModeDetail
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected open URL command")
	}

	cmd()

	if openedURL != "https://gitlab.example/issue/84" {
		t.Fatalf("expected issue URL opened, got %q", openedURL)
	}
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

	if cmd == nil {
		t.Fatal("expected close command")
	}

	updated, _ = model.Update(cmd())

	model = updated.(Model)
	if !closed || reopened || model.issueItems[0].State != "closed" {
		t.Fatalf("expected close to update state, closed=%t reopened=%t issue=%+v", closed, reopened, model.issueItems[0])
	}

	if !strings.Contains(model.renderRight(), "🔴 closed") {
		t.Fatalf("expected closed summary, got %q", model.renderRight())
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected reopen command")
	}

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
	if !strings.Contains(view, "Needs work") {
		t.Fatalf("expected issue discussion body, got %q", view)
	}

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

	if model.issueCommentInput {
		t.Fatal("expected issue comment input to close")
	}

	if cmd == nil {
		t.Fatal("expected issue comment command")
	}

	cmd()

	if !called {
		t.Fatal("expected PostIssueComment to be called")
	}
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
	if model.replyInput {
		t.Fatal("expected d to be ignored for issue discussions")
	}
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
