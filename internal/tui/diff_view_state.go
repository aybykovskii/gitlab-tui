//nolint:mnd // Diff column widths and gutter sizes are intentional UI layout constants.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
	"github.com/aybykovskii/gitlab-tui/pkg/diff"
)

type DiffViewState struct {
	viewport.Model
	diffFiles          []mr.ChangedFile
	selectedFile       int
	diffCursor         int
	fileDiffReturnTab  DetailTab
	rangeStart         int
	threadPanelVisible bool
	threadPanelCursor  int
	diffDiscussions    []mr.Discussion
	diffDrafts         []mr.DraftComment
	emoji              config.EmojiConfig
}

func NewDiffViewState() DiffViewState {
	return DiffViewState{
		Model:              viewport.New(0, 0),
		rangeStart:         -1,
		threadPanelVisible: true,
	}
}

func (s *DiffViewState) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.Model, cmd = s.Model.Update(msg)

	return cmd
}

func (s *DiffViewState) View(layout LayoutState) string {
	s.Width = layout.Width - 4
	s.Height = max(1, layout.Height-2)
	s.SetContent(s.content(layout))

	return s.Model.View()
}

func (s DiffViewState) content(layout LayoutState) string {
	if len(s.diffFiles) == 0 {
		return "No files"
	}

	file := s.diffFiles[clamp(s.selectedFile, 0, len(s.diffFiles)-1)]
	lines := []string{fmt.Sprintf("Diff %s", file.Path), ""}
	annotated := diff.ProjectDiscussions(file.Diff, s.diffDiscussions, file.Path)

	colWidth := max(10, (max(20, layout.Width-4)-22)/2)
	rowFmt := fmt.Sprintf("%%s │ %%-%ds │ %%s │ %%-%ds", colWidth, colWidth)

	for i, arow := range annotated {
		cursor := "  "
		if i == s.diffCursor {
			cursor = "> "
		}

		draftMarker := s.draftGutterMarker(file.Path, arow.NewLine)
		if draftMarker == " " && s.isActiveDraftRangeRow(i) {
			draftMarker = "·"
		}

		discussionMarker := s.discussionGutterMarker(arow.Discussions)

		var oldNum, newNum, oldContent, newContent, rowColor string

		switch {
		case arow.OldLine == 0 && arow.NewLine > 0:
			oldNum = "    "
			newNum = fmt.Sprintf("%4d", arow.NewLine)
			oldContent = strings.Repeat(" ", colWidth)
			newContent = "+ " + arow.NewText
			rowColor = "2"
		case arow.NewLine == 0 && arow.OldLine > 0:
			oldNum = fmt.Sprintf("%4d", arow.OldLine)
			newNum = "    "
			oldContent = "- " + arow.OldText
			newContent = ""
			rowColor = "1"
		default:
			oldNum = fmt.Sprintf("%4d", arow.OldLine)
			newNum = fmt.Sprintf("%4d", arow.NewLine)
			oldContent = "  " + arow.OldText
			newContent = "  " + arow.NewText
			rowColor = "240"
		}

		if runes := []rune(oldContent); len(runes) > colWidth {
			oldContent = string(runes[:colWidth])
		}

		if runes := []rune(newContent); len(runes) > colWidth {
			newContent = string(runes[:colWidth])
		}

		lineContent := fmt.Sprintf(rowFmt, oldNum, oldContent, newNum, newContent)
		lines = append(lines, cursor+draftMarker+discussionMarker+" "+ansiColor(rowColor, lineContent))
	}

	discussion, draft := s.threadAtCursor()
	allDiscussions := s.discussionsAtCursor()

	if (discussion != nil || draft != nil) && s.threadPanelVisible {
		lines = append(lines, s.threadPanelLines(discussion, draft, len(allDiscussions), layout.Width-4)...)
	}

	return strings.Join(lines, "\n")
}

func ansiColor(color string, text string) string {
	return "\x1b[38;5;" + color + "m" + text + "\x1b[0m"
}

func (s DiffViewState) draftGutterMarker(path string, newLine int) string {
	if newLine == 0 {
		return " "
	}

	for _, dr := range s.diffDrafts {
		if dr.Position == nil || dr.Position.NewPath != path {
			continue
		}

		startLine := dr.Position.NewLine
		endLine := dr.EndLine

		if endLine == 0 {
			endLine = startLine
		}

		if newLine >= startLine && newLine <= endLine {
			if s.emoji.Enabled {
				icon := s.emoji.Resolve().Draft
				if icon != "" {
					return icon
				}
			}

			return "●"
		}
	}

	return " "
}

func (s DiffViewState) discussionGutterMarker(discussions []mr.Discussion) string {
	if len(discussions) == 0 {
		return " "
	}

	if s.emoji.Enabled {
		return "💬"
	}

	return "○"
}

func (s DiffViewState) isActiveDraftRangeRow(index int) bool {
	if s.rangeStart < 0 {
		return false
	}

	start, end := s.rangeStart, s.diffCursor
	if start > end {
		start, end = end, start
	}

	return index >= start && index <= end
}

func (s DiffViewState) cursorRow() (mr.ChangedFile, mr.DiffRow, bool) {
	if len(s.diffFiles) <= s.selectedFile {
		return mr.ChangedFile{}, mr.DiffRow{}, false
	}

	file := s.diffFiles[s.selectedFile]
	if s.diffCursor >= len(file.Diff) {
		return mr.ChangedFile{}, mr.DiffRow{}, false
	}

	return file, file.Diff[s.diffCursor], true
}

func (s DiffViewState) discussionsAtCursor() []mr.Discussion {
	file, row, ok := s.cursorRow()
	if !ok || row.NewLine == 0 {
		return nil
	}

	var result []mr.Discussion

	for _, discussion := range s.diffDiscussions {
		if discussion.Position != nil && discussion.Position.NewPath == file.Path && discussion.Position.NewLine == row.NewLine {
			result = append(result, discussion)
		}
	}

	return result
}

func (s DiffViewState) threadAtCursor() (*mr.Discussion, *mr.DraftComment) {
	discussions := s.discussionsAtCursor()
	if len(discussions) > 0 {
		idx := clamp(s.threadPanelCursor, 0, len(discussions)-1)
		discussion := discussions[idx]

		return &discussion, nil
	}

	file, row, ok := s.cursorRow()
	if !ok || row.NewLine == 0 {
		return nil, nil
	}

	for i := range s.diffDrafts {
		draft := &s.diffDrafts[i]
		if draft.Position != nil && draft.Position.NewPath == file.Path && draft.Position.NewLine == row.NewLine {
			return nil, draft
		}
	}

	return nil, nil
}

func (s DiffViewState) threadPanelLines(discussion *mr.Discussion, draft *mr.DraftComment, total int, width int) []string {
	sep := strings.Repeat("─", max(4, width))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	lines := []string{sep}

	if discussion != nil {
		header := "Discussion"
		if total > 1 {
			header = fmt.Sprintf("Discussion [%d/%d  [/]: switch]", s.threadPanelCursor+1, total)
		}

		if discussion.Resolved {
			header = "✅ " + header + " (resolved)"
		}

		lines = append(lines, renderDiscussionBlock(*discussion, header, "  ", true, true)...)
	} else if draft != nil {
		lines = append(lines, dimStyle.Render("📝 Draft"))
		lines = append(lines, dimStyle.Render("  "+draft.Body))
	}

	return lines
}
