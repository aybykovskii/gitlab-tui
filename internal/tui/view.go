package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var body string
	if m.mode == ModeProjectSelect || m.mode == ModeProjectInput {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderAppContextPane(), m.renderProjectPicker())
	} else if m.mode == ModeSections {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderProjectList(), m.renderSections())
	} else if m.mode == ModeEntityList {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderSectionsContext(), m.renderEntityListPane())
	} else if m.mode == ModeFileDiff {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderChangedFilesPane(), m.renderFileDiffPane())
	} else if m.mode == ModeLabelSelect {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderList(), m.renderLabelSelector())
	} else {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderList(), m.renderRight())
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, m.renderKeyBar())
}
