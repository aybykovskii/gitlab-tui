package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

var errTestRefresh = errors.New("refresh failed")

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

func TestRecentProjectSelectionOpensProject(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{Recents: []string{"group/one", "group/two"}})
	if model.mode != ModeProjectSelect {
		t.Fatalf("expected project select mode, got %v", model.mode)
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.projectPath != "group/two" {
		t.Fatalf("expected selected recent project, got %q", model.projectPath)
	}
	if model.mode != ModeDetail {
		t.Fatalf("expected detail mode, got %v", model.mode)
	}
}

func TestManualProjectInputOpensProject(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{})
	if model.mode != ModeProjectInput {
		t.Fatalf("expected project input mode, got %v", model.mode)
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("group/project")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.projectPath != "group/project" {
		t.Fatalf("expected manual project, got %q", model.projectPath)
	}
	if model.mode != ModeDetail {
		t.Fatalf("expected detail mode, got %v", model.mode)
	}
}

func TestRefreshKeyReturnsCommand(t *testing.T) {
	model := NewModelWithProject(FakeMergeRequests(), ProjectOptions{
		Path: "group/project",
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
