package tui

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aybykovskii/gitlab-tui/internal/diff"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type Mode int

const (
	ModeProjectSelect Mode = iota
	ModeProjectInput
	ModeSections
	ModeDetail
	ModeDiff
	ModeFileDiff
)

type DetailTab int

const (
	TabSummary DetailTab = iota
	TabDiscussions
	TabFiles
)

type Section string

const (
	SectionMergeRequests Section = "mr"
	SectionIssues        Section = "issue"
	SectionPipelines     Section = "pipeline"
)

type sectionDef struct {
	label     string
	id        Section
	available bool
}

var tuiSections = []sectionDef{
	{label: "Merge Requests", id: SectionMergeRequests, available: true},
	{label: "Issues", id: SectionIssues, available: false},
	{label: "Pipelines", id: SectionPipelines, available: false},
}

type Focus int

const (
	FocusList Focus = iota
	FocusDetail
	FocusFilter
)

type RefreshFunc func() ([]mr.MergeRequest, error)
type DiffFunc func(iid int) ([]mr.DiffRow, error)
type ProjectLoadFunc func(path string) (ProjectData, error)
type LoadDiscussionsFunc func(iid int) ([]mr.Discussion, error)
type LoadFilesFunc func(iid int) ([]mr.ChangedFile, error)
type SubmitDraftsFunc       func(iid int, drafts []mr.DraftComment) error
type DiscardDraftsFunc      func(iid int) error
type ReplyToDiscussionFunc   func(iid int, discussionID string, body string) error
type DraftReplyFunc          func(iid int, discussionID string, body string) error
type ResolveDiscussionFunc   func(iid int, discussionID string) error
type UnresolveDiscussionFunc func(iid int, discussionID string) error
type PostInlineCommentFunc   func(iid int, position mr.DiffPosition, body string) error
type PostMRCommentFunc       func(iid int, body string) error
type ApproveMRFunc           func(iid int) error
type MergeMRFunc             func(iid int) error
type EditMRFunc              func(iid int, title, description string) error
type OpenURLFunc             func(url string) error
type OpenEditorFunc          func(path string, line int) error

type ProjectData struct {
	Items           []mr.MergeRequest
	Refresh         RefreshFunc
	LoadDiff        DiffFunc
	LoadDiscussions LoadDiscussionsFunc
	LoadFiles       LoadFilesFunc
}

type ProjectOptions struct {
	Path                string
	Recents             []string
	Projects            []string
	Section             Section
	EntityID            string
	Refresh             RefreshFunc
	LoadDiff            DiffFunc
	LoadProject         ProjectLoadFunc
	LoadDiscussions     LoadDiscussionsFunc
	LoadFiles           LoadFilesFunc
	SubmitDrafts        SubmitDraftsFunc
	DiscardDrafts       DiscardDraftsFunc
	ReplyToDiscussion   ReplyToDiscussionFunc
	DraftReply          DraftReplyFunc
	ResolveDiscussion   ResolveDiscussionFunc
	UnresolveDiscussion UnresolveDiscussionFunc
	PostInlineComment   PostInlineCommentFunc
	PostMRComment       PostMRCommentFunc
	ApproveMR           ApproveMRFunc
	MergeMR             MergeMRFunc
	EditMR              EditMRFunc
	OpenURL             OpenURLFunc
	OpenEditor          OpenEditorFunc
}

type projectStartedMsg struct {
	path string
}

type projectFinishedMsg struct {
	path string
	data ProjectData
	err  error
}

type discussionsStartedMsg struct{ iid int }
type discussionsFinishedMsg struct {
	iid         int
	discussions []mr.Discussion
	err         error
}

type filesStartedMsg struct{ iid int }
type filesFinishedMsg struct {
	iid   int
	files []mr.ChangedFile
	err   error
}

type draftAddedMsg struct {
	iid   int
	draft mr.DraftComment
}

type draftsSubmittedMsg struct {
	iid int
	err error
}

type draftsDiscardedMsg struct {
	iid int
	err error
}

type replyFinishedMsg struct {
	iid          int
	discussionID string
	body         string
	draft        bool
	err          error
}

type resolveFinishedMsg struct {
	iid          int
	discussionID string
	resolved     bool
	err          error
}

type inlineCommentFinishedMsg struct {
	iid int
	err error
}

type mrCommentFinishedMsg struct {
	iid int
	err error
}

type approveMRFinishedMsg struct {
	iid int
	err error
}

type mergeMRFinishedMsg struct {
	iid int
	err error
}

type editMRFinishedMsg struct {
	iid         int
	title       string
	description string
	err         error
}

type openURLMsg struct {
	url string
	err error
}

type openEditorMsg struct {
	err error
}

type refreshStartedMsg struct{}

type refreshFinishedMsg struct {
	items []mr.MergeRequest
	err   error
}

type diffStartedMsg struct{}

type diffFinishedMsg struct {
	iid  int
	rows []mr.DiffRow
	err  error
}

type Model struct {
	items          []mr.MergeRequest
	query          string
	selected       int
	mode           Mode
	focus          Focus
	width          int
	height         int
	listTop        int
	rightTop       int
	projectPath    string
	recentProjects []string
	gitlabProjects []string
	projectList    []string
	section        Section
	sectionCursor  int
	entityID       string
	projectLoaded  bool
	activeTab           DetailTab
	discussions         map[int][]mr.Discussion
	changedFiles        map[int][]mr.ChangedFile
	selectedFile        int
	fileDiffTop         int
	diffCursor          int
	rangeStart          int
	commentInput        bool
	commentInstant      bool
	commentBuffer       string
	commentError        string
	mrCommentInput      bool
	mrCommentBuffer     string
	mrCommentError      string
	mergeConfirmPending bool
	editInput           bool
	editField           string
	editBuffer          string
	editTitle           string
	actionError         string
	postInlineComment   PostInlineCommentFunc
	postMRComment       PostMRCommentFunc
	approveMR           ApproveMRFunc
	mergeMR             MergeMRFunc
	editMR              EditMRFunc
	openURL             OpenURLFunc
	openEditor          OpenEditorFunc
	drafts              map[int][]mr.DraftComment
	submitDrafts        SubmitDraftsFunc
	discardDrafts       DiscardDraftsFunc
	discussionCursor    int
	replyInput          bool
	replyDraft          bool
	replyDiscussionID   string
	replyBuffer         string
	replyToDiscussion   ReplyToDiscussionFunc
	draftReply          DraftReplyFunc
	resolveDiscussion   ResolveDiscussionFunc
	unresolveDiscussion UnresolveDiscussionFunc
	discussionsLoading  bool
	filesLoading        bool
	discussionsError    string
	filesError          string
	loadDiscussions     LoadDiscussionsFunc
	loadFiles           LoadFilesFunc
	projectInput   string
	refresh        RefreshFunc
	loadDiff       DiffFunc
	loadProject    ProjectLoadFunc
	loading        bool
	projectLoading bool
	projectError   bool
	diffLoading    bool
	errorMessage   string
}

func NewFakeModel() Model {
	return NewModel(FakeMergeRequests())
}

func NewModel(items []mr.MergeRequest) Model {
	return NewModelWithProject(items, ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
}

func NewModelWithProject(items []mr.MergeRequest, options ProjectOptions) Model {
	model := Model{
		items:          items,
		focus:          FocusList,
		width:          100,
		height:         30,
		projectPath:    options.Path,
		recentProjects: options.Recents,
		gitlabProjects: options.Projects,
		projectList:    buildProjectList(options.Path, options.Recents, options.Projects),
		section:         options.Section,
		entityID:        options.EntityID,
		refresh:         options.Refresh,
		loadDiff:        options.LoadDiff,
		loadProject:     options.LoadProject,
		loadDiscussions: options.LoadDiscussions,
		loadFiles:       options.LoadFiles,
		discussions:     map[int][]mr.Discussion{},
		changedFiles:    map[int][]mr.ChangedFile{},
		drafts:          map[int][]mr.DraftComment{},
		rangeStart:      -1,
		submitDrafts:        options.SubmitDrafts,
		discardDrafts:       options.DiscardDrafts,
		replyToDiscussion:   options.ReplyToDiscussion,
		draftReply:          options.DraftReply,
		resolveDiscussion:   options.ResolveDiscussion,
		unresolveDiscussion: options.UnresolveDiscussion,
		postInlineComment:   options.PostInlineComment,
		postMRComment:       options.PostMRComment,
		approveMR:           options.ApproveMR,
		mergeMR:             options.MergeMR,
		editMR:              options.EditMR,
		openURL:             options.OpenURL,
		openEditor:          options.OpenEditor,
	}
	if model.projectPath == "" {
		if len(model.projectList) > 0 {
			model.mode = ModeProjectSelect
		} else {
			model.mode = ModeProjectInput
			model.focus = FocusFilter
		}
	} else if model.section == SectionMergeRequests {
		model.selectEntity()
		model.mode = ModeDetail
	} else {
		model.mode = ModeSections
	}
	return model
}

func Run(stdout io.Writer) error {
	return RunWithProject(stdout, ProjectOptions{Path: "group/project"})
}

func RunWithProject(stdout io.Writer, options ProjectOptions) error {
	program := tea.NewProgram(NewModelWithProject(nil, options), tea.WithMouseCellMotion(), tea.WithOutput(stdout))
	_, err := program.Run()
	return err
}

func (m Model) Init() tea.Cmd {
	if m.projectPath != "" && m.loadProject != nil && !m.projectLoaded && m.section == SectionMergeRequests {
		_, cmd := m.openProjectCommand(m.projectPath)
		return cmd
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		return m.updateKey(msg)
	case tea.MouseMsg:
		return m.updateMouse(msg)
	case projectStartedMsg:
		m.projectPath = msg.path
		m.mode = ModeDetail
		m.focus = FocusList
		m.loading = true
		m.projectLoading = true
		m.projectLoaded = false
		m.projectError = false
		m.items = nil
		m.errorMessage = ""
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
		m.refresh = msg.data.Refresh
		m.loadDiff = msg.data.LoadDiff
		m.loadDiscussions = msg.data.LoadDiscussions
		m.loadFiles = msg.data.LoadFiles
		m.selected = clampSelection(0, len(m.filtered()))
		m.selectEntity()
		m.listTop = 0
		m.rightTop = 0
		if m.section == SectionMergeRequests {
			m.mode = ModeDetail
		} else {
			m.mode = ModeSections
		}
		m.focus = FocusList
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
	if m.mode == ModeProjectSelect {
		switch msg.String() {
		case "up", "k":
			m.selected = clamp(m.selected-1, 0, len(m.projectList)-1)
		case "down", "j":
			m.selected = clamp(m.selected+1, 0, len(m.projectList)-1)
		case "enter":
			if len(m.projectList) > 0 {
				return m.openProjectCommand(m.projectList[m.selected])
			}
		case "i":
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
				return m.openProjectCommand(strings.TrimSpace(m.projectInput))
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
					m.mode = ModeDetail
					m.focus = FocusList
					return m, nil
				}
				return m.openProjectCommand(m.projectPath)
			}
		case "esc", "backspace":
			m.returnToProjectPicker()
		}
		return m, nil
	}

	if m.focus == FocusFilter {
		switch msg.Type {
		case tea.KeyEsc, tea.KeyEnter:
			m.focus = FocusList
		case tea.KeyBackspace:
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.selected = clampSelection(m.selected, len(m.filtered()))
			}
		case tea.KeyRunes:
			m.query += msg.String()
			m.selected = clampSelection(m.selected, len(m.filtered()))
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
			case tea.KeyRunes:
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
			case tea.KeyRunes:
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
		case "down", "j":
			rowCount := 0
			if len(files) > m.selectedFile {
				rowCount = len(files[m.selectedFile].Diff)
			}
			m.diffCursor = clamp(m.diffCursor+1, 0, max(0, rowCount-1))
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
			item, ok := m.selectedItem()
			if !ok {
				break
			}
			files := m.currentFiles()
			if len(files) <= m.selectedFile {
				break
			}
			file := files[m.selectedFile]
			if m.diffCursor >= len(file.Diff) {
				break
			}
			row := file.Diff[m.diffCursor]
			ds := m.discussions[item.IID]
			for _, d := range ds {
				if d.Position != nil && d.Position.NewPath == file.Path && d.Position.NewLine == row.NewLine && row.NewLine != 0 {
					m.replyInput = true
					m.replyDraft = isDraft
					m.replyDiscussionID = d.ID
					m.replyBuffer = ""
					break
				}
			}
		case "x":
			item, ok := m.selectedItem()
			if !ok {
				break
			}
			files := m.currentFiles()
			if len(files) <= m.selectedFile {
				break
			}
			file := files[m.selectedFile]
			if m.diffCursor >= len(file.Diff) {
				break
			}
			row := file.Diff[m.diffCursor]
			ds := m.discussions[item.IID]
			for i, d := range ds {
				if d.Position != nil && d.Position.NewPath == file.Path && d.Position.NewLine == row.NewLine && row.NewLine != 0 {
					iid := item.IID
					dID := d.ID
					resolved := !d.Resolved
					if resolved {
						fn := m.resolveDiscussion
						if fn == nil {
							m.discussions[iid][i].Resolved = true
							return m, nil
						}
						return m, func() tea.Msg {
							err := fn(iid, dID)
							return resolveFinishedMsg{iid: iid, discussionID: dID, resolved: true, err: err}
						}
					}
					fn := m.unresolveDiscussion
					if fn == nil {
						m.discussions[iid][i].Resolved = false
						return m, nil
					}
					return m, func() tea.Msg {
						err := fn(iid, dID)
						return resolveFinishedMsg{iid: iid, discussionID: dID, resolved: false, err: err}
					}
				}
			}
		case "p":
			item, ok := m.selectedItem()
			if ok && m.submitDrafts != nil {
				drafts := m.drafts[item.IID]
				submit := m.submitDrafts
				iid := item.IID
				return m, tea.Sequence(
					func() tea.Msg {
						err := submit(iid, drafts)
						return draftsSubmittedMsg{iid: iid, err: err}
					},
				)
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
		case "esc", "backspace":
			if m.rangeStart >= 0 {
				m.rangeStart = -1
				return m, nil
			}
			m.mode = ModeDetail
			m.activeTab = TabFiles
			m.fileDiffTop = 0
		}
		return m, nil
	}

	if m.mode == ModeDetail && m.editInput {
		return m.updateMREdit(msg)
	}

	if m.mode == ModeDetail && m.mrCommentInput {
		return m.updateMRCommentInput(msg)
	}

	if m.mode == ModeDetail && m.mergeConfirmPending && msg.String() != "M" {
		m.mergeConfirmPending = false
		return m, nil
	}

	if m.mode == ModeDetail && m.activeTab == TabDiscussions {
		return m.updateDiscussionsTab(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.projectError || (m.projectPath != "" && len(m.items) == 0) {
			m.errorMessage = ""
			m.returnToProjectPicker()
			return m, nil
		}
		if m.mode == ModeDiff {
			m.mode = ModeDetail
			m.rightTop = 0
		}
	case "/":
		m.focus = FocusFilter
	case "r":
		if m.projectError && m.projectPath != "" {
			return m.openProjectCommand(m.projectPath)
		}
		return m, m.refreshCommand()
	case "m":
		if m.mode == ModeDetail {
			m.mrCommentInput = true
			m.mrCommentBuffer = ""
			m.mrCommentError = ""
		}
	case "A":
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
	case "M":
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
	case "o":
		if m.mode == ModeDetail {
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
	case "e":
		if m.mode == ModeDetail {
			item, ok := m.selectedItem()
			if ok {
				m.editInput = true
				m.editField = "title"
				m.editBuffer = item.Title
				m.editTitle = ""
			}
		}
	case "tab":
		if m.mode == ModeDetail {
			m.activeTab = (m.activeTab + 1) % (TabFiles + 1)
			return m.onTabEntered()
		}
	case "up", "k":
		m.moveSelection(-1)
	case "down", "j":
		m.moveSelection(1)
	case "enter":
		if m.mode == ModeDetail && m.activeTab == TabFiles {
			if item, ok := m.selectedItem(); ok {
				if files, loaded := m.changedFiles[item.IID]; loaded && len(files) > 0 {
					m.mode = ModeFileDiff
					m.selectedFile = 0
					m.fileDiffTop = 0
					return m, nil
				}
			}
		}
		if item, ok := m.selectedItem(); ok {
			return m.openDiffCommand(item)
		}
	case "backspace":
		if m.mode == ModeDiff {
			m.mode = ModeDetail
			m.rightTop = 0
		}
	}

	return m, nil
}

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
	case tea.KeyRunes:
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

func (m Model) updateMRCommentInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mrCommentInput = false
		m.mrCommentBuffer = ""
	case tea.KeyBackspace:
		if len(m.mrCommentBuffer) > 0 {
			m.mrCommentBuffer = m.mrCommentBuffer[:len(m.mrCommentBuffer)-1]
		}
	case tea.KeyRunes:
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

func (m Model) focusedDiscussion() (mr.Discussion, bool) {
	item, ok := m.selectedItem()
	if !ok {
		return mr.Discussion{}, false
	}
	ds := m.discussions[item.IID]
	if m.discussionCursor < 0 || m.discussionCursor >= len(ds) {
		return mr.Discussion{}, false
	}
	return ds[m.discussionCursor], true
}

func (m Model) updateDiscussionsTab(msg tea.KeyMsg) (Model, tea.Cmd) {
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
		case tea.KeyRunes:
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

	item, ok := m.selectedItem()
	if !ok {
		return m, nil
	}
	ds := m.discussions[item.IID]
	count := len(ds)

	switch msg.String() {
	case "j", "down":
		m.discussionCursor = clamp(m.discussionCursor+1, 0, max(0, count-1))
	case "k", "up":
		m.discussionCursor = clamp(m.discussionCursor-1, 0, max(0, count-1))
	case "r":
		if d, ok := m.focusedDiscussion(); ok {
			m.replyInput = true
			m.replyDraft = false
			m.replyDiscussionID = d.ID
			m.replyBuffer = ""
		}
	case "d":
		if d, ok := m.focusedDiscussion(); ok {
			m.replyInput = true
			m.replyDraft = true
			m.replyDiscussionID = d.ID
			m.replyBuffer = ""
		}
	case "x":
		if d, ok := m.focusedDiscussion(); ok {
			iid := item.IID
			dID := d.ID
			resolved := !d.Resolved
			if resolved {
				fn := m.resolveDiscussion
				if fn == nil {
					m.discussions[iid][m.discussionCursor].Resolved = true
					return m, nil
				}
				return m, func() tea.Msg {
					err := fn(iid, dID)
					return resolveFinishedMsg{iid: iid, discussionID: dID, resolved: true, err: err}
				}
			}
			fn := m.unresolveDiscussion
			if fn == nil {
				m.discussions[iid][m.discussionCursor].Resolved = false
				return m, nil
			}
			return m, func() tea.Msg {
				err := fn(iid, dID)
				return resolveFinishedMsg{iid: iid, discussionID: dID, resolved: false, err: err}
			}
		}
	case "tab":
		m.activeTab = (m.activeTab + 1) % (TabFiles + 1)
		return m.onTabEntered()
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) returnToProjectPicker() {
	m.projectPath = ""
	if len(m.projectList) > 0 {
		m.mode = ModeProjectSelect
		m.focus = FocusList
		return
	}
	m.mode = ModeProjectInput
	m.focus = FocusFilter
}

func (m Model) openProjectCommand(path string) (Model, tea.Cmd) {
	m.projectPath = path
	m.mode = ModeDetail
	m.focus = FocusList
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

func (m Model) openDiffCommand(item mr.MergeRequest) (Model, tea.Cmd) {
	if m.loadDiff == nil || len(item.Diff) > 0 {
		m.mode = ModeDiff
		m.focus = FocusDetail
		m.rightTop = 0
		return m, nil
	}
	loadDiff := m.loadDiff
	return m, tea.Sequence(
		func() tea.Msg { return diffStartedMsg{} },
		func() tea.Msg {
			rows, err := loadDiff(item.IID)
			return diffFinishedMsg{iid: item.IID, rows: rows, err: err}
		},
	)
}

func (m Model) refreshCommand() tea.Cmd {
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

func (m Model) updateMouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	if m.mode == ModeProjectSelect {
		if msg.Button == tea.MouseButtonLeft && msg.Y >= 2 {
			idx := msg.Y - 2
			if idx >= 0 && idx < len(m.projectList) {
				m.selected = idx
				return m.openProjectCommand(m.projectList[idx])
			}
		}
		if msg.Button == tea.MouseButtonWheelUp {
			m.selected = clamp(m.selected-1, 0, len(m.projectList)-1)
		}
		if msg.Button == tea.MouseButtonWheelDown {
			m.selected = clamp(m.selected+1, 0, len(m.projectList)-1)
		}
		return m, nil
	}

	leftWidth := m.leftWidth()
	if msg.X < leftWidth {
		m.focus = FocusList
	} else {
		m.focus = FocusDetail
	}

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

func (m Model) View() string {
	if m.mode == ModeProjectSelect || m.mode == ModeProjectInput {
		return m.renderProjectPicker()
	}
	if m.mode == ModeSections {
		return lipgloss.JoinHorizontal(lipgloss.Top, m.renderProjectList(), m.renderSections())
	}
	if m.mode == ModeFileDiff {
		return lipgloss.JoinHorizontal(lipgloss.Top, m.renderChangedFilesPane(), m.renderFileDiffPane())
	}

	left := m.renderList()
	right := m.renderRight()
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func buildProjectList(opened string, recents []string, projects []string) []string {
	seen := map[string]bool{}
	list := []string{}
	candidates := []string{}
	if opened != "" {
		candidates = append(candidates, opened)
	}
	candidates = append(candidates, recents...)
	candidates = append(candidates, projects...)
	for _, project := range candidates {
		if project == "" || seen[project] {
			continue
		}
		seen[project] = true
		list = append(list, project)
	}
	return list
}

func (m Model) renderProjectList() string {
	width := m.leftWidth()
	style := paneStyle(width, max(8, m.height), false)
	lines := []string{"Projects", ""}
	for _, project := range m.projectList {
		prefix := "  "
		if project == m.projectPath {
			prefix = "> "
		}
		lines = append(lines, prefix+project)
	}
	if len(m.projectList) == 0 {
		lines = append(lines, "No projects")
	}
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderSections() string {
	width := max(20, m.width-m.leftWidth())
	style := paneStyle(width, max(8, m.height), true)
	lines := []string{"Sections", ""}
	for i, sec := range tuiSections {
		prefix := "  "
		if i == m.sectionCursor {
			prefix = "> "
		}
		label := sec.label
		if !sec.available {
			label += " (soon)"
		}
		lines = append(lines, prefix+label)
	}
	hint := "Enter: open  Esc: projects"
	if !tuiSections[m.sectionCursor].available {
		hint = "Not yet implemented  Esc: projects"
	}
	lines = append(lines, "", hint)
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderProjectPicker() string {
	style := paneStyle(max(40, m.width), max(8, m.height), true)
	if m.mode == ModeProjectInput {
		return style.Render(strings.Join([]string{
			"Open GitLab project",
			"",
			"Project path:",
			m.projectInput,
			"",
			"Enter: open project",
		}, "\n"))
	}

	lines := []string{"Projects", ""}
	for i, project := range m.projectList {
		prefix := "  "
		if i == m.selected {
			prefix = "> "
		}
		lines = append(lines, prefix+project)
	}
	lines = append(lines, "", "Enter/click: open  i: manual input")
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderList() string {
	width := m.leftWidth()
	height := max(8, m.height)
	style := paneStyle(width, height, m.focus == FocusList || m.focus == FocusFilter)
	lines := []string{"Project: " + m.projectPath, "Merge Requests", "Filter: " + m.query}
	if m.projectLoading {
		lines = append(lines, "Loading project…")
	} else if m.loading {
		lines = append(lines, "Refreshing…")
	}
	if m.errorMessage != "" {
		lines = append(lines, "Error: "+m.errorMessage)
		if m.projectError {
			lines = append(lines, "Esc: choose project")
		}
	}
	if m.diffLoading {
		lines = append(lines, "Loading diff…")
	}
	items := m.filtered()
	if len(items) == 0 {
		lines = append(lines, "No opened MRs")
		if len(m.items) == 0 && m.projectPath != "" {
			lines = append(lines, "r refresh  Esc: choose project")
		}
	} else {
		visible := max(1, height-5)
		end := min(len(items), m.listTop+visible)
		for i := m.listTop; i < end; i++ {
			prefix := "  "
			if i == m.selected {
				prefix = "> "
			}
			item := items[i]
			lines = append(lines, fmt.Sprintf("%s%s !%d %s", prefix, pipelineIcon(item.Pipeline), item.IID, item.Title))
			lines = append(lines, fmt.Sprintf("  %s %s → %s", item.Author, item.SourceBranch, item.TargetBranch))
		}
	}
	lines = append(lines, "", "↑/↓ select  / filter  r refresh  Enter diff")
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderRight() string {
	width := max(20, m.width-m.leftWidth())
	height := max(8, m.height)
	style := paneStyle(width, height, m.focus == FocusDetail)
	items := m.filtered()
	if len(items) == 0 {
		return style.Render("No MR selected")
	}
	item := items[clampSelection(m.selected, len(items))]
	if m.mode == ModeDiff {
		return style.Render(m.renderDiff(item))
	}
	tabs := "[Summary] [Discussions] [Files]"
	switch m.activeTab {
	case TabDiscussions:
		tabs = "[Summary] [>Discussions<] [Files]"
	case TabFiles:
		tabs = "[Summary] [Discussions] [>Files<]"
	default:
		tabs = "[>Summary<] [Discussions] [Files]"
	}
	header := fmt.Sprintf("!%d %s\n%s", item.IID, item.Title, tabs)

	switch m.activeTab {
	case TabDiscussions:
		return style.Render(header + "\n\n" + m.renderDiscussions(item))
	case TabFiles:
		return style.Render(header + "\n\n" + m.renderFiles(item))
	default:
		lines := []string{
			header,
			"",
			"Author: " + formatAuthor(item),
			"Branches: " + item.SourceBranch + " → " + item.TargetBranch,
			"State: " + item.State,
			"Pipeline: " + pipelineIcon(item.Pipeline) + " " + item.Pipeline,
			"Approvals: " + item.Approvals,
		}
		if item.WebURL != "" {
			lines = append(lines, "URL: "+item.WebURL)
		}
		if m.actionError != "" {
			lines = append(lines, "", "Action: "+m.actionError)
		}
		if m.mergeConfirmPending {
			lines = append(lines, "", "Press M again to confirm merge  (any other key cancels)")
		}
		if m.mrCommentError != "" {
			lines = append(lines, "", "Comment error: "+m.mrCommentError)
		}
		if m.editInput {
			lines = append(lines, "", fmt.Sprintf("Edit %s: %s█", m.editField, m.editBuffer), "Tab: next field  Enter: save  Esc: cancel")
		} else if m.mrCommentInput {
			lines = append(lines, "", "MR comment: "+m.mrCommentBuffer+"█", "Enter: send  Esc: cancel")
		} else {
			lines = append(lines, "", item.Description, "", "A: approve  M: merge  e: edit  o: open  m: comment  Tab: next tab")
		}
		return style.Render(strings.Join(lines, "\n"))
	}
}

func (m Model) renderDiscussions(item mr.MergeRequest) string {
	if m.discussionsLoading {
		return "Loading discussions…"
	}
	if m.discussionsError != "" {
		return "Error: " + m.discussionsError + "\n\nr retry"
	}
	discussions, loaded := m.discussions[item.IID]
	if !loaded {
		return "Tab to load discussions"
	}
	if len(discussions) == 0 {
		return "No discussions"
	}
	lines := []string{}
	for _, d := range discussions {
		status := "open"
		if d.Resolved {
			status = "resolved"
		}
		noteCount := len(d.Notes)
		firstAuthor := ""
		firstBody := ""
		if noteCount > 0 {
			firstAuthor = d.Notes[0].Author
			firstBody = d.Notes[0].Body
		}
		lines = append(lines, fmt.Sprintf("[%s] %s (%d notes)", status, firstAuthor, noteCount))
		lines = append(lines, "  "+firstBody)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderFiles(item mr.MergeRequest) string {
	if m.filesLoading {
		return "Loading files…"
	}
	if m.filesError != "" {
		return "Error: " + m.filesError + "\n\nr retry"
	}
	files, loaded := m.changedFiles[item.IID]
	if !loaded {
		return "Tab to load files"
	}
	if len(files) == 0 {
		return "No changed files"
	}
	lines := []string{}
	for _, f := range files {
		marker := " "
		if f.IsNew {
			marker = "A"
		} else if f.IsDeleted {
			marker = "D"
		} else if f.IsRenamed {
			marker = "R"
		}
		lines = append(lines, fmt.Sprintf("%s %s  +%d -%d", marker, f.Path, f.AddedLines, f.RemovedLines))
	}
	return strings.Join(lines, "\n")
}

func (m Model) currentFiles() []mr.ChangedFile {
	item, ok := m.selectedItem()
	if !ok {
		return nil
	}
	return m.changedFiles[item.IID]
}

func (m Model) renderChangedFilesPane() string {
	width := m.leftWidth()
	height := max(8, m.height)
	style := paneStyle(width, height, false)
	files := m.currentFiles()
	lines := []string{"Changed Files", ""}
	for i, f := range files {
		prefix := "  "
		if i == m.selectedFile {
			prefix = "> "
		}
		lines = append(lines, prefix+f.Path)
	}
	lines = append(lines, "", "←/→ files  Esc: back")
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderFileDiffPane() string {
	width := max(20, m.width-m.leftWidth())
	height := max(8, m.height)
	style := paneStyle(width, height, true)
	files := m.currentFiles()
	if len(files) == 0 {
		return style.Render("No files")
	}
	file := files[m.selectedFile]
	lines := []string{fmt.Sprintf("Diff %s", file.Path), ""}
	item, _ := m.selectedItem()
	discussions := m.discussions[item.IID]
	draftsForMR := m.drafts[item.IID]
	annotated := diff.ProjectDiscussions(file.Diff, discussions, file.Path)

	colWidth := max(10, (max(20, m.width-m.leftWidth())-20)/2)
	rowFmt := fmt.Sprintf("%%4d │ %%-%ds │ %%4d │ %%s", colWidth)
	for i, arow := range annotated {
		cursor := "  "
		if i == m.diffCursor {
			cursor = "> "
		}
		line := cursor + fmt.Sprintf(rowFmt, arow.OldLine, arow.OldText, arow.NewLine, arow.NewText)
		if len(arow.Discussions) > 0 {
			line += " 💬"
		}
		if m.rangeStart >= 0 && i >= m.rangeStart && i <= m.diffCursor {
			line += " ▌"
		}
		lines = append(lines, line)
		for _, d := range arow.Discussions {
			author := ""
			body := ""
			if len(d.Notes) > 0 {
				author = d.Notes[0].Author
				body = d.Notes[0].Body
			}
			lines = append(lines, fmt.Sprintf("  ↳ [%s] %s", author, body))
		}
		for _, dr := range draftsForMR {
			if dr.Position == nil || dr.Position.NewPath != file.Path || arow.NewLine == 0 {
				continue
			}
			startLine := dr.Position.NewLine
			endLine := dr.EndLine
			if endLine == 0 {
				endLine = startLine
			}
			if arow.NewLine >= startLine && arow.NewLine <= endLine {
				lines = append(lines, fmt.Sprintf("  [DRAFT] %s", dr.Body))
			}
		}
	}
	if m.commentError != "" {
		lines = append(lines, "", "Error: "+m.commentError)
	}
	if m.commentInput {
		prompt := "Comment"
		if m.commentInstant {
			prompt = "Instant comment"
		}
		lines = append(lines, "", prompt+": "+m.commentBuffer+"█", "Enter: send  Esc: cancel")
	} else {
		lines = append(lines, "", "i: instant  c: draft  v: range  p: publish  D: discard  Esc: back")
	}
	if m.fileDiffTop >= len(lines) {
		m.fileDiffTop = max(0, len(lines)-1)
	}
	visible := max(1, height-2)
	end := min(len(lines), m.fileDiffTop+visible)
	return style.Render(strings.Join(lines[m.fileDiffTop:end], "\n"))
}

func (m Model) renderDiff(item mr.MergeRequest) string {
	lines := []string{fmt.Sprintf("Diff !%d %s", item.IID, item.Title), ""}
	for _, row := range item.Diff {
		lines = append(lines, fmt.Sprintf("%4d │ %-36s │ %4d │ %s", row.OldLine, row.OldText, row.NewLine, row.NewText))
	}
	lines = append(lines, "", "Esc/backspace: back to detail")
	if m.rightTop >= len(lines) {
		m.rightTop = max(0, len(lines)-1)
	}
	visible := max(1, m.height-2)
	end := min(len(lines), m.rightTop+visible)
	return strings.Join(lines[m.rightTop:end], "\n")
}

func formatAuthor(item mr.MergeRequest) string {
	if item.AuthorUsername == "" || item.AuthorUsername == item.Author {
		return item.Author
	}
	return item.Author + " @" + item.AuthorUsername
}

func pipelineIcon(status string) string {
	switch status {
	case "success":
		return "✓"
	case "failed":
		return "✗"
	case "running":
		return "●"
	case "pending":
		return "○"
	default:
		return "–"
	}
}

func (m Model) filtered() []mr.MergeRequest {
	return mr.Filter(m.items, m.query)
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

func paneStyle(width int, height int, focused bool) lipgloss.Style {
	color := lipgloss.Color("240")
	if focused {
		color = lipgloss.Color("63")
	}
	return lipgloss.NewStyle().Width(width-2).Height(height-2).Border(lipgloss.RoundedBorder()).BorderForeground(color).Padding(0, 1)
}

func clampSelection(selected int, count int) int {
	if count <= 0 {
		return 0
	}
	return clamp(selected, 0, count-1)
}

func clamp(v int, minValue int, maxValue int) int {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
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
