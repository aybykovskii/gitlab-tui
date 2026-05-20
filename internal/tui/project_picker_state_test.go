package tui

import (
	"strings"
	"testing"
)

func TestProjectPickerStateViewRendersRecentProjectsAtTop(t *testing.T) {
	t.Parallel()

	state := NewProjectPickerState(
		[]RecentProjectOption{{Path: "group/recent-project", Account: "acme"}},
		nil,
	)

	view := state.View(LayoutState{Mode: ModeProjectSelect})

	for _, want := range []string{"Recent", "group/recent-project (acme)"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected %q in view, got:\n%s", want, view)
		}
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

	if !strings.Contains(view, "[myaccount]  gitlab.example.com") {
		t.Fatalf("expected account section header, got:\n%s", view)
	}
}

func TestProjectPickerStateViewShowsLoadingIndicatorForPendingAccount(t *testing.T) {
	t.Parallel()

	state := NewProjectPickerState(nil, []AccountProjectLoader{
		{ID: "acme", Host: "gitlab.acme.io"},
	})
	// initialAccountProjectStates sets loading=true by default

	view := state.View(LayoutState{Mode: ModeProjectSelect})

	if !strings.Contains(view, "Loading…") {
		t.Fatalf("expected loading indicator, got:\n%s", view)
	}
}

func TestProjectPickerStateViewShowsErrorIndicatorForFailedAccount(t *testing.T) {
	t.Parallel()

	state := NewProjectPickerState(nil, []AccountProjectLoader{
		{ID: "acme", Host: "gitlab.acme.io"},
	})
	state.accountProjectStates["acme"] = accountProjectState{host: "gitlab.acme.io", err: "connection refused"}
	state.rebuildRows()

	view := state.View(LayoutState{Mode: ModeProjectSelect})

	if !strings.Contains(view, "Error: connection refused") {
		t.Fatalf("expected error indicator, got:\n%s", view)
	}
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

	if !strings.Contains(view, "group/visible-project") {
		t.Fatalf("expected visible project in filtered view, got:\n%s", view)
	}

	if strings.Contains(view, "group/hidden-project") {
		t.Fatalf("expected hidden project to be filtered out, got:\n%s", view)
	}
}

func TestProjectPickerStateViewRendersInputFormInProjectInputMode(t *testing.T) {
	t.Parallel()

	state := NewProjectPickerState(nil, nil)
	state.projectInput = "group/typed-path"

	view := state.View(LayoutState{Mode: ModeProjectInput})

	for _, want := range []string{"Open GitLab project", "Project path:", "group/typed-path"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected %q in input mode view, got:\n%s", want, view)
		}
	}
}
