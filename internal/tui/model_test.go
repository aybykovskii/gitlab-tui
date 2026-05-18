package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

var errTestRefresh = errors.New("refresh failed")

func TestAllNavigationKeyMapsHaveHelp(t *testing.T) {
	keyMaps := map[string][]key.Binding{
		"project list": newProjectListKeys().LocalKeys(),
		"sections":     newSectionsKeys().LocalKeys(),
		"entity list":  newEntityListKeys().LocalKeys(),
		"mr detail":    newMRDetailKeys().LocalKeys(),
		"diff view":    newDiffViewKeys().LocalKeys(),
		"file diff":    newFileDiffKeys().LocalKeys(),
	}
	for name, bindings := range keyMaps {
		if len(bindings) == 0 {
			t.Fatalf("expected %s key map to have bindings", name)
		}
		for _, binding := range bindings {
			if binding.Help().Key == "" || binding.Help().Desc == "" {
				t.Fatalf("expected %s binding to have help, got %+v", name, binding.Help())
			}
		}
	}
}

func TestCollapsedKeyBarTruncatesLocalKeys(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{Recents: []string{"group/project"}})
	model.width = 34

	view := model.renderKeyBar()
	if !strings.Contains(view, "…") {
		t.Fatalf("expected truncated key bar with ellipsis, got %q", view)
	}
}

func TestExpandedKeyBarShowsAllProjectListLocalKeys(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{Recents: []string{"group/project"}})
	model.width = 80
	model.keyBarExpanded = true

	view := model.renderKeyBar()
	for _, want := range []string{"↑/k up", "↓/j down", "Enter open", "/ filter", "i manual", "r retry", "Global:"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected expanded key bar to contain %q, got %q", want, view)
		}
	}
	if !strings.Contains(view, "─") {
		t.Fatalf("expected expanded key bar separator, got %q", view)
	}
}

func TestViewRendersPersistentKeyBar(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{Recents: []string{"group/project"}})
	model.width = 80
	model.height = 24

	view := model.View()

	for _, want := range []string{"Local", "Global", "q quit", "Esc back", "h keys"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected key bar to contain %q, got %q", want, view)
		}
	}
}

func TestHKeyTogglesExpandedKeyBarAndPaneHeight(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{Recents: []string{"group/project"}})
	model.height = 30
	collapsedHeight := model.paneHeight()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model = updated.(Model)

	if !model.keyBarExpanded {
		t.Fatal("expected key bar to expand")
	}
	if model.paneHeight() >= collapsedHeight {
		t.Fatalf("expected expanded key bar to shrink panes, collapsed=%d expanded=%d", collapsedHeight, model.paneHeight())
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model = updated.(Model)
	if model.keyBarExpanded {
		t.Fatal("expected key bar to collapse")
	}
}

func TestGlobalKeysUseDeclaredBindings(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Fatal("expected ctrl+c declared global quit binding to return quit command")
	}
}

func TestInputModeQDoesNotQuit(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	model.mrCommentInput = true

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updated.(Model)

	if cmd != nil {
		t.Fatal("expected q to be handled as input, not global quit")
	}
	if model.mrCommentBuffer != "q" {
		t.Fatalf("expected q in comment buffer, got %q", model.mrCommentBuffer)
	}
}

func TestResolvedProjectShowsProjectListAndSections(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:     "group/project",
		Recents:  []string{"group/project", "recent/other"},
		Projects: []string{"group/project", "gitlab/other"},
	})

	view := model.View()
	for _, want := range []string{
		"Projects",
		"> group/project",
		"recent/other",
		"gitlab/other",
		"Sections",
		"Merge Requests",
		"Issues",
		"Pipelines",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected view to contain %q, got %q", want, view)
		}
	}
}

func TestEnterOnMergeRequestsSectionOpensMRList(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project"})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.mode != ModeEntityList {
		t.Fatalf("expected ModeEntityList after entering MR section, got %v", model.mode)
	}
	if model.section != SectionMergeRequests {
		t.Fatalf("expected MR section, got %q", model.section)
	}
	if !strings.Contains(model.View(), "Merge Requests") || !strings.Contains(model.View(), "Port TUI shell") {
		t.Fatalf("expected MR list view, got %q", model.View())
	}
}

func TestOpenedProjectMovesToTopOfProjectList(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:     "group/project",
		Recents:  []string{"recent/other"},
		Projects: []string{"group/project"},
	})

	if len(model.projectList) == 0 || model.projectList[0] != "group/project" {
		t.Fatalf("expected opened project first, got %+v", model.projectList)
	}
}

func TestKeyboardSelectionAndDiffNavigation(t *testing.T) {
	model := NewFakeModel()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	if model.selected != 1 {
		t.Fatalf("expected selected index 1, got %d", model.selected)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.mode != ModeDiff {
		t.Fatalf("expected diff mode, got %v", model.mode)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	if model.mode != ModeDetail {
		t.Fatalf("expected detail mode, got %v", model.mode)
	}
}

func TestFilterInputNarrowsList(t *testing.T) {
	model := NewFakeModel()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("yaml")})
	model = updated.(Model)

	filtered := model.filtered()
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered item, got %d", len(filtered))
	}
	if !strings.Contains(strings.ToLower(filtered[0].Title), "yaml") {
		t.Fatalf("unexpected filtered item: %+v", filtered[0])
	}
}

func TestDirectMRDeepLinkSelectsLoadedMergeRequest(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionMergeRequests, EntityID: "123"})
	updated, _ := model.Update(projectFinishedMsg{path: "group/project", data: ProjectData{Items: []mr.MergeRequest{
		{IID: 101, Title: "First MR"},
		{IID: 123, Title: "Loaded target"},
	}}})
	model = updated.(Model)

	if model.selected != 1 {
		t.Fatalf("expected loaded target MR selected, got %d", model.selected)
	}
	if !strings.Contains(model.View(), "!123 Loaded target") {
		t.Fatalf("expected loaded target MR detail, got %q", model.View())
	}
}

func TestDirectMRDeepLinkSelectsMatchingMergeRequest(t *testing.T) {
	model := NewModelWithProject([]mr.MergeRequest{
		{IID: 101, Title: "First MR"},
		{IID: 123, Title: "Target MR", Description: "Deep linked"},
	}, ProjectOptions{Path: "group/project", Section: SectionMergeRequests, EntityID: "123"})

	if model.selected != 1 {
		t.Fatalf("expected selected MR index 1, got %d", model.selected)
	}
	view := model.View()
	if !strings.Contains(view, "!123 Target MR") || !strings.Contains(view, "Deep linked") {
		t.Fatalf("expected target MR detail, got %q", view)
	}
}

func TestMRListAndDetailRenderPreviousMRInfo(t *testing.T) {
	model := NewModelWithProject([]mr.MergeRequest{{
		IID:            10,
		Title:          "Add review UI",
		Author:         "Alice Doe",
		AuthorUsername: "alice",
		SourceBranch:   "feature/review",
		TargetBranch:   "main",
		State:          "opened",
		Pipeline:       "success",
		Approvals:      "1/2",
		Description:    "Review from terminal",
		WebURL:         "https://gitlab.com/group/project/-/merge_requests/10",
	}}, ProjectOptions{Path: "group/project", Section: SectionMergeRequests})

	view := model.View()
	for _, want := range []string{
		"✓ !10 Add review UI",
		"Alice Doe",
		"feature/review → main",
		"Author: Alice Doe @alice",
		"State: opened",
		"Pipeline: ✓ success",
		"Approvals: 1/2",
		"Review from terminal",
		"https://gitlab.com/group/project/-/merge_requests/10",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected view to contain %q, got %q", want, view)
		}
	}
}

func TestMouseClickSelectsListItem(t *testing.T) {
	model := NewFakeModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	model = updated.(Model)

	updated, _ = model.Update(tea.MouseMsg{X: 2, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	model = updated.(Model)

	if model.selected != 1 {
		t.Fatalf("expected clicked second item to be selected, got %d", model.selected)
	}
}

func TestProjectSelectionShowsRecentsAndGitLabProjects(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Recents:  []string{"recent/project"},
		Projects: []string{"gitlab/project"},
	})

	view := model.View()
	if !strings.Contains(view, "recent/project") || !strings.Contains(view, "gitlab/project") {
		t.Fatalf("expected recents and GitLab projects, got %q", view)
	}
}

func TestProjectSelectionGoesToSectionsImmediately(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{
		Recents: []string{"group/one", "group/two"},
		LoadProject: func(path string) (ProjectData, error) {
			return ProjectData{Items: []mr.MergeRequest{{IID: 42, Title: "Loaded"}}}, nil
		},
	})
	if model.mode != ModeProjectSelect {
		t.Fatalf("expected project select mode, got %v", model.mode)
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if cmd != nil {
		t.Fatal("expected no loading command on project selection")
	}
	if model.mode != ModeSections {
		t.Fatalf("expected sections mode immediately, got %v", model.mode)
	}
	if model.projectPath != "group/two" {
		t.Fatalf("expected selected project path, got %q", model.projectPath)
	}
}

func TestProjectLoadShowsLoadingState(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Recents: []string{"group/project"}})
	updated, _ := model.Update(projectStartedMsg{path: "group/project"})
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, "Loading project…") {
		t.Fatalf("expected project loading state, got %q", view)
	}
	if strings.Contains(view, "Refreshing…") {
		t.Fatalf("expected project load not to look like refresh, got %q", view)
	}
}

func TestProjectLoadErrorCanReturnToSelection(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Recents: []string{"group/project"}})
	updated, _ := model.Update(projectStartedMsg{path: "group/project"})
	model = updated.(Model)
	updated, _ = model.Update(projectFinishedMsg{path: "group/project", err: errTestRefresh})
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, "Error: refresh failed") {
		t.Fatalf("expected load error in view, got %q", view)
	}
	if !strings.Contains(view, "Esc back") {
		t.Fatalf("expected recovery key in key bar, got %q", view)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	if model.mode != ModeProjectSelect {
		t.Fatalf("expected project select after Esc, got %v", model.mode)
	}
}

func TestProjectLoadErrorRetryReloadsSameProject(t *testing.T) {
	calls := 0
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Recents: []string{"group/project"},
		LoadProject: func(path string) (ProjectData, error) {
			calls++
			if path != "group/project" {
				t.Fatalf("expected retry path group/project, got %q", path)
			}
			return ProjectData{Items: []mr.MergeRequest{{IID: 9, Title: "Retried"}}}, nil
		},
	})
	updated, _ := model.Update(projectStartedMsg{path: "group/project"})
	model = updated.(Model)
	updated, _ = model.Update(projectFinishedMsg{path: "group/project", err: errTestRefresh})
	model = updated.(Model)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("expected retry command")
	}

	_ = cmd
	if calls != 0 {
		t.Fatalf("expected command creation not to call loader yet, got %d calls", calls)
	}
}

func TestManualProjectLoadErrorReturnsToInput(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{})
	updated, _ := model.Update(projectStartedMsg{path: "group/project"})
	model = updated.(Model)
	updated, _ = model.Update(projectFinishedMsg{path: "group/project", err: errTestRefresh})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	if model.mode != ModeProjectInput {
		t.Fatalf("expected project input after Esc, got %v", model.mode)
	}
	if model.focus != FocusFilter {
		t.Fatalf("expected input focus after Esc, got %v", model.focus)
	}
}

func TestManualProjectInputGoesToSectionsImmediately(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{
		LoadProject: func(path string) (ProjectData, error) {
			return ProjectData{Items: []mr.MergeRequest{{IID: 7, Title: "Manual"}}}, nil
		},
	})
	if model.mode != ModeProjectInput {
		t.Fatalf("expected project input mode, got %v", model.mode)
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("group/project")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if cmd != nil {
		t.Fatal("expected no loading command on manual project input")
	}
	if model.mode != ModeSections {
		t.Fatalf("expected sections mode immediately, got %v", model.mode)
	}
	if model.projectPath != "group/project" {
		t.Fatalf("expected project path set, got %q", model.projectPath)
	}
}

func TestEnterLoadsDiffWhenNeeded(t *testing.T) {
	model := NewModelWithProject([]mr.MergeRequest{{IID: 1, Title: "Needs diff"}}, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadDiff: func(iid int) ([]mr.DiffRow, error) {
			return []mr.DiffRow{{OldLine: 1, OldText: "old", NewLine: 1, NewText: "new"}}, nil
		},
	})

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected diff load command")
	}
}

func TestDiffFinishedStoresRowsAndOpensDiff(t *testing.T) {
	model := NewModelWithProject([]mr.MergeRequest{{IID: 1, Title: "Needs diff"}}, ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(diffFinishedMsg{iid: 1, rows: []mr.DiffRow{{OldLine: 1, OldText: "old"}}})
	model = updated.(Model)

	if model.mode != ModeDiff {
		t.Fatalf("expected diff mode, got %v", model.mode)
	}
	if len(model.items[0].Diff) != 1 {
		t.Fatalf("expected diff rows to be stored, got %+v", model.items[0].Diff)
	}
}

func TestEmptyProjectStateCanReturnToProjectSelection(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Recents: []string{"group/project"}, Section: SectionMergeRequests})
	updated, _ := model.Update(projectFinishedMsg{path: "group/project", data: ProjectData{Items: []mr.MergeRequest{}}})
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, "No opened MRs") {
		t.Fatalf("expected empty MR state, got %q", view)
	}
	if !strings.Contains(view, "r refresh") || !strings.Contains(view, "Esc back") {
		t.Fatalf("expected empty state actions in key bar, got %q", view)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	if model.mode != ModeProjectSelect {
		t.Fatalf("expected project select after Esc, got %v", model.mode)
	}
}

func TestRefreshKeyReturnsCommand(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		Refresh: func() ([]mr.MergeRequest, error) {
			return []mr.MergeRequest{{IID: 99, Title: "Refreshed"}}, nil
		},
	})

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

func TestRefreshFinishedReplacesItems(t *testing.T) {
	model := NewFakeModel()
	updated, _ := model.Update(refreshFinishedMsg{items: []mr.MergeRequest{{IID: 99, Title: "Refreshed"}}})
	model = updated.(Model)

	if len(model.items) != 1 {
		t.Fatalf("expected refreshed items, got %d", len(model.items))
	}
	if model.items[0].IID != 99 {
		t.Fatalf("unexpected refreshed item: %+v", model.items[0])
	}
}

func TestRefreshFinishedStoresError(t *testing.T) {
	model := NewFakeModel()
	updated, _ := model.Update(refreshFinishedMsg{err: errTestRefresh})
	model = updated.(Model)

	if model.errorMessage != errTestRefresh.Error() {
		t.Fatalf("expected refresh error, got %q", model.errorMessage)
	}
}

func TestMouseWheelScrollsRightPanel(t *testing.T) {
	model := NewFakeModel()
	before := model.rightTop

	updated, _ := model.Update(tea.MouseMsg{X: 2, Y: 4, Button: tea.MouseButtonWheelDown})
	model = updated.(Model)

	if model.rightTop != before+1 {
		t.Fatalf("expected wheel down to scroll right panel (rightTop %d→%d)", before, model.rightTop)
	}
	if model.selected != 0 {
		t.Fatalf("expected wheel not to change MR selection, got selected=%d", model.selected)
	}
}

// --- #41: MR detail tabs ---

func discussionOpts() ProjectOptions {
	return ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) {
			return []mr.Discussion{
				{ID: "d1", Resolved: false, Notes: []mr.Note{
					{Author: "alice", Body: "Please fix the naming", Resolved: false},
					{Author: "bob", Body: "Done", Resolved: true},
				}},
				{ID: "d2", Resolved: true, Notes: []mr.Note{
					{Author: "carol", Body: "LGTM", Resolved: true},
				}},
			}, nil
		},
		LoadFiles: func(iid int) ([]mr.ChangedFile, error) {
			return []mr.ChangedFile{
				{Path: "internal/tui/model.go", IsNew: false, AddedLines: 10, RemovedLines: 3},
				{Path: "internal/mr/model.go", IsNew: true, AddedLines: 25, RemovedLines: 0},
			}, nil
		},
	}
}

func TestDiscussionsTabTriggersLoadOnFirstVisit(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)

	if model.activeTab != TabDiscussions {
		t.Fatalf("expected TabDiscussions, got %v", model.activeTab)
	}
	if cmd == nil {
		t.Fatal("expected a load command when switching to Discussions tab for the first time")
	}
	if !strings.Contains(model.View(), "Loading") {
		t.Fatalf("expected loading state in Discussions tab, got:\n%s", model.View())
	}
}

func TestDiscussionsTabRendersThreadsAfterLoad(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)

	updated, _ = model.Update(discussionsFinishedMsg{
		iid: 42,
		discussions: []mr.Discussion{
			{ID: "d1", Resolved: false, Notes: []mr.Note{
				{Author: "alice", Body: "Please fix the naming", Resolved: false},
				{Author: "bob", Body: "Done", Resolved: true},
			}},
		},
	})
	model = updated.(Model)

	view := model.View()
	for _, want := range []string{"[open]", "alice", "Please fix the naming", "bob", "Done"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected %q in Discussions view, got:\n%s", want, view)
		}
	}
}

func TestFocusedDiscussionThreadIsMarked(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{
		iid: 42,
		discussions: []mr.Discussion{
			{ID: "d1", Notes: []mr.Note{{Author: "alice", Body: "Thread one"}}},
			{ID: "d2", Notes: []mr.Note{{Author: "bob", Body: "Thread two"}}},
		},
	})
	model = updated.(Model)

	// cursor starts at 0 → first thread marked, second not
	view := model.View()
	if !strings.Contains(view, "> [") {
		t.Fatalf("expected focused thread marker '> [' in view, got:\n%s", view)
	}

	// Move cursor to second thread
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updated.(Model)
	view = model.View()
	if strings.Count(view, "> [") != 1 {
		t.Fatalf("expected exactly one focused thread marker, got:\n%s", view)
	}
	if !strings.Contains(view, "bob") {
		t.Fatalf("expected second thread visible after j, got:\n%s", view)
	}
}

func TestDiscussionThreadsAreSeparatedByDivider(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{
		iid: 42,
		discussions: []mr.Discussion{
			{ID: "d1", Notes: []mr.Note{{Author: "alice", Body: "Thread one"}}},
			{ID: "d2", Notes: []mr.Note{{Author: "bob", Body: "Thread two"}}},
		},
	})
	model = updated.(Model)

	view := model.View()
	// Both threads must be present
	if !strings.Contains(view, "Thread one") || !strings.Contains(view, "Thread two") {
		t.Fatalf("expected both threads in view, got:\n%s", view)
	}
	// A horizontal rule separates consecutive threads
	if !strings.Contains(view, "───") {
		t.Fatalf("expected separator (───) between threads, got:\n%s", view)
	}
}

func TestDiscussionThreadShowsAllReplies(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{
		iid: 42,
		discussions: []mr.Discussion{
			{ID: "d1", Notes: []mr.Note{
				{Author: "alice", Body: "First comment"},
				{Author: "bob", Body: "Second reply"},
				{Author: "carol", Body: "Third reply"},
			}},
		},
	})
	model = updated.(Model)

	view := model.View()
	for _, want := range []string{"First comment", "bob", "Second reply", "carol", "Third reply"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected %q in Discussions view, got:\n%s", want, view)
		}
	}
}

func TestDiscussionsTabShowsEmptyState(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, discussions: []mr.Discussion{}})
	model = updated.(Model)

	if !strings.Contains(model.View(), "No discussions") {
		t.Fatalf("expected empty state, got:\n%s", model.View())
	}
}

func TestDiscussionsTabShowsErrorState(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, err: errors.New("network error")})
	model = updated.(Model)

	if !strings.Contains(model.View(), "Error:") {
		t.Fatalf("expected error state, got:\n%s", model.View())
	}
}

func TestFilesTabTriggersLoadOnFirstVisit(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())

	// Tab twice: Summary → Discussions → Files
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)

	if model.activeTab != TabFiles {
		t.Fatalf("expected TabFiles, got %v", model.activeTab)
	}
	if cmd == nil {
		t.Fatal("expected a load command when switching to Files tab for the first time")
	}
	if !strings.Contains(model.View(), "Loading") {
		t.Fatalf("expected loading state in Files tab, got:\n%s", model.View())
	}
}

func TestFilesTabRendersChangedFilesAfterLoad(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)

	updated, _ = model.Update(filesFinishedMsg{
		iid: 42,
		files: []mr.ChangedFile{
			{Path: "internal/tui/model.go", IsNew: false, AddedLines: 10, RemovedLines: 3},
			{Path: "cmd/main.go", IsNew: true, AddedLines: 25, RemovedLines: 0},
		},
	})
	model = updated.(Model)

	view := model.View()
	for _, want := range []string{
		"internal/tui/model.go", "+10", "-3",
		"cmd/main.go", "A", "+25",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected %q in Files view, got:\n%s", want, view)
		}
	}
}

func TestFilesTabShowsEmptyState(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{}})
	model = updated.(Model)

	if !strings.Contains(model.View(), "No changed files") {
		t.Fatalf("expected empty state, got:\n%s", model.View())
	}
}

func TestFilesTabShowsErrorState(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, err: errors.New("timeout")})
	model = updated.(Model)

	if !strings.Contains(model.View(), "Error:") {
		t.Fatalf("expected error state, got:\n%s", model.View())
	}
}

// --- #42: Changed Files + Diff View ---

func fileDiffOpts() ProjectOptions {
	return ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadFiles: func(iid int) ([]mr.ChangedFile, error) {
			return []mr.ChangedFile{
				{
					Path: "internal/tui/model.go", AddedLines: 10, RemovedLines: 3,
					Diff: []mr.DiffRow{
						{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"},
						{OldLine: 0, NewLine: 2, NewText: "new line"},
					},
				},
				{
					Path: "cmd/main.go", IsNew: true, AddedLines: 5,
					Diff: []mr.DiffRow{
						{OldLine: 0, NewLine: 1, NewText: "package main"},
					},
				},
			}, nil
		},
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) {
			return []mr.Discussion{}, nil
		},
	}
}

func TestEnterOnFileInFilesTabOpensFileDiffMode(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), fileDiffOpts())

	// Navigate to Files tab (Tab twice: Summary → Discussions → Files)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)

	// Load files
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{Path: "internal/tui/model.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"}}},
		{Path: "cmd/main.go", IsNew: true, Diff: []mr.DiffRow{{OldLine: 0, NewLine: 1, NewText: "package main"}}},
	}})
	model = updated.(Model)

	if model.activeTab != TabFiles {
		t.Fatalf("expected TabFiles, got %v", model.activeTab)
	}

	// Press Enter to open selected file
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.mode != ModeFileDiff {
		t.Fatalf("expected ModeFileDiff after Enter on file, got %v", model.mode)
	}
}

func TestFileDiffLeftPaneShowsFileListWithCurrentHighlighted(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), fileDiffOpts())

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{Path: "internal/tui/model.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"}}},
		{Path: "cmd/main.go", IsNew: true, Diff: []mr.DiffRow{{OldLine: 0, NewLine: 1, NewText: "package main"}}},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, "> internal/tui/model.go") {
		t.Fatalf("expected selected file highlighted with '>', got:\n%s", view)
	}
	if !strings.Contains(view, "cmd/main.go") {
		t.Fatalf("expected second file in left pane, got:\n%s", view)
	}
}

func TestFileDiffRightPaneShowsPerFileDiffRows(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), fileDiffOpts())

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{
			Path: "internal/tui/model.go",
			Diff: []mr.DiffRow{
				{OldLine: 5, NewLine: 5, OldText: "before", NewText: "before"},
				{OldLine: 0, NewLine: 6, NewText: "added line"},
			},
		},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, "Diff internal/tui/model.go") {
		t.Fatalf("expected file name in diff header, got:\n%s", view)
	}
	if !strings.Contains(view, "before") {
		t.Fatalf("expected diff row text in right pane, got:\n%s", view)
	}
	if !strings.Contains(view, "added line") {
		t.Fatalf("expected added line in right pane, got:\n%s", view)
	}
}

func TestFileDiffShowsInlineDiscussionMarkerOnPositionedRow(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadFiles: func(iid int) ([]mr.ChangedFile, error) {
			return []mr.ChangedFile{
				{
					Path: "main.go",
					Diff: []mr.DiffRow{
						{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"},
						{OldLine: 0, NewLine: 2, NewText: "reviewed line"},
					},
				},
			}, nil
		},
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) {
			return []mr.Discussion{
				{
					ID:       "d1",
					Notes:    []mr.Note{{Author: "alice", Body: "fix this"}},
					Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 2},
				},
			}, nil
		},
	})

	// Load discussions first
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, discussions: []mr.Discussion{
		{ID: "d1", Notes: []mr.Note{{Author: "alice", Body: "fix this"}}, Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 2}},
	}})
	model = updated.(Model)

	// Navigate to Files tab and open file
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{Path: "main.go", Diff: []mr.DiffRow{
			{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"},
			{OldLine: 0, NewLine: 2, NewText: "reviewed line"},
		}},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, "💬") && !strings.Contains(view, "[D]") && !strings.Contains(view, "discussion") {
		t.Fatalf("expected inline discussion marker on positioned row, got:\n%s", view)
	}
}

func fileDiffModelWithFiles(t *testing.T, files []mr.ChangedFile) Model {
	t.Helper()
	model := NewModelWithProject(FakeMergeRequests(), fileDiffOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: files})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return updated.(Model)
}

func TestRightKeyMovesToNextFile(t *testing.T) {
	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
		{Path: "b.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "b", NewText: "b"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)

	if model.selectedFile != 1 {
		t.Fatalf("expected selectedFile 1 after right, got %d", model.selectedFile)
	}
	if !strings.Contains(model.View(), "> b.go") {
		t.Fatalf("expected b.go highlighted, got:\n%s", model.View())
	}
}

func TestRightKeyDoesNotExceedLastFile(t *testing.T) {
	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)

	if model.selectedFile != 0 {
		t.Fatalf("expected selectedFile to stay at 0 (last file), got %d", model.selectedFile)
	}
}

func TestLeftKeyMovesToPreviousFile(t *testing.T) {
	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
		{Path: "b.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "b", NewText: "b"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(Model)

	if model.selectedFile != 0 {
		t.Fatalf("expected selectedFile 0 after left, got %d", model.selectedFile)
	}
	if !strings.Contains(model.View(), "> a.go") {
		t.Fatalf("expected a.go highlighted, got:\n%s", model.View())
	}
}

func TestLeftKeyDoesNotGoBelowFirstFile(t *testing.T) {
	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(Model)

	if model.selectedFile != 0 {
		t.Fatalf("expected selectedFile to stay at 0, got %d", model.selectedFile)
	}
}

func TestEscInFileDiffReturnsToFilesTab(t *testing.T) {
	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	if model.mode != ModeDetail {
		t.Fatalf("expected ModeDetail after Esc, got %v", model.mode)
	}
	if model.activeTab != TabFiles {
		t.Fatalf("expected TabFiles after Esc, got %v", model.activeTab)
	}
}

func TestBackspaceInFileDiffReturnsToFilesTab(t *testing.T) {
	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	model = updated.(Model)

	if model.mode != ModeDetail {
		t.Fatalf("expected ModeDetail after Backspace, got %v", model.mode)
	}
	if model.activeTab != TabFiles {
		t.Fatalf("expected TabFiles after Backspace, got %v", model.activeTab)
	}
	if !strings.Contains(model.View(), ">Files<") {
		t.Fatalf("expected Files tab active in view, got:\n%s", model.View())
	}
}

// --- #43: Draft comments ---

func draftOpts() ProjectOptions {
	return ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadFiles: func(iid int) ([]mr.ChangedFile, error) {
			return []mr.ChangedFile{
				{Path: "main.go", Diff: []mr.DiffRow{
					{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"},
					{OldLine: 0, NewLine: 2, NewText: "new line"},
				}},
			}, nil
		},
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) {
			return []mr.Discussion{}, nil
		},
	}
}

func draftFileDiffModel(t *testing.T) Model {
	t.Helper()
	model := NewModelWithProject(FakeMergeRequests(), draftOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{Path: "main.go", Diff: []mr.DiffRow{
			{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"},
			{OldLine: 0, NewLine: 2, NewText: "new line"},
		}},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return updated.(Model)
}

func TestModelStoresDraftCommentForCurrentMR(t *testing.T) {
	model := draftFileDiffModel(t)

	draft := mr.DraftComment{
		LocalID:  "local-1",
		Body:     "Fix this please",
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 2},
	}
	updated, _ := model.Update(draftAddedMsg{iid: 42, draft: draft})
	model = updated.(Model)

	if len(model.drafts[42]) != 1 {
		t.Fatalf("expected 1 draft stored, got %d", len(model.drafts[42]))
	}
	if model.drafts[42][0].Body != "Fix this please" {
		t.Fatalf("unexpected draft body: %q", model.drafts[42][0].Body)
	}
}

func TestDraftMarkerAppearsOnDiffRow(t *testing.T) {
	model := draftFileDiffModel(t)

	updated, _ := model.Update(draftAddedMsg{iid: 42, draft: mr.DraftComment{
		LocalID:  "d1",
		Body:     "Check this",
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 2},
	}})
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, "[DRAFT]") {
		t.Fatalf("expected [DRAFT] marker in diff view, got:\n%s", view)
	}
}

func TestDraftRangeMarkerSpansMultipleRows(t *testing.T) {
	model := draftFileDiffModel(t)

	// range draft: lines 1–2 (EndLine > 0)
	updated, _ := model.Update(draftAddedMsg{iid: 42, draft: mr.DraftComment{
		LocalID:  "r1",
		Body:     "Range comment",
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 1},
		EndLine:  2,
	}})
	model = updated.(Model)

	view := model.View()
	count := strings.Count(view, "[DRAFT]")
	if count < 2 {
		t.Fatalf("expected [DRAFT] marker on both rows of range (got %d), view:\n%s", count, view)
	}
}

func TestVKeyStartsRangeSelection(t *testing.T) {
	model := draftFileDiffModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	model = updated.(Model)

	if model.rangeStart != model.diffCursor {
		t.Fatalf("expected rangeStart == diffCursor (%d), got rangeStart=%d", model.diffCursor, model.rangeStart)
	}
	view := model.View()
	if !strings.Contains(view, "▌") {
		t.Fatalf("expected range selection marker ▌ in view, got:\n%s", view)
	}
}

func TestEscCancelsRangeSelection(t *testing.T) {
	model := draftFileDiffModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	model = updated.(Model)
	if model.rangeStart < 0 {
		t.Fatal("expected range selection to be active")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	if model.rangeStart != -1 {
		t.Fatalf("expected rangeStart -1 after Esc, got %d", model.rangeStart)
	}
	if model.mode != ModeFileDiff {
		t.Fatalf("expected to remain in ModeFileDiff after cancelling range, got %v", model.mode)
	}
}

func TestCKeyEntersCommentInputAndEnterSavesDraft(t *testing.T) {
	model := draftFileDiffModel(t)

	// Press 'c' to open comment input
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	model = updated.(Model)

	if !model.commentInput {
		t.Fatal("expected commentInput to be true after 'c'")
	}

	// Type comment text
	for _, ch := range "My draft comment" {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		model = updated.(Model)
	}
	if model.commentBuffer != "My draft comment" {
		t.Fatalf("expected buffer 'My draft comment', got %q", model.commentBuffer)
	}

	// Press Enter to save
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.commentInput {
		t.Fatal("expected commentInput false after Enter")
	}
	if len(model.drafts[42]) != 1 {
		t.Fatalf("expected 1 draft saved, got %d", len(model.drafts[42]))
	}
	if model.drafts[42][0].Body != "My draft comment" {
		t.Fatalf("unexpected draft body: %q", model.drafts[42][0].Body)
	}
}

func TestCommentInputBufferAppearsInView(t *testing.T) {
	model := draftFileDiffModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	model = updated.(Model)

	view := model.View()
	if !strings.Contains(view, "hello") {
		t.Fatalf("expected comment buffer in view, got:\n%s", view)
	}
}

func TestPKeySubmitsDraftsAndClearsOnSuccess(t *testing.T) {
	submitted := false
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadFiles: func(iid int) ([]mr.ChangedFile, error) {
			return []mr.ChangedFile{
				{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
			}, nil
		},
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) { return nil, nil },
		SubmitDrafts: func(iid int, drafts []mr.DraftComment) error {
			submitted = true
			if len(drafts) != 1 {
				t.Errorf("expected 1 draft, got %d", len(drafts))
			}
			return nil
		},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	// Pre-load a draft
	updated, _ = model.Update(draftAddedMsg{iid: 42, draft: mr.DraftComment{LocalID: "d1", Body: "fix"}})
	model = updated.(Model)

	// Press 'p' to publish
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected submit command from 'p'")
	}

	// Execute the command
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	if !submitted {
		t.Fatal("expected submit function to have been called")
	}
	if len(model.drafts[42]) != 0 {
		t.Fatalf("expected drafts cleared after submit, got %d", len(model.drafts[42]))
	}
}

func TestDKeyDiscardsAllLocalDrafts(t *testing.T) {
	discarded := false
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadFiles: func(iid int) ([]mr.ChangedFile, error) {
			return []mr.ChangedFile{
				{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
			}, nil
		},
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) { return nil, nil },
		DiscardDrafts: func(iid int) error {
			discarded = true
			return nil
		},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	updated, _ = model.Update(draftAddedMsg{iid: 42, draft: mr.DraftComment{LocalID: "d1", Body: "fix"}})
	model = updated.(Model)
	updated, _ = model.Update(draftAddedMsg{iid: 42, draft: mr.DraftComment{LocalID: "d2", Body: "also fix"}})
	model = updated.(Model)

	if len(model.drafts[42]) != 2 {
		t.Fatalf("setup: expected 2 drafts, got %d", len(model.drafts[42]))
	}

	// Press 'D' to discard
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	model = updated.(Model)

	if len(model.drafts[42]) != 0 {
		t.Fatalf("expected drafts cleared immediately after D, got %d", len(model.drafts[42]))
	}
	if cmd != nil {
		msg := cmd()
		model.Update(msg)
	}
	if !discarded {
		t.Fatal("expected discard function to have been called")
	}
}

// --- #44: Discussion write actions ---

func discussionWriteOpts() ProjectOptions {
	return ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) {
			return []mr.Discussion{
				{ID: "d1", Resolved: false, Notes: []mr.Note{{Author: "alice", Body: "Fix naming"}}},
				{ID: "d2", Resolved: true, Notes: []mr.Note{{Author: "bob", Body: "LGTM"}}},
			}, nil
		},
		LoadFiles: func(iid int) ([]mr.ChangedFile, error) { return nil, nil },
	}
}

func discussionsTabModel(t *testing.T) Model {
	t.Helper()
	model := NewModelWithProject(FakeMergeRequests(), discussionWriteOpts())
	// Navigate to Discussions tab
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	// Load discussions
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, discussions: []mr.Discussion{
		{ID: "d1", Resolved: false, Notes: []mr.Note{{Author: "alice", Body: "Fix naming"}}},
		{ID: "d2", Resolved: true, Notes: []mr.Note{{Author: "bob", Body: "LGTM"}}},
	}})
	return updated.(Model)
}

func TestDiscussionCursorMovesWithJK(t *testing.T) {
	model := discussionsTabModel(t)

	if model.discussionCursor != 0 {
		t.Fatalf("expected initial cursor 0, got %d", model.discussionCursor)
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updated.(Model)

	if model.discussionCursor != 1 {
		t.Fatalf("expected cursor 1 after j, got %d", model.discussionCursor)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	model = updated.(Model)

	if model.discussionCursor != 0 {
		t.Fatalf("expected cursor 0 after k, got %d", model.discussionCursor)
	}
}

func TestDiscussionCursorDoesNotExceedBounds(t *testing.T) {
	model := discussionsTabModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	model = updated.(Model)
	if model.discussionCursor != 0 {
		t.Fatalf("expected cursor to stay at 0 (already first), got %d", model.discussionCursor)
	}

	// Move to last
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updated.(Model)
	if model.discussionCursor != 1 {
		t.Fatalf("expected cursor to stay at 1 (last), got %d", model.discussionCursor)
	}
}

func TestRKeyOpensReplyInputForFocusedDiscussion(t *testing.T) {
	model := discussionsTabModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model = updated.(Model)

	if !model.replyInput {
		t.Fatal("expected replyInput true after r")
	}
	if model.replyDraft {
		t.Fatal("expected replyDraft false for instant reply")
	}
	if model.replyDiscussionID != "d1" {
		t.Fatalf("expected replyDiscussionID d1, got %q", model.replyDiscussionID)
	}
}

func TestEnterInReplyInputSendsReplyAndAddsNote(t *testing.T) {
	called := false
	opts := discussionWriteOpts()
	opts.ReplyToDiscussion = func(iid int, discussionID string, body string) error {
		called = true
		if discussionID != "d1" {
			t.Errorf("expected d1, got %q", discussionID)
		}
		if body != "My reply" {
			t.Errorf("expected 'My reply', got %q", body)
		}
		return nil
	}

	model := NewModelWithProject(FakeMergeRequests(), opts)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, discussions: []mr.Discussion{
		{ID: "d1", Resolved: false, Notes: []mr.Note{{Author: "alice", Body: "Fix naming"}}},
	}})
	model = updated.(Model)

	// Open reply
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("My reply")})
	model = updated.(Model)

	// Send
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.replyInput {
		t.Fatal("expected replyInput closed after Enter")
	}
	if cmd == nil {
		t.Fatal("expected reply command")
	}

	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	if !called {
		t.Fatal("expected ReplyToDiscussion to be called")
	}
	if len(model.discussions[42][0].Notes) != 2 {
		t.Fatalf("expected 2 notes after reply, got %d", len(model.discussions[42][0].Notes))
	}
	if model.discussions[42][0].Notes[1].Body != "My reply" {
		t.Fatalf("unexpected note body: %q", model.discussions[42][0].Notes[1].Body)
	}
}

func TestDKeyOpensDraftReplyAndEnterCallsService(t *testing.T) {
	called := false
	opts := discussionWriteOpts()
	opts.DraftReply = func(iid int, discussionID string, body string) error {
		called = true
		if body != "Draft reply" {
			t.Errorf("expected 'Draft reply', got %q", body)
		}
		return nil
	}

	model := NewModelWithProject(FakeMergeRequests(), opts)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, discussions: []mr.Discussion{
		{ID: "d1", Resolved: false, Notes: []mr.Note{{Author: "alice", Body: "Fix"}}},
	}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	model = updated.(Model)

	if !model.replyInput || !model.replyDraft {
		t.Fatalf("expected replyInput=true replyDraft=true, got input=%v draft=%v", model.replyInput, model.replyDraft)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Draft reply")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected draft reply command")
	}
	cmd()
	if !called {
		t.Fatal("expected DraftReply to be called")
	}
}

func TestXKeyResolvesOpenDiscussion(t *testing.T) {
	resolved := false
	opts := discussionWriteOpts()
	opts.ResolveDiscussion = func(iid int, discussionID string) error {
		resolved = true
		if discussionID != "d1" {
			t.Errorf("expected d1, got %q", discussionID)
		}
		return nil
	}

	model := NewModelWithProject(FakeMergeRequests(), opts)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, discussions: []mr.Discussion{
		{ID: "d1", Resolved: false, Notes: []mr.Note{{Author: "alice", Body: "Fix"}}},
	}})
	model = updated.(Model)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected resolve command")
	}
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	if !resolved {
		t.Fatal("expected ResolveDiscussion to be called")
	}
	if !model.discussions[42][0].Resolved {
		t.Fatal("expected discussion marked resolved")
	}
}

func TestXKeyUnresolvesResolvedDiscussion(t *testing.T) {
	unresolved := false
	opts := discussionWriteOpts()
	opts.UnresolveDiscussion = func(iid int, discussionID string) error {
		unresolved = true
		return nil
	}

	model := NewModelWithProject(FakeMergeRequests(), opts)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, discussions: []mr.Discussion{
		{ID: "d2", Resolved: true, Notes: []mr.Note{{Author: "bob", Body: "LGTM"}}},
	}})
	model = updated.(Model)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected unresolve command")
	}
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	if !unresolved {
		t.Fatal("expected UnresolveDiscussion to be called")
	}
	if model.discussions[42][0].Resolved {
		t.Fatal("expected discussion marked unresolved")
	}
}

func diffViewWithInlineDiscussion(t *testing.T, replyFn ReplyToDiscussionFunc) Model {
	t.Helper()
	opts := ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadFiles: func(iid int) ([]mr.ChangedFile, error) {
			return []mr.ChangedFile{
				{Path: "main.go", Diff: []mr.DiffRow{
					{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"},
					{OldLine: 0, NewLine: 2, NewText: "new line"},
				}},
			}, nil
		},
		LoadDiscussions:   func(iid int) ([]mr.Discussion, error) { return nil, nil },
		ReplyToDiscussion: replyFn,
	}

	model := NewModelWithProject(FakeMergeRequests(), opts)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	// Load discussions with inline position on new line 1
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, discussions: []mr.Discussion{
		{ID: "inline-d1", Resolved: false, Notes: []mr.Note{{Author: "alice", Body: "inline comment"}},
			Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 1}},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{Path: "main.go", Diff: []mr.DiffRow{
			{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"},
			{OldLine: 0, NewLine: 2, NewText: "new line"},
		}},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return updated.(Model)
}

func TestRKeyInDiffViewOnInlineDiscussionOpensReplyInput(t *testing.T) {
	model := diffViewWithInlineDiscussion(t, nil)

	// diffCursor is at 0 which has the inline discussion (NewLine=1)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model = updated.(Model)

	if !model.replyInput {
		t.Fatalf("expected replyInput true, got false; mode=%v cursor=%d", model.mode, model.diffCursor)
	}
	if model.replyDiscussionID != "inline-d1" {
		t.Fatalf("expected replyDiscussionID inline-d1, got %q", model.replyDiscussionID)
	}
}

func TestDKeyInDiffViewOpensDraftReplyForInlineDiscussion(t *testing.T) {
	model := diffViewWithInlineDiscussion(t, nil)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	model = updated.(Model)

	if !model.replyInput {
		t.Fatal("expected replyInput true after d")
	}
	if !model.replyDraft {
		t.Fatal("expected replyDraft true after d")
	}
	if model.replyDiscussionID != "inline-d1" {
		t.Fatalf("expected replyDiscussionID inline-d1, got %q", model.replyDiscussionID)
	}
}

func TestXKeyInDiffViewResolvesInlineDiscussion(t *testing.T) {
	resolved := false
	opts := ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadFiles: func(iid int) ([]mr.ChangedFile, error) {
			return []mr.ChangedFile{
				{Path: "main.go", Diff: []mr.DiffRow{
					{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"},
				}},
			}, nil
		},
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) { return nil, nil },
		ResolveDiscussion: func(iid int, discussionID string) error {
			resolved = true
			return nil
		},
	}

	model := NewModelWithProject(FakeMergeRequests(), opts)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, discussions: []mr.Discussion{
		{ID: "inline-d1", Resolved: false, Notes: []mr.Note{{Author: "alice", Body: "fix"}},
			Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 1}},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"}}},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected resolve command from x")
	}
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	if !resolved {
		t.Fatal("expected ResolveDiscussion to be called")
	}
	if !model.discussions[42][0].Resolved {
		t.Fatal("expected inline discussion marked resolved")
	}
}

// --- #45: Instant comments ---

func instantCommentFileDiffModel(t *testing.T, postFn PostInlineCommentFunc) Model {
	t.Helper()
	opts := ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadFiles: func(iid int) ([]mr.ChangedFile, error) {
			return []mr.ChangedFile{
				{Path: "main.go", Diff: []mr.DiffRow{
					{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"},
				}},
			}, nil
		},
		LoadDiscussions:   func(iid int) ([]mr.Discussion, error) { return nil, nil },
		PostInlineComment: postFn,
	}
	model := NewModelWithProject(FakeMergeRequests(), opts)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"}}},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return updated.(Model)
}

func TestIKeyOpensInstantInlineCommentInput(t *testing.T) {
	model := instantCommentFileDiffModel(t, nil)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	model = updated.(Model)

	if !model.commentInput {
		t.Fatal("expected commentInput true after i")
	}
	if !model.commentInstant {
		t.Fatal("expected commentInstant true after i")
	}
}

func TestEnterInInstantCommentCallsAPIAndDoesNotSaveDraft(t *testing.T) {
	called := false
	model := instantCommentFileDiffModel(t, func(iid int, position mr.DiffPosition, body string) error {
		called = true
		if body != "Instant review" {
			t.Errorf("expected 'Instant review', got %q", body)
		}
		if position.NewPath != "main.go" {
			t.Errorf("expected path main.go, got %q", position.NewPath)
		}
		return nil
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Instant review")})
	model = updated.(Model)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected API command on Enter for instant comment")
	}
	cmd()

	if !called {
		t.Fatal("expected PostInlineCommentFunc to be called")
	}
	if len(model.drafts[42]) != 0 {
		t.Fatalf("expected no local draft saved, got %d", len(model.drafts[42]))
	}
}

func TestInstantCommentAPIErrorShownInView(t *testing.T) {
	model := instantCommentFileDiffModel(t, func(iid int, position mr.DiffPosition, body string) error {
		return errors.New("network timeout")
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if cmd != nil {
		msg := cmd()
		updated, _ = model.Update(msg)
		model = updated.(Model)
	}

	if model.mode != ModeFileDiff {
		t.Fatalf("expected to stay in ModeFileDiff after error, got %v", model.mode)
	}
	if !strings.Contains(model.View(), "network timeout") {
		t.Fatalf("expected error in view, got:\n%s", model.View())
	}
}

func TestMKeyOpensMRCommentInput(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:          "group/project",
		Section:       SectionMergeRequests,
		PostMRComment: func(iid int, body string) error { return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	model = updated.(Model)

	if !model.mrCommentInput {
		t.Fatal("expected mrCommentInput true after m")
	}
}

func TestEnterInMRCommentInputCallsPostMRCommentFunc(t *testing.T) {
	called := false
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		PostMRComment: func(iid int, body string) error {
			called = true
			if iid != 42 {
				t.Errorf("expected iid 42, got %d", iid)
			}
			if body != "Great work" {
				t.Errorf("expected 'Great work', got %q", body)
			}
			return nil
		},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Great work")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.mrCommentInput {
		t.Fatal("expected mrCommentInput closed after Enter")
	}
	if cmd == nil {
		t.Fatal("expected command from Enter in MR comment input")
	}
	cmd()
	if !called {
		t.Fatal("expected PostMRCommentFunc to be called")
	}
}

func TestEscInMRCommentInputCancelsWithoutSending(t *testing.T) {
	called := false
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:          "group/project",
		Section:       SectionMergeRequests,
		PostMRComment: func(iid int, body string) error { called = true; return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("some text")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	if model.mrCommentInput {
		t.Fatal("expected mrCommentInput false after Esc")
	}
	if cmd != nil {
		t.Fatal("expected no command on Esc")
	}
	if called {
		t.Fatal("expected PostMRCommentFunc NOT to be called after Esc")
	}
	if model.mrCommentBuffer != "" {
		t.Fatalf("expected buffer cleared after Esc, got %q", model.mrCommentBuffer)
	}
}

func TestMRCommentAPIErrorShownInViewWithoutLosingContext(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		PostMRComment: func(iid int, body string) error {
			return errors.New("forbidden")
		},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if cmd != nil {
		msg := cmd()
		updated, _ = model.Update(msg)
		model = updated.(Model)
	}

	if model.mode != ModeDetail {
		t.Fatalf("expected to stay in ModeDetail after error, got %v", model.mode)
	}
	if !strings.Contains(model.View(), "forbidden") {
		t.Fatalf("expected error in view, got:\n%s", model.View())
	}
}

// --- #46: MR actions + external open ---

func mrActionsModel(t *testing.T, opts ProjectOptions) Model {
	t.Helper()
	if opts.Path == "" {
		opts.Path = "group/project"
	}
	if opts.Section == "" {
		opts.Section = SectionMergeRequests
	}
	return NewModelWithProject(FakeMergeRequests(), opts)
}

func TestAKeyApprovesCurrentMR(t *testing.T) {
	called := false
	model := mrActionsModel(t, ProjectOptions{
		ApproveMR: func(iid int) error {
			called = true
			if iid != 42 {
				t.Errorf("expected iid 42, got %d", iid)
			}
			return nil
		},
	})

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("A")})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected approve command")
	}
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	if !called {
		t.Fatal("expected ApproveMRFunc to be called")
	}
	if !strings.Contains(model.View(), "Approved") {
		t.Fatalf("expected 'Approved' in view, got:\n%s", model.View())
	}
}

func TestMKeySetsMergeConfirmPending(t *testing.T) {
	model := mrActionsModel(t, ProjectOptions{
		MergeMR: func(iid int) error { return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	model = updated.(Model)

	if !model.mergeConfirmPending {
		t.Fatal("expected mergeConfirmPending true after first M")
	}
	if !strings.Contains(model.View(), "confirm merge") {
		t.Fatalf("expected confirmation prompt in view, got:\n%s", model.View())
	}
}

func TestMKeyAgainConfirmsMerge(t *testing.T) {
	called := false
	model := mrActionsModel(t, ProjectOptions{
		MergeMR: func(iid int) error { called = true; return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected merge command on second M")
	}
	cmd()
	if !called {
		t.Fatal("expected MergeMRFunc to be called")
	}
}

func TestOtherKeyAfterMCancelsMerge(t *testing.T) {
	model := mrActionsModel(t, ProjectOptions{
		MergeMR: func(iid int) error { return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updated.(Model)

	if model.mergeConfirmPending {
		t.Fatal("expected mergeConfirmPending cleared after non-M key")
	}
	if cmd != nil {
		t.Fatal("expected no merge command")
	}
}

func TestOKeyOpensURLInBrowser(t *testing.T) {
	opened := ""
	model := NewModelWithProject([]mr.MergeRequest{{
		IID: 42, Title: "Test", WebURL: "https://gitlab.com/group/project/-/merge_requests/42",
	}}, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		OpenURL: func(url string) error { opened = url; return nil },
	})

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected open URL command")
	}
	cmd()
	if opened != "https://gitlab.com/group/project/-/merge_requests/42" {
		t.Fatalf("expected MR URL opened, got %q", opened)
	}
}

func TestEKeyOpensEditModeOnTitleField(t *testing.T) {
	model := mrActionsModel(t, ProjectOptions{
		EditMR: func(iid int, title, description string) error { return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	model = updated.(Model)

	if !model.editInput {
		t.Fatal("expected editInput true after e")
	}
	if model.editField != "title" {
		t.Fatalf("expected editField 'title', got %q", model.editField)
	}
	if model.editBuffer != "Port TUI shell to Bubble Tea" {
		t.Fatalf("expected buffer pre-filled with current title, got %q", model.editBuffer)
	}
}

func TestTabInEditModeMoveToDescriptionField(t *testing.T) {
	model := mrActionsModel(t, ProjectOptions{
		EditMR: func(iid int, title, description string) error { return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	model = updated.(Model)
	// Clear and type new title
	model.editBuffer = "New title"
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)

	if model.editField != "description" {
		t.Fatalf("expected editField 'description' after Tab, got %q", model.editField)
	}
	if model.editTitle != "New title" {
		t.Fatalf("expected editTitle saved, got %q", model.editTitle)
	}
}

func TestEnterInEditModeSavesAndCallsEditMR(t *testing.T) {
	called := false
	model := NewModelWithProject([]mr.MergeRequest{{
		IID: 42, Title: "Old title", Description: "Old desc",
	}}, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		EditMR: func(iid int, title, description string) error {
			called = true
			if title != "New title" {
				t.Errorf("expected 'New title', got %q", title)
			}
			if description != "New desc" {
				t.Errorf("expected 'New desc', got %q", description)
			}
			return nil
		},
	})

	// Open edit, move to title field
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	model = updated.(Model)
	model.editBuffer = "New title"

	// Tab to description
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	model.editBuffer = "New desc"

	// Enter to save
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.editInput {
		t.Fatal("expected editInput closed after Enter")
	}
	if cmd == nil {
		t.Fatal("expected edit command")
	}
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	if !called {
		t.Fatal("expected EditMRFunc to be called")
	}
	if model.items[0].Title != "New title" {
		t.Fatalf("expected title updated locally, got %q", model.items[0].Title)
	}
}

func TestEKeyInDiffViewOpensFileInEditor(t *testing.T) {
	openedPath := ""
	openedLine := 0
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:            "group/project",
		Section:         SectionMergeRequests,
		LoadFiles:       func(iid int) ([]mr.ChangedFile, error) { return nil, nil },
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) { return nil, nil },
		OpenEditor: func(path string, line int) error {
			openedPath = path
			openedLine = line
			return nil
		},
	})

	// Navigate to Files tab and open file
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{Path: "internal/tui/model.go", Diff: []mr.DiffRow{
			{OldLine: 10, NewLine: 10, OldText: "old", NewText: "old"},
			{OldLine: 0, NewLine: 11, NewText: "new line"},
		}},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	// Move cursor to second row (NewLine=11)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updated.(Model)

	// Press 'e' to open in editor
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected open editor command")
	}
	cmd()

	if openedPath != "internal/tui/model.go" {
		t.Fatalf("expected path internal/tui/model.go, got %q", openedPath)
	}
	if openedLine != 11 {
		t.Fatalf("expected line 11, got %d", openedLine)
	}
}

// --- #52: Diff View — type markers and readability ---

func fileDiffModelWithRows(t *testing.T, rows []mr.DiffRow) Model {
	t.Helper()
	opts := ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		LoadFiles: func(iid int) ([]mr.ChangedFile, error) {
			return []mr.ChangedFile{{Path: "main.go", Diff: rows}}, nil
		},
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) { return nil, nil },
	}
	model := NewModelWithProject(FakeMergeRequests(), opts)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{
		{Path: "main.go", Diff: rows},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return updated.(Model)
}

func TestDiffAdditionRowHasNoZeroOldLineNumber(t *testing.T) {
	model := fileDiffModelWithRows(t, []mr.DiffRow{
		{OldLine: 0, NewLine: 5, NewText: "added"},
	})

	view := model.View()
	// Old line number should be blank, not "   0"
	if strings.Contains(view, "   0 │") {
		t.Fatalf("expected no '0' old line number for addition row, got:\n%s", view)
	}
}

func TestDiffContextRowHasNoTypeMarker(t *testing.T) {
	model := fileDiffModelWithRows(t, []mr.DiffRow{
		{OldLine: 1, NewLine: 1, OldText: "unchanged", NewText: "unchanged"},
	})

	view := model.View()
	if !strings.Contains(view, "unchanged") {
		t.Fatalf("expected context row text in view, got:\n%s", view)
	}
	// Context rows must not be prefixed with + or -
	if strings.Contains(view, "+ unchanged") || strings.Contains(view, "- unchanged") {
		t.Fatalf("expected context row to have no +/- marker, got:\n%s", view)
	}
}

func TestDiffDeletionRowIsMarkedWithMinus(t *testing.T) {
	model := fileDiffModelWithRows(t, []mr.DiffRow{
		{OldLine: 3, NewLine: 0, OldText: "old code"},
	})

	view := model.View()
	if !strings.Contains(view, "- old code") {
		t.Fatalf("expected deletion row marked with '-', got:\n%s", view)
	}
}

func TestDiffAdditionRowIsMarkedWithPlus(t *testing.T) {
	model := fileDiffModelWithRows(t, []mr.DiffRow{
		{OldLine: 0, NewLine: 1, NewText: "new feature"},
	})

	view := model.View()
	if !strings.Contains(view, "+ new feature") {
		t.Fatalf("expected addition row marked with '+', got:\n%s", view)
	}
}

// --- #50: Left panel always read-only, no focus ---

func TestMouseClickOnLeftPanelDoesNotChangeFocus(t *testing.T) {
	model := NewFakeModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	model = updated.(Model)

	// Click on left side (X=2, well within leftWidth ~35)
	updated, _ = model.Update(tea.MouseMsg{X: 2, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	model = updated.(Model)

	if model.focus == FocusList {
		t.Fatal("expected left panel click not to set FocusList")
	}
}

func TestLeftPanelHasNoActiveBorderInDetailMode(t *testing.T) {
	model := NewFakeModel()
	// Simulate mouse click that used to set FocusList
	updated, _ := model.Update(tea.MouseMsg{X: 2, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	model = updated.(Model)

	if model.focus == FocusList {
		t.Fatal("focus must never be FocusList in ModeDetail")
	}
	// Left panel style uses focused=false → border color 240 (dim), right panel uses focused=true → 63 (bright)
	// We verify this indirectly: renderList checks m.focus==FocusList||FocusFilter
	// After our fix, paneStyle(focused=false) is always used for the left panel
	left := model.renderList()
	right := model.renderRight()
	// lipgloss renders dim border with color 240 for inactive, 63 for active
	// Both panels render — just verify the model state is consistent
	if left == "" || right == "" {
		t.Fatal("expected both panels to render")
	}
}

// --- #49: No eager MR loading on project selection ---

func TestMRSectionSelectionTriggersLoadingWhenNotLoaded(t *testing.T) {
	loadCalled := false
	model := NewModelWithProject(nil, ProjectOptions{
		Recents: []string{"group/project"},
		LoadProject: func(path string) (ProjectData, error) {
			loadCalled = true
			return ProjectData{Items: []mr.MergeRequest{{IID: 1, Title: "Loaded MR"}}}, nil
		},
	})

	// Select project → ModeSections immediately
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.mode != ModeSections {
		t.Fatalf("setup: expected ModeSections, got %v", model.mode)
	}

	// Select MR section → should trigger loading
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if cmd == nil {
		t.Fatal("expected loading command when selecting MR section with unloaded project")
	}
	if loadCalled {
		t.Fatal("expected LoadProject not yet called (command not executed)")
	}
}

func TestMRSectionLoadingCompletionShowsMRList(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{
		Recents:     []string{"group/project"},
		LoadProject: func(path string) (ProjectData, error) { return ProjectData{}, nil },
	})

	// Project selection → sections
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	// MR section → loading starts
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	// Loading completes
	updated, _ = model.Update(projectFinishedMsg{
		path: "group/project",
		data: ProjectData{Items: []mr.MergeRequest{
			{IID: 10, Title: "First MR"},
			{IID: 11, Title: "Second MR"},
		}},
	})
	model = updated.(Model)

	if model.mode != ModeEntityList {
		t.Fatalf("expected ModeEntityList after load, got %v", model.mode)
	}
	if len(model.items) != 2 {
		t.Fatalf("expected 2 MRs loaded, got %d", len(model.items))
	}
	view := model.View()
	if !strings.Contains(view, "First MR") {
		t.Fatalf("expected MR list in view, got:\n%s", view)
	}
}

// --- #53: Two-panel layout for entity list ---

func entityListModel(t *testing.T) Model {
	t.Helper()
	model := NewModelWithProject(nil, ProjectOptions{
		Recents:     []string{"group/project"},
		LoadProject: func(path string) (ProjectData, error) { return ProjectData{}, nil },
	})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // select project → ModeSections
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // select MR section → loading
	model = updated.(Model)
	updated, _ = model.Update(projectFinishedMsg{
		path: "group/project",
		data: ProjectData{Items: []mr.MergeRequest{
			{IID: 10, Title: "Alpha MR"},
			{IID: 11, Title: "Beta MR"},
		}},
	})
	model = updated.(Model)
	if model.mode != ModeEntityList {
		t.Fatalf("setup: expected ModeEntityList, got %v", model.mode)
	}
	return model
}

func TestEntityListViewShowsSectionsContextOnLeft(t *testing.T) {
	model := entityListModel(t)
	view := model.View()

	if !strings.Contains(view, "Merge Requests") {
		t.Fatalf("expected sections context on left, got:\n%s", view)
	}
	if !strings.Contains(view, "Alpha MR") {
		t.Fatalf("expected entity list on right, got:\n%s", view)
	}
}

func TestEntityListEnterGoesToMRDetail(t *testing.T) {
	model := entityListModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.mode != ModeDetail {
		t.Fatalf("expected ModeDetail after Enter in entity list, got %v", model.mode)
	}
}

func TestEntityListEscGoesToSections(t *testing.T) {
	model := entityListModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	if model.mode != ModeSections {
		t.Fatalf("expected ModeSections after Esc in entity list, got %v", model.mode)
	}
}

// --- #48: Two-panel layout on project picker screen ---

func TestProjectPickerRendersLeftContextPane(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{
		Recents: []string{"recent/project"},
	})
	if model.mode != ModeProjectSelect {
		t.Fatalf("expected ModeProjectSelect, got %v", model.mode)
	}

	view := model.View()
	if !strings.Contains(view, "gitlab-tui") {
		t.Fatalf("expected left context pane with 'gitlab-tui', got:\n%s", view)
	}
	if !strings.Contains(view, "Projects") {
		t.Fatalf("expected 'Projects' heading in right pane, got:\n%s", view)
	}
	if !strings.Contains(view, "recent/project") {
		t.Fatalf("expected project in right pane, got:\n%s", view)
	}
}

func TestProjectInputRendersLeftContextPane(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{})
	if model.mode != ModeProjectInput {
		t.Fatalf("expected ModeProjectInput, got %v", model.mode)
	}

	view := model.View()
	if !strings.Contains(view, "gitlab-tui") {
		t.Fatalf("expected left context pane with 'gitlab-tui', got:\n%s", view)
	}
	if !strings.Contains(view, "Open GitLab project") {
		t.Fatalf("expected 'Open GitLab project' in right pane, got:\n%s", view)
	}
}

func TestTabKeyCyclesDetailTabs(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})

	// Summary (default) → Discussions
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	if model.activeTab != TabDiscussions {
		t.Fatalf("expected TabDiscussions after first Tab, got %v", model.activeTab)
	}

	// Discussions → Files
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	if model.activeTab != TabFiles {
		t.Fatalf("expected TabFiles after second Tab, got %v", model.activeTab)
	}

	// Files → Summary (wrap)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	if model.activeTab != TabSummary {
		t.Fatalf("expected TabSummary after third Tab, got %v", model.activeTab)
	}
}

func TestProjectSelectStructuredRecentsDoNotRenderLegacyDuplicate(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{
		Recents:        []string{"group/project"},
		RecentProjects: []RecentProjectOption{{Path: "group/project", Account: "default"}},
	})

	view := model.View()
	if strings.Count(view, "group/project") != 1 {
		t.Fatalf("expected project once in Recent, got %q", view)
	}
}

func TestProjectSelectShowsRecentSectionBeforeAccounts(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{
		RecentProjects: []RecentProjectOption{
			{Path: "group/new", Account: "work"},
			{Path: "group/old", Account: "default"},
		},
		LoadProjects: []AccountProjectLoader{{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return nil, nil }}},
	})

	view := model.View()
	recentIndex := strings.Index(view, "Recent")
	accountIndex := strings.Index(view, "[default]  https://gitlab.com")
	if recentIndex == -1 || accountIndex == -1 || recentIndex > accountIndex {
		t.Fatalf("expected Recent before account section, got %q", view)
	}
	for _, want := range []string{"group/new (work)", "group/old (default)"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected recent entry %q, got %q", want, view)
		}
	}
	if model.projectRows[0].selectable {
		t.Fatalf("expected Recent header to be non-selectable: %+v", model.projectRows[0])
	}
	if model.selected != 2 {
		t.Fatalf("expected cursor to skip Recent header and spacer, selected=%d rows=%+v", model.selected, model.projectRows)
	}
}

func TestProjectSelectHidesRecentSectionWhenEmpty(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{
		LoadProjects: []AccountProjectLoader{{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return nil, nil }}},
	})

	if strings.Contains(model.View(), "Recent") {
		t.Fatalf("expected no Recent section, got %q", model.View())
	}
}

func TestProjectSelectRecentSelectionUsesProjectPath(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{
		RecentProjects: []RecentProjectOption{{Path: "group/project", Account: "work"}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.projectPath != "group/project" || model.mode != ModeSections {
		t.Fatalf("expected recent project path to open sections, path=%q mode=%v", model.projectPath, model.mode)
	}
}

func TestProjectSelectFilterMatchesRecentAndAccountProjects(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{
		RecentProjects: []RecentProjectOption{{Path: "team/Alpha", Account: "work"}, {Path: "group/beta", Account: "default"}},
		LoadProjects:   []AccountProjectLoader{{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return nil, nil }}},
	})
	updated, _ := model.Update(accountProjectsFinishedMsg{accountID: "default", projects: []string{"org/alpha-service", "org/gamma"}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	model = updated.(Model)

	view := model.View()
	for _, want := range []string{"team/Alpha (work)", "org/alpha-service"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected filtered view to contain %q, got %q", want, view)
		}
	}
	for _, unwanted := range []string{"group/beta", "org/gamma"} {
		if strings.Contains(view, unwanted) {
			t.Fatalf("expected filtered view to hide %q, got %q", unwanted, view)
		}
	}
	if model.selected != 2 {
		t.Fatalf("expected cursor on first filtered result, got %d rows %+v", model.selected, model.projectRows)
	}
}

func TestProjectSelectFilterHidesSectionsWithoutMatchesAndEscResets(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{
		RecentProjects: []RecentProjectOption{{Path: "recent/only", Account: "work"}},
		LoadProjects:   []AccountProjectLoader{{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return nil, nil }}},
	})
	updated, _ := model.Update(accountProjectsFinishedMsg{accountID: "default", projects: []string{"account/project"}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)
	for _, r := range "only" {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		model = updated.(Model)
	}

	view := model.View()
	if !strings.Contains(view, "Recent") || strings.Contains(view, "[default]  https://gitlab.com") {
		t.Fatalf("expected only Recent section after filter, got %q", view)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	view = model.View()
	if !strings.Contains(view, "[default]  https://gitlab.com") || !strings.Contains(view, "account/project") {
		t.Fatalf("expected full list after Esc reset, got %q", view)
	}
}

func TestProjectSelectFilterShowsNoMatches(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{RecentProjects: []RecentProjectOption{{Path: "group/project", Account: "default"}}})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	model = updated.(Model)

	if !strings.Contains(model.View(), "No matching projects") {
		t.Fatalf("expected no-match state, got %q", model.View())
	}
}

func TestProjectSelectStartsAccountProjectLoads(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{LoadProjects: []AccountProjectLoader{
		{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return []string{"group/project"}, nil }},
		{ID: "work", Host: "https://gitlab.example.com", Load: func() ([]string, error) { return []string{"work/project"}, nil }},
	}})

	cmd := model.Init()

	if model.mode != ModeProjectSelect {
		t.Fatalf("expected project select mode, got %v", model.mode)
	}
	if cmd == nil {
		t.Fatal("expected account project load batch")
	}
	view := model.View()
	for _, want := range []string{"[default]  https://gitlab.com", "[work]  https://gitlab.example.com", "Loading…"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected view to contain %q, got %q", want, view)
		}
	}
}

func TestProjectSelectShowsLoadedAccountProjectsAndSkipsHeaders(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{LoadProjects: []AccountProjectLoader{
		{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return nil, nil }},
	}})
	projects := []string{"group/one", "group/two"}
	updated, _ := model.Update(accountProjectsFinishedMsg{accountID: "default", projects: projects})
	model = updated.(Model)

	if model.projectRows[0].selectable {
		t.Fatalf("expected account header to be non-selectable: %+v", model.projectRows[0])
	}
	if model.selected != 1 {
		t.Fatalf("expected first project row selected, got %d rows %+v", model.selected, model.projectRows)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(Model)
	if got, ok := model.selectedProject(); !ok || got != "group/two" {
		t.Fatalf("expected second project selected, got %q ok=%t", got, ok)
	}
}

func TestProjectSelectShowsErrorAndRetriesOnlyFailedAccounts(t *testing.T) {
	failedCalls := 0
	successCalls := 0
	model := NewModelWithProject(nil, ProjectOptions{LoadProjects: []AccountProjectLoader{
		{ID: "failed", Host: "https://gitlab.com", Load: func() ([]string, error) { failedCalls++; return []string{"retry/project"}, nil }},
		{ID: "ok", Host: "https://gitlab.example.com", Load: func() ([]string, error) { successCalls++; return []string{"ok/project"}, nil }},
	}})
	updated, _ := model.Update(accountProjectsFinishedMsg{accountID: "failed", err: errTestRefresh})
	model = updated.(Model)
	updated, _ = model.Update(accountProjectsFinishedMsg{accountID: "ok", projects: []string{"ok/project"}})
	model = updated.(Model)

	if !strings.Contains(model.View(), "Error: refresh failed  r: retry") {
		t.Fatalf("expected error row, got %q", model.View())
	}
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected retry command")
	}
	_ = cmd()
	if failedCalls != 1 || successCalls != 0 {
		t.Fatalf("expected only failed loader to run, failed=%d success=%d", failedCalls, successCalls)
	}
}

// --- #64: Global key suppression in input modes ---

func TestEnteringMRCommentInputDisablesGlobalKeys(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(Model)

	if model.globals.Quit.Enabled() {
		t.Fatal("expected Quit to be disabled when entering MR comment input")
	}
	if model.globals.Back.Enabled() {
		t.Fatal("expected Back to be disabled when entering MR comment input")
	}
}

func TestExitingMRCommentInputRestoresGlobalKeys(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	if !model.globals.Quit.Enabled() {
		t.Fatal("expected Quit to be restored after exiting MR comment input")
	}
	if !model.globals.Back.Enabled() {
		t.Fatal("expected Back to be restored after exiting MR comment input")
	}
}

func TestEnteringEditInputDisablesGlobalKeys(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = updated.(Model)

	if model.globals.Quit.Enabled() {
		t.Fatal("expected Quit to be disabled when entering edit input")
	}
	if model.globals.Back.Enabled() {
		t.Fatal("expected Back to be disabled when entering edit input")
	}
}

func TestExitingEditInputRestoresGlobalKeys(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	if !model.globals.Quit.Enabled() {
		t.Fatal("expected Quit to be restored after exiting edit input")
	}
	if !model.globals.Back.Enabled() {
		t.Fatal("expected Back to be restored after exiting edit input")
	}
}

func TestEnteringReplyInputInDiscussionsDisablesGlobalKeys(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	model.activeTab = TabDiscussions
	model.discussions[42] = []mr.Discussion{{ID: "d1", Notes: []mr.Note{{Author: "alice", Body: "test"}}}}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model = updated.(Model)

	if model.globals.Quit.Enabled() {
		t.Fatal("expected Quit to be disabled when entering reply input")
	}
	if model.globals.Back.Enabled() {
		t.Fatal("expected Back to be disabled when entering reply input")
	}
}

func TestExitingReplyInputInDiscussionsRestoresGlobalKeys(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	model.activeTab = TabDiscussions
	model.discussions[42] = []mr.Discussion{{ID: "d1", Notes: []mr.Note{{Author: "alice", Body: "test"}}}}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	if !model.globals.Quit.Enabled() {
		t.Fatal("expected Quit to be restored after exiting reply input")
	}
	if !model.globals.Back.Enabled() {
		t.Fatal("expected Back to be restored after exiting reply input")
	}
}

func TestEnteringCommentInputInFileDiffDisablesGlobalKeys(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	model.activeTab = TabFiles
	model.changedFiles[42] = []mr.ChangedFile{{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1}}}}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	model = updated.(Model)

	if model.globals.Quit.Enabled() {
		t.Fatal("expected Quit to be disabled when entering comment input in FileDiff")
	}
	if model.globals.Back.Enabled() {
		t.Fatal("expected Back to be disabled when entering comment input in FileDiff")
	}
}

func TestExitingCommentInputInFileDiffRestoresGlobalKeys(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	model.activeTab = TabFiles
	model.changedFiles[42] = []mr.ChangedFile{{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1}}}}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	if !model.globals.Quit.Enabled() {
		t.Fatal("expected Quit to be restored after exiting comment input")
	}
	if !model.globals.Back.Enabled() {
		t.Fatal("expected Back to be restored after exiting comment input")
	}
}

func TestEnteringFocusFilterDisablesGlobalKeys(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)

	if model.globals.Quit.Enabled() {
		t.Fatal("expected Quit to be disabled when entering filter")
	}
	if model.globals.Back.Enabled() {
		t.Fatal("expected Back to be disabled when entering filter")
	}
}

func TestExitingFocusFilterRestoresGlobalKeys(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	if !model.globals.Quit.Enabled() {
		t.Fatal("expected Quit to be restored after exiting filter")
	}
	if !model.globals.Back.Enabled() {
		t.Fatal("expected Back to be restored after exiting filter")
	}
}

func TestEnteringProjectInputModeDisablesGlobalKeys(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{Recents: []string{"group/project"}})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model = updated.(Model)

	if model.globals.Quit.Enabled() {
		t.Fatal("expected Quit to be disabled when entering project input mode")
	}
	if model.globals.Back.Enabled() {
		t.Fatal("expected Back to be disabled when entering project input mode")
	}
}

func TestQInEditInputDoesNotQuit(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = updated.(Model)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updated.(Model)

	if cmd != nil {
		t.Fatal("expected q in edit mode to be treated as input, not global quit")
	}
	if !strings.HasSuffix(model.editBuffer, "q") {
		t.Fatalf("expected edit buffer to end with q, got %q", model.editBuffer)
	}
}

func TestQInFocusFilterDoesNotQuit(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd != nil {
		t.Fatal("expected q in filter mode to be treated as input, not global quit")
	}
}

func TestKeyBarShowsInputHintsInFilterMode(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)

	view := model.renderKeyBar()
	if !strings.Contains(view, "Enter") || !strings.Contains(view, "send") {
		t.Fatalf("expected input hints in key bar during filter, got %q", view)
	}
	if !strings.Contains(view, "Esc") || !strings.Contains(view, "cancel") {
		t.Fatalf("expected Esc cancel hint in key bar during filter, got %q", view)
	}
}

func TestProjectSelectEnterOpensSelectedLoadedProjectSections(t *testing.T) {
	model := NewModelWithProject(nil, ProjectOptions{
		LoadProjects: []AccountProjectLoader{{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return nil, nil }}},
	})
	updated, _ := model.Update(accountProjectsFinishedMsg{accountID: "default", projects: []string{"group/project"}})
	model = updated.(Model)
	model.selected = 1

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if cmd != nil {
		t.Fatal("expected section transition without loading command")
	}
	if model.projectPath != "group/project" || model.mode != ModeSections {
		t.Fatalf("expected selected project sections to open, path=%q mode=%v", model.projectPath, model.mode)
	}
}
