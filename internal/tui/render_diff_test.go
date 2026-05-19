package tui

import (
	"strings"
	"testing"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func TestBuildFileTreeFlat(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{{Path: "main.go"}, {Path: "go.mod"}}
	tree := buildFileTree(files)

	if len(tree.children) != 2 {
		t.Fatalf("expected 2 root children, got %d", len(tree.children))
	}

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

	if len(tree.children) != 2 {
		t.Fatalf("expected 2 root children (internal, cmd), got %d", len(tree.children))
	}

	internal := tree.children[0]
	if internal.name != "internal" || internal.fileIdx != -1 {
		t.Fatalf("expected directory node 'internal', got %q fileIdx=%d", internal.name, internal.fileIdx)
	}

	if len(internal.children) != 2 {
		t.Fatalf("expected 2 children under internal, got %d", len(internal.children))
	}
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
	if len(pkg.children) != 2 {
		t.Fatalf("expected 2 children under pkg, got %d", len(pkg.children))
	}
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
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestRenderFileTreeLinesTruncation(t *testing.T) {
	t.Parallel()

	files := []mr.ChangedFile{{Path: "very_long_filename_that_exceeds_limit.go"}}
	tree := buildFileTree(files)
	lines := renderFileTreeLines(tree, "", false, files, -1, 10)

	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "…") {
		t.Fatalf("expected truncation with …, got: %q", lines[0])
	}

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

	if !strings.Contains(out, "\x1b[38;5;2m") {
		t.Fatalf("expected green color for new file")
	}

	if !strings.Contains(out, "\x1b[38;5;1m") {
		t.Fatalf("expected red color for deleted file")
	}

	if !strings.Contains(out, "\x1b[38;5;3m") {
		t.Fatalf("expected yellow color for renamed file")
	}
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

	if !strings.Contains(linesA[0], "a.go") {
		t.Fatalf("expected a.go in first line, got: %q", linesA[0])
	}

	if !strings.Contains(linesB[1], "b.go") {
		t.Fatalf("expected b.go in second line, got: %q", linesB[1])
	}

	// selected line contains lipgloss background escape
	if !strings.Contains(linesA[0], "\x1b[") {
		t.Fatalf("expected ANSI escape in selected line a.go, got: %q", linesA[0])
	}
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
