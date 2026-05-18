package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateMREdit(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.editInput = false
		m.editBuffer = ""
		m.editTitle = ""
	case tea.KeyBackspace:
		if len(m.editBuffer) > 0 {
			m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
		}
	case tea.KeyRunes, tea.KeySpace:
		m.editBuffer += msg.String()
	case tea.KeyTab:
		if m.editField == "title" {
			m.editTitle = m.editBuffer
			item, _ := m.selectedItem()
			m.editField = "description"
			m.editBuffer = item.Description
		}
	case tea.KeyEnter:
		title := m.editTitle
		desc := m.editBuffer
		if m.editField == "title" {
			title = m.editBuffer
			item, _ := m.selectedItem()
			desc = item.Description
		}
		m.editInput = false
		m.editBuffer = ""
		m.editTitle = ""
		item, ok := m.selectedItem()
		if !ok || m.editMR == nil {
			return m, nil
		}
		fn := m.editMR
		iid := item.IID
		return m, func() tea.Msg {
			err := fn(iid, title, desc)
			return editMRFinishedMsg{iid: iid, title: title, description: desc, err: err}
		}
	}
	return m, nil
}

func (m Model) updateIssueEdit(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.editInput = false
		m.editBuffer = ""
		m.editTitle = ""
	case tea.KeyBackspace:
		if len(m.editBuffer) > 0 {
			m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
		}
	case tea.KeyRunes, tea.KeySpace:
		m.editBuffer += msg.String()
	case tea.KeyTab:
		if m.editField == "title" {
			m.editTitle = m.editBuffer
			item, _ := m.selectedIssue()
			m.editField = "description"
			m.editBuffer = item.Description
		}
	case tea.KeyEnter:
		title := m.editTitle
		desc := m.editBuffer
		if m.editField == "title" {
			title = m.editBuffer
			item, _ := m.selectedIssue()
			desc = item.Description
		}
		m.editInput = false
		m.editBuffer = ""
		m.editTitle = ""
		item, ok := m.selectedIssue()
		if !ok || m.editIssue == nil {
			return m, nil
		}
		fn := m.editIssue
		iid := item.IID
		return m, func() tea.Msg {
			err := fn(iid, title, desc)
			return editIssueFinishedMsg{iid: iid, title: title, description: desc, err: err}
		}
	}
	return m, nil
}

func (m Model) assignOrUnassignIssueCommand() (Model, tea.Cmd) {
	item, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	assigned := false
	for _, assignee := range item.Assignees {
		if assignee == "me" {
			assigned = true
			break
		}
	}
	fn := m.assignSelfIssue
	assignees := append([]string(nil), item.Assignees...)
	if assigned {
		fn = m.unassignSelfIssue
		assignees = nil
		for _, assignee := range item.Assignees {
			if assignee != "me" {
				assignees = append(assignees, assignee)
			}
		}
	} else {
		assignees = append(assignees, "me")
	}
	if fn == nil {
		return m, nil
	}
	iid := item.IID
	return m, func() tea.Msg {
		err := fn(iid)
		return issueAssigneeFinishedMsg{iid: iid, assignees: assignees, err: err}
	}
}

func (m Model) closeOrReopenIssueCommand() (Model, tea.Cmd) {
	item, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	state := "closed"
	fn := m.closeIssue
	if item.State == "closed" {
		state = "opened"
		fn = m.reopenIssue
	}
	if fn == nil {
		return m, nil
	}
	iid := item.IID
	return m, func() tea.Msg {
		err := fn(iid)
		return issueStateFinishedMsg{iid: iid, state: state, err: err}
	}
}

func (m Model) updateIssueCommentInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.issueCommentInput = false
		m.issueCommentBuffer = ""
	case tea.KeyBackspace:
		if len(m.issueCommentBuffer) > 0 {
			m.issueCommentBuffer = m.issueCommentBuffer[:len(m.issueCommentBuffer)-1]
		}
	case tea.KeyRunes, tea.KeySpace:
		m.issueCommentBuffer += msg.String()
	case tea.KeyEnter:
		body := m.issueCommentBuffer
		m.issueCommentInput = false
		m.issueCommentBuffer = ""
		item, ok := m.selectedIssue()
		if !ok || m.postIssueComment == nil {
			return m, nil
		}
		fn := m.postIssueComment
		iid := item.IID
		return m, func() tea.Msg {
			err := fn(iid, body)
			return mrCommentFinishedMsg{iid: iid, err: err}
		}
	}
	return m, nil
}

func (m Model) updateMRCommentInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mrCommentInput = false
		m.mrCommentBuffer = ""
	case tea.KeyBackspace:
		if len(m.mrCommentBuffer) > 0 {
			m.mrCommentBuffer = m.mrCommentBuffer[:len(m.mrCommentBuffer)-1]
		}
	case tea.KeyRunes, tea.KeySpace:
		m.mrCommentBuffer += msg.String()
	case tea.KeyEnter:
		body := m.mrCommentBuffer
		m.mrCommentInput = false
		m.mrCommentBuffer = ""
		item, ok := m.selectedItem()
		if !ok || m.postMRComment == nil {
			return m, nil
		}
		fn := m.postMRComment
		iid := item.IID
		return m, func() tea.Msg {
			err := fn(iid, body)
			return mrCommentFinishedMsg{iid: iid, err: err}
		}
	}
	return m, nil
}

func (m Model) issueStateLabel() string {
	if m.issueState == "" {
		return "all"
	}
	return m.issueState
}

func (m *Model) cycleIssueState() {
	switch m.issueState {
	case "opened":
		m.issueState = "closed"
	case "closed":
		m.issueState = ""
	default:
		m.issueState = "opened"
	}
	m.query = ""
}
