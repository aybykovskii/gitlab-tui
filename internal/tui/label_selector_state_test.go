package tui

import (
	"strings"
	"testing"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func TestLabelSelectorStateViewShowsEmptyMessage(t *testing.T) {
	t.Parallel()

	state := NewLabelSelectorState()

	view := state.View(LayoutState{})

	if !strings.Contains(view, "No project labels") {
		t.Fatalf("expected empty message, got:\n%s", view)
	}
}

func TestLabelSelectorStateViewRendersHeaderHint(t *testing.T) {
	t.Parallel()

	state := NewLabelSelectorState()

	view := state.View(LayoutState{})

	if !strings.Contains(view, "Space toggle") || !strings.Contains(view, "Enter save") || !strings.Contains(view, "Esc cancel") {
		t.Fatalf("expected hint header in view, got:\n%s", view)
	}
}

func TestLabelSelectorStateViewShowsUnselectedMarkerForUnpendingLabel(t *testing.T) {
	t.Parallel()

	state := NewLabelSelectorState()
	state.labels = []mr.Label{{Name: "bug", Color: "#d73a4a"}}

	view := state.View(LayoutState{})

	if !strings.Contains(view, "○") {
		t.Fatalf("expected unselected marker ○, got:\n%s", view)
	}
}

func TestLabelSelectorStateViewShowsSelectedMarkerForPendingLabel(t *testing.T) {
	t.Parallel()

	state := NewLabelSelectorState()
	state.labels = []mr.Label{{Name: "bug", Color: "#d73a4a"}}
	state.pending = []string{"bug"}

	view := state.View(LayoutState{})

	if !strings.Contains(view, "●") {
		t.Fatalf("expected selected marker ●, got:\n%s", view)
	}
}

func TestLabelSelectorStateViewMarksCursorRow(t *testing.T) {
	t.Parallel()

	state := NewLabelSelectorState()
	state.labels = []mr.Label{
		{Name: "bug", Color: "#d73a4a"},
		{Name: "enhancement", Color: "#a2eeef"},
	}
	state.cursor = 1

	view := state.View(LayoutState{})

	lines := strings.Split(view, "\n")
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "> ") && strings.Contains(line, "enhancement") {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected cursor marker on 'enhancement' row, got:\n%s", view)
	}
}

func TestLabelSelectorStateViewOnlyMarksOneLabelSelected(t *testing.T) {
	t.Parallel()

	state := NewLabelSelectorState()
	state.labels = []mr.Label{
		{Name: "bug", Color: "#d73a4a"},
		{Name: "enhancement", Color: "#a2eeef"},
	}
	state.pending = []string{"bug"}

	view := state.View(LayoutState{})

	if strings.Count(view, "●") != 1 {
		t.Fatalf("expected exactly one selected marker, got:\n%s", view)
	}

	if strings.Count(view, "○") != 1 {
		t.Fatalf("expected exactly one unselected marker, got:\n%s", view)
	}
}
