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
					ID:    "d1",
					Notes: []mr.Note{{Author: "alice", Body: "fix this"}},
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
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) { return nil, nil },
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
