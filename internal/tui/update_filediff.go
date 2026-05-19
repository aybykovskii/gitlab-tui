package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func (m Model) updateFileDiff(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.replyInput {
		return m.updateFileDiffReplyInput(msg)
	}

	if m.commentInput {
		return m.updateFileDiffCommentInput(msg)
	}

	return m.updateFileDiffKeys(msg)
}

//nolint:dupl // File diff reply flow mirrors discussion replies intentionally.
func (m Model) updateFileDiffReplyInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.replyInput = false
		m.replyBuffer = ""
		m.replyDiscussionID = ""

		return m, nil

	case tea.KeyBackspace:
		if len(m.replyBuffer) > 0 {
			m.replyBuffer = m.replyBuffer[:len(m.replyBuffer)-1]
		}

		return m, nil

	case tea.KeyRunes, tea.KeySpace:
		m.replyBuffer += msg.String()
		return m, nil

	case tea.KeyEnter:
		body := m.replyBuffer
		discussionID := m.replyDiscussionID
		isDraft := m.replyDraft

		m.replyInput = false
		m.replyBuffer = ""
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
	}

	return m, nil
}

func (m Model) updateFileDiffCommentInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.commentInput = false
		m.commentBuffer = ""

		return m, nil

	case tea.KeyBackspace:
		if len(m.commentBuffer) > 0 {
			m.commentBuffer = m.commentBuffer[:len(m.commentBuffer)-1]
		}

		return m, nil

	case tea.KeyRunes, tea.KeySpace:
		m.commentBuffer += msg.String()
		return m, nil

	case tea.KeyEnter:
		body := m.commentBuffer
		instant := m.commentInstant

		m.commentInput = false
		m.commentInstant = false
		m.commentBuffer = ""

		item, ok := m.selectedItem()
		if !ok {
			return m, nil
		}

		files := m.currentFiles()

		var filePath string
		if len(files) > m.selectedFile {
			filePath = files[m.selectedFile].Path
		}

		startLine := m.diffCursor
		if m.rangeStart >= 0 {
			startLine = m.rangeStart
		}

		var newLine int
		if len(files) > m.selectedFile && startLine < len(files[m.selectedFile].Diff) {
			newLine = files[m.selectedFile].Diff[startLine].NewLine
		}

		var endNewLine int
		if m.rangeStart >= 0 && len(files) > m.selectedFile && m.diffCursor < len(files[m.selectedFile].Diff) {
			endNewLine = files[m.selectedFile].Diff[m.diffCursor].NewLine
		}

		m.rangeStart = -1

		if instant {
			callback := m.postInlineComment
			if callback == nil {
				return m, nil
			}

			pos := mr.DiffPosition{NewPath: filePath, NewLine: newLine}
			iid := item.IID

			return m, func() tea.Msg {
				err := callback(iid, pos, body)
				return inlineCommentFinishedMsg{iid: iid, err: err}
			}
		}

		draft := mr.DraftComment{
			LocalID:  fmt.Sprintf("local-%d", len(m.drafts[item.IID])+1),
			Body:     body,
			Position: &mr.DiffPosition{NewPath: filePath, NewLine: newLine},
			EndLine:  endNewLine,
		}
		m.drafts[item.IID] = append(m.drafts[item.IID], draft)
	}

	return m, nil
}

//nolint:gocyclo // File diff key handler maps many UI shortcuts explicitly.
func (m Model) updateFileDiffKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	files := m.currentFiles()

	switch msg.String() {
	case "right", "l":
		if m.rangeStart < 0 {
			m.selectedFile = clamp(m.selectedFile+1, 0, len(files)-1)
			m.DiffViewState.YOffset = 0
			m.diffCursor = 0
		}

	case "left", "h":
		if m.rangeStart < 0 {
			m.selectedFile = clamp(m.selectedFile-1, 0, len(files)-1)
			m.DiffViewState.YOffset = 0
			m.diffCursor = 0
		}

	case "up", "k":
		rowCount := 0
		if len(files) > m.selectedFile {
			rowCount = len(files[m.selectedFile].Diff)
		}

		m.diffCursor = clamp(m.diffCursor-1, 0, max(0, rowCount-1))
		m.threadPanelCursor = 0

	case "down", "j":
		rowCount := 0
		if len(files) > m.selectedFile {
			rowCount = len(files[m.selectedFile].Diff)
		}

		m.diffCursor = clamp(m.diffCursor+1, 0, max(0, rowCount-1))
		m.threadPanelCursor = 0

	case "v":
		if m.rangeStart >= 0 {
			m.rangeStart = -1
		} else {
			m.rangeStart = m.diffCursor
		}

	case "i":
		m.commentInput = true
		m.commentInstant = true
		m.commentBuffer = ""
		m.commentError = ""

	case "c":
		m.commentInput = true
		m.commentInstant = false
		m.commentBuffer = ""
		m.commentError = ""

	case "r", "d":
		discussions := m.discussionsAtCursor()
		if len(discussions) > 0 {
			idx := clamp(m.threadPanelCursor, 0, len(discussions)-1)
			m.replyInput = true
			m.replyDraft = msg.String() == "d"
			m.replyDiscussionID = discussions[idx].ID
			m.replyBuffer = ""
		}

	case "x":
		return m.toggleDiscussionResolveAtCursor()

	case "p":
		return m.submitDraftsCommand()

	case "e":
		return m.openEditorAtCursorCommand()

	case "D":
		return m.discardDraftsCommand()

	case "t":
		m.threadPanelVisible = !m.threadPanelVisible

	case "[":
		if m.threadPanelCursor > 0 {
			m.threadPanelCursor--
		}

	case "]":
		discussions := m.discussionsAtCursor()
		if m.threadPanelCursor < len(discussions)-1 {
			m.threadPanelCursor++
		}

	case "esc", "backspace":
		if m.rangeStart >= 0 {
			m.rangeStart = -1
			return m, nil
		}

		m.mode = ModeDetail
		m.activeTab = m.fileDiffReturnTab
		m.DiffViewState.YOffset = 0
	}

	return m, nil
}

func (m Model) toggleDiscussionResolveAtCursor() (Model, tea.Cmd) {
	item, ok := m.selectedItem()
	if !ok {
		return m, nil
	}

	discussions := m.discussionsAtCursor()
	if len(discussions) == 0 {
		return m, nil
	}

	idx := clamp(m.threadPanelCursor, 0, len(discussions)-1)
	activeID := discussions[idx].ID
	iid := item.IID

	for i, discussion := range m.discussions[iid] {
		if discussion.ID != activeID {
			continue
		}

		resolved := !discussion.Resolved

		if resolved {
			callback := m.resolveDiscussion
			if callback == nil {
				m.discussions[iid][i].Resolved = true
				return m, nil
			}

			return m, func() tea.Msg {
				err := callback(iid, activeID)
				return resolveFinishedMsg{iid: iid, discussionID: activeID, resolved: true, err: err}
			}
		}

		callback := m.unresolveDiscussion
		if callback == nil {
			m.discussions[iid][i].Resolved = false
			return m, nil
		}

		return m, func() tea.Msg {
			err := callback(iid, activeID)
			return resolveFinishedMsg{iid: iid, discussionID: activeID, resolved: false, err: err}
		}
	}

	return m, nil
}

func (m Model) submitDraftsCommand() (Model, tea.Cmd) {
	item, ok := m.selectedItem()
	if !ok || m.submitDrafts == nil {
		return m, nil
	}

	drafts := m.drafts[item.IID]
	submit := m.submitDrafts
	post := m.postMRComment
	summary := strings.TrimSpace(m.reviewSummary)
	iid := item.IID

	return m, func() tea.Msg {
		err := submit(iid, drafts)
		if err == nil && summary != "" && post != nil {
			err = post(iid, summary)
		}

		return draftsSubmittedMsg{iid: iid, err: err}
	}
}

func (m Model) openEditorAtCursorCommand() (Model, tea.Cmd) {
	files := m.currentFiles()
	if len(files) <= m.selectedFile || m.openEditor == nil {
		return m, nil
	}

	file := files[m.selectedFile]
	line := 0

	if m.diffCursor < len(file.Diff) {
		line = file.Diff[m.diffCursor].NewLine
	}

	callback := m.openEditor
	path := file.Path

	return m, func() tea.Msg {
		err := callback(path, line)
		return openEditorMsg{err: err}
	}
}

func (m Model) discardDraftsCommand() (Model, tea.Cmd) {
	item, ok := m.selectedItem()
	if !ok {
		return m, nil
	}

	m.drafts[item.IID] = nil

	if m.discardDrafts == nil {
		return m, nil
	}

	discard := m.discardDrafts
	iid := item.IID

	return m, func() tea.Msg {
		return draftsDiscardedMsg{iid: iid, err: discard(iid)}
	}
}
