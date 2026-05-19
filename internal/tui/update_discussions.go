package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func (m Model) selectedIssue() (issue.Issue, bool) {
	items := m.filteredIssues()
	if len(items) == 0 {
		return issue.Issue{}, false
	}

	return items[clampSelection(m.EntityListState.selected, len(items))], true
}

func (m Model) focusedIssueDiscussion() (issue.Discussion, bool) {
	item, ok := m.selectedIssue()
	if !ok {
		return issue.Discussion{}, false
	}

	discussions := m.IssueDetailState.discussions[item.IID]
	if m.discussionCursor < 0 || m.discussionCursor >= len(discussions) {
		return issue.Discussion{}, false
	}

	return discussions[m.discussionCursor], true
}

func (m Model) focusedDiscussion() (mr.Discussion, bool) {
	item, ok := m.selectedItem()
	if !ok {
		return mr.Discussion{}, false
	}

	discussions := m.discussions[item.IID]
	if m.discussionCursor < 0 || m.discussionCursor >= len(discussions) {
		return mr.Discussion{}, false
	}

	return discussions[m.discussionCursor], true
}

func (m Model) updateIssueDiscussionsTab(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.replyInput {
		switch msg.Type {
		case tea.KeyEsc:
			m.replyInput = false
			m.Reset()
			m.replyDiscussionID = ""
		case tea.KeyEnter:
			m.replyInput = false
			m.Reset()
			m.replyDiscussionID = ""
		case tea.KeyBackspace:
			m.Backspace()
		case tea.KeyRunes, tea.KeySpace:
			m.Append(msg.String())
		}

		return m, nil
	}

	switch {
	case msg.String() == "j" || msg.String() == "down":
		if item, ok := m.selectedIssue(); ok {
			count := len(m.IssueDetailState.discussions[item.IID])
			m.discussionCursor = clamp(m.discussionCursor+1, 0, max(0, count-1))
		}
	case msg.String() == "k" || msg.String() == "up":
		m.discussionCursor = clamp(m.discussionCursor-1, 0, max(0, m.discussionCursor))
	case msg.String() == "r":
		if discussion, ok := m.focusedIssueDiscussion(); ok {
			m.replyInput = true
			m.replyDraft = false
			m.replyDiscussionID = discussion.ID
			m.Begin()
		}
	}

	return m, nil
}

//nolint:gocyclo,dupl // Discussion tab input flow mirrors file diff replies intentionally.
func (m Model) updateDiscussionsTab(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.replyInput {
		switch msg.Type {
		case tea.KeyEsc:
			m.replyInput = false
			m.Reset()
			m.replyDiscussionID = ""
		case tea.KeyEnter:
			body := m.Value()
			discussionID := m.replyDiscussionID
			isDraft := m.replyDraft
			m.replyInput = false
			m.Reset()
			m.replyDiscussionID = ""
			m.replyDraft = false
			item, ok := m.selectedItem()

			if !ok {
				return m, nil
			}

			iid := item.IID

			if isDraft {
				callback := m.draftReply
				if callback == nil {
					return m, nil
				}

				return m, func() tea.Msg {
					err := callback(iid, discussionID, body)
					return replyFinishedMsg{iid: iid, discussionID: discussionID, body: body, draft: true, err: err}
				}
			}

			callback := m.replyToDiscussion
			if callback == nil {
				return m, nil
			}

			return m, func() tea.Msg {
				err := callback(iid, discussionID, body)
				return replyFinishedMsg{iid: iid, discussionID: discussionID, body: body, draft: false, err: err}
			}
		case tea.KeyBackspace:
			m.Backspace()
		case tea.KeyRunes, tea.KeySpace:
			m.Append(msg.String())
		}

		return m, nil
	}

	item, ok := m.selectedItem()
	if !ok {
		return m, nil
	}

	discussions := m.discussions[item.IID]
	count := len(discussions)

	switch {
	case msg.String() == "j" || msg.String() == "down":
		m.discussionCursor = clamp(m.discussionCursor+1, 0, max(0, count-1))
	case msg.String() == "k" || msg.String() == "up":
		m.discussionCursor = clamp(m.discussionCursor-1, 0, max(0, count-1))
	case msg.String() == "r":
		if discussion, ok := m.focusedDiscussion(); ok {
			m.replyInput = true
			m.replyDraft = false
			m.replyDiscussionID = discussion.ID
			m.Begin()
		}
	case msg.String() == "d":
		if discussion, ok := m.focusedDiscussion(); ok {
			m.replyInput = true
			m.replyDraft = true
			m.replyDiscussionID = discussion.ID
			m.Begin()
		}
	case msg.String() == "x":
		if discussion, ok := m.focusedDiscussion(); ok {
			iid := item.IID
			discussionID := discussion.ID

			resolved := !discussion.Resolved
			if resolved {
				callback := m.resolveDiscussion
				if callback == nil {
					m.discussions[iid][m.discussionCursor].Resolved = true
					return m, nil
				}

				return m, func() tea.Msg {
					err := callback(iid, discussionID)
					return resolveFinishedMsg{iid: iid, discussionID: discussionID, resolved: true, err: err}
				}
			}

			callback := m.unresolveDiscussion
			if callback == nil {
				m.discussions[iid][m.discussionCursor].Resolved = false
				return m, nil
			}

			return m, func() tea.Msg {
				err := callback(iid, discussionID)
				return resolveFinishedMsg{iid: iid, discussionID: discussionID, resolved: false, err: err}
			}
		}
	case msg.String() == "tab":
		m.activeTab = (m.activeTab + 1) % (TabFiles + 1)
		return m.onTabEntered()
	case key.Matches(msg, m.globals.Quit):
		return m, tea.Quit
	}

	return m, nil
}
