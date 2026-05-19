package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateProjectSelect(msg tea.KeyMsg) (Model, tea.Cmd) {
	pp := &m.ProjectPickerState
	if pp.projectFilterActive {
		switch msg.Type {
		case tea.KeyEsc:
			pp.query = ""
			pp.projectFilterActive = false
			pp.rebuildRows()
			pp.selected = pp.nearestSelectable(0)

			return m, nil
		case tea.KeyBackspace:
			if len(pp.query) > 0 {
				pp.query = pp.query[:len(pp.query)-1]
				pp.rebuildRows()
				pp.selected = pp.nearestSelectable(0)
			}

			return m, nil
		case tea.KeyRunes:
			pp.query += msg.String()
			pp.rebuildRows()
			pp.selected = pp.nearestSelectable(0)

			return m, nil
		}
	}

	switch {
	case key.Matches(msg, m.projectListKeys.Filter):
		pp.projectFilterActive = true
		pp.query = ""
		pp.rebuildRows()
		pp.selected = pp.nearestSelectable(0)
	case key.Matches(msg, m.globals.Back):
		pp.query = ""
		pp.projectFilterActive = false
		pp.rebuildRows()
		pp.selected = pp.nearestSelectable(0)
	case key.Matches(msg, m.projectListKeys.Up):
		pp.selected = pp.nextSelectable(pp.selected, -1)
	case key.Matches(msg, m.projectListKeys.Down):
		pp.selected = pp.nextSelectable(pp.selected, 1)
	case key.Matches(msg, m.projectListKeys.Open):
		if project, ok := pp.selectedProject(); ok {
			return m.selectProject(project)
		}
	case key.Matches(msg, m.projectListKeys.Retry):
		return m, m.retryFailedProjectLoads()
	case key.Matches(msg, m.projectListKeys.Input):
		m.mode = ModeProjectInput
		m.focus = FocusFilter
		m.ProjectPickerState.projectInput = ""
	}

	return m, nil
}

func (m Model) updateProjectInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	pp := &m.ProjectPickerState

	switch msg.Type {
	case tea.KeyEnter:
		if trimmed := strings.TrimSpace(pp.projectInput); trimmed != "" {
			return m.selectProject(trimmed)
		}
	case tea.KeyBackspace:
		if len(pp.projectInput) > 0 {
			pp.projectInput = pp.projectInput[:len(pp.projectInput)-1]
		}
	case tea.KeyRunes:
		pp.projectInput += msg.String()
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
	el := &m.EntityListState

	switch msg.Type {
	case tea.KeyEsc, tea.KeyEnter:
		m.focus = FocusDetail
	case tea.KeyBackspace:
		if len(el.query) > 0 {
			el.query = el.query[:len(el.query)-1]
			el.selected = m.clampEntitySelection(el.selected)
		}
	case tea.KeyRunes:
		el.query += msg.String()
		el.selected = m.clampEntitySelection(el.selected)
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
