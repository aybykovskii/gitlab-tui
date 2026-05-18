package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var body string

	switch m.mode {
	case ModeProjectSelect, ModeProjectInput:
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderAppContextPane(), m.renderProjectPicker())
	case ModeSections:
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderProjectList(), m.renderSections())
	case ModeEntityList:
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderSectionsContext(), m.renderEntityListPane())
	case ModeFileDiff:
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderChangedFilesPane(), m.renderFileDiffPane())
	case ModeLabelSelect:
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderList(), m.renderLabelSelector())
	default:
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderList(), m.renderRight())
	}

	return lipgloss.JoinVertical(lipgloss.Left, body, m.renderKeyBar())
}
