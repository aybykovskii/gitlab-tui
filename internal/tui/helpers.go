//nolint:mnd,unparam // UI helper constants and generic clamp bounds are intentional.
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
	return m.EntityListState.filteredMRs()
}

func (m Model) filteredIssues() []issue.Issue {
	return m.EntityListState.filteredIssues()
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
