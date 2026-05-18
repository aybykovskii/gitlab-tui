package tui

import (
	"strings"
	"testing"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

// Cycle 1: luminance formula — dark background → white foreground
func TestForegroundForBlackBackgroundIsWhite(t *testing.T) {
	fg := foregroundForBackground("#000000")
	if fg != "255" {
		t.Fatalf("expected white (255) for black background, got %q", fg)
	}
}

// Cycle 2: light background → black foreground
func TestForegroundForWhiteBackgroundIsBlack(t *testing.T) {
	fg := foregroundForBackground("#FFFFFF")
	if fg != "0" {
		t.Fatalf("expected black (0) for white background, got %q", fg)
	}
}

// Cycle 3: boundary — pure blue (#0000FF, L≈0.072) → white
func TestForegroundForPureBlueIsWhite(t *testing.T) {
	fg := foregroundForBackground("#0000FF")
	if fg != "255" {
		t.Fatalf("expected white for pure blue #0000FF (L=0.072<0.179), got %q", fg)
	}
}

// Cycle 3b: boundary — light yellow (#FFFF00, L≈0.93) → black
func TestForegroundForYellowIsBlack(t *testing.T) {
	fg := foregroundForBackground("#FFFF00")
	if fg != "0" {
		t.Fatalf("expected black for yellow #FFFF00 (L>0.179), got %q", fg)
	}
}

// Cycle 3c: malformed hex → falls back to white (safe default)
func TestForegroundForMalformedHexFallsBackToWhite(t *testing.T) {
	fg := foregroundForBackground("notahex")
	if fg != "255" {
		t.Fatalf("expected white fallback for malformed hex, got %q", fg)
	}
}

// Cycle 4: pill contains the label name
func TestRenderLabelPillContainsName(t *testing.T) {
	pill := renderLabelPill("bug", "#EE0701")
	if !strings.Contains(pill, "bug") {
		t.Fatalf("expected pill to contain label name 'bug', got %q", pill)
	}
}

// Cycle 5: labels are cached in model after projectFinishedMsg
func TestProjectLabelsStoredAfterProjectLoad(t *testing.T) {
	labels := []mr.Label{
		{Name: "bug", Color: "#EE0701"},
		{Name: "frontend", Color: "#0075CA"},
	}
	model := NewModelWithProject(nil, ProjectOptions{Path: "group/project"})

	updated, _ := model.Update(projectFinishedMsg{
		path: "group/project",
		data: ProjectData{
			Items:  nil,
			Labels: labels,
		},
	})
	model = updated.(Model)

	if len(model.projectLabels) != 2 {
		t.Fatalf("expected 2 cached labels, got %d", len(model.projectLabels))
	}

	if model.projectLabels[0].Name != "bug" {
		t.Fatalf("expected first label 'bug', got %q", model.projectLabels[0].Name)
	}
}

// Cycle 6: Summary renders label pills when MR has labels and labels are cached
func TestSummaryRendersLabelPillWithCachedColor(t *testing.T) {
	items := []mr.MergeRequest{{
		IID: 42, Title: "Fix login", State: "opened",
		SourceBranch: "feat", TargetBranch: "main",
		Approvals: "0/1",
		Labels:    []string{"bug"},
	}}
	model := NewModelWithProject(items, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
	})
	model.projectLabels = []mr.Label{{Name: "bug", Color: "#EE0701"}}

	view := model.renderRight()

	if !strings.Contains(view, "bug") {
		t.Fatalf("expected label 'bug' in summary, got:\n%s", view)
	}
}

// Cycle 7: MR without labels hides labels line
func TestSummaryHidesLabelsLineWhenNoLabels(t *testing.T) {
	items := []mr.MergeRequest{{
		IID: 42, Title: "Fix login", State: "opened",
		SourceBranch: "feat", TargetBranch: "main",
		Approvals: "0/1",
		Labels:    nil,
	}}
	model := NewModelWithProject(items, ProjectOptions{
		Path:    "group/project",
		Section: SectionMergeRequests,
	})
	model.projectLabels = []mr.Label{{Name: "bug", Color: "#EE0701"}}

	view := model.renderRight()

	if strings.Contains(view, "bug") {
		t.Fatalf("expected no label line for MR without labels, got:\n%s", view)
	}
}
