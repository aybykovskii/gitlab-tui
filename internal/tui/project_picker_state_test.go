package tui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectPickerStateViewRendersRecentProjectsAtTop(t *testing.T) {
	t.Parallel()

	state := NewProjectPickerState(
		[]RecentProjectOption{{Path: "group/recent-project", Account: "acme"}},
		nil,
	)

	view := state.View(LayoutState{Mode: ModeProjectSelect})

	for _, want := range []string{"Recent", "group/recent-project (acme)"} {
		assert.Contains(t, view, want)
	}
}

func TestProjectPickerStateViewMarksSelectedRow(t *testing.T) {
	t.Parallel()

	state := NewProjectPickerState(
		[]RecentProjectOption{{Path: "group/project"}},
		nil,
	)
	state.selected = 2 // after "Recent" header row and blank separator row

	view := state.View(LayoutState{Mode: ModeProjectSelect})

	if !strings.Contains(view, "group/project") || !strings.Contains(view, "│") {
		t.Fatalf("expected fancylist selection marker on project row, got:\n%s", view)
	}
}

func TestProjectPickerStateViewRendersAccountSectionHeader(t *testing.T) {
	t.Parallel()

	state := NewProjectPickerState(nil, []AccountProjectLoader{
		{ID: "myaccount", Host: "gitlab.example.com"},
	})

	view := state.View(LayoutState{Mode: ModeProjectSelect})

	assert.Contains(t, view, "[myaccount]  gitlab.example.com")
}

func TestProjectPickerStateViewShowsLoadingIndicatorForPendingAccount(t *testing.T) {
	t.Parallel()

	state := NewProjectPickerState(nil, []AccountProjectLoader{
		{ID: "acme", Host: "gitlab.acme.io"},
	})
	// initialAccountProjectStates sets loading=true by default

	view := state.View(LayoutState{Mode: ModeProjectSelect})

	assert.Contains(t, view, "Loading…")
}

func TestProjectPickerStateViewShowsErrorIndicatorForFailedAccount(t *testing.T) {
	t.Parallel()

	state := NewProjectPickerState(nil, []AccountProjectLoader{
		{ID: "acme", Host: "gitlab.acme.io"},
	})
	state.accountProjectStates["acme"] = accountProjectState{host: "gitlab.acme.io", err: "connection refused"}
	state.rebuildRows()

	view := state.View(LayoutState{Mode: ModeProjectSelect})

	assert.Contains(t, view, "Error: connection refused")
}

func TestProjectPickerStateViewFiltersProjectsByQuery(t *testing.T) {
	t.Parallel()

	state := NewProjectPickerState(
		[]RecentProjectOption{
			{Path: "group/visible-project"},
			{Path: "group/hidden-project"},
		},
		nil,
	)
	state.query = "visible"
	state.rebuildRows()

	view := state.View(LayoutState{Mode: ModeProjectSelect})

	assert.Contains(t, view, "group/visible-project")

	assert.NotContains(t, view, "group/hidden-project")
}

func TestProjectPickerStateViewRendersInputFormInProjectInputMode(t *testing.T) {
	t.Parallel()

	state := NewProjectPickerState(nil, nil)
	state.projectInput = "group/typed-path"

	view := state.View(LayoutState{Mode: ModeProjectInput})

	for _, want := range []string{"Open GitLab project", "Project path:", "group/typed-path"} {
		assert.Contains(t, view, want)
	}
}
