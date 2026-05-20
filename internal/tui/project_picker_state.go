//nolint:mnd // Project picker page size is an intentional UI limit.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

type ProjectPickerState struct {
	recentProjectOptions []RecentProjectOption
	staticProjects       []string
	projectRows          []projectListRow
	loadProjects         []AccountProjectLoader
	accountProjectStates map[string]accountProjectState
	projectInput         string
	projectFilterActive  bool
	selected             int
	query                string
	rowList              list.Model
}

func NewProjectPickerState(recents []RecentProjectOption, loaders []AccountProjectLoader, staticProjects ...string) ProjectPickerState {
	s := ProjectPickerState{
		recentProjectOptions: recents,
		staticProjects:       staticProjects,
		loadProjects:         loaders,
		accountProjectStates: initialAccountProjectStates(loaders),
		rowList:              newCompactFancyList("Projects", newProjectPickerDelegate()),
	}
	s.rebuildRows()

	return s
}

func (s ProjectPickerState) View(layout LayoutState) string {
	if layout.Mode == ModeProjectInput {
		return strings.Join([]string{
			"Open GitLab project",
			"",
			"Project path:",
			s.projectInput,
		}, "\n")
	}

	var header []string
	if s.projectFilterActive || s.query != "" {
		header = append(header, "Filter: "+s.query)
	}

	footer := []string{"", "Enter/click: open  i: manual input  r: retry"}

	if len(s.projectRows) == 0 {
		return strings.Join(append(append(header, "Projects", "No matching projects"), footer...), "\n")
	}

	s.rowList.Select(s.selected)

	height := layout.Height
	if height == 0 {
		height = 40
	}

	width := layout.Width
	if width == 0 {
		width = 80
	}

	listH := max(1, height-len(header)-len(footer))
	s.rowList.SetSize(width-4, listH)

	headerStr := ""
	if len(header) > 0 {
		headerStr = strings.Join(header, "\n") + "\n"
	}

	return headerStr + s.rowList.View() + strings.Join(footer, "\n")
}

func (s *ProjectPickerState) rebuildRows() {
	s.projectRows = nil

	if filtered := s.filteredRecents(); len(filtered) > 0 {
		s.projectRows = append(s.projectRows, projectListRow{label: "Recent"}, projectListRow{})

		for _, recent := range filtered {
			label := recent.Path
			if recent.Account != "" {
				label += " (" + recent.Account + ")"
			}

			s.projectRows = append(s.projectRows, projectListRow{project: recent.Path, label: label, selectable: true})
		}
	}

	for _, project := range s.staticProjects {
		if s.matchesFilter(project) {
			s.projectRows = append(s.projectRows, projectListRow{project: project, label: project, selectable: true})
		}
	}

	if len(s.projectRows) > 0 && len(s.loadProjects) > 0 {
		s.projectRows = append(s.projectRows, projectListRow{})
	}

	for _, loader := range s.loadProjects {
		state := s.accountProjectStates[loader.ID]
		projects := filteredProjectPaths(state.projects[:min(len(state.projects), 15)], s.query)
		showStatus := !s.projectFilterActive && len(projects) == 0

		if len(projects) == 0 && !showStatus {
			continue
		}

		header := fmt.Sprintf("[%s]  %s", loader.ID, state.host)
		s.projectRows = append(s.projectRows, projectListRow{label: header})

		if state.loading && showStatus {
			s.projectRows = append(s.projectRows, projectListRow{label: "Loading…"})
			continue
		}

		if state.err != "" && showStatus {
			s.projectRows = append(s.projectRows, projectListRow{label: "Error: " + state.err + "  r: retry"})
			continue
		}

		for _, project := range projects {
			s.projectRows = append(s.projectRows, projectListRow{project: project, label: project, selectable: true})
		}
	}

	items := make([]list.Item, len(s.projectRows))
	for i, row := range s.projectRows {
		items[i] = row
	}

	_ = s.rowList.SetItems(items)
}

func (s ProjectPickerState) selectedProject() (string, bool) {
	if s.selected < 0 || s.selected >= len(s.projectRows) || !s.projectRows[s.selected].selectable {
		return "", false
	}

	return s.projectRows[s.selected].project, true
}

func (s ProjectPickerState) nextSelectable(from int, delta int) int {
	if len(s.projectRows) == 0 {
		return 0
	}

	for i := clamp(from+delta, 0, len(s.projectRows)-1); i >= 0 && i < len(s.projectRows); i += delta {
		if s.projectRows[i].selectable {
			return i
		}

		if i == 0 && delta < 0 || i == len(s.projectRows)-1 && delta > 0 {
			break
		}
	}

	return from
}

func (s ProjectPickerState) nearestSelectable(index int) int {
	if len(s.projectRows) == 0 {
		return 0
	}

	if index >= 0 && index < len(s.projectRows) && s.projectRows[index].selectable {
		return index
	}

	if next := s.nextSelectable(index, 1); next != index {
		return next
	}

	return s.nextSelectable(index, -1)
}

func (s ProjectPickerState) filteredRecents() []RecentProjectOption {
	projects := make([]RecentProjectOption, 0, len(s.recentProjectOptions))

	for _, recent := range s.recentProjectOptions {
		if s.matchesFilter(recent.Path) {
			projects = append(projects, recent)
		}
	}

	return projects
}

func (s ProjectPickerState) matchesFilter(project string) bool {
	if strings.TrimSpace(s.query) == "" {
		return true
	}

	return strings.Contains(strings.ToLower(project), strings.ToLower(s.query))
}
