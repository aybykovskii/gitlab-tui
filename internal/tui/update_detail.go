package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

//nolint:gocyclo // Detail key handler maps many UI shortcuts explicitly.
func (m Model) updateDetailKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.globals.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.globals.Back):
		return m.handleBack()

	case msg.String() == "/":
		m.focus = FocusFilter

	case msg.String() == "r":
		if m.projectError && m.projectPath != "" {
			return m.openProjectCommand(m.projectPath)
		}

		return m, m.refreshCommand()

	case msg.String() == "m":
		if m.mode == ModeDetail {
			m.openCommentInput()
		}

	case msg.String() == "c":
		if m.mode == ModeDetail && m.section == SectionIssues {
			return m.closeOrReopenIssueCommand()
		}

	case msg.String() == "a":
		if m.mode == ModeDetail && m.section == SectionIssues {
			return m.assignOrUnassignIssueCommand()
		}

	case msg.String() == "A":
		if m.mode == ModeDetail {
			return m.approveMRCommand()
		}

	case msg.String() == "M":
		if m.mode == ModeDetail {
			return m.mergeMRCommand()
		}

	case msg.String() == "o":
		if m.mode == ModeDetail {
			return m.openURLCommand()
		}

	case msg.String() == "e":
		if m.mode == ModeDetail {
			m.openEditInput()
		}

	case msg.String() == "l":
		if m.mode == ModeDetail {
			return m.openLabelSelector()
		}

	case msg.String() == "d":
		if m.mode == ModeDetail && m.activeTab != TabDiscussions {
			return m.toggleDraftMRCommand()
		}

	case msg.String() == "tab":
		if m.mode == ModeDetail {
			return m.cycleTab()
		}

	case msg.String() == "up" || msg.String() == "k":
		if m.mode == ModeDetail {
			m.rightTop = max(0, m.rightTop-1)
		} else {
			m.moveSelection(-1)
		}

	case msg.String() == "down" || msg.String() == "j":
		if m.mode == ModeDetail {
			m.rightTop = max(0, m.rightTop+1)
		} else {
			m.moveSelection(1)
		}

	case msg.String() == "enter":
		if m.mode == ModeDetail && m.activeTab == TabFiles {
			return m.openFileDiff()
		}
	}

	return m, nil
}

func (m Model) handleBack() (Model, tea.Cmd) {
	if m.projectError || (m.projectPath != "" && len(m.items) == 0) {
		m.errorMessage = ""
		m.returnToProjectPicker()

		return m, nil
	}

	if m.mode == ModeDetail {
		m.mode = ModeEntityList
		m.focus = FocusDetail
		m.rightTop = 0
	}

	return m, nil
}

func (m *Model) openCommentInput() {
	if m.section == SectionIssues {
		m.issueCommentInput = true
		m.issueCommentBuffer = ""
		m.issueCommentError = ""
	} else {
		m.mrCommentInput = true
		m.mrCommentBuffer = ""
		m.mrCommentError = ""
	}
}

func (m Model) approveMRCommand() (Model, tea.Cmd) {
	item, ok := m.selectedItem()
	if !ok || m.approveMR == nil {
		return m, nil
	}

	callback := m.approveMR
	iid := item.IID

	return m, func() tea.Msg {
		err := callback(iid)
		return approveMRFinishedMsg{iid: iid, err: err}
	}
}

func (m Model) mergeMRCommand() (Model, tea.Cmd) {
	if !m.mergeConfirmPending {
		m.mergeConfirmPending = true
		return m, nil
	}

	item, ok := m.selectedItem()
	if !ok || m.mergeMR == nil {
		m.mergeConfirmPending = false
		return m, nil
	}

	callback := m.mergeMR
	iid := item.IID
	m.mergeConfirmPending = false

	return m, func() tea.Msg {
		err := callback(iid)
		return mergeMRFinishedMsg{iid: iid, err: err}
	}
}

func (m Model) openURLCommand() (Model, tea.Cmd) {
	if m.openURL == nil {
		return m, nil
	}

	var webURL string

	if m.section == SectionIssues {
		item, ok := m.selectedIssue()
		if !ok || item.WebURL == "" {
			return m, nil
		}

		webURL = item.WebURL
	} else {
		item, ok := m.selectedItem()
		if !ok || item.WebURL == "" {
			return m, nil
		}

		webURL = item.WebURL
	}

	callback := m.openURL
	url := webURL

	return m, func() tea.Msg {
		err := callback(url)
		return openURLMsg{url: url, err: err}
	}
}

func (m *Model) openEditInput() {
	if m.section == SectionIssues {
		item, ok := m.selectedIssue()
		if !ok {
			return
		}

		m.editInput = true
		m.editField = "title"
		m.editBuffer = item.Title
		m.editTitle = ""
	} else {
		item, ok := m.selectedItem()
		if !ok {
			return
		}

		m.editInput = true
		m.editField = "title"
		m.editBuffer = item.Title
		m.editTitle = ""
	}
}

func (m Model) openLabelSelector() (Model, tea.Cmd) {
	if m.section == SectionIssues {
		item, ok := m.selectedIssue()
		if !ok {
			return m, nil
		}

		m.mode = ModeLabelSelect
		m.labelCursor = 0
		pending := make([]string, len(item.Labels))
		copy(pending, item.Labels)
		m.labelPending = pending

		return m, nil
	}

	if m.activeTab != TabSummary {
		return m, nil
	}

	item, ok := m.selectedItem()
	if !ok {
		return m, nil
	}

	m.mode = ModeLabelSelect
	m.labelCursor = 0
	pending := make([]string, len(item.Labels))
	copy(pending, item.Labels)
	m.labelPending = pending

	return m, nil
}

func (m Model) toggleDraftMRCommand() (Model, tea.Cmd) {
	item, ok := m.selectedItem()
	if !ok {
		return m, nil
	}

	prev := item.Draft

	for i := range m.items {
		if m.items[i].IID == item.IID {
			m.items[i].Draft = !prev
			break
		}
	}

	if m.toggleDraftMR == nil {
		return m, nil
	}

	callback := m.toggleDraftMR
	iid := item.IID

	return m, func() tea.Msg {
		err := callback(iid)
		return toggleDraftFinishedMsg{iid: iid, prev: prev, err: err}
	}
}

func (m Model) cycleTab() (Model, tea.Cmd) {
	if m.section == SectionIssues {
		m.activeTab = (m.activeTab + 1) % (TabDiscussions + 1)
		return m, m.loadIssueDiscussionsCommand()
	}

	m.activeTab = (m.activeTab + 1) % (TabReview + 1)

	return m.onTabEntered()
}

func (m Model) openFileDiff() (Model, tea.Cmd) {
	item, ok := m.selectedItem()
	if !ok {
		return m, nil
	}

	files, loaded := m.changedFiles[item.IID]
	if !loaded || len(files) == 0 {
		return m, nil
	}

	m.mode = ModeFileDiff
	m.fileDiffReturnTab = TabFiles
	m.selectedFile = 0
	m.fileDiffTop = 0
	m.diffCursor = 0
	m.threadPanelCursor = 0

	return m, m.ensureDiscussionsLoaded(item.IID)
}
