package tui

import (
	"strings"
)

func (m Model) renderProjectList() string {
	width := m.leftWidth()
	style := paneStyle(width, m.paneHeight(), false)
	lines := []string{"Projects", ""}
	for _, project := range m.projectList {
		prefix := "  "
		if project == m.projectPath {
			prefix = "> "
		}
		lines = append(lines, prefix+project)
	}
	if len(m.projectList) == 0 {
		lines = append(lines, "No projects")
	}
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderSections() string {
	width := max(20, m.width-m.leftWidth())
	style := paneStyle(width, m.paneHeight(), true)
	lines := []string{"Sections", ""}
	for i, sec := range tuiSections {
		prefix := "  "
		if i == m.sectionCursor {
			prefix = "> "
		}
		label := sec.label
		if !sec.available {
			label += " (soon)"
		}
		lines = append(lines, prefix+label)
	}
	if !tuiSections[m.sectionCursor].available {
		lines = append(lines, "", "Not yet implemented")
	}
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderSectionsContext() string {
	width := m.leftWidth()
	style := paneStyle(width, m.paneHeight(), false)
	lines := []string{"Sections", ""}
	for _, sec := range tuiSections {
		prefix := "  "
		if sec.id == m.section {
			prefix = "> "
		}
		lines = append(lines, prefix+sec.label)
	}
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderAppContextPane() string {
	width := m.leftWidth()
	style := paneStyle(width, m.paneHeight(), false)
	return style.Render("gitlab-tui")
}

func (m Model) renderProjectPicker() string {
	width := max(20, m.width-m.leftWidth())
	style := paneStyle(width, m.paneHeight(), true)
	if m.mode == ModeProjectInput {
		return style.Render(strings.Join([]string{
			"Open GitLab project",
			"",
			"Project path:",
			m.projectInput,
		}, "\n"))
	}

	lines := []string{"Projects", ""}
	if m.projectFilterActive || m.query != "" {
		lines = append(lines, "Filter: "+m.query, "")
	}
	if len(m.projectRows) == 0 {
		lines = append(lines, "No matching projects")
	}
	for i, row := range m.projectRows {
		prefix := "  "
		if i == m.selected && row.selectable {
			prefix = "> "
		}
		lines = append(lines, prefix+row.label)
	}
	lines = append(lines, "", "Enter/click: open  i: manual input  r: retry")
	return style.Render(strings.Join(lines, "\n"))
}
