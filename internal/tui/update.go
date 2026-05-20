//nolint:mnd,gocritic // Interactive state machine keeps UI constants and branches explicit.
package tui

import (
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

//nolint:gocyclo // Bubble Tea update loop centralizes many message types.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		next, cmd := m.updateKey(msg)

		if next.mode == ModeDetail && next.section == SectionMergeRequests {
			cmd = batchCommands(cmd, next.MRDetailState.Update(msg))
		}

		if next.mode == ModeDetail && next.section == SectionIssues {
			cmd = batchCommands(cmd, next.IssueDetailState.Update(msg))
		}

		if next.mode == ModeFileDiff {
			cmd = batchCommands(cmd, next.DiffViewState.Update(msg))
		}

		next.syncGlobalKeys()

		return next, cmd
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
		state := m.ProjectPickerState.accountProjectStates[msg.accountID]
		state.loading = true
		state.err = ""
		m.ProjectPickerState.accountProjectStates[msg.accountID] = state
		m.rebuildProjectRows()

		return m, nil
	case accountProjectsFinishedMsg:
		state := m.ProjectPickerState.accountProjectStates[msg.accountID]
		state.loading = false
		state.projects = msg.projects
		state.err = ""

		if msg.err != nil {
			state.err = msg.err.Error()
			state.projects = nil
		}

		m.ProjectPickerState.accountProjectStates[msg.accountID] = state
		m.rebuildProjectRows()
		m.ProjectPickerState.selected = m.ProjectPickerState.nearestSelectable(m.ProjectPickerState.selected)

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
		m.loadDiscussions = msg.data.LoadDiscussions
		m.loadFiles = msg.data.LoadFiles
		m.closeIssue = msg.data.CloseIssue
		m.reopenIssue = msg.data.ReopenIssue
		m.editIssue = msg.data.EditIssue
		m.assignSelfIssue = msg.data.AssignSelfIssue
		m.unassignSelfIssue = msg.data.UnassignSelfIssue
		m.EntityListState.mrList.Select(0)
		m.EntityListState.syncMRList()
		m.selectEntity()
		m.MRDetailState.GotoTop()

		switch m.section {
		case SectionMergeRequests:
			if m.entityID != "" {
				m.mode = ModeDetail
			} else {
				m.mode = ModeEntityList
			}
		case SectionIssues:
			m.mode = ModeEntityList
		default:
			m.mode = ModeSections
		}

		m.focus = FocusDetail

		return m, nil
	case discussionsStartedMsg, discussionsFinishedMsg, filesStartedMsg:
		return m, m.MRDetailState.Update(msg)
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
			if discussions, ok := m.discussions[msg.iid]; ok {
				for i, discussion := range discussions {
					if discussion.ID == msg.discussionID {
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
			if discussions, ok := m.discussions[msg.iid]; ok {
				for i, discussion := range discussions {
					if discussion.ID == msg.discussionID {
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
		return m, m.MRDetailState.Update(msg)
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
		m.EntityListState.syncMRList()

		return m, nil
	case issuesFinishedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			return m, nil
		}

		m.issueItems = msg.items
		m.EntityListState.syncIssueList()

		return m, nil
	case issueDiscussionsFinishedMsg:
		return m, m.IssueDetailState.Update(msg)
	}

	return m, nil
}

//nolint:gocyclo // Keyboard state machine has many intentional shortcuts.
func (m Model) updateKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if key.Matches(msg, m.globals.ToggleKeyBar) && !m.inputActive() {
		m.keyBarExpanded = !m.keyBarExpanded
		return m, nil
	}

	if !m.inputActive() && key.Matches(msg, m.globals.Quit) {
		return m, tea.Quit
	}

	switch m.mode {
	case ModeLabelSelect:
		return m.updateLabelSelect(msg)
	case ModeProjectSelect:
		return m.updateProjectSelect(msg)
	case ModeProjectInput:
		return m.updateProjectInput(msg)
	case ModeSections:
		return m.updateSections(msg)
	case ModeFileDiff:
		return m.updateFileDiff(msg)
	}

	if m.focus == FocusFilter {
		return m.updateFilter(msg)
	}

	if m.mode == ModeEntityList {
		return m.updateEntityList(msg)
	}

	if m.mode == ModeDetail {
		if m.editInput {
			if m.section == SectionIssues {
				return m.updateIssueEdit(msg)
			}

			return m.updateMREdit(msg)
		}

		if m.issueCommentInput {
			return m.updateIssueCommentInput(msg)
		}

		if m.mrCommentInput {
			return m.updateMRCommentInput(msg)
		}

		if m.section == SectionMergeRequests && m.activeTab == TabReview && msg.String() != "tab" {
			return m.updateReviewTab(msg)
		}

		if m.mergeConfirmPending && msg.String() != "M" {
			m.mergeConfirmPending = false
			return m, nil
		}

		if m.section == SectionIssues && m.IssueDetailState.activeTab == TabDiscussions && msg.String() != "tab" {
			return m.updateIssueDiscussionsTab(msg)
		}

		if m.section == SectionMergeRequests && m.activeTab == TabDiscussions && msg.String() != "tab" {
			return m.updateDiscussionsTab(msg)
		}
	}

	return m.updateDetailKeys(msg)
}

func (m *Model) selectEntity() {
	if m.entityID == "" {
		return
	}

	iid, err := strconv.Atoi(m.entityID)
	if err != nil {
		return
	}

	for i, item := range m.EntityListState.filteredMRs() {
		if item.IID == iid {
			m.EntityListState.mrList.Select(i)
			return
		}
	}
}

func (m *Model) moveSelection(delta int) {
	if m.section == SectionIssues {
		if delta > 0 {
			m.EntityListState.issueList.CursorDown()
		} else {
			m.EntityListState.issueList.CursorUp()
		}

		return
	}

	if delta > 0 {
		m.EntityListState.mrList.CursorDown()
	} else {
		m.EntityListState.mrList.CursorUp()
	}
}

func (m Model) inputActive() bool {
	return m.ProjectPickerState.projectFilterActive || m.mode == ModeProjectInput || m.commentInput || m.mrCommentInput || m.issueCommentInput || m.editInput || m.replyInput || m.focus == FocusFilter
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
	items := m.EntityListState.filteredMRs()
	if len(items) == 0 {
		return mr.MergeRequest{}, false
	}

	return items[clampSelection(m.EntityListState.mrList.Index(), len(items))], true
}

const defaultLeftPanelWidthPct = 35

func (m Model) leftWidth() int {
	if m.width <= 0 {
		return 40
	}

	w := m.layout.LeftPanelWidth
	var pct int

	switch m.mode {
	case ModeFileDiff:
		pct = w.FileDiff
	case ModeEntityList:
		pct = w.EntityList
	case ModeDetail, ModeLabelSelect:
		pct = w.Detail
	case ModeSections:
		pct = w.Sections
	default: // ModeProjectSelect, ModeProjectInput
		pct = w.Projects
	}

	if pct == 0 {
		pct = defaultLeftPanelWidthPct
	}

	return max(24, m.width*pct/100)
}
