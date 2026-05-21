package tui

import (
	"strings"
	"testing"


	"github.com/stretchr/testify/assert"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func TestBuildFileTreeFlat(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{{Path: "main.go"}, {Path: "go.mod"}}
	tree := buildFileTree(files)

	assert.Len(t, tree.children, 2)

	if tree.children[0].name != "main.go" || tree.children[0].fileIdx != 0 {
		t.Fatalf("expected first child main.go (idx 0), got %q (idx %d)", tree.children[0].name, tree.children[0].fileIdx)
	}
}

func TestBuildFileTreeNested(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{
		{Path: "internal/tui/render.go"},
		{Path: "internal/mr/model.go"},
		{Path: "cmd/main.go"},
	}
	tree := buildFileTree(files)

	assert.Len(t, tree.children, 2)

	internal := tree.children[0]
	if internal.name != "internal" || internal.fileIdx != -1 {
		t.Fatalf("expected directory node 'internal', got %q fileIdx=%d", internal.name, internal.fileIdx)
	}

	assert.Len(t, internal.children, 2)
}

func TestBuildFileTreeSharedPrefix(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{
		{Path: "pkg/a/foo.go"},
		{Path: "pkg/b/bar.go"},
	}
	tree := buildFileTree(files)

	if len(tree.children) != 1 || tree.children[0].name != "pkg" {
		t.Fatalf("expected single 'pkg' root child, got %d children", len(tree.children))
	}

	pkg := tree.children[0]
	assert.Len(t, pkg.children, 2)
}

func TestRenderFileTreeLinesConnectors(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{
		{Path: "internal/render.go"},
		{Path: "internal/update.go"},
		{Path: "cmd/main.go"},
	}
	tree := buildFileTree(files)
	lines := renderFileTreeLines(tree, "", false, files, -1, 60)
	out := strings.Join(lines, "\n")

	for _, want := range []string{"├──", "└──", "│", "internal", "cmd", "render.go", "update.go", "main.go"} {
		assert.Contains(t, out, want)
	}
}

func TestRenderFileTreeLinesTruncation(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{{Path: "very_long_filename_that_exceeds_limit.go"}}
	tree := buildFileTree(files)
	lines := renderFileTreeLines(tree, "", false, files, -1, 10)

	assert.Len(t, lines, 1)

	assert.Contains(t, lines[0], "…")

	runes := []rune(lines[0])
	if len(runes) > 10 {
		t.Fatalf("expected line to be at most 10 runes, got %d: %q", len(runes), lines[0])
	}
}

func TestRenderFileTreeLinesColors(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{
		{Path: "new.go", IsNew: true},
		{Path: "del.go", IsDeleted: true},
		{Path: "ren.go", IsRenamed: true},
		{Path: "mod.go"},
	}
	tree := buildFileTree(files)
	lines := renderFileTreeLines(tree, "", false, files, -1, 60)
	out := strings.Join(lines, "\n")

	assert.Contains(t, out, "\x1b[38;5;2m")

	assert.Contains(t, out, "\x1b[38;5;1m")

	assert.Contains(t, out, "\x1b[38;5;3m")
}

func TestRenderFileTreeLinesSelectedHighlight(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{
		{Path: "a.go"},
		{Path: "b.go"},
	}
	tree := buildFileTree(files)

	linesA := renderFileTreeLines(tree, "", false, files, 0, 60)
	linesB := renderFileTreeLines(tree, "", false, files, 1, 60)

	assert.Contains(t, linesA[0], "a.go")

	assert.Contains(t, linesB[1], "b.go")

	// selected line contains lipgloss background escape
	assert.Contains(t, linesA[0], "\x1b[")
}

func TestFindSelectedFileLine(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{
		{Path: "internal/tui/render.go"},
		{Path: "internal/tui/update.go"},
		{Path: "cmd/main.go"},
	}
	tree := buildFileTree(files)

	// internal=0, tui=1, render.go=2, update.go=3, cmd=4, main.go=5
	cases := []struct {
		selectedFile int
		wantLine     int
	}{
		{0, 2}, // render.go
		{1, 3}, // update.go
		{2, 5}, // main.go
	}

	for _, tc := range cases {
		got := findSelectedFileLine(tree, tc.selectedFile)
		if got != tc.wantLine {
			t.Errorf("findSelectedFileLine(selectedFile=%d) = %d, want %d", tc.selectedFile, got, tc.wantLine)
		}
	}
}

func TestFindSelectedFileLineFlatTree(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{{Path: "a.go"}, {Path: "b.go"}, {Path: "c.go"}}
	tree := buildFileTree(files)

	for i := range files {
		got := findSelectedFileLine(tree, i)
		if got != i {
			t.Errorf("flat tree: findSelectedFileLine(%d) = %d, want %d", i, got, i)
		}
	}
}

func TestTruncateRunes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input  string
		max    int
		expect string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 4, "hel…"},
		{"hello", 1, "…"},
		{"├──file", 4, "├──…"},
	}

	for _, tc := range cases {
		got := truncateRunes(tc.input, tc.max)
		if got != tc.expect {
			t.Errorf("truncateRunes(%q, %d) = %q, want %q", tc.input, tc.max, got, tc.expect)
		}
	}
}
