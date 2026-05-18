package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func (m Model) updateReviewTab(msg tea.KeyMsg) (Model, tea.Cmd) {
	item, ok := m.selectedItem()
	if !ok {
		return m, nil
	}
	drafts := m.drafts[item.IID]
	if m.reviewSummaryInput {
		switch msg.Type {
		case tea.KeyEnter:
			m.reviewSummaryInput = false
		case tea.KeyEsc:
			m.reviewSummaryInput = false
		case tea.KeyBackspace:
			if len(m.reviewSummary) > 0 {
				m.reviewSummary = m.reviewSummary[:len(m.reviewSummary)-1]
			}
		case tea.KeyRunes, tea.KeySpace:
			m.reviewSummary += msg.String()
		}
		return m, nil
	}
	switch msg.String() {
	case "up", "k":
		m.reviewCursor = max(0, m.reviewCursor-1)
	case "down", "j":
		if len(drafts) == 0 || m.reviewCursor >= len(drafts)-1 {
			m.reviewSummaryInput = true
		} else {
			m.reviewCursor++
		}
	case "enter":
		if len(drafts) == 0 || m.reviewCursor >= len(drafts) {
			m.reviewSummaryInput = true
			return m, nil
		}
		m.openDraftInDiff(drafts[m.reviewCursor])
	case "p":
		if m.submitDrafts == nil {
			return m, nil
		}
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
	case "D":
		m.drafts[item.IID] = nil
		if m.discardDrafts == nil {
			return m, nil
		}
		discard := m.discardDrafts
		iid := item.IID
		return m, func() tea.Msg { return draftsDiscardedMsg{iid: iid, err: discard(iid)} }
	}
	return m, nil
}

func (m *Model) openDraftInDiff(draft mr.DraftComment) {
	if draft.Position == nil {
		return
	}
	files := m.currentFiles()
	for fileIndex, file := range files {
		if file.Path != draft.Position.NewPath {
			continue
		}
		m.mode = ModeFileDiff
		m.fileDiffReturnTab = TabReview
		m.selectedFile = fileIndex
		m.fileDiffTop = 0
		m.diffCursor = 0
		for rowIndex, row := range file.Diff {
			if row.NewLine == draft.Position.NewLine {
				m.diffCursor = rowIndex
				break
			}
		}
		return
	}
}
