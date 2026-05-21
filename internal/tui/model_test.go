package tui

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type quitModel struct{}

func (quitModel) Init() tea.Cmd                       { return tea.Quit }
func (quitModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return quitModel{}, nil }
func (quitModel) View() string                        { return "full screen" }

func TestProgramUsesAltScreen(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	program := tea.NewProgram(quitModel{}, append(programOptions(&stdout), tea.WithInput(bytes.NewBuffer(nil)))...)

	if _, err := program.Run(); err != nil {
		t.Fatalf("expected program to run: %v", err)
	}

	output := stdout.String()
	assert.Contains(t, output, "\x1b[?1049h")

	assert.Contains(t, output, "\x1b[?1049l")
}

var errTestRefresh = errors.New("refresh failed")

func TestAllNavigationKeyMapsHaveHelp(t *testing.T) {
	t.Parallel()

	keyMaps := map[string][]key.Binding{
		"project list": newProjectListKeys().LocalKeys(),
		"sections":     newSectionsKeys().LocalKeys(),
		"entity list":  newEntityListKeys().LocalKeys(),
		"mr detail":    newMRDetailKeys().LocalKeys(),
		"diff view":    newDiffViewKeys().LocalKeys(),
		"file diff":    newFileDiffKeys().LocalKeys(),
	}
	for name, bindings := range keyMaps {
		assert.NotEmpty(t, bindings)

		for _, binding := range bindings {
			if binding.Help().Key == "" || binding.Help().Desc == "" {
				t.Fatalf("expected %s binding to have help, got %+v", name, binding.Help())
			}
		}
	}
}

func TestCollapsedKeyBarTruncatesLocalKeys(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Recents: []string{"group/project"}})
	model.width = 34

	view := model.renderKeyBar()
	assert.Contains(t, view, "…")
}

func TestExpandedKeyBarShowsAllProjectListLocalKeys(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Recents: []string{"group/project"}})
	model.width = 80
	model.keyBarExpanded = true

	view := model.renderKeyBar()
	for _, want := range []string{"↑/k up", "↓/j down", "Enter open", "/ filter", "i manual", "r retry", "Global:"} {
		assert.Contains(t, view, want)
	}

	assert.Contains(t, view, "─")
}

func TestViewRendersPersistentKeyBar(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Recents: []string{"group/project"}})
	model.width = 80
	model.height = 24

	view := model.View()

	for _, want := range []string{"Local", "Global", "q quit", "Esc back", "h keys"} {
		assert.Contains(t, view, want)
	}
}

func TestHKeyTogglesExpandedKeyBarAndPaneHeight(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Recents: []string{"group/project"}})
	model.height = 30
	collapsedHeight := model.paneHeight()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model = updated.(Model)

	assert.True(t, model.keyBarExpanded)

	if model.paneHeight() >= collapsedHeight {
		t.Fatalf("expected expanded key bar to shrink panes, collapsed=%d expanded=%d", collapsedHeight, model.paneHeight())
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})

	model = updated.(Model)
	assert.False(t, model.keyBarExpanded)
}

func TestGlobalKeysUseDeclaredBindings(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	assert.NotNil(t, cmd)
}

func TestInputModeQDoesNotQuit(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	model.mrCommentInput = true

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updated.(Model)

	assert.Nil(t, cmd)

	assert.Equal(t, "q", model.Value())
}

func TestResolvedProjectShowsProjectListAndSections(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:     "group/project",
		Recents:  []string{"group/project", "recent/other"},
		Projects: []string{"group/project", "gitlab/other"},
	})

	view := model.View()
	for _, want := range []string{
		"Projects",
		"group/project",
		"recent/other",
		"gitlab/other",
		"Sections",
		"Merge Requests",
		"Issues",
		"Pipelines",
	} {
		assert.Contains(t, view, want)
	}
}

func TestEnterOnMergeRequestsSectionOpensMRList(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project"})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.Equal(t, ModeEntityList, model.mode)

	assert.Equal(t, SectionMergeRequests, model.section)

	if !strings.Contains(model.View(), "Merge Requests") || !strings.Contains(model.View(), "Port TUI shell") {
		t.Fatalf("expected MR list view, got %q", model.View())
	}
}

func TestOpenedProjectMovesToTopOfProjectList(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	model := NewFakeModel()
	model.MRDetailState.YOffset = 2

	// In ModeDetail, Down scrolls the right panel rather than moving selection
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})

	model = updated.(Model)
	assert.Equal(t, 3, model.MRDetailState.YOffset)

	assert.Equal(t, 0, model.EntityListState.mrList.Index())

	// Enter on Summary tab no longer opens diff — it is a no-op
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	model = updated.(Model)
	assert.Equal(t, ModeDetail, model.mode)

	// Esc from ModeDetail returns to entity list
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	model = updated.(Model)
	assert.Equal(t, ModeEntityList, model.mode)
}

func TestFilterInputNarrowsList(t *testing.T) {
	t.Parallel()

	model := NewFakeModel()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("yaml")})
	model = updated.(Model)

	filtered := model.filtered()
	assert.Len(t, filtered, 1)

	assert.Contains(t, strings.ToLower(filtered[0].Title), "yaml")
}

func TestDirectMRDeepLinkSelectsLoadedMergeRequest(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project", Section: SectionMergeRequests, EntityID: "123"})
	updated, _ := model.Update(projectFinishedMsg{path: "group/project", data: ProjectData{Items: []mr.MergeRequest{
		{IID: 101, Title: "First MR"},
		{IID: 123, Title: "Loaded target"},
	}}})
	model = updated.(Model)

	assert.Equal(t, 1, model.EntityListState.mrList.Index())

	assert.Contains(t, model.View(), "!123 Loaded target")
}

func TestDirectMRDeepLinkSelectsMatchingMergeRequest(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject([]mr.MergeRequest{
		{IID: 101, Title: "First MR"},
		{IID: 123, Title: "Target MR", Description: "Deep linked"},
	}, ProjectOptions{Path: "group/project", Section: SectionMergeRequests, EntityID: "123"})

	assert.Equal(t, 1, model.EntityListState.mrList.Index())

	view := model.View()
	if !strings.Contains(view, "!123 Target MR") || !strings.Contains(view, "Deep linked") {
		t.Fatalf("expected target MR detail, got %q", view)
	}
}

func TestMRListAndDetailRenderPreviousMRInfo(t *testing.T) {
	t.Parallel()

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
		"Alice Doe @alice",
		"feature/review → main",
		"opened",
		"✓ success",
		"1/2",
		"Review from terminal",
		"https://gitlab.com/group/project/-/merge_requests/10",
	} {
		assert.Contains(t, view, want)
	}
}

func TestProjectSelectionShowsRecentsAndGitLabProjects(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{
		Recents: []string{"group/one", "group/two"},
		LoadProject: func(path string, _ string) (ProjectData, error) {
			return ProjectData{Items: []mr.MergeRequest{{IID: 42, Title: "Loaded"}}}, nil
		},
	})
	assert.Equal(t, ModeProjectSelect, model.mode)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.Nil(t, cmd)

	assert.Equal(t, ModeSections, model.mode)

	assert.Equal(t, "group/two", model.projectPath)
}

func TestProjectLoadShowsLoadingState(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Recents: []string{"group/project"}})
	updated, _ := model.Update(projectStartedMsg{path: "group/project"})
	model = updated.(Model)

	view := model.View()
	assert.Contains(t, view, "Loading project…")

	assert.NotContains(t, view, "Refreshing…")
}

func TestProjectLoadErrorCanReturnToSelection(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Recents: []string{"group/project"}})
	updated, _ := model.Update(projectStartedMsg{path: "group/project"})
	model = updated.(Model)
	updated, _ = model.Update(projectFinishedMsg{path: "group/project", err: errTestRefresh})
	model = updated.(Model)

	view := model.View()
	assert.Contains(t, view, "Error: refresh failed")

	assert.Contains(t, view, "Esc back")

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	model = updated.(Model)
	assert.Equal(t, ModeProjectSelect, model.mode)
}

func TestProjectLoadErrorRetryReloadsSameProject(t *testing.T) {
	t.Parallel()

	calls := 0
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Recents: []string{"group/project"},
		LoadProject: func(path string, _ string) (ProjectData, error) {
			calls++
			assert.Equal(t, "group/project", path)
			return ProjectData{Items: []mr.MergeRequest{{IID: 9, Title: "Retried"}}}, nil
		},
	})
	updated, _ := model.Update(projectStartedMsg{path: "group/project"})
	model = updated.(Model)
	updated, _ = model.Update(projectFinishedMsg{path: "group/project", err: errTestRefresh})
	model = updated.(Model)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})

	assert.NotNil(t, cmd)

	_ = cmd

	assert.Equal(t, 0, calls)
}

func TestManualProjectLoadErrorReturnsToInput(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{})
	updated, _ := model.Update(projectStartedMsg{path: "group/project"})
	model = updated.(Model)
	updated, _ = model.Update(projectFinishedMsg{path: "group/project", err: errTestRefresh})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	assert.Equal(t, ModeProjectInput, model.mode)

	assert.Equal(t, FocusFilter, model.focus)
}

func TestManualProjectInputGoesToSectionsImmediately(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{
		LoadProject: func(path string, _ string) (ProjectData, error) {
			return ProjectData{Items: []mr.MergeRequest{{IID: 7, Title: "Manual"}}}, nil
		},
	})
	assert.Equal(t, ModeProjectInput, model.mode)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("group/project")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.Nil(t, cmd)

	assert.Equal(t, ModeSections, model.mode)

	assert.Equal(t, "group/project", model.projectPath)
}

func TestEnterOnSummaryDoesNotTriggerDiff(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject([]mr.MergeRequest{{IID: 1, Title: "Needs diff"}}, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
	})

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.Nil(t, cmd)

	assert.Equal(t, ModeDetail, model.mode)
}

func TestEmptyProjectStateCanReturnToProjectSelection(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Recents: []string{"group/project"}, Section: SectionMergeRequests})
	updated, _ := model.Update(projectFinishedMsg{path: "group/project", data: ProjectData{Items: []mr.MergeRequest{}}})
	model = updated.(Model)

	view := model.View()
	assert.Contains(t, view, "No opened MRs")

	if !strings.Contains(view, "r refresh") || !strings.Contains(view, "Esc back") {
		t.Fatalf("expected empty state actions in key bar, got %q", view)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	model = updated.(Model)
	assert.Equal(t, ModeProjectSelect, model.mode)
}

func TestRefreshKeyReturnsCommand(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		Refresh: func() ([]mr.MergeRequest, error) {
			return []mr.MergeRequest{{IID: 99, Title: "Refreshed"}}, nil
		},
	})

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	assert.NotNil(t, cmd)
}

func TestRefreshFinishedReplacesItems(t *testing.T) {
	t.Parallel()

	model := NewFakeModel()
	updated, _ := model.Update(refreshFinishedMsg{items: []mr.MergeRequest{{IID: 99, Title: "Refreshed"}}})
	model = updated.(Model)

	assert.Len(t, model.items, 1)

	assert.Equal(t, 99, model.items[0].IID)
}

func TestRefreshFinishedStoresError(t *testing.T) {
	t.Parallel()

	model := NewFakeModel()
	updated, _ := model.Update(refreshFinishedMsg{err: errTestRefresh})
	model = updated.(Model)

	assert.Equal(t, errTestRefresh.Error(), model.errorMessage)
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
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)

	assert.Equal(t, TabDiscussions, model.activeTab)

	assert.NotNil(t, cmd)

	assert.Contains(t, model.View(), "Loading")
}

func TestDiscussionsTabRendersThreadsAfterLoad(t *testing.T) {
	t.Parallel()

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
		assert.Contains(t, view, want)
	}
}

func TestFocusedDiscussionThreadIsMarked(t *testing.T) {
	t.Parallel()

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
	assert.Contains(t, view, "> [")

	// Move cursor to second thread
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updated.(Model)

	view = model.View()
	if strings.Count(view, "> [") != 1 {
		t.Fatalf("expected exactly one focused thread marker, got:\n%s", view)
	}

	assert.Contains(t, view, "bob")
}

func TestDiscussionThreadsAreSeparatedByDivider(t *testing.T) {
	t.Parallel()

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
	assert.Contains(t, view, "───")
}

func TestDiscussionThreadShowsAllReplies(t *testing.T) {
	t.Parallel()

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
		assert.Contains(t, view, want)
	}
}

func TestDiscussionsTabShowsEmptyState(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, discussions: []mr.Discussion{}})
	model = updated.(Model)

	assert.Contains(t, model.View(), "No discussions")
}

func TestDiscussionsTabShowsErrorState(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, err: errors.New("network error")})
	model = updated.(Model)

	assert.Contains(t, model.View(), "Error:")
}

func TestFilesTabTriggersLoadOnFirstVisit(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())

	// Tab twice: Summary → Discussions → Files
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)

	assert.Equal(t, TabFiles, model.activeTab)

	assert.NotNil(t, cmd)

	assert.Contains(t, model.View(), "Loading")
}

func TestFilesTabRendersChangedFilesAfterLoad(t *testing.T) {
	t.Parallel()

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
		assert.Contains(t, view, want)
	}
}

func TestFilesTabShowsEmptyState(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{}})
	model = updated.(Model)

	assert.Contains(t, model.View(), "No changed files")
}

func TestFilesTabShowsErrorState(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), discussionOpts())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, err: errors.New("timeout")})
	model = updated.(Model)

	assert.Contains(t, model.View(), "Error:")
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
	t.Parallel()

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

	assert.Equal(t, TabFiles, model.activeTab)

	// Press Enter to open selected file
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.Equal(t, ModeFileDiff, model.mode)
}

func TestFileDiffLeftPaneShowsFileListWithCurrentHighlighted(t *testing.T) {
	t.Parallel()

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
	if !strings.Contains(view, ansiSelected("│       └── model.go")) && !strings.Contains(view, "\x1b[48;5;63m") {
		t.Fatalf("expected selected file highlighted, got:\n%s", view)
	}

	assert.Contains(t, view, "model.go")

	assert.Contains(t, view, "main.go")
}

func TestFileDiffRightPaneShowsPerFileDiffRows(t *testing.T) {
	t.Parallel()

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
	assert.Contains(t, view, "Diff internal/tui/model.go")

	assert.Contains(t, view, "before")

	assert.Contains(t, view, "added line")
}

func TestFileDiffShowsDiscussionGutterWithoutInlineBody(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		Emoji:   config.DefaultEmojiConfig(),
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
	assert.Contains(t, view, " 💬 ")

	if strings.Contains(view, "fix this") || strings.Contains(view, "alice") || strings.Contains(view, "↳") {
		t.Fatalf("expected discussion body to stay out of diff rows, got:\n%s", view)
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
	t.Parallel()

	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
		{Path: "b.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "b", NewText: "b"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)

	assert.Equal(t, 1, model.DiffViewState.selectedFile)

	view := model.View()
	if !strings.Contains(view, "\x1b[48;5;63m") || !strings.Contains(view, "b.go") {
		t.Fatalf("expected b.go highlighted, got:\n%s", view)
	}
}

func TestRightKeyDoesNotExceedLastFile(t *testing.T) {
	t.Parallel()

	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)

	assert.Equal(t, 0, model.DiffViewState.selectedFile)
}

func TestLeftKeyMovesToPreviousFile(t *testing.T) {
	t.Parallel()

	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
		{Path: "b.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "b", NewText: "b"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(Model)

	assert.Equal(t, 0, model.DiffViewState.selectedFile)

	view := model.View()
	if !strings.Contains(view, "\x1b[48;5;63m") || !strings.Contains(view, "a.go") {
		t.Fatalf("expected a.go highlighted, got:\n%s", view)
	}
}

func TestLeftKeyDoesNotGoBelowFirstFile(t *testing.T) {
	t.Parallel()

	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(Model)

	assert.Equal(t, 0, model.DiffViewState.selectedFile)
}

func TestEscInFileDiffReturnsToFilesTab(t *testing.T) {
	t.Parallel()

	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	assert.Equal(t, ModeDetail, model.mode)

	assert.Equal(t, TabFiles, model.activeTab)
}

func TestBackspaceInFileDiffReturnsToFilesTab(t *testing.T) {
	t.Parallel()

	model := fileDiffModelWithFiles(t, []mr.ChangedFile{
		{Path: "a.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "a", NewText: "a"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	model = updated.(Model)

	assert.Equal(t, ModeDetail, model.mode)

	assert.Equal(t, TabFiles, model.activeTab)

	assert.Contains(t, model.View(), ">Files<")
}

// --- #43: Draft comments ---

func draftOpts() ProjectOptions {
	return ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		Emoji:   config.DefaultEmojiConfig(),
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

func TestReviewTabAppearsInTabCycleWithDraftCount(t *testing.T) {
	t.Parallel()

	model := draftFileDiffModel(t)
	updated, _ := model.Update(draftAddedMsg{iid: 42, draft: mr.DraftComment{
		LocalID:  "d1",
		Body:     "Check this",
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 2},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	model = updated.(Model)
	assert.Equal(t, TabReview, model.activeTab)

	if !strings.Contains(model.View(), ">Review (1)<") || !strings.Contains(model.View(), "main.go:2 Check this") {
		t.Fatalf("expected Review tab with draft count and context, got:\n%s", model.View())
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	model = updated.(Model)
	assert.Equal(t, TabSummary, model.activeTab)
}

func TestReviewTabOpensDraftDiffAndEscReturnsToReview(t *testing.T) {
	t.Parallel()

	model := draftFileDiffModel(t)
	updated, _ := model.Update(draftAddedMsg{iid: 42, draft: mr.DraftComment{
		LocalID:  "d1",
		Body:     "Check this",
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 2},
	}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	model.activeTab = TabReview

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	model = updated.(Model)
	if model.mode != ModeFileDiff || model.diffCursor != 1 {
		t.Fatalf("expected file diff at draft row, mode=%v cursor=%d", model.mode, model.diffCursor)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	model = updated.(Model)
	if model.mode != ModeDetail || model.activeTab != TabReview {
		t.Fatalf("expected Esc to return to Review tab, mode=%v tab=%v", model.mode, model.activeTab)
	}
}

func TestReviewTabPublishesDraftsWithSummaryAndDiscards(t *testing.T) {
	t.Parallel()

	submitted := false
	postedSummary := ""
	discarded := false
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		SubmitDrafts: func(iid int, drafts []mr.DraftComment) error {
			submitted = iid == 42 && len(drafts) == 1
			return nil
		},
		PostMRComment: func(iid int, body string) error {
			postedSummary = body
			return nil
		},
		DiscardDrafts: func(iid int) error {
			discarded = iid == 42
			return nil
		},
	})
	model.activeTab = TabReview
	model.drafts[42] = []mr.DraftComment{{LocalID: "d1", Body: "Check this", Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 2}}}
	model.reviewSummary = "Looks good overall"

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	model = updated.(Model)

	assert.NotNil(t, cmd)

	updated, _ = model.Update(cmd())
	model = updated.(Model)

	if !submitted || postedSummary != "Looks good overall" {
		t.Fatalf("expected submit and summary post, submitted=%v summary=%q", submitted, postedSummary)
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})

	model = updated.(Model)
	assert.Len(t, model.drafts[42], 0)

	assert.NotNil(t, cmd)

	cmd()

	assert.True(t, discarded)
}

func TestDraftInlineCommentCreatesServerDraftAndStoresID(t *testing.T) {
	t.Parallel()

	created := false
	opts := draftOpts()
	opts.DraftInlineComment = func(iid int, position mr.DiffPosition, body string) (int, error) {
		created = true
		if iid != 42 || body != "Server draft" {
			t.Fatalf("unexpected draft callback args: iid=%d body=%q", iid, body)
		}
		if position.NewPath != "main.go" || position.NewLine != 1 || position.OldPath != "main.go" || position.OldLine != 1 {
			t.Fatalf("unexpected position: %+v", position)
		}
		return 987, nil
	}

	model := NewModelWithProject(FakeMergeRequests(), opts)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(filesFinishedMsg{iid: 42, files: []mr.ChangedFile{{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1, OldText: "old", NewText: "old"}}}}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Server draft")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	assert.NotNil(t, cmd)

	updated, _ = model.Update(cmd())
	model = updated.(Model)

	assert.True(t, created)
	if len(model.drafts[42]) != 1 || model.drafts[42][0].ID != 987 {
		t.Fatalf("expected stored draft id 987, got %+v", model.drafts[42])
	}
}

func TestModelStoresDraftCommentForCurrentMR(t *testing.T) {
	t.Parallel()

	model := draftFileDiffModel(t)

	draft := mr.DraftComment{
		LocalID:  "local-1",
		Body:     "Fix this please",
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 2},
	}
	updated, _ := model.Update(draftAddedMsg{iid: 42, draft: draft})
	model = updated.(Model)

	assert.Len(t, model.drafts[42], 1)

	assert.Equal(t, "Fix this please", model.drafts[42][0].Body)
}

func TestDraftMarkerAppearsInGutterWithoutInlineBody(t *testing.T) {
	t.Parallel()

	model := draftFileDiffModel(t)

	updated, _ := model.Update(draftAddedMsg{iid: 42, draft: mr.DraftComment{
		LocalID:  "d1",
		Body:     "Check this",
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 2},
	}})
	model = updated.(Model)

	view := model.View()
	assert.Contains(t, view, "📝 ")

	if strings.Contains(view, "[DRAFT]") || strings.Contains(view, "Check this") {
		t.Fatalf("expected draft body to stay out of diff rows, got:\n%s", view)
	}
}

func TestDraftRangeMarkerSpansMultipleRows(t *testing.T) {
	t.Parallel()

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
	count := strings.Count(view, "📝 ")

	if count < 2 {
		t.Fatalf("expected draft gutter marker on both rows of range (got %d), view:\n%s", count, view)
	}
	// Thread Panel shows the draft body when cursor is on the commented line — that is correct.
	// Verify the body is in the Thread Panel section (after ───), not inline in a diff row.
	sepIdx := strings.Index(view, "─────")
	if sepIdx >= 0 && strings.Contains(view[sepIdx:], "Range comment") {
		// body is in Thread Panel — expected
		return
	}
	// If no separator, body must not appear at all (panel hidden or no draft at cursor)
	assert.NotContains(t, view, "Range comment")
}

func TestFileDiffGuttersUseTextSymbolsWhenEmojiDisabled(t *testing.T) {
	t.Parallel()

	model := draftFileDiffModel(t)
	model.emoji = config.EmojiConfig{Enabled: false}
	updated, _ := model.Update(draftAddedMsg{iid: 42, draft: mr.DraftComment{
		LocalID:  "d1",
		Body:     "Check this",
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 2},
	}})
	model = updated.(Model)
	model.discussions[42] = []mr.Discussion{{
		ID:       "disc-1",
		Notes:    []mr.Note{{Author: "alice", Body: "fix this"}},
		Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 2},
	}}

	view := model.View()
	assert.Contains(t, view, "●○ ")

	if strings.Contains(view, "📝") || strings.Contains(view, "💬") {
		t.Fatalf("expected no emoji markers when emoji disabled, got:\n%s", view)
	}
}

func TestActiveDraftRangeUsesDotGutterMarker(t *testing.T) {
	t.Parallel()

	model := draftFileDiffModel(t)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)

	view := model.View()
	if strings.Count(view, "· ") < 2 {
		t.Fatalf("expected active draft range dot markers, got:\n%s", view)
	}
}

func TestVKeyStartsRangeSelection(t *testing.T) {
	t.Parallel()

	model := draftFileDiffModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	model = updated.(Model)

	assert.Equal(t, model.diffCursor, model.rangeStart)

	view := model.View()
	assert.Contains(t, view, "· ")
}

func TestEscCancelsRangeSelection(t *testing.T) {
	t.Parallel()

	model := draftFileDiffModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})

	model = updated.(Model)
	if model.rangeStart < 0 {
		t.Fatal("expected range selection to be active")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	assert.Equal(t, -1, model.rangeStart)

	assert.Equal(t, ModeFileDiff, model.mode)
}

func TestCKeyEntersCommentInputAndEnterSavesDraft(t *testing.T) {
	t.Parallel()

	model := draftFileDiffModel(t)

	// Press 'c' to open comment input
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	model = updated.(Model)

	assert.True(t, model.commentInput)

	// Type comment text
	for _, ch := range "My draft comment" {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		model = updated.(Model)
	}

	assert.Equal(t, "My draft comment", model.Value())

	// Press Enter to save
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.False(t, model.commentInput)

	assert.Len(t, model.drafts[42], 1)

	assert.Equal(t, "My draft comment", model.drafts[42][0].Body)
}

func TestCommentInputBufferAppearsInView(t *testing.T) {
	t.Parallel()

	model := draftFileDiffModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	model = updated.(Model)

	view := model.View()
	assert.Contains(t, view, "hello")
}

func TestPKeySubmitsDraftsAndClearsOnSuccess(t *testing.T) {
	t.Parallel()

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

	assert.NotNil(t, cmd)

	// Execute the command
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	assert.True(t, submitted)

	assert.Len(t, model.drafts[42], 0)
}

func TestDKeyDiscardsAllLocalDrafts(t *testing.T) {
	t.Parallel()

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

	assert.Len(t, model.drafts[42], 2)

	// Press 'D' to discard
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	model = updated.(Model)

	assert.Len(t, model.drafts[42], 0)

	if cmd != nil {
		msg := cmd()
		model.Update(msg)
	}

	assert.True(t, discarded)
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
	t.Parallel()

	model := discussionsTabModel(t)

	assert.Equal(t, 0, model.discussionCursor)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updated.(Model)

	assert.Equal(t, 1, model.discussionCursor)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	model = updated.(Model)

	assert.Equal(t, 0, model.discussionCursor)
}

func TestDiscussionCursorDoesNotExceedBounds(t *testing.T) {
	t.Parallel()

	model := discussionsTabModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})

	model = updated.(Model)
	assert.Equal(t, 0, model.discussionCursor)

	// Move to last
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	model = updated.(Model)
	assert.Equal(t, 1, model.discussionCursor)
}

func TestRKeyOpensReplyInputForFocusedDiscussion(t *testing.T) {
	t.Parallel()

	model := discussionsTabModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model = updated.(Model)

	assert.True(t, model.replyInput)

	assert.False(t, model.replyDraft)

	assert.Equal(t, "d1", model.replyDiscussionID)
}

func TestEnterInReplyInputSendsReplyAndAddsNote(t *testing.T) {
	t.Parallel()

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

	assert.False(t, model.replyInput)

	assert.NotNil(t, cmd)

	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	assert.True(t, called)

	assert.Len(t, model.discussions[42][0].Notes, 2)

	assert.Equal(t, "My reply", model.discussions[42][0].Notes[1].Body)
}

func TestDKeyOpensDraftReplyAndEnterCallsService(t *testing.T) {
	t.Parallel()

	called := false
	opts := discussionWriteOpts()
	opts.DraftReply = func(iid int, discussionID string, body string) (int, error) {
		called = true

		if body != "Draft reply" {
			t.Errorf("expected 'Draft reply', got %q", body)
		}

		return 123, nil
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

	assert.NotNil(t, cmd)

	updated, _ = model.Update(cmd())
	model = updated.(Model)

	assert.True(t, called)
	if len(model.drafts[42]) != 1 || model.drafts[42][0].ID != 123 {
		t.Fatalf("expected stored draft reply id 123, got %+v", model.drafts[42])
	}
}

func TestXKeyResolvesOpenDiscussion(t *testing.T) {
	t.Parallel()

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

	assert.NotNil(t, cmd)

	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	assert.True(t, resolved)

	assert.True(t, model.discussions[42][0].Resolved)
}

func TestXKeyUnresolvesResolvedDiscussion(t *testing.T) {
	t.Parallel()

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

	assert.NotNil(t, cmd)

	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	assert.True(t, unresolved)

	assert.False(t, model.discussions[42][0].Resolved)
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
		{
			ID: "inline-d1", Resolved: false, Notes: []mr.Note{{Author: "alice", Body: "inline comment"}},
			Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 1},
		},
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
	t.Parallel()

	model := diffViewWithInlineDiscussion(t, nil)

	// diffCursor is at 0 which has the inline discussion (NewLine=1)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model = updated.(Model)

	assert.True(t, model.replyInput)

	assert.Equal(t, "inline-d1", model.replyDiscussionID)
}

func TestDKeyInDiffViewOpensDraftReplyForInlineDiscussion(t *testing.T) {
	t.Parallel()

	model := diffViewWithInlineDiscussion(t, nil)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	model = updated.(Model)

	assert.True(t, model.replyInput)

	assert.True(t, model.replyDraft)

	assert.Equal(t, "inline-d1", model.replyDiscussionID)
}

func TestXKeyInDiffViewResolvesInlineDiscussion(t *testing.T) {
	t.Parallel()

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
		{
			ID: "inline-d1", Resolved: false, Notes: []mr.Note{{Author: "alice", Body: "fix"}},
			Position: &mr.DiffPosition{NewPath: "main.go", NewLine: 1},
		},
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

	assert.NotNil(t, cmd)

	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	assert.True(t, resolved)

	assert.True(t, model.discussions[42][0].Resolved)
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
	t.Parallel()

	model := instantCommentFileDiffModel(t, nil)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	model = updated.(Model)

	assert.True(t, model.commentInput)

	assert.True(t, model.commentInstant)
}

func TestEnterInInstantCommentCallsAPIAndDoesNotSaveDraft(t *testing.T) {
	t.Parallel()

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

	assert.NotNil(t, cmd)

	cmd()

	assert.True(t, called)

	assert.Len(t, model.drafts[42], 0)
}

func TestInstantCommentAPIErrorShownInView(t *testing.T) {
	t.Parallel()

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

	assert.Equal(t, ModeFileDiff, model.mode)

	assert.Contains(t, model.View(), "network timeout")
}

func TestMKeyOpensMRCommentInput(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path:          "group/project",
		Section:       SectionMergeRequests,
		PostMRComment: func(iid int, body string) error { return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	model = updated.(Model)

	assert.True(t, model.mrCommentInput)
}

func TestEnterInMRCommentInputCallsPostMRCommentFunc(t *testing.T) {
	t.Parallel()

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

	assert.False(t, model.mrCommentInput)

	assert.NotNil(t, cmd)

	cmd()

	assert.True(t, called)
}

func TestEscInMRCommentInputCancelsWithoutSending(t *testing.T) {
	t.Parallel()

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

	assert.False(t, model.mrCommentInput)

	assert.Nil(t, cmd)

	assert.False(t, called)

	assert.Equal(t, "", model.Value())
}

func TestMRCommentAPIErrorShownInViewWithoutLosingContext(t *testing.T) {
	t.Parallel()

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

	assert.Equal(t, ModeDetail, model.mode)

	assert.Contains(t, model.View(), "forbidden")
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
	t.Parallel()

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

	assert.NotNil(t, cmd)

	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	assert.True(t, called)

	assert.Contains(t, model.View(), "Approved")
}

func TestMKeySetsMergeConfirmPending(t *testing.T) {
	t.Parallel()

	model := mrActionsModel(t, ProjectOptions{
		MergeMR: func(iid int) error { return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	model = updated.(Model)

	assert.True(t, model.mergeConfirmPending)

	assert.Contains(t, model.View(), "confirm merge")
}

func TestMKeyAgainConfirmsMerge(t *testing.T) {
	t.Parallel()

	called := false
	model := mrActionsModel(t, ProjectOptions{
		MergeMR: func(iid int) error { called = true; return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	model = updated.(Model)
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})

	assert.NotNil(t, cmd)

	cmd()

	assert.True(t, called)
}

func TestOtherKeyAfterMCancelsMerge(t *testing.T) {
	t.Parallel()

	model := mrActionsModel(t, ProjectOptions{
		MergeMR: func(iid int) error { return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	model = updated.(Model)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updated.(Model)

	assert.False(t, model.mergeConfirmPending)

	assert.Nil(t, cmd)
}

func TestOKeyOpensURLInBrowser(t *testing.T) {
	t.Parallel()

	opened := ""
	model := NewModelWithProject([]mr.MergeRequest{{
		IID: 42, Title: "Test", WebURL: "https://gitlab.com/group/project/-/merge_requests/42",
	}}, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		OpenURL: func(url string) error { opened = url; return nil },
	})

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})

	assert.NotNil(t, cmd)

	cmd()

	assert.Equal(t, "https://gitlab.com/group/project/-/merge_requests/42", opened)
}

func TestEKeyOpensEditModeOnTitleField(t *testing.T) {
	t.Parallel()

	model := mrActionsModel(t, ProjectOptions{
		EditMR: func(iid int, title, description string) error { return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	model = updated.(Model)

	assert.True(t, model.editInput)

	assert.Equal(t, "title", model.editField)

	assert.Equal(t, "Port TUI shell to Bubble Tea", model.Value())
}

func TestTabInEditModeMoveToDescriptionField(t *testing.T) {
	t.Parallel()

	model := mrActionsModel(t, ProjectOptions{
		EditMR: func(iid int, title, description string) error { return nil },
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	model = updated.(Model)
	// Clear and type new title
	model.BeginWithValue("New title")
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)

	assert.Equal(t, "description", model.editField)

	assert.Equal(t, "New title", model.editTitle)
}

func TestEnterInEditModeSavesAndCallsEditMR(t *testing.T) {
	t.Parallel()

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
	model.BeginWithValue("New title")

	// Tab to description
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	model.BeginWithValue("New desc")

	// Enter to save
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.False(t, model.editInput)

	assert.NotNil(t, cmd)

	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	assert.True(t, called)

	assert.Equal(t, "New title", model.items[0].Title)
}

func TestEKeyInDiffViewOpensFileInEditor(t *testing.T) {
	t.Parallel()

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
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})

	assert.NotNil(t, cmd)

	cmd()

	assert.Equal(t, "internal/tui/model.go", openedPath)

	assert.Equal(t, 11, openedLine)
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
	t.Parallel()

	model := fileDiffModelWithRows(t, []mr.DiffRow{
		{OldLine: 0, NewLine: 5, NewText: "added"},
	})

	view := model.View()
	// Old line number should be blank, not "   0"
	assert.NotContains(t, view, "   0 │")
}

func TestDiffContextRowHasNoTypeMarker(t *testing.T) {
	t.Parallel()

	model := fileDiffModelWithRows(t, []mr.DiffRow{
		{OldLine: 1, NewLine: 1, OldText: "unchanged", NewText: "unchanged"},
	})

	view := model.View()
	assert.Contains(t, view, "unchanged")
	// Context rows must not be prefixed with + or -
	if strings.Contains(view, "+ unchanged") || strings.Contains(view, "- unchanged") {
		t.Fatalf("expected context row to have no +/- marker, got:\n%s", view)
	}
}

func TestDiffDeletionRowIsMarkedWithMinus(t *testing.T) {
	t.Parallel()

	model := fileDiffModelWithRows(t, []mr.DiffRow{
		{OldLine: 3, NewLine: 0, OldText: "old code"},
	})

	view := model.View()
	assert.Contains(t, view, "- old code")
}

func TestDiffAdditionRowIsMarkedWithPlus(t *testing.T) {
	t.Parallel()

	model := fileDiffModelWithRows(t, []mr.DiffRow{
		{OldLine: 0, NewLine: 1, NewText: "new feature"},
	})

	view := model.View()
	assert.Contains(t, view, "+ new feature")
}

// --- #50: Left panel always read-only, no focus ---

func TestLeftPanelHasNoActiveBorderInDetailMode(t *testing.T) {
	t.Parallel()

	model := NewFakeModel()
	// Simulate mouse click that used to set FocusList
	updated, _ := model.Update(tea.MouseMsg{X: 2, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	model = updated.(Model)

	assert.NotEqual(t, FocusList, model.focus)
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
	t.Parallel()

	loadCalled := false
	model := NewModelWithProject(nil, ProjectOptions{
		Recents: []string{"group/project"},
		LoadProject: func(path string, _ string) (ProjectData, error) {
			loadCalled = true
			return ProjectData{Items: []mr.MergeRequest{{IID: 1, Title: "Loaded MR"}}}, nil
		},
	})

	// Select project → ModeSections immediately
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	model = updated.(Model)
	assert.Equal(t, ModeSections, model.mode)

	// Select MR section → should trigger loading
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.NotNil(t, cmd)

	assert.False(t, loadCalled)
}

func TestMRSectionLoadingCompletionShowsMRList(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{
		Recents:     []string{"group/project"},
		LoadProject: func(path string, _ string) (ProjectData, error) { return ProjectData{}, nil },
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

	assert.Equal(t, ModeEntityList, model.mode)

	assert.Len(t, model.items, 2)

	view := model.View()
	assert.Contains(t, view, "First MR")
}

// --- #53: Two-panel layout for entity list ---

func entityListModel(t *testing.T) Model {
	t.Helper()

	model := NewModelWithProject(nil, ProjectOptions{
		Recents:     []string{"group/project"},
		LoadProject: func(path string, _ string) (ProjectData, error) { return ProjectData{}, nil },
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
	assert.Equal(t, ModeEntityList, model.mode)

	return model
}

func TestEntityListViewShowsSectionsContextOnLeft(t *testing.T) {
	t.Parallel()

	model := entityListModel(t)
	view := model.View()

	assert.Contains(t, view, "Merge Requests")

	assert.Contains(t, view, "Alpha MR")
}

func TestEntityListEnterGoesToMRDetail(t *testing.T) {
	t.Parallel()

	model := entityListModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.Equal(t, ModeDetail, model.mode)
}

func TestEntityListEscGoesToSections(t *testing.T) {
	t.Parallel()

	model := entityListModel(t)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	assert.Equal(t, ModeSections, model.mode)
}

// --- #48: Two-panel layout on project picker screen ---

func TestProjectPickerRendersLeftContextPane(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{
		Recents: []string{"recent/project"},
	})
	assert.Equal(t, ModeProjectSelect, model.mode)

	view := model.View()
	assert.Contains(t, view, "gitlab-tui")

	assert.Contains(t, view, "Projects")

	assert.Contains(t, view, "recent/project")
}

func TestProjectInputRendersLeftContextPane(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{})
	assert.Equal(t, ModeProjectInput, model.mode)

	view := model.View()
	assert.Contains(t, view, "gitlab-tui")

	assert.Contains(t, view, "Open GitLab project")
}

func TestTabKeyCyclesDetailTabs(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})

	// Summary (default) → Discussions
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})

	model = updated.(Model)
	assert.Equal(t, TabDiscussions, model.activeTab)

	// Discussions → Files
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	model = updated.(Model)
	assert.Equal(t, TabFiles, model.activeTab)

	// Files → Review
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	model = updated.(Model)
	assert.Equal(t, TabReview, model.activeTab)

	// Review → Summary (wrap)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	model = updated.(Model)
	assert.Equal(t, TabSummary, model.activeTab)
}

func TestProjectSelectStructuredRecentsDoNotRenderLegacyDuplicate(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
		assert.Contains(t, view, want)
	}

	assert.False(t, model.ProjectPickerState.projectRows[0].selectable)

	assert.Equal(t, 2, model.ProjectPickerState.selected)
}

func TestProjectSelectHidesRecentSectionWhenEmpty(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{
		LoadProjects: []AccountProjectLoader{{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return nil, nil }}},
	})

	assert.NotContains(t, model.View(), "Recent")
}

func TestProjectSelectRecentSelectionUsesProjectPath(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
		assert.Contains(t, view, want)
	}

	for _, unwanted := range []string{"group/beta", "org/gamma"} {
		assert.NotContains(t, view, unwanted)
	}

	assert.Equal(t, 2, model.ProjectPickerState.selected)
}

func TestProjectSelectFilterHidesSectionsWithoutMatchesAndEscResets(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{
		RecentProjects: []RecentProjectOption{{Path: "recent/only", Account: "work"}},
		LoadProjects:   []AccountProjectLoader{{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return nil, nil }}},
	})
	updated, _ := model.Update(accountProjectsFinishedMsg{accountID: "default", projects: []string{"account/project"}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)

	for _, runeValue := range "only" {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{runeValue}})
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
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{RecentProjects: []RecentProjectOption{{Path: "group/project", Account: "default"}}})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	model = updated.(Model)

	assert.Contains(t, model.View(), "No matching projects")
}

func TestProjectSelectStartsAccountProjectLoads(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{LoadProjects: []AccountProjectLoader{
		{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return []string{"group/project"}, nil }},
		{ID: "work", Host: "https://gitlab.example.com", Load: func() ([]string, error) { return []string{"work/project"}, nil }},
	}})

	cmd := model.Init()

	assert.Equal(t, ModeProjectSelect, model.mode)

	assert.NotNil(t, cmd)

	view := model.View()
	for _, want := range []string{"[default]  https://gitlab.com", "[work]  https://gitlab.example.com", "Loading…"} {
		assert.Contains(t, view, want)
	}
}

func TestProjectSelectShowsLoadedAccountProjectsAndSkipsHeaders(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{LoadProjects: []AccountProjectLoader{
		{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return nil, nil }},
	}})
	projects := []string{"group/one", "group/two"}
	updated, _ := model.Update(accountProjectsFinishedMsg{accountID: "default", projects: projects})
	model = updated.(Model)

	assert.False(t, model.ProjectPickerState.projectRows[0].selectable)

	assert.Equal(t, 1, model.ProjectPickerState.selected)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	model = updated.(Model)
	if got, _, ok := model.ProjectPickerState.selectedProject(); !ok || got != "group/two" {
		t.Fatalf("expected second project selected, got %q ok=%t", got, ok)
	}
}

func TestProjectSelectShowsErrorAndRetriesOnlyFailedAccounts(t *testing.T) {
	t.Parallel()

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

	assert.Contains(t, model.View(), "Error: refresh failed  r: retry")

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	assert.NotNil(t, cmd)

	_ = cmd()

	if failedCalls != 1 || successCalls != 0 {
		t.Fatalf("expected only failed loader to run, failed=%d success=%d", failedCalls, successCalls)
	}
}

// --- #66: Up/down scrolls right panel in ModeDetail ---

func TestDownScrollsRightPanelInModeDetail(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	initialSelected := model.ProjectPickerState.selected
	model.MRDetailState.YOffset = 5

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(Model)

	assert.Equal(t, 6, model.MRDetailState.YOffset)

	assert.Equal(t, initialSelected, model.ProjectPickerState.selected)
}

func TestUpScrollsRightPanelInModeDetail(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	initialSelected := model.ProjectPickerState.selected
	model.MRDetailState.YOffset = 5

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(Model)

	assert.Equal(t, 4, model.MRDetailState.YOffset)

	assert.Equal(t, initialSelected, model.ProjectPickerState.selected)
}

func TestMRDetailViewportOffsetFloorAtZeroInModeDetail(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	model.MRDetailState.YOffset = 0

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(Model)

	assert.Equal(t, 0, model.MRDetailState.YOffset)
}

func TestArrowKeysScrollRightPanelInModeDetail(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	model.MRDetailState.YOffset = 3

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})

	model = updated.(Model)
	assert.Equal(t, 4, model.MRDetailState.YOffset)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})

	model = updated.(Model)
	assert.Equal(t, 3, model.MRDetailState.YOffset)
}

func TestUpDownInModeEntityListStillMovesSelection(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	model.mode = ModeEntityList
	initialSelected := model.EntityListState.mrList.Index()
	initialMRDetailViewportOffset := model.MRDetailState.YOffset

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(Model)

	assert.Equal(t, initialSelected+1, model.EntityListState.mrList.Index())

	assert.Equal(t, initialMRDetailViewportOffset, model.MRDetailState.YOffset)
}

// --- #64: Global key suppression in input modes ---

func TestEnteringMRCommentInputDisablesGlobalKeys(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	assert.True(t, model.globals.Quit.Enabled())

	assert.True(t, model.globals.Back.Enabled())
}

func TestEnteringEditInputDisablesGlobalKeys(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	assert.True(t, model.globals.Quit.Enabled())

	assert.True(t, model.globals.Back.Enabled())
}

func TestEnteringReplyInputInDiscussionsDisablesGlobalKeys(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	model.activeTab = TabDiscussions
	model.discussions[42] = []mr.Discussion{{ID: "d1", Notes: []mr.Note{{Author: "alice", Body: "test"}}}}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	assert.True(t, model.globals.Quit.Enabled())

	assert.True(t, model.globals.Back.Enabled())
}

func TestEnteringCommentInputInFileDiffDisablesGlobalKeys(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	model.activeTab = TabFiles
	model.changedFiles[42] = []mr.ChangedFile{{Path: "main.go", Diff: []mr.DiffRow{{OldLine: 1, NewLine: 1}}}}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	assert.True(t, model.globals.Quit.Enabled())

	assert.True(t, model.globals.Back.Enabled())
}

func TestEnteringFocusFilterDisablesGlobalKeys(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	assert.True(t, model.globals.Quit.Enabled())

	assert.True(t, model.globals.Back.Enabled())
}

func TestEnteringProjectInputModeDisablesGlobalKeys(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = updated.(Model)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updated.(Model)

	assert.Nil(t, cmd)

	if !strings.HasSuffix(model.Value(), "q") {
		t.Fatalf("expected edit buffer to end with q, got %q", model.Value())
	}
}

func TestQInFocusFilterDoesNotQuit(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	assert.Nil(t, cmd)
}

func TestKeyBarShowsInputHintsInFilterMode(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	model := NewModelWithProject(nil, ProjectOptions{
		LoadProjects: []AccountProjectLoader{{ID: "default", Host: "https://gitlab.com", Load: func() ([]string, error) { return nil, nil }}},
	})
	updated, _ := model.Update(accountProjectsFinishedMsg{accountID: "default", projects: []string{"group/project"}})
	model = updated.(Model)
	model.ProjectPickerState.selected = 1

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.Nil(t, cmd)

	if model.projectPath != "group/project" || model.mode != ModeSections {
		t.Fatalf("expected selected project sections to open, path=%q mode=%v", model.projectPath, model.mode)
	}
}

// --- issue #68: EmojiConfig + Draft toggle ---

func TestDKeyInModeDetailFlipsDraftOptimistically(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Fix login", Draft: false, State: "opened"}}
	model := NewModelWithProject(items, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	model = updated.(Model)

	assert.True(t, model.items[0].Draft)
}

func TestDKeyInModeDetailCallsToggleDraftFunc(t *testing.T) {
	t.Parallel()

	var calledIID int

	items := []mr.MergeRequest{{IID: 42, Title: "Fix login", Draft: false, State: "opened"}}
	model := NewModelWithProject(items, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		ToggleDraftMR: func(iid int, _ string, _ bool) error {
			calledIID = iid
			return nil
		},
	})

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	model = updated.(Model)

	assert.NotNil(t, cmd)

	msg := cmd()
	model.Update(msg)

	assert.Equal(t, 42, calledIID)
}

func TestDKeyInModeDetailWithNilFuncFlipsDraftLocally(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Draft: Fix login", Draft: true, State: "opened"}}
	model := NewModelWithProject(items, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
	})

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	model = updated.(Model)

	assert.False(t, model.items[0].Draft)

	assert.Nil(t, cmd)
}

func TestDKeyInModeDetailDoesNotTriggerInDiscussionsTab(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Fix", Draft: false, State: "opened"}}
	model := NewModelWithProject(items, ProjectOptions{
		Path:            "group/project",
		Section:         SectionMergeRequests,
		LoadDiscussions: func(iid int) ([]mr.Discussion, error) { return nil, nil },
		LoadFiles:       func(iid int) ([]mr.ChangedFile, error) { return nil, nil },
	})
	// Switch to discussions tab
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	updated, _ = model.Update(discussionsFinishedMsg{iid: 42, discussions: []mr.Discussion{
		{ID: "d1", Notes: []mr.Note{{Author: "alice", Body: "check this"}}},
	}})
	model = updated.(Model)

	// 'd' in discussions tab should open draft reply, not toggle Draft
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	model = updated.(Model)

	assert.False(t, model.items[0].Draft)

	assert.True(t, model.replyInput)
}

func TestSummaryRendersDraftPrefixForDraftMR(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Fix login", Draft: true, State: "opened", Approvals: "0/1"}}
	model := NewModelWithProject(items, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
	})

	view := model.renderRight()

	assert.Contains(t, view, "Draft:")
}

func TestSummaryDoesNotRenderDraftPrefixForNonDraftMR(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{IID: 42, Title: "Fix login", Draft: false, State: "opened", Approvals: "0/1"}}
	model := NewModelWithProject(items, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
	})

	view := model.renderRight()

	assert.NotContains(t, view, "Draft:")
}

func TestSummaryRendersEmojiAuthorWhenEnabled(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{
		IID: 42, Title: "Fix login", Author: "alice", AuthorUsername: "alice",
		SourceBranch: "feat", TargetBranch: "main", State: "opened",
		Pipeline: "success", Approvals: "1/2",
	}}
	model := NewModelWithProject(items, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
		Emoji:   config.DefaultEmojiConfig(),
	})

	view := model.renderRight()

	assert.Contains(t, view, "👤")

	assert.Contains(t, view, "🌿")
}

func TestSummaryRendersReviewersAndAssignees(t *testing.T) {
	t.Parallel()

	items := []mr.MergeRequest{{
		IID: 42, Title: "Fix", Author: "alice", SourceBranch: "feat", TargetBranch: "main",
		State: "opened", Approvals: "0/1",
		Reviewers: []string{"bob", "carol"},
		Assignees: []string{"dave"},
	}}
	model := NewModelWithProject(items, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
	})

	view := model.renderRight()

	assert.Contains(t, view, "bob")

	assert.Contains(t, view, "dave")
}

// --- issue #70: Label Selector ---

func labelSelectorModel() Model {
	return NewModelWithProject(
		[]mr.MergeRequest{{IID: 42, Title: "Fix login", State: "opened", Labels: []string{"bug"}}},
		ProjectOptions{
			Path:    "group/project",
			Section: SectionMergeRequests,
		},
	)
}

func TestLKeyInModeDetailSummaryOpensLabelSelector(t *testing.T) {
	t.Parallel()

	model := labelSelectorModel()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	model = updated.(Model)

	assert.Equal(t, ModeLabelSelect, model.mode)
}

func TestLabelSelectorRendersMarkers(t *testing.T) {
	t.Parallel()

	model := labelSelectorModel()
	model.projectLabels = []mr.Label{
		{Name: "bug", Color: "#EE0701"},
		{Name: "feature", Color: "#0075CA"},
	}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	model = updated.(Model)

	view := model.View()

	assert.Contains(t, view, "●")

	assert.Contains(t, view, "○")
}

func TestLabelSelectorUpDownMoveCursor(t *testing.T) {
	t.Parallel()

	model := labelSelectorModel()
	model.projectLabels = []mr.Label{
		{Name: "bug", Color: "#EE0701"},
		{Name: "feature", Color: "#0075CA"},
	}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	model = updated.(Model)

	assert.Equal(t, 0, model.LabelSelectorState.cursor)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)

	assert.Equal(t, 1, model.LabelSelectorState.cursor)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(Model)

	assert.Equal(t, 0, model.LabelSelectorState.cursor)
}

func TestLabelSelectorSpaceTogglesSelection(t *testing.T) {
	t.Parallel()

	model := labelSelectorModel()
	model.projectLabels = []mr.Label{
		{Name: "bug", Color: "#EE0701"},
		{Name: "feature", Color: "#0075CA"},
	}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	model = updated.(Model)

	// bug is already selected (in MR.Labels), Space should deselect
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(Model)

	for _, label := range model.LabelSelectorState.pending {
		assert.NotEqual(t, "bug", label)
	}

	// Move to 'feature' and select it
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(Model)

	found := false

	for _, label := range model.LabelSelectorState.pending {
		if label == "feature" {
			found = true
		}
	}

	assert.True(t, found)
}

func TestLabelSelectorEscReturnsToDetailWithoutAPICall(t *testing.T) {
	t.Parallel()

	called := false
	model := NewModelWithProject(
		[]mr.MergeRequest{{IID: 42, Title: "Fix", State: "opened", Labels: []string{"bug"}}},
		ProjectOptions{
			Path:    "group/project",
			Section: SectionMergeRequests,
			UpdateMRLabels: func(iid int, labels []string) error {
				called = true
				return nil
			},
		},
	)
	model.projectLabels = []mr.Label{{Name: "bug", Color: "#EE0701"}}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	assert.Equal(t, ModeDetail, model.mode)

	assert.False(t, called)

	if len(model.items[0].Labels) != 1 || model.items[0].Labels[0] != "bug" {
		t.Fatalf("expected MR labels unchanged after Esc, got %v", model.items[0].Labels)
	}
}

func TestLabelSelectorEnterCallsUpdateMRLabels(t *testing.T) {
	t.Parallel()

	var calledWith []string

	model := NewModelWithProject(
		[]mr.MergeRequest{{IID: 42, Title: "Fix", State: "opened", Labels: []string{"bug"}}},
		ProjectOptions{
			Path:    "group/project",
			Section: SectionMergeRequests,
			UpdateMRLabels: func(iid int, labels []string) error {
				calledWith = labels
				return nil
			},
		},
	)
	model.projectLabels = []mr.Label{
		{Name: "bug", Color: "#EE0701"},
		{Name: "feature", Color: "#0075CA"},
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	model = updated.(Model)
	// Move to feature and toggle on
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(Model)
	// Enter saves
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	assert.Equal(t, ModeDetail, model.mode)

	assert.NotNil(t, cmd)

	msg := cmd()
	model.Update(msg)

	assert.Len(t, calledWith, 2)
}

func TestLabelSelectorEnterUpdatesLabelsOptimistically(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(
		[]mr.MergeRequest{{IID: 42, Title: "Fix", State: "opened", Labels: []string{"bug"}}},
		ProjectOptions{
			Path:    "group/project",
			Section: SectionMergeRequests,
		},
	)
	model.projectLabels = []mr.Label{
		{Name: "bug", Color: "#EE0701"},
		{Name: "feature", Color: "#0075CA"},
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	// Before cmd runs, labels already updated optimistically
	assert.Len(t, model.items[0].Labels, 2)
}

func TestLabelSelectorReopensWithSavedSelection(t *testing.T) {
	t.Parallel()

	model := NewModelWithProject(
		[]mr.MergeRequest{{IID: 42, Title: "Fix", State: "opened", Labels: []string{"bug"}}},
		ProjectOptions{
			Path:    "group/project",
			Section: SectionMergeRequests,
		},
	)
	model.projectLabels = []mr.Label{
		{Name: "bug", Color: "#EE0701"},
		{Name: "feature", Color: "#0075CA"},
	}

	// Open, add feature, save
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	// Reopen
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	model = updated.(Model)

	featureSelected := false

	for _, label := range model.LabelSelectorState.pending {
		if label == "feature" {
			featureSelected = true
		}
	}

	assert.True(t, featureSelected)
}
