package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

// Cycle 1: luminance formula — dark background → white foreground.
func TestForegroundForBlackBackgroundIsWhite(t *testing.T) {
	t.Parallel()

	fg := foregroundForBackground("#000000")
	assert.EqualValues(t, "255", fg)
}

// Cycle 2: light background → black foreground.
func TestForegroundForWhiteBackgroundIsBlack(t *testing.T) {
	t.Parallel()

	fg := foregroundForBackground("#FFFFFF")
	assert.EqualValues(t, "0", fg)
}

// Cycle 3: boundary — pure blue (#0000FF, L≈0.072) → white.
func TestForegroundForPureBlueIsWhite(t *testing.T) {
	t.Parallel()

	fg := foregroundForBackground("#0000FF")
	assert.EqualValues(t, "255", fg)
}

// Cycle 3b: boundary — light yellow (#FFFF00, L≈0.93) → black.
func TestForegroundForYellowIsBlack(t *testing.T) {
	t.Parallel()

	fg := foregroundForBackground("#FFFF00")
	assert.EqualValues(t, "0", fg)
}

// Cycle 3c: malformed hex → falls back to white (safe default).
func TestForegroundForMalformedHexFallsBackToWhite(t *testing.T) {
	t.Parallel()

	fg := foregroundForBackground("notahex")
	assert.EqualValues(t, "255", fg)
}

// Cycle 4: pill contains the label name.
func TestRenderLabelPillContainsName(t *testing.T) {
	t.Parallel()

	pill := renderLabelPill("bug", "#EE0701")
	assert.Contains(t, pill, "bug")
}

// Cycle 5: labels are cached in model after projectFinishedMsg.
func TestProjectLabelsStoredAfterProjectLoad(t *testing.T) {
	t.Parallel()

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

	assert.Len(t, model.projectLabels, 2)

	assert.Equal(t, "bug", model.projectLabels[0].Name)
}

// Cycle 6: Summary renders label pills when MR has labels and labels are cached.
func TestSummaryRendersLabelPillWithCachedColor(t *testing.T) {
	t.Parallel()

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

	assert.Contains(t, view, "bug")
}

// Cycle 7: MR without labels hides labels line.
func TestSummaryHidesLabelsLineWhenNoLabels(t *testing.T) {
	t.Parallel()

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

	assert.NotContains(t, view, "bug")
}
