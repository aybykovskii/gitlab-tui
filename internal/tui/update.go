package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		next, cmd := m.updateKey(msg)
		next.syncGlobalKeys()
		return next, cmd
	case tea.MouseMsg:
		return m.updateMouse(msg)
	case projectStartedMsg:
		m.projectPath = msg.path
		m.mode = ModeEntityList
		m.focus = FocusDetail
		m.loading = true
		m.projectLoading = true
		m.projectLoaded = false
		m.projectError = false
		m.items = nil
		m.errorMessage = ""
		return m, nil
	case accountProjectsStartedMsg:
		state := m.accountProjectStates[msg.accountID]
		state.loading = true
		state.err = ""
		m.accountProjectStates[msg.accountID] = state
		m.rebuildProjectRows()
		return m, nil
	case accountProjectsFinishedMsg:
		state := m.accountProjectStates[msg.accountID]
		state.loading = false
		state.projects = msg.projects
		state.err = ""
		if msg.err != nil {
			state.err = msg.err.Error()
			state.projects = nil
		}
		m.accountProjectStates[msg.accountID] = state
		m.rebuildProjectRows()
		m.selected = m.nearestSelectable(m.selected)
		return m, nil
	case projectFinishedMsg:
		m.loading = false
		m.projectLoading = false
		if msg.err != nil {
			m.projectError = true
			m.projectLoaded = false
			m.items = nil
			m.errorMessage = msg.err.Error()
			return m, nil
		}
		m.projectError = false
		m.projectLoaded = true
		m.projectPath = msg.path
		m.items = msg.data.Items
		m.projectLabels = msg.data.Labels
		if msg.data.UpdateMRLabels != nil {
			m.updateMRLabels = msg.data.UpdateMRLabels
		}
		m.issueItems = msg.data.Issues
		m.refresh = msg.data.Refresh
		m.loadIssues = msg.data.LoadIssues
		m.postIssueComment = msg.data.PostIssueComment
		m.loadIssueDiscussions = msg.data.LoadIssueDiscussions
		m.loadDiff = msg.data.LoadDiff
		m.loadDiscussions = msg.data.LoadDiscussions
		m.loadFiles = msg.data.LoadFiles
		m.closeIssue = msg.data.CloseIssue
		m.reopenIssue = msg.data.ReopenIssue
		m.editIssue = msg.data.EditIssue
		m.assignSelfIssue = msg.data.AssignSelfIssue
		m.unassignSelfIssue = msg.data.UnassignSelfIssue
		m.selected = clampSelection(0, len(m.filtered()))
		m.selectEntity()
		m.listTop = 0
		m.rightTop = 0
		if m.section == SectionMergeRequests {
			if m.entityID != "" {
				m.mode = ModeDetail
			} else {
				m.mode = ModeEntityList
			}
		} else if m.section == SectionIssues {
			m.mode = ModeEntityList
		} else {
			m.mode = ModeSections
		}
		m.focus = FocusDetail
		return m, nil
	case discussionsStartedMsg:
		m.discussionsLoading = true
		m.discussionsError = ""
		return m, nil
	case discussionsFinishedMsg:
		m.discussionsLoading = false
		if msg.err != nil {
			m.discussionsError = msg.err.Error()
			return m, nil
		}
		m.discussions[msg.iid] = msg.discussions
		return m, nil
	case filesStartedMsg:
		m.filesLoading = true
		m.filesError = ""
		return m, nil
	case approveMRFinishedMsg:
		if msg.err != nil {
			m.actionError = msg.err.Error()
		} else {
			m.actionError = "Approved"
		}
		return m, nil
	case mergeMRFinishedMsg:
		m.mergeConfirmPending = false
		if msg.err != nil {
			m.actionError = msg.err.Error()
		} else {
			m.actionError = "Merged"
		}
		return m, nil
	case editMRFinishedMsg:
		if msg.err != nil {
			m.actionError = msg.err.Error()
		} else {
			for i, item := range m.items {
				if item.IID == msg.iid {
					m.items[i].Title = msg.title
					m.items[i].Description = msg.description
				}
			}
		}
		return m, nil
	case openURLMsg:
		if msg.err != nil {
			m.actionError = msg.err.Error()
		}
		return m, nil
	case openEditorMsg:
		if msg.err != nil {
			m.actionError = msg.err.Error()
		}
		return m, nil
	case updateMRLabelsFinishedMsg:
		if msg.err != nil {
			for i := range m.items {
				if m.items[i].IID == msg.iid {
					m.items[i].Labels = msg.prev
					break
				}
			}
			m.actionError = msg.err.Error()
		}
		return m, nil
	case toggleDraftFinishedMsg:
		if msg.err != nil {
			for i := range m.items {
				if m.items[i].IID == msg.iid {
					m.items[i].Draft = msg.prev
					break
				}
			}
			m.actionError = msg.err.Error()
		}
		return m, nil
	case inlineCommentFinishedMsg:
		if msg.err != nil {
			m.commentError = msg.err.Error()
		} else {
			m.commentError = ""
		}
		return m, nil
	case mrCommentFinishedMsg:
		if msg.err != nil {
			m.mrCommentError = msg.err.Error()
		} else {
			m.mrCommentError = ""
		}
		return m, nil
	case issueStateFinishedMsg:
		if msg.err != nil {
			m.actionError = msg.err.Error()
			return m, nil
		}
		for i := range m.issueItems {
			if m.issueItems[i].IID == msg.iid {
				m.issueItems[i].State = msg.state
				break
			}
		}
		return m, nil
	case editIssueFinishedMsg:
		if msg.err != nil {
			m.actionError = msg.err.Error()
			return m, nil
		}
		for i := range m.issueItems {
			if m.issueItems[i].IID == msg.iid {
				m.issueItems[i].Title = msg.title
				m.issueItems[i].Description = msg.description
				break
			}
		}
		return m, nil
	case issueAssigneeFinishedMsg:
		if msg.err != nil {
			m.actionError = msg.err.Error()
			return m, nil
		}
		for i := range m.issueItems {
			if m.issueItems[i].IID == msg.iid {
				m.issueItems[i].Assignees = msg.assignees
				break
			}
		}
		return m, nil
	case replyFinishedMsg:
		if msg.err == nil && !msg.draft {
			if ds, ok := m.discussions[msg.iid]; ok {
				for i, d := range ds {
					if d.ID == msg.discussionID {
						m.discussions[msg.iid][i].Notes = append(m.discussions[msg.iid][i].Notes, mr.Note{
							Author: "me",
							Body:   msg.body,
						})
					}
				}
			}
		}
		return m, nil
	case resolveFinishedMsg:
		if msg.err == nil {
			if ds, ok := m.discussions[msg.iid]; ok {
				for i, d := range ds {
					if d.ID == msg.discussionID {
						m.discussions[msg.iid][i].Resolved = msg.resolved
					}
				}
			}
		}
		return m, nil
	case draftAddedMsg:
		m.drafts[msg.iid] = append(m.drafts[msg.iid], msg.draft)
		return m, nil
	case draftsSubmittedMsg:
		if msg.err == nil {
			m.drafts[msg.iid] = nil
		}
		return m, nil
	case draftsDiscardedMsg:
		m.drafts[msg.iid] = nil
		return m, nil
	case filesFinishedMsg:
		m.filesLoading = false
		if msg.err != nil {
			m.filesError = msg.err.Error()
			return m, nil
		}
		m.changedFiles[msg.iid] = msg.files
		return m, nil
	case refreshStartedMsg:
		m.loading = true
		m.errorMessage = ""
		return m, nil
	case refreshFinishedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			return m, nil
		}
		m.items = msg.items
		m.selected = clampSelection(m.selected, len(m.filtered()))
		m.listTop = 0
		return m, nil
	case issuesFinishedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			return m, nil
		}
		m.issueItems = msg.items
		m.selected = clampSelection(m.selected, len(m.filteredIssues()))
		m.listTop = 0
		return m, nil
	case issueDiscussionsFinishedMsg:
		if msg.err != nil {
			m.discussionsError = msg.err.Error()
			return m, nil
		}
		m.issueDiscussions[msg.iid] = msg.discussions
		return m, nil
	case diffStartedMsg:
		m.diffLoading = true
		m.errorMessage = ""
		return m, nil
	case diffFinishedMsg:
		m.diffLoading = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			return m, nil
		}
		m.setDiffRows(msg.iid, msg.rows)
		m.mode = ModeDiff
		m.focus = FocusDetail
		m.rightTop = 0
		return m, nil
	}

	return m, nil
}

func (m Model) updateKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if key.Matches(msg, m.globals.ToggleKeyBar) && !m.inputActive() {
		m.keyBarExpanded = !m.keyBarExpanded
		return m, nil
	}
	if m.mode == ModeLabelSelect {
		return m.updateLabelSelect(msg)
	}
	if m.mode == ModeProjectSelect {
		if m.projectFilterActive {
			switch msg.Type {
			case tea.KeyEsc:
				m.query = ""
				m.projectFilterActive = false
				m.rebuildProjectRows()
				m.selected = m.nearestSelectable(0)
				return m, nil
			case tea.KeyBackspace:
				if len(m.query) > 0 {
					m.query = m.query[:len(m.query)-1]
					m.rebuildProjectRows()
					m.selected = m.nearestSelectable(0)
				}
				return m, nil
			case tea.KeyRunes:
				m.query += msg.String()
				m.rebuildProjectRows()
				m.selected = m.nearestSelectable(0)
				return m, nil
			}
		}
		switch {
		case key.Matches(msg, m.projectListKeys.Filter):
			m.projectFilterActive = true
			m.query = ""
			m.rebuildProjectRows()
			m.selected = m.nearestSelectable(0)
		case key.Matches(msg, m.globals.Back):
			m.query = ""
			m.projectFilterActive = false
			m.rebuildProjectRows()
			m.selected = m.nearestSelectable(0)
		case key.Matches(msg, m.projectListKeys.Up):
			m.selected = m.nextSelectable(m.selected, -1)
		case key.Matches(msg, m.projectListKeys.Down):
			m.selected = m.nextSelectable(m.selected, 1)
		case key.Matches(msg, m.projectListKeys.Open):
			if project, ok := m.selectedProject(); ok {
				return m.selectProject(project)
			}
		case key.Matches(msg, m.projectListKeys.Retry):
			return m, m.retryFailedProjectLoads()
		case key.Matches(msg, m.projectListKeys.Input):
			m.mode = ModeProjectInput
			m.focus = FocusFilter
			m.projectInput = ""
		}
		return m, nil
	}

	if m.mode == ModeProjectInput {
		switch msg.Type {
		case tea.KeyEnter:
			if strings.TrimSpace(m.projectInput) != "" {
				return m.selectProject(strings.TrimSpace(m.projectInput))
			}
		case tea.KeyBackspace:
			if len(m.projectInput) > 0 {
				m.projectInput = m.projectInput[:len(m.projectInput)-1]
			}
		case tea.KeyRunes:
			m.projectInput += msg.String()
		}
		return m, nil
	}

	if m.mode == ModeSections {
		switch msg.String() {
		case "up", "k":
			m.sectionCursor = clamp(m.sectionCursor-1, 0, len(tuiSections)-1)
		case "down", "j":
			m.sectionCursor = clamp(m.sectionCursor+1, 0, len(tuiSections)-1)
		case "enter":
			sec := tuiSections[m.sectionCursor]
			if sec.available && sec.id == SectionMergeRequests {
				m.section = SectionMergeRequests
				if m.projectLoaded {
					m.mode = ModeEntityList
					m.focus = FocusDetail
					return m, nil
				}
				return m.openProjectCommand(m.projectPath)
			}
			if sec.available && sec.id == SectionIssues {
				m.section = SectionIssues
				m.mode = ModeEntityList
				m.focus = FocusDetail
				return m, m.loadIssuesCommand()
			}
		case "esc", "backspace":
			m.returnToProjectPicker()
		}
		return m, nil
	}

	if m.focus == FocusFilter {
		switch msg.Type {
		case tea.KeyEsc, tea.KeyEnter:
			m.focus = FocusDetail
		case tea.KeyBackspace:
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.selected = m.clampEntitySelection(m.selected)
			}
		case tea.KeyRunes:
			m.query += msg.String()
			m.selected = m.clampEntitySelection(m.selected)
		}
		return m, nil
	}

	if m.mode == ModeEntityList {
		switch msg.String() {
		case "up", "k":
			m.moveSelection(-1)
		case "down", "j":
			m.moveSelection(1)
		case "enter":
			m.mode = ModeDetail
			m.focus = FocusDetail
			m.activeTab = TabSummary
		case "esc", "backspace":
			if m.projectError || (m.projectPath != "" && len(m.items) == 0) {
				m.errorMessage = ""
				m.returnToProjectPicker()
			} else {
				m.mode = ModeSections
			}
		case "/":
			m.focus = FocusFilter
		case "r":
			if m.projectError && m.projectPath != "" {
				return m.openProjectCommand(m.projectPath)
			}
			return m, m.refreshCommand()
		case "s":
			if m.section == SectionIssues {
				m.cycleIssueState()
				return m, m.loadIssuesCommand()
			}
		}
		return m, nil
	}

	if m.mode == ModeFileDiff {
		if m.replyInput {
			switch msg.Type {
			case tea.KeyEsc:
				m.replyInput = false
				m.replyBuffer = ""
				m.replyDiscussionID = ""
			case tea.KeyBackspace:
				if len(m.replyBuffer) > 0 {
					m.replyBuffer = m.replyBuffer[:len(m.replyBuffer)-1]
				}
			case tea.KeyRunes, tea.KeySpace:
				m.replyBuffer += msg.String()
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
					fn := m.draftReply
					if fn == nil {
						return m, nil
					}
					return m, func() tea.Msg {
						err := fn(iid, discussionID, body)
						return replyFinishedMsg{iid: iid, discussionID: discussionID, body: body, draft: true, err: err}
					}
				}
				fn := m.replyToDiscussion
				if fn == nil {
					return m, nil
				}
				return m, func() tea.Msg {
					err := fn(iid, discussionID, body)
					return replyFinishedMsg{iid: iid, discussionID: discussionID, body: body, draft: false, err: err}
				}
			}
			return m, nil
		}
		if m.commentInput {
			switch msg.Type {
			case tea.KeyEsc:
				m.commentInput = false
				m.commentBuffer = ""
			case tea.KeyEnter:
				body := m.commentBuffer
				instant := m.commentInstant
				m.commentInput = false
				m.commentInstant = false
				m.commentBuffer = ""
				item, ok := m.selectedItem()
				if ok {
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
						fn := m.postInlineComment
						if fn != nil {
							pos := mr.DiffPosition{NewPath: filePath, NewLine: newLine}
							iid := item.IID
							return m, func() tea.Msg {
								err := fn(iid, pos, body)
								return inlineCommentFinishedMsg{iid: iid, err: err}
							}
						}
					} else {
						draft := mr.DraftComment{
							LocalID:  fmt.Sprintf("local-%d", len(m.drafts[item.IID])+1),
							Body:     body,
							Position: &mr.DiffPosition{NewPath: filePath, NewLine: newLine},
							EndLine:  endNewLine,
						}
						m.drafts[item.IID] = append(m.drafts[item.IID], draft)
					}
				}
			case tea.KeyBackspace:
				if len(m.commentBuffer) > 0 {
					m.commentBuffer = m.commentBuffer[:len(m.commentBuffer)-1]
				}
			case tea.KeyRunes, tea.KeySpace:
				m.commentBuffer += msg.String()
			}
			return m, nil
		}

		files := m.currentFiles()
		switch msg.String() {
		case "right", "l":
			if m.rangeStart < 0 {
				m.selectedFile = clamp(m.selectedFile+1, 0, len(files)-1)
				m.fileDiffTop = 0
				m.diffCursor = 0
			}
		case "left", "h":
			if m.rangeStart < 0 {
				m.selectedFile = clamp(m.selectedFile-1, 0, len(files)-1)
				m.fileDiffTop = 0
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
			isDraft := msg.String() == "d"
			ds := m.discussionsAtCursor()
			if len(ds) > 0 {
				idx := clamp(m.threadPanelCursor, 0, len(ds)-1)
				m.replyInput = true
				m.replyDraft = isDraft
				m.replyDiscussionID = ds[idx].ID
				m.replyBuffer = ""
			}
		case "x":
			item, ok := m.selectedItem()
			if !ok {
				break
			}
			ds := m.discussionsAtCursor()
			if len(ds) == 0 {
				break
			}
			idx := clamp(m.threadPanelCursor, 0, len(ds)-1)
			activeID := ds[idx].ID
			iid := item.IID
			for i, d := range m.discussions[iid] {
				if d.ID != activeID {
					continue
				}
				resolved := !d.Resolved
				if resolved {
					fn := m.resolveDiscussion
					if fn == nil {
						m.discussions[iid][i].Resolved = true
						return m, nil
					}
					return m, func() tea.Msg {
						err := fn(iid, activeID)
						return resolveFinishedMsg{iid: iid, discussionID: activeID, resolved: true, err: err}
					}
				}
				fn := m.unresolveDiscussion
				if fn == nil {
					m.discussions[iid][i].Resolved = false
					return m, nil
				}
				return m, func() tea.Msg {
					err := fn(iid, activeID)
					return resolveFinishedMsg{iid: iid, discussionID: activeID, resolved: false, err: err}
				}
			}
		case "p":
			item, ok := m.selectedItem()
			if ok && m.submitDrafts != nil {
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
		case "e":
			files := m.currentFiles()
			if len(files) > m.selectedFile && m.openEditor != nil {
				file := files[m.selectedFile]
				line := 0
				if m.diffCursor < len(file.Diff) {
					line = file.Diff[m.diffCursor].NewLine
				}
				fn := m.openEditor
				path := file.Path
				return m, func() tea.Msg {
					err := fn(path, line)
					return openEditorMsg{err: err}
				}
			}
		case "D":
			item, ok := m.selectedItem()
			if ok {
				m.drafts[item.IID] = nil
				if m.discardDrafts != nil {
					discard := m.discardDrafts
					iid := item.IID
					return m, func() tea.Msg {
						return draftsDiscardedMsg{iid: iid, err: discard(iid)}
					}
				}
			}
		case "t":
			m.threadPanelVisible = !m.threadPanelVisible
		case "[":
			if m.threadPanelCursor > 0 {
				m.threadPanelCursor--
			}
		case "]":
			ds := m.discussionsAtCursor()
			if m.threadPanelCursor < len(ds)-1 {
				m.threadPanelCursor++
			}
		case "esc", "backspace":
			if m.rangeStart >= 0 {
				m.rangeStart = -1
				return m, nil
			}
			m.mode = ModeDetail
			m.activeTab = m.fileDiffReturnTab
			m.fileDiffTop = 0
		}
		return m, nil
	}

	if m.mode == ModeDetail && m.editInput {
		if m.section == SectionIssues {
			return m.updateIssueEdit(msg)
		}
		return m.updateMREdit(msg)
	}

	if m.mode == ModeDetail && m.issueCommentInput {
		return m.updateIssueCommentInput(msg)
	}
	if m.mode == ModeDetail && m.mrCommentInput {
		return m.updateMRCommentInput(msg)
	}
	if m.mode == ModeDetail && m.activeTab == TabReview && msg.String() != "tab" {
		return m.updateReviewTab(msg)
	}

	if m.mode == ModeDetail && m.mergeConfirmPending && msg.String() != "M" {
		m.mergeConfirmPending = false
		return m, nil
	}

	if m.mode == ModeDetail && m.activeTab == TabDiscussions && msg.String() != "tab" {
		if m.section == SectionIssues {
			return m.updateIssueDiscussionsTab(msg)
		}
		return m.updateDiscussionsTab(msg)
	}

	switch {
	case key.Matches(msg, m.globals.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.globals.Back):
		if m.projectError || (m.projectPath != "" && len(m.items) == 0) {
			m.errorMessage = ""
			m.returnToProjectPicker()
			return m, nil
		}
		if m.mode == ModeDiff {
			m.mode = ModeDetail
			m.rightTop = 0
		} else if m.mode == ModeDetail {
			m.mode = ModeEntityList
			m.focus = FocusDetail
			m.rightTop = 0
		}
	case msg.String() == "/":
		m.focus = FocusFilter
	case msg.String() == "r":
		if m.projectError && m.projectPath != "" {
			return m.openProjectCommand(m.projectPath)
		}
		return m, m.refreshCommand()
	case msg.String() == "m":
		if m.mode == ModeDetail {
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
			item, ok := m.selectedItem()
			if ok && m.approveMR != nil {
				fn := m.approveMR
				iid := item.IID
				return m, func() tea.Msg {
					err := fn(iid)
					return approveMRFinishedMsg{iid: iid, err: err}
				}
			}
		}
	case msg.String() == "M":
		if m.mode == ModeDetail {
			if m.mergeConfirmPending {
				item, ok := m.selectedItem()
				if ok && m.mergeMR != nil {
					fn := m.mergeMR
					iid := item.IID
					return m, func() tea.Msg {
						err := fn(iid)
						return mergeMRFinishedMsg{iid: iid, err: err}
					}
				}
				m.mergeConfirmPending = false
			} else {
				m.mergeConfirmPending = true
			}
		}
	case msg.String() == "o":
		if m.mode == ModeDetail {
			if m.section == SectionIssues {
				item, ok := m.selectedIssue()
				if ok && item.WebURL != "" && m.openURL != nil {
					fn := m.openURL
					url := item.WebURL
					return m, func() tea.Msg { return openURLMsg{url: url, err: fn(url)} }
				}
			} else {
				item, ok := m.selectedItem()
				if ok && item.WebURL != "" && m.openURL != nil {
					fn := m.openURL
					url := item.WebURL
					return m, func() tea.Msg {
						err := fn(url)
						return openURLMsg{url: url, err: err}
					}
				}
			}
		}
	case msg.String() == "e":
		if m.mode == ModeDetail {
			if m.section == SectionIssues {
				item, ok := m.selectedIssue()
				if ok {
					m.editInput = true
					m.editField = "title"
					m.editBuffer = item.Title
					m.editTitle = ""
				}
			} else {
				item, ok := m.selectedItem()
				if ok {
					m.editInput = true
					m.editField = "title"
					m.editBuffer = item.Title
					m.editTitle = ""
				}
			}
		}
	case msg.String() == "l":
		if m.mode == ModeDetail {
			if m.section == SectionIssues {
				issueItem, ok := m.selectedIssue()
				if ok {
					m.mode = ModeLabelSelect
					m.labelCursor = 0
					pending := make([]string, len(issueItem.Labels))
					copy(pending, issueItem.Labels)
					m.labelPending = pending
				}
			} else if m.activeTab == TabSummary {
				item, ok := m.selectedItem()
				if ok {
					m.mode = ModeLabelSelect
					m.labelCursor = 0
					pending := make([]string, len(item.Labels))
					copy(pending, item.Labels)
					m.labelPending = pending
				}
			}
		}
	case msg.String() == "d":
		if m.mode == ModeDetail && m.activeTab != TabDiscussions {
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
			fn := m.toggleDraftMR
			iid := item.IID
			return m, func() tea.Msg {
				err := fn(iid)
				return toggleDraftFinishedMsg{iid: iid, prev: prev, err: err}
			}
		}
	case msg.String() == "tab":
		if m.mode == ModeDetail {
			if m.section == SectionIssues {
				m.activeTab = (m.activeTab + 1) % (TabDiscussions + 1)
				return m, m.loadIssueDiscussionsCommand()
			}
			m.activeTab = (m.activeTab + 1) % (TabReview + 1)
			return m.onTabEntered()
		}
	case msg.String() == "up" || msg.String() == "k":
		if m.mode == ModeDetail || m.mode == ModeDiff {
			m.rightTop = max(0, m.rightTop-1)
		} else {
			m.moveSelection(-1)
		}
	case msg.String() == "down" || msg.String() == "j":
		if m.mode == ModeDetail || m.mode == ModeDiff {
			m.rightTop = max(0, m.rightTop+1)
		} else {
			m.moveSelection(1)
		}
	case msg.String() == "enter":
		if m.mode == ModeDetail && m.activeTab == TabFiles {
			if item, ok := m.selectedItem(); ok {
				if files, loaded := m.changedFiles[item.IID]; loaded && len(files) > 0 {
					m.mode = ModeFileDiff
					m.fileDiffReturnTab = TabFiles
					m.selectedFile = 0
					m.fileDiffTop = 0
					m.diffCursor = 0
					m.threadPanelCursor = 0
					return m, m.ensureDiscussionsLoaded(item.IID)
				}
			}
		}
	case msg.String() == "backspace":
		if m.mode == ModeDiff {
			m.mode = ModeDetail
			m.rightTop = 0
		}
	}

	return m, nil
}

func (m Model) updateMouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	if m.mode == ModeProjectSelect {
		if msg.Button == tea.MouseButtonLeft && msg.Y >= 2 {
			idx := msg.Y - 2
			if idx >= 0 && idx < len(m.projectRows) && m.projectRows[idx].selectable {
				m.selected = idx
				return m.selectProject(m.projectRows[idx].project)
			}
		}
		if msg.Button == tea.MouseButtonWheelUp {
			m.selected = m.nextSelectable(m.selected, -1)
		}
		if msg.Button == tea.MouseButtonWheelDown {
			m.selected = m.nextSelectable(m.selected, 1)
		}
		return m, nil
	}

	leftWidth := m.leftWidth()
	m.focus = FocusDetail

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.scrollFocused(-1)
	case tea.MouseButtonWheelDown:
		m.scrollFocused(1)
	case tea.MouseButtonLeft:
		if msg.X < leftWidth && msg.Y >= 4 {
			idx := m.listTop + msg.Y - 4
			if idx >= 0 && idx < len(m.filtered()) {
				m.selected = idx
				m.mode = ModeDetail
			}
		} else if msg.X >= leftWidth {
			if item, ok := m.selectedItem(); ok {
				return m.openDiffCommand(item)
			}
		}
	}

	return m, nil
}

func (m *Model) selectEntity() {
	if m.entityID == "" {
		return
	}
	iid, err := strconv.Atoi(m.entityID)
	if err != nil {
		return
	}
	for i, item := range m.filtered() {
		if item.IID == iid {
			m.selected = i
			return
		}
	}
}

func (m *Model) moveSelection(delta int) {
	count := len(m.filtered())
	if m.section == SectionIssues {
		count = len(m.filteredIssues())
	}
	if count == 0 {
		m.selected = 0
		return
	}

	m.selected = clamp(m.selected+delta, 0, count-1)
	visible := max(1, m.height-4)
	if m.selected < m.listTop {
		m.listTop = m.selected
	}
	if m.selected >= m.listTop+visible {
		m.listTop = m.selected - visible + 1
	}
}

func (m *Model) scrollFocused(delta int) {
	if m.focus == FocusList {
		m.moveSelection(delta)
		return
	}
	m.rightTop = max(0, m.rightTop+delta)
}

func (m Model) inputActive() bool {
	return m.projectFilterActive || m.mode == ModeProjectInput || m.commentInput || m.mrCommentInput || m.issueCommentInput || m.editInput || m.replyInput || m.focus == FocusFilter
}

func (m *Model) syncGlobalKeys() {
	active := m.inputActive()
	m.globals.Quit.SetEnabled(!active)
	m.globals.Back.SetEnabled(!active)
}

func (m Model) localKeys() []key.Binding {
	if m.mode == ModeProjectSelect {
		return m.projectListKeys.LocalKeys()
	}
	switch m.mode {
	case ModeSections:
		return newSectionsKeys().LocalKeys()
	case ModeEntityList:
		return newEntityListKeys().LocalKeys()
	case ModeDetail:
		if m.section == SectionIssues {
			return newIssueDetailKeys().LocalKeys()
		}
		return newMRDetailKeys().LocalKeys()
	case ModeLabelSelect:
		return newMRDetailKeys().LocalKeys()
	case ModeDiff:
		return newDiffViewKeys().LocalKeys()
	case ModeFileDiff:
		return newFileDiffKeys().LocalKeys()
	default:
		return newEntityListKeys().LocalKeys()
	}
}

func (m Model) globalKeys() []key.Binding {
	if m.inputActive() {
		return []key.Binding{key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "send")), key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "cancel"))}
	}
	return []key.Binding{m.globals.Quit, m.globals.Back, m.globals.ToggleKeyBar}
}

func (m Model) keyBarHeight() int {
	content := 2
	if m.keyBarExpanded {
		content = max(4, (len(m.localKeys())+1)/2+2)
	}
	return content + 2
}

func (m Model) paneHeight() int {
	return max(8, m.height-m.keyBarHeight())
}

func (m Model) selectedItem() (mr.MergeRequest, bool) {
	items := m.filtered()
	if len(items) == 0 {
		return mr.MergeRequest{}, false
	}
	return items[clampSelection(m.selected, len(items))], true
}

func (m *Model) setDiffRows(iid int, rows []mr.DiffRow) {
	for i := range m.items {
		if m.items[i].IID == iid {
			m.items[i].Diff = rows
			return
		}
	}
}

func (m Model) leftWidth() int {
	if m.width <= 0 {
		return 40
	}
	return max(24, m.width*35/100)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
