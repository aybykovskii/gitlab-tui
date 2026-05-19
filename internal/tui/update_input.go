package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

//nolint:dupl // MR and issue edit flows intentionally mirror each other.
func (m Model) updateMREdit(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.editInput = false
		m.Reset()
		m.editTitle = ""
	case tea.KeyTab:
		if m.editField == "title" {
			m.editTitle = m.Value()
			item, _ := m.selectedItem()
			m.editField = "description"
			m.BeginWithValue(item.Description)
		}
	case tea.KeyEnter:
		title := m.editTitle
		description := m.Value()

		if m.editField == "title" {
			title = m.Value()
			item, _ := m.selectedItem()
			description = item.Description
		}

		m.editInput = false
		m.Reset()
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
	case tea.KeyBackspace:
		m.Backspace()
	case tea.KeyRunes, tea.KeySpace:
		m.Append(msg.String())
	}

	return m, nil
}

//nolint:dupl // Issue and MR edit flows intentionally mirror each other.
func (m Model) updateIssueEdit(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.editInput = false
		m.Reset()
		m.editTitle = ""
	case tea.KeyTab:
		if m.editField == "title" {
			m.editTitle = m.Value()
			item, _ := m.selectedIssue()
			m.editField = "description"
			m.BeginWithValue(item.Description)
		}
	case tea.KeyEnter:
		title := m.editTitle
		description := m.Value()

		if m.editField == "title" {
			title = m.Value()
			item, _ := m.selectedIssue()
			description = item.Description
		}

		m.editInput = false
		m.Reset()
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
	case tea.KeyBackspace:
		m.Backspace()
	case tea.KeyRunes, tea.KeySpace:
		m.Append(msg.String())
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
		m.Reset()
	case tea.KeyEnter:
		body := m.Value()
		m.issueCommentInput = false
		m.Reset()
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
	case tea.KeyBackspace:
		m.Backspace()
	case tea.KeyRunes, tea.KeySpace:
		m.Append(msg.String())
	}

	return m, nil
}

//nolint:dupl // MR and issue comment flows intentionally mirror each other.
func (m Model) updateMRCommentInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mrCommentInput = false
		m.Reset()
	case tea.KeyEnter:
		body := m.Value()
		m.mrCommentInput = false
		m.Reset()
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
	case tea.KeyBackspace:
		m.Backspace()
	case tea.KeyRunes, tea.KeySpace:
		m.Append(msg.String())
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
