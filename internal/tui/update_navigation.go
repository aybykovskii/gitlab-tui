package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateProjectSelect(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.projectFilterActive {
		switch msg.Type {
		case tea.KeyEsc:
			m.query = ""
			m.projectFilterActive = false
			m.rebuildProjectRows()
			m.selected = m.nearestSelectable(0)

			return m, nil
		case tea.KeyBackspace:
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.rebuildProjectRows()
				m.selected = m.nearestSelectable(0)
			}

			return m, nil
		case tea.KeyRunes:
			m.query += msg.String()
			m.rebuildProjectRows()
			m.selected = m.nearestSelectable(0)

			return m, nil
		}
	}

	switch {
	case key.Matches(msg, m.projectListKeys.Filter):
		m.projectFilterActive = true
		m.query = ""
		m.rebuildProjectRows()
		m.selected = m.nearestSelectable(0)
	case key.Matches(msg, m.globals.Back):
		m.query = ""
		m.projectFilterActive = false
		m.rebuildProjectRows()
		m.selected = m.nearestSelectable(0)
	case key.Matches(msg, m.projectListKeys.Up):
		m.selected = m.nextSelectable(m.selected, -1)
	case key.Matches(msg, m.projectListKeys.Down):
		m.selected = m.nextSelectable(m.selected, 1)
	case key.Matches(msg, m.projectListKeys.Open):
		if project, ok := m.selectedProject(); ok {
			return m.selectProject(project)
		}
	case key.Matches(msg, m.projectListKeys.Retry):
		return m, m.retryFailedProjectLoads()
	case key.Matches(msg, m.projectListKeys.Input):
		m.mode = ModeProjectInput
		m.focus = FocusFilter
		m.projectInput = ""
	}

	return m, nil
}

func (m Model) updateProjectInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if trimmed := strings.TrimSpace(m.projectInput); trimmed != "" {
			return m.selectProject(trimmed)
		}
	case tea.KeyBackspace:
		if len(m.projectInput) > 0 {
			m.projectInput = m.projectInput[:len(m.projectInput)-1]
		}
	case tea.KeyRunes:
		m.projectInput += msg.String()
	}

	return m, nil
}

func (m Model) updateSections(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.sectionCursor = clamp(m.sectionCursor-1, 0, len(tuiSections)-1)
	case "down", "j":
		m.sectionCursor = clamp(m.sectionCursor+1, 0, len(tuiSections)-1)
	case "enter":
		sec := tuiSections[m.sectionCursor]
		if sec.available && sec.id == SectionMergeRequests {
			m.section = SectionMergeRequests
			if m.projectLoaded {
				m.mode = ModeEntityList
				m.focus = FocusDetail

				return m, nil
			}

			return m.openProjectCommand(m.projectPath)
		}

		if sec.available && sec.id == SectionIssues {
			m.section = SectionIssues
			m.mode = ModeEntityList
			m.focus = FocusDetail

			return m, m.loadIssuesCommand()
		}
	case "esc", "backspace":
		m.returnToProjectPicker()
	}

	return m, nil
}

func (m Model) updateFilter(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyEnter:
		m.focus = FocusDetail
	case tea.KeyBackspace:
		if len(m.query) > 0 {
			m.query = m.query[:len(m.query)-1]
			m.selected = m.clampEntitySelection(m.selected)
		}
	case tea.KeyRunes:
		m.query += msg.String()
		m.selected = m.clampEntitySelection(m.selected)
	}

	return m, nil
}

func (m Model) updateEntityList(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.moveSelection(-1)
	case "down", "j":
		m.moveSelection(1)
	case "enter":
		m.mode = ModeDetail
		m.focus = FocusDetail
		if m.section == SectionIssues {
			m.IssueDetailState.activeTab = TabSummary
		} else {
			m.activeTab = TabSummary
		}
	case "esc", "backspace":
		if m.projectError || (m.projectPath != "" && len(m.items) == 0) {
			m.errorMessage = ""
			m.returnToProjectPicker()
		} else {
			m.mode = ModeSections
		}
	case "/":
		m.focus = FocusFilter
	case "r":
		if m.projectError && m.projectPath != "" {
			return m.openProjectCommand(m.projectPath)
		}

		return m, m.refreshCommand()
	case "s":
		if m.section == SectionIssues {
			m.cycleIssueState()
			return m, m.loadIssuesCommand()
		}
	}

	return m, nil
}
