//nolint:mnd,prealloc,gocritic // UI sizing constants and append composition favor readability.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type DiscussionListOptions struct {
	ShowStatus bool
	Separator  bool
}

func (m Model) renderDiscussions(item mr.MergeRequest) string {
	if m.discussionsLoading {
		return "Loading discussions…"
	}

	if m.discussionsError != "" {
		return "Error: " + m.discussionsError + "\n\nr retry"
	}

	discussions, loaded := m.discussions[item.IID]
	if !loaded {
		return "Tab to load discussions"
	}

	return renderDiscussionList(discussions, m.discussionCursor, DiscussionListOptions{ShowStatus: true, Separator: true})
}

func renderDiscussionList(discussions []mr.Discussion, cursor int, opts DiscussionListOptions) string {
	if len(discussions) == 0 {
		return "No discussions"
	}

	sep := "─────────────────────────────────────────"
	lines := []string{}

	for i, discussion := range discussions {
		if opts.Separator && i > 0 {
			lines = append(lines, sep)
		}

		cursorPrefix := "  "
		if i == cursor {
			cursorPrefix = "> "
		}

		firstAuthor := ""
		if len(discussion.Notes) > 0 {
			firstAuthor = discussion.Notes[0].Author
		}

		header := firstAuthor

		if opts.ShowStatus {
			status := "open"
			if discussion.Resolved {
				status = "resolved"
			}

			header = fmt.Sprintf("[%s] %s", status, firstAuthor)
		}

		lines = append(lines, renderDiscussionBlock(discussion, header, cursorPrefix, false, false)...)
	}

	return strings.Join(lines, "\n")
}

func (m Model) currentFiles() []mr.ChangedFile {
	item, ok := m.selectedItem()
	if !ok {
		return nil
	}

	return m.changedFiles[item.IID]
}

type fileTreeNode struct {
	name     string
	children []*fileTreeNode
	fileIdx  int // -1 for directories, ≥0 for leaf files
}

func buildFileTree(files []mr.ChangedFile) *fileTreeNode {
	root := &fileTreeNode{fileIdx: -1}

	for i, file := range files {
		parts := strings.Split(file.Path, "/")
		cur := root

		for j, part := range parts {
			isFile := j == len(parts)-1

			var child *fileTreeNode
			for _, c := range cur.children {
				if c.name == part {
					child = c
					break
				}
			}

			if child == nil {
				idx := -1
				if isFile {
					idx = i
				}
				child = &fileTreeNode{name: part, fileIdx: idx}
				cur.children = append(cur.children, child)
			}

			cur = child
		}
	}

	return root
}

func fileStatusColor(file mr.ChangedFile) string {
	switch {
	case file.IsNew:
		return "2"
	case file.IsDeleted:
		return "1"
	case file.IsRenamed:
		return "3"
	default:
		return ""
	}
}

func renderFileTreeLines(node *fileTreeNode, prefix string, isLast bool, files []mr.ChangedFile, selectedFile int, innerWidth int) []string {
	var lines []string

	if node.name != "" {
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		linePrefix := prefix + connector
		prefixWidth := len([]rune(linePrefix))
		name := truncateRunes(node.name, max(1, innerWidth-prefixWidth))

		if node.fileIdx >= 0 {
			if node.fileIdx == selectedFile {
				lines = append(lines, ansiSelected(linePrefix+name))
			} else {
				color := fileStatusColor(files[node.fileIdx])
				if color != "" {
					lines = append(lines, linePrefix+ansiColor(color, name))
				} else {
					lines = append(lines, linePrefix+name)
				}
			}
		} else {
			lines = append(lines, linePrefix+name)
		}
	}

	childPrefix := prefix
	if node.name != "" {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for i, child := range node.children {
		lines = append(lines, renderFileTreeLines(child, childPrefix, i == len(node.children)-1, files, selectedFile, innerWidth)...)
	}

	return lines
}

// findSelectedFileLine returns the 0-based line index of selectedFile in the tree output,
// matching the DFS traversal order of renderFileTreeLines. Returns -1 if not found.
func findSelectedFileLine(node *fileTreeNode, selectedFile int) int {
	count := 0
	return dfsCountTo(node, selectedFile, &count)
}

func dfsCountTo(node *fileTreeNode, target int, count *int) int {
	if node.name != "" {
		if node.fileIdx == target {
			return *count
		}

		(*count)++
	}

	for _, child := range node.children {
		if pos := dfsCountTo(child, target, count); pos >= 0 {
			return pos
		}
	}

	return -1
}

func (m Model) renderChangedFilesPane() string {
	width := m.leftWidth()
	height := m.paneHeight()
	style := paneStyle(width, height, false)
	files := m.currentFiles()
	innerWidth := width - 4 // border (1+1) + padding (1+1)
	lines := []string{"Changed Files", ""}

	if len(files) > 0 {
		tree := buildFileTree(files)
		treeLines := renderFileTreeLines(tree, "", false, files, m.selectedFile, innerWidth)

		treeHeight := max(1, height-4) // pane inner area minus 2 header lines
		if len(treeLines) > treeHeight {
			selectedLine := findSelectedFileLine(tree, m.selectedFile)
			offset := clamp(selectedLine-treeHeight/2, 0, len(treeLines)-treeHeight)
			treeLines = treeLines[offset:min(offset+treeHeight, len(treeLines))]
		}

		lines = append(lines, treeLines...)
	}

	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderFileDiffPane() string {
	width := max(20, m.width-m.leftWidth())
	height := m.paneHeight()
	style := paneStyle(width, height, true)
	files := m.currentFiles()

	if len(files) == 0 {
		return style.Render("No files")
	}

	item, _ := m.selectedItem()
	m.DiffViewState.diffFiles = files
	m.DiffViewState.diffDiscussions = m.discussions[item.IID]
	m.DiffViewState.diffDrafts = m.drafts[item.IID]
	m.DiffViewState.emoji = m.emoji

	var inputLines []string

	if m.commentError != "" {
		inputLines = append(inputLines, "", "Error: "+m.commentError)
	}

	if m.commentInput {
		prompt := "Comment"
		if m.commentInstant {
			prompt = "Instant comment"
		}

		inputLines = append(inputLines, "", prompt+": "+m.Value()+"█")
	}

	diffHeight := height - len(inputLines)
	view := m.DiffViewState.View(LayoutState{Width: width, Height: diffHeight, Focus: m.focus, Mode: m.mode})

	if len(inputLines) > 0 {
		view += "\n" + strings.Join(inputLines, "\n")
	}

	return style.Render(view)
}

func (m Model) cursorRow() (mr.ChangedFile, mr.DiffRow, bool) {
	files := m.currentFiles()
	if len(files) <= m.selectedFile {
		return mr.ChangedFile{}, mr.DiffRow{}, false
	}

	file := files[m.selectedFile]
	if m.diffCursor >= len(file.Diff) {
		return mr.ChangedFile{}, mr.DiffRow{}, false
	}

	return file, file.Diff[m.diffCursor], true
}

func (m Model) discussionsAtCursor() []mr.Discussion {
	file, row, ok := m.cursorRow()
	if !ok || row.NewLine == 0 {
		return nil
	}

	item, ok := m.selectedItem()
	if !ok {
		return nil
	}

	var result []mr.Discussion

	for _, discussion := range m.discussions[item.IID] {
		if discussion.Position != nil && discussion.Position.NewPath == file.Path && discussion.Position.NewLine == row.NewLine {
			result = append(result, discussion)
		}
	}

	return result
}

func renderDiscussionBlock(discussion mr.Discussion, header string, cursor string, dimResolved bool, authorInFirstNote bool) []string {
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	dim := dimResolved && discussion.Resolved
	apply := func(s string) string {
		if dim {
			return dimStyle.Render(s)
		}

		return s
	}
	lines := []string{apply(cursor + header)}

	for j, note := range discussion.Notes {
		var entry string

		if j == 0 {
			if authorInFirstNote {
				entry = fmt.Sprintf("  [%s] %s", note.Author, note.Body)
			} else {
				entry = "  " + note.Body
			}
		} else {
			entry = fmt.Sprintf("  ↳ %s: %s", note.Author, note.Body)
		}

		lines = append(lines, apply(entry))
	}

	return lines
}
