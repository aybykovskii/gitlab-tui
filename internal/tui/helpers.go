package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func toggleStringSlice(slice []string, val string) []string {
	for i, value := range slice {
		if value == val {
			result := make([]string, 0, len(slice)-1)
			result = append(result, slice[:i]...)
			result = append(result, slice[i+1:]...)

			return result
		}
	}

	return append(slice, val)
}

func initialAccountProjectStates(loaders []AccountProjectLoader) map[string]accountProjectState {
	states := map[string]accountProjectState{}
	for _, loader := range loaders {
		states[loader.ID] = accountProjectState{host: loader.Host, loading: true}
	}

	return states
}

func loadAccountProjectsCommand(loader AccountProjectLoader) tea.Cmd {
	return func() tea.Msg {
		projects, err := loader.Load()
		return accountProjectsFinishedMsg{accountID: loader.ID, projects: projects, err: err}
	}
}

func buildRecentProjectOptions(recents []string, recentProjects []RecentProjectOption) []RecentProjectOption {
	if len(recentProjects) > 0 {
		return recentProjects
	}

	options := make([]RecentProjectOption, 0, len(recents))
	for _, recent := range recents {
		options = append(options, RecentProjectOption{Path: recent})
	}

	return options
}

func buildProjectList(opened string, recents []string, projects []string) []string {
	seen := map[string]bool{}
	list := []string{}

	candidates := []string{}
	if opened != "" {
		candidates = append(candidates, opened)
	}

	candidates = append(candidates, recents...)
	candidates = append(candidates, projects...)

	for _, project := range candidates {
		if project == "" || seen[project] {
			continue
		}

		seen[project] = true

		list = append(list, project)
	}

	return list
}

func filteredProjectPaths(projects []string, query string) []string {
	if strings.TrimSpace(query) == "" {
		return projects
	}

	filtered := make([]string, 0, len(projects))

	needle := strings.ToLower(query)
	for _, project := range projects {
		if strings.Contains(strings.ToLower(project), needle) {
			filtered = append(filtered, project)
		}
	}

	return filtered
}

func (m Model) nextSelectable(from int, delta int) int {
	if len(m.projectRows) == 0 {
		return 0
	}

	for i := clamp(from+delta, 0, len(m.projectRows)-1); i >= 0 && i < len(m.projectRows); i += delta {
		if m.projectRows[i].selectable {
			return i
		}

		if i == 0 && delta < 0 || i == len(m.projectRows)-1 && delta > 0 {
			break
		}
	}

	return from
}

func (m Model) nearestSelectable(index int) int {
	if len(m.projectRows) == 0 {
		return 0
	}

	if index >= 0 && index < len(m.projectRows) && m.projectRows[index].selectable {
		return index
	}

	if next := m.nextSelectable(index, 1); next != index {
		return next
	}

	return m.nextSelectable(index, -1)
}

func oneLinePreview(text string) string {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return ""
	}

	preview := strings.Join(fields, " ")
	if len(preview) > 60 {
		return preview[:57] + "..."
	}

	return preview
}

func (m Model) filtered() []mr.MergeRequest {
	return mr.Filter(m.items, m.query)
}

func (m Model) filteredIssues() []issue.Issue {
	query := strings.ToLower(strings.TrimSpace(m.query))
	if query == "" {
		return m.issueItems
	}

	filtered := make([]issue.Issue, 0, len(m.issueItems))

	for _, item := range m.issueItems {
		text := strings.ToLower(item.Title + " " + item.Author)
		if strings.Contains(text, query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func (m Model) clampEntitySelection(selected int) int {
	if m.section == SectionIssues {
		return clampSelection(selected, len(m.filteredIssues()))
	}

	return clampSelection(selected, len(m.filtered()))
}

func clampSelection(selected int, count int) int {
	if count <= 0 {
		return 0
	}

	return clamp(selected, 0, count-1)
}

func clamp(v int, minValue int, maxValue int) int {
	if v < minValue {
		return minValue
	}

	if v > maxValue {
		return maxValue
	}

	return v
}
