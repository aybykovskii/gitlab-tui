package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/issue"
)

func TestEnterOnIssuesSectionOpensIssueList(t *testing.T) {
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
	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeDetail
	model.issueItems = []issue.Issue{{IID: 81, Title: "Issue Detail"}}

	for _, want := range []DetailTab{TabDiscussions, TabSummary, TabDiscussions} {
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
		model = updated.(Model)
		if model.activeTab != want {
			t.Fatalf("expected tab %v, got %v", want, model.activeTab)
		}
	}
}

func TestIssueDetailKeyBarUsesIssueKeys(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionIssues})
	model.mode = ModeDetail
	model.issueItems = []issue.Issue{{IID: 81, Title: "Issue Detail"}}

	view := model.renderKeyBar()
	if !strings.Contains(view, "Tab next tab") || strings.Contains(view, "approve") || strings.Contains(view, "merge") {
		t.Fatalf("expected issue detail local keys, got %q", view)
	}
}

func TestIssueFilterNarrowsByTitleAndAuthor(t *testing.T) {
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
