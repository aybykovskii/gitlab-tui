package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) returnToProjectPicker() {
	m.projectPath = ""
	if len(m.projectList) > 0 {
		m.mode = ModeProjectSelect
		m.focus = FocusDetail

		return
	}

	m.mode = ModeProjectInput
	m.focus = FocusFilter
}

func (m Model) selectProject(path string) (Model, tea.Cmd) {
	m.projectPath = path
	m.mode = ModeSections
	m.focus = FocusDetail
	m.selected = 0
	m.listTop = 0
	m.rightTop = 0
	m.projectLoaded = false
	m.items = nil
	found := false

	for _, project := range m.projectList {
		if project == path {
			found = true
			break
		}
	}

	if !found {
		m.projectList = append([]string{path}, m.projectList...)
	}

	return m, nil
}

func (m Model) openProjectCommand(path string) (Model, tea.Cmd) {
	m.projectPath = path
	m.mode = ModeEntityList
	m.focus = FocusDetail
	m.selected = 0
	m.listTop = 0
	m.rightTop = 0

	if m.loadProject == nil {
		return m, nil
	}

	loadProject := m.loadProject

	return m, tea.Sequence(
		func() tea.Msg { return projectStartedMsg{path: path} },
		func() tea.Msg {
			data, err := loadProject(path)
			return projectFinishedMsg{path: path, data: data, err: err}
		},
	)
}

func (m Model) onTabEntered() (Model, tea.Cmd) {
	item, ok := m.selectedItem()
	if !ok {
		return m, nil
	}

	switch m.activeTab {
	case TabDiscussions:
		if _, loaded := m.discussions[item.IID]; loaded {
			return m, nil
		}

		if m.loadDiscussions == nil {
			return m, nil
		}

		m.discussionsLoading = true
		m.discussionsError = ""
		load := m.loadDiscussions
		iid := item.IID

		return m, tea.Sequence(
			func() tea.Msg { return discussionsStartedMsg{iid: iid} },
			func() tea.Msg {
				items, err := load(iid)
				return discussionsFinishedMsg{iid: iid, discussions: items, err: err}
			},
		)
	case TabFiles:
		if _, loaded := m.changedFiles[item.IID]; loaded {
			return m, nil
		}

		if m.loadFiles == nil {
			return m, nil
		}

		m.filesLoading = true
		m.filesError = ""
		load := m.loadFiles
		iid := item.IID

		return m, tea.Sequence(
			func() tea.Msg { return filesStartedMsg{iid: iid} },
			func() tea.Msg {
				files, err := load(iid)
				return filesFinishedMsg{iid: iid, files: files, err: err}
			},
		)
	}

	return m, nil
}

func (m Model) ensureDiscussionsLoaded(iid int) tea.Cmd {
	if _, loaded := m.discussions[iid]; loaded {
		return nil
	}

	if m.loadDiscussions == nil {
		return nil
	}

	load := m.loadDiscussions

	return tea.Sequence(
		func() tea.Msg { return discussionsStartedMsg{iid: iid} },
		func() tea.Msg {
			items, err := load(iid)
			return discussionsFinishedMsg{iid: iid, discussions: items, err: err}
		},
	)
}

func (m Model) refreshCommand() tea.Cmd {
	if m.section == SectionIssues {
		return m.loadIssuesCommand()
	}

	if m.refresh == nil || m.loading {
		return nil
	}

	refresh := m.refresh

	return tea.Sequence(
		func() tea.Msg { return refreshStartedMsg{} },
		func() tea.Msg {
			items, err := refresh()
			return refreshFinishedMsg{items: items, err: err}
		},
	)
}

func (m Model) loadIssuesCommand() tea.Cmd {
	if m.loadIssues == nil || m.loading {
		return nil
	}

	loadIssues := m.loadIssues
	state := m.issueState

	return tea.Sequence(
		func() tea.Msg { return refreshStartedMsg{} },
		func() tea.Msg {
			items, err := loadIssues(state, "")
			return issuesFinishedMsg{items: items, err: err}
		},
	)
}

func (m Model) loadIssueDiscussionsCommand() tea.Cmd {
	if m.activeTab != TabDiscussions || m.loadIssueDiscussions == nil {
		return nil
	}

	item, ok := m.selectedIssue()
	if !ok {
		return nil
	}

	load := m.loadIssueDiscussions
	iid := item.IID

	return func() tea.Msg {
		discussions, err := load(iid)
		return issueDiscussionsFinishedMsg{iid: iid, discussions: discussions, err: err}
	}
}

func (m Model) selectedProject() (string, bool) {
	if m.selected < 0 || m.selected >= len(m.projectRows) || !m.projectRows[m.selected].selectable {
		return "", false
	}

	return m.projectRows[m.selected].project, true
}

func (m Model) retryFailedProjectLoads() tea.Cmd {
	cmds := []tea.Cmd{}

	for _, loader := range m.loadProjects {
		if state := m.accountProjectStates[loader.ID]; state.err != "" {
			cmds = append(cmds, loadAccountProjectsCommand(loader))
		}
	}

	if len(cmds) == 1 {
		return cmds[0]
	}

	return tea.Batch(cmds...)
}

func (m *Model) rebuildProjectRows() {
	m.projectRows = nil
	if len(m.filteredRecentProjects()) > 0 {
		m.projectRows = append(m.projectRows, projectListRow{label: "Recent"}, projectListRow{})
		for _, recent := range m.filteredRecentProjects() {
			label := recent.Path
			if recent.Account != "" {
				label += " (" + recent.Account + ")"
			}

			m.projectRows = append(m.projectRows, projectListRow{project: recent.Path, label: label, selectable: true})
		}
	}

	for _, project := range m.projectList {
		if m.matchesProjectFilter(project) {
			m.projectRows = append(m.projectRows, projectListRow{project: project, label: project, selectable: true})
		}
	}

	if len(m.projectRows) > 0 && len(m.loadProjects) > 0 {
		m.projectRows = append(m.projectRows, projectListRow{})
	}

	for _, loader := range m.loadProjects {
		state := m.accountProjectStates[loader.ID]
		projects := filteredProjectPaths(state.projects[:min(len(state.projects), 15)], m.query)
		showStatus := !m.projectFilterActive && len(projects) == 0

		if len(projects) == 0 && !showStatus {
			continue
		}

		header := fmt.Sprintf("[%s]  %s", loader.ID, state.host)

		m.projectRows = append(m.projectRows, projectListRow{label: header})
		if state.loading && showStatus {
			m.projectRows = append(m.projectRows, projectListRow{label: "Loading…"})
			continue
		}

		if state.err != "" && showStatus {
			m.projectRows = append(m.projectRows, projectListRow{label: "Error: " + state.err + "  r: retry"})
			continue
		}

		for _, project := range projects {
			m.projectRows = append(m.projectRows, projectListRow{project: project, label: project, selectable: true})
		}
	}
}

func (m Model) filteredRecentProjects() []RecentProjectOption {
	projects := make([]RecentProjectOption, 0, len(m.recentProjectOptions))

	for _, recent := range m.recentProjectOptions {
		if m.matchesProjectFilter(recent.Path) {
			projects = append(projects, recent)
		}
	}

	return projects
}

func (m Model) matchesProjectFilter(project string) bool {
	if strings.TrimSpace(m.query) == "" {
		return true
	}

	return strings.Contains(strings.ToLower(project), strings.ToLower(m.query))
}
