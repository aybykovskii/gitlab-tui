//nolint:mnd,prealloc // UI sizing and small slice growth are intentional.
package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

func (m Model) renderProjectList() string {
	width := m.leftWidth()
	style := paneStyle(width, m.paneHeight(), false)
	itemStyles := list.NewDefaultItemStyles()
	titleStyle := list.DefaultStyles().Title

	lines := []string{titleStyle.Render("Projects"), ""}
	for _, project := range m.projectList {
		if project == m.projectPath {
			lines = append(lines, itemStyles.SelectedTitle.Render(project))
		} else {
			lines = append(lines, itemStyles.NormalTitle.Render(project))
		}
	}

	if len(m.projectList) == 0 {
		lines = append(lines, "  No projects")
	}

	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderSections() string {
	width := max(20, m.width-m.leftWidth())
	height := m.paneHeight()
	style := paneStyle(width, height, true)

	idx := m.sectionList.Index()
	m.sectionList.SetSize(width-4, max(1, height-2))
	content := m.sectionList.View()

	if idx >= 0 && idx < len(tuiSections) && !tuiSections[idx].available {
		content += "\n\nNot yet implemented"
	}

	return style.Render(content)
}

func (m Model) renderSectionsContext() string {
	width := m.leftWidth()
	height := m.paneHeight()
	style := paneStyle(width, height, false)

	m.sectionList.SetSize(width-4, max(1, height-2))

	return style.Render(m.sectionList.View())
}

func (m Model) renderAppContextPane() string {
	width := m.leftWidth()
	style := paneStyle(width, m.paneHeight(), false)

	return style.Render("gitlab-tui")
}

func (m Model) renderProjectPicker() string {
	width := max(20, m.width-m.leftWidth())
	height := m.paneHeight()
	style := paneStyle(width, height, true)

	return style.Render(m.ProjectPickerState.View(LayoutState{Width: width, Height: height, Mode: m.mode}))
}
