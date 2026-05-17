package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

var errTestRefresh = errors.New("refresh failed")

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

	if model.mode != ModeDetail {
		t.Fatalf("expected MR list/detail mode, got %v", model.mode)
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

func TestProjectSelectionOpensSectionsAfterLoad(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Recents: []string{"group/one", "group/two"},
		LoadProject: func(path string) (ProjectData, error) {
			if path != "group/two" {
				t.Fatalf("expected selected recent project path, got %q", path)
			}
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

	if cmd == nil {
		t.Fatal("expected project load command")
	}
	updated, _ = model.Update(projectFinishedMsg{path: "group/two", data: ProjectData{Items: []mr.MergeRequest{{IID: 42, Title: "Loaded"}}}})
	model = updated.(Model)
	if model.projectPath != "group/two" {
		t.Fatalf("expected selected recent project, got %q", model.projectPath)
	}
	if model.mode != ModeSections {
		t.Fatalf("expected sections mode, got %v", model.mode)
	}
	if len(model.items) != 1 || model.items[0].IID != 42 {
		t.Fatalf("expected loaded items, got %+v", model.items)
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
	if !strings.Contains(view, "Esc: choose project") {
		t.Fatalf("expected recovery hint in view, got %q", view)
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

func TestManualProjectInputLoadsProject(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		LoadProject: func(path string) (ProjectData, error) {
			if path != "group/project" {
				t.Fatalf("expected manual project path, got %q", path)
			}
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

	if cmd == nil {
		t.Fatal("expected project load command")
	}
	updated, _ = model.Update(projectFinishedMsg{path: "group/project", data: ProjectData{Items: []mr.MergeRequest{{IID: 7, Title: "Manual"}}}})
	model = updated.(Model)
	if model.projectPath != "group/project" {
		t.Fatalf("expected manual project, got %q", model.projectPath)
	}
	if model.mode != ModeSections {
		t.Fatalf("expected sections mode, got %v", model.mode)
	}
	if len(model.items) != 1 || model.items[0].IID != 7 {
		t.Fatalf("expected loaded items, got %+v", model.items)
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
	if !strings.Contains(view, "r refresh") || !strings.Contains(view, "Esc: choose project") {
		t.Fatalf("expected empty state actions, got %q", view)
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

func TestMouseWheelMovesSelection(t *testing.T) {
	model := NewFakeModel()
	updated, _ := model.Update(tea.MouseMsg{X: 2, Y: 4, Button: tea.MouseButtonWheelDown})
	model = updated.(Model)

	if model.selected != 1 {
		t.Fatalf("expected wheel down to select next item, got %d", model.selected)
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
	for _, want := range []string{"[open]", "alice", "2 notes", "Please fix the naming"} {
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
