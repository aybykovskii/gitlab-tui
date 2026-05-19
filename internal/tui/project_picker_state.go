//nolint:mnd // Project picker page size is an intentional UI limit.
package tui

import (
	"fmt"
	"strings"
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
}

func NewProjectPickerState(recents []RecentProjectOption, loaders []AccountProjectLoader, staticProjects ...string) ProjectPickerState {
	s := ProjectPickerState{
		recentProjectOptions: recents,
		staticProjects:       staticProjects,
		loadProjects:         loaders,
		accountProjectStates: initialAccountProjectStates(loaders),
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

	lines := []string{"Projects", ""}
	if s.projectFilterActive || s.query != "" {
		lines = append(lines, "Filter: "+s.query, "")
	}

	if len(s.projectRows) == 0 {
		lines = append(lines, "No matching projects")
	}

	for i, row := range s.projectRows {
		prefix := "  "
		if i == s.selected && row.selectable {
			prefix = "> "
		}

		lines = append(lines, prefix+row.label)
	}

	lines = append(lines, "", "Enter/click: open  i: manual input  r: retry")

	return strings.Join(lines, "\n")
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
