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

func (m Model) renderChangedFilesPane() string {
	width := m.leftWidth()
	height := m.paneHeight()
	style := paneStyle(width, height, false)
	files := m.currentFiles()
	lines := []string{"Changed Files", ""}

	for i, file := range files {
		prefix := "  "
		if i == m.selectedFile {
			prefix = "> "
		}

		lines = append(lines, prefix+file.Path)
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

	view := m.DiffViewState.View(LayoutState{Width: width, Height: height, Focus: m.focus, Mode: m.mode})

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
