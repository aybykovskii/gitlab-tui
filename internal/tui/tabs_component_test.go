package tui

import "testing"

func TestTabsComponentMarksActiveTab(t *testing.T) {
	t.Parallel()

	component := TabsComponent{Labels: []string{"Summary", "Discussions"}, Active: 1}

	if got, want := component.View(), "[Summary] [>Discussions<]"; got != want {
		t.Fatalf("expected active tab marker, got %q want %q", got, want)
	}
}

func TestTabsComponentRendersAllInactiveWhenActiveIndexIsOutOfRange(t *testing.T) {
	t.Parallel()

	component := TabsComponent{Labels: []string{"Summary", "Discussions"}, Active: -1}

	if got, want := component.View(), "[Summary] [Discussions]"; got != want {
		t.Fatalf("expected all tabs inactive, got %q want %q", got, want)
	}
}

func TestTabsComponentRendersSingleTab(t *testing.T) {
	t.Parallel()

	component := TabsComponent{Labels: []string{"Summary"}, Active: 0}

	if got, want := component.View(), "[>Summary<]"; got != want {
		t.Fatalf("expected single active tab, got %q want %q", got, want)
	}
}

func TestTabsComponentJoinsLabelsWithSingleSpace(t *testing.T) {
	t.Parallel()

	component := TabsComponent{Labels: []string{"Summary", "Discussions", "Files"}, Active: 2}

	if got, want := component.View(), "[Summary] [Discussions] [>Files<]"; got != want {
		t.Fatalf("expected tabs joined by spaces, got %q want %q", got, want)
	}
}
