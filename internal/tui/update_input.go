package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

//nolint:dupl // MR and issue edit flows intentionally mirror each other.
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
		description := m.editBuffer

		if m.editField == "title" {
			title = m.editBuffer
			item, _ := m.selectedItem()
			description = item.Description
		}

		m.editInput = false
		m.editBuffer = ""
		m.editTitle = ""
		item, ok := m.selectedItem()

		if !ok || m.editMR == nil {
			return m, nil
		}

		callback := m.editMR
		iid := item.IID

		return m, func() tea.Msg {
			err := callback(iid, title, description)
			return editMRFinishedMsg{iid: iid, title: title, description: description, err: err}
		}
	}

	return m, nil
}

//nolint:dupl // Issue and MR edit flows intentionally mirror each other.
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
		description := m.editBuffer

		if m.editField == "title" {
			title = m.editBuffer
			item, _ := m.selectedIssue()
			description = item.Description
		}

		m.editInput = false
		m.editBuffer = ""
		m.editTitle = ""
		item, ok := m.selectedIssue()

		if !ok || m.editIssue == nil {
			return m, nil
		}

		callback := m.editIssue
		iid := item.IID

		return m, func() tea.Msg {
			err := callback(iid, title, description)
			return editIssueFinishedMsg{iid: iid, title: title, description: description, err: err}
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

	callback := m.assignSelfIssue

	assignees := append([]string(nil), item.Assignees...)

	if assigned {
		callback = m.unassignSelfIssue
		assignees = nil

		for _, assignee := range item.Assignees {
			if assignee != "me" {
				assignees = append(assignees, assignee)
			}
		}
	} else {
		assignees = append(assignees, "me")
	}

	if callback == nil {
		return m, nil
	}

	iid := item.IID

	return m, func() tea.Msg {
		err := callback(iid)
		return issueAssigneeFinishedMsg{iid: iid, assignees: assignees, err: err}
	}
}

func (m Model) closeOrReopenIssueCommand() (Model, tea.Cmd) {
	item, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}

	state := "closed"
	callback := m.closeIssue

	if item.State == "closed" {
		state = "opened"
		callback = m.reopenIssue
	}

	if callback == nil {
		return m, nil
	}

	iid := item.IID

	return m, func() tea.Msg {
		err := callback(iid)
		return issueStateFinishedMsg{iid: iid, state: state, err: err}
	}
}

//nolint:dupl // Issue and MR comment flows intentionally mirror each other.
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

		callback := m.postIssueComment
		iid := item.IID

		return m, func() tea.Msg {
			err := callback(iid, body)
			return mrCommentFinishedMsg{iid: iid, err: err}
		}
	}

	return m, nil
}

//nolint:dupl // MR and issue comment flows intentionally mirror each other.
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

		callback := m.postMRComment
		iid := item.IID

		return m, func() tea.Msg {
			err := callback(iid, body)
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

	m.EntityListState.query = ""
}
