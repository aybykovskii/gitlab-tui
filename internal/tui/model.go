package tui

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/diff"
	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type Mode int

const (
	ModeProjectSelect Mode = iota
	ModeProjectInput
	ModeSections
	ModeEntityList
	ModeDetail
	ModeDiff
	ModeFileDiff
	ModeLabelSelect
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
	{label: "Issues", id: SectionIssues, available: true},
	{label: "Pipelines", id: SectionPipelines, available: false},
}

type Focus int

const (
	FocusList Focus = iota
	FocusDetail
	FocusFilter
)

type RefreshFunc func() ([]mr.MergeRequest, error)
type LoadIssuesFunc func(state string, search string) ([]issue.Issue, error)
type LoadIssueDiscussionsFunc func(iid int) ([]issue.Discussion, error)
type DiffFunc func(iid int) ([]mr.DiffRow, error)
type ProjectLoadFunc func(path string) (ProjectData, error)
type AccountProjectsLoadFunc func() ([]string, error)
type LoadDiscussionsFunc func(iid int) ([]mr.Discussion, error)
type LoadFilesFunc func(iid int) ([]mr.ChangedFile, error)
type SubmitDraftsFunc func(iid int, drafts []mr.DraftComment) error
type DiscardDraftsFunc func(iid int) error
type ReplyToDiscussionFunc func(iid int, discussionID string, body string) error
type DraftReplyFunc func(iid int, discussionID string, body string) error
type ResolveDiscussionFunc func(iid int, discussionID string) error
type UnresolveDiscussionFunc func(iid int, discussionID string) error
type PostInlineCommentFunc func(iid int, position mr.DiffPosition, body string) error
type PostMRCommentFunc func(iid int, body string) error
type PostIssueCommentFunc func(iid int, body string) error
type IssueStateActionFunc func(iid int) error
type EditIssueFunc func(iid int, title, description string) error
type ApproveMRFunc func(iid int) error
type MergeMRFunc func(iid int) error
type EditMRFunc func(iid int, title, description string) error
type OpenURLFunc func(url string) error
type OpenEditorFunc func(path string, line int) error
type ToggleDraftMRFunc func(iid int) error
type UpdateMRLabelsFunc func(iid int, labels []string) error

type GlobalKeys struct {
	Quit         key.Binding
	Back         key.Binding
	ToggleKeyBar key.Binding
}

type ProjectListKeys struct {
	Up     key.Binding
	Down   key.Binding
	Open   key.Binding
	Filter key.Binding
	Input  key.Binding
	Retry  key.Binding
}

type SectionsKeys struct {
	Up   key.Binding
	Down key.Binding
	Open key.Binding
}

type EntityListKeys struct {
	Up      key.Binding
	Down    key.Binding
	Open    key.Binding
	Filter  key.Binding
	Refresh key.Binding
}

type MRDetailKeys struct {
	Approve     key.Binding
	Merge       key.Binding
	Edit        key.Binding
	OpenURL     key.Binding
	Comment     key.Binding
	NextTab     key.Binding
	ToggleDraft key.Binding
	LabelSelect key.Binding
}

type IssueDetailKeys struct {
	NextTab key.Binding
}

type DiffViewKeys struct {
	Up      key.Binding
	Down    key.Binding
	Comment key.Binding
	Range   key.Binding
	Publish key.Binding
}

type FileDiffKeys struct {
	PrevFile key.Binding
	NextFile key.Binding
	Up       key.Binding
	Down     key.Binding
	Comment  key.Binding
	Reply    key.Binding
}

func newGlobalKeys() GlobalKeys {
	return GlobalKeys{
		Quit:         key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Back:         key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "back")),
		ToggleKeyBar: key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "keys")),
	}
}

func newProjectListKeys() ProjectListKeys {
	return ProjectListKeys{
		Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Open:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "open")),
		Filter: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Input:  key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "manual")),
		Retry:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
	}
}

func (k ProjectListKeys) LocalKeys() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Open, k.Filter, k.Input, k.Retry}
}

func newSectionsKeys() SectionsKeys {
	return SectionsKeys{
		Up:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Open: key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "open")),
	}
}

func (k SectionsKeys) LocalKeys() []key.Binding { return []key.Binding{k.Up, k.Down, k.Open} }

func newEntityListKeys() EntityListKeys {
	return EntityListKeys{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Open:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "open")),
		Filter:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	}
}

func (k EntityListKeys) LocalKeys() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Open, k.Filter, k.Refresh}
}

func newMRDetailKeys() MRDetailKeys {
	return MRDetailKeys{
		Approve:     key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "approve")),
		Merge:       key.NewBinding(key.WithKeys("M"), key.WithHelp("M", "merge")),
		Edit:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		OpenURL:     key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open")),
		Comment:     key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "comment")),
		NextTab:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "next tab")),
		ToggleDraft: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "toggle draft")),
		LabelSelect: key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "labels")),
	}
}

func (k MRDetailKeys) LocalKeys() []key.Binding {
	return []key.Binding{k.Approve, k.Merge, k.Edit, k.OpenURL, k.Comment, k.NextTab, k.ToggleDraft, k.LabelSelect}
}

func newIssueDetailKeys() IssueDetailKeys {
	return IssueDetailKeys{NextTab: key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "next tab"))}
}

func (k IssueDetailKeys) LocalKeys() []key.Binding { return []key.Binding{k.NextTab} }

func newDiffViewKeys() DiffViewKeys {
	return DiffViewKeys{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Comment: key.NewBinding(key.WithKeys("i", "c"), key.WithHelp("i/c", "comment")),
		Range:   key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "range")),
		Publish: key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "publish")),
	}
}

func (k DiffViewKeys) LocalKeys() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Comment, k.Range, k.Publish}
}

func newFileDiffKeys() FileDiffKeys {
	return FileDiffKeys{
		PrevFile: key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev file")),
		NextFile: key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next file")),
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Comment:  key.NewBinding(key.WithKeys("i", "c"), key.WithHelp("i/c", "comment")),
		Reply:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reply")),
	}
}

func (k FileDiffKeys) LocalKeys() []key.Binding {
	return []key.Binding{k.PrevFile, k.NextFile, k.Up, k.Down, k.Comment, k.Reply}
}

type ProjectData struct {
	Items                []mr.MergeRequest
	Issues               []issue.Issue
	Labels               []mr.Label
	Refresh              RefreshFunc
	LoadIssues           LoadIssuesFunc
	PostIssueComment     PostIssueCommentFunc
	LoadIssueDiscussions LoadIssueDiscussionsFunc
	LoadDiff             DiffFunc
	LoadDiscussions      LoadDiscussionsFunc
	LoadFiles            LoadFilesFunc
	CloseIssue           IssueStateActionFunc
	ReopenIssue          IssueStateActionFunc
	EditIssue            EditIssueFunc
	AssignSelfIssue      IssueStateActionFunc
	UnassignSelfIssue    IssueStateActionFunc
	UpdateMRLabels       UpdateMRLabelsFunc
}

type AccountProjectLoader struct {
	ID   string
	Host string
	Load AccountProjectsLoadFunc
}

type RecentProjectOption struct {
	Path    string
	Account string
}

type ProjectOptions struct {
	Path                 string
	Recents              []string
	RecentProjects       []RecentProjectOption
	Projects             []string
	LoadProjects         []AccountProjectLoader
	Section              Section
	EntityID             string
	Issues               []issue.Issue
	Refresh              RefreshFunc
	LoadIssues           LoadIssuesFunc
	LoadDiff             DiffFunc
	LoadProject          ProjectLoadFunc
	LoadDiscussions      LoadDiscussionsFunc
	LoadFiles            LoadFilesFunc
	SubmitDrafts         SubmitDraftsFunc
	DiscardDrafts        DiscardDraftsFunc
	ReplyToDiscussion    ReplyToDiscussionFunc
	DraftReply           DraftReplyFunc
	ResolveDiscussion    ResolveDiscussionFunc
	UnresolveDiscussion  UnresolveDiscussionFunc
	PostInlineComment    PostInlineCommentFunc
	PostMRComment        PostMRCommentFunc
	PostIssueComment     PostIssueCommentFunc
	LoadIssueDiscussions LoadIssueDiscussionsFunc
	CloseIssue           IssueStateActionFunc
	ReopenIssue          IssueStateActionFunc
	EditIssue            EditIssueFunc
	AssignSelfIssue      IssueStateActionFunc
	UnassignSelfIssue    IssueStateActionFunc
	ApproveMR            ApproveMRFunc
	MergeMR              MergeMRFunc
	EditMR               EditMRFunc
	OpenURL              OpenURLFunc
	OpenEditor           OpenEditorFunc
	ToggleDraftMR        ToggleDraftMRFunc
	UpdateMRLabels       UpdateMRLabelsFunc
	Emoji                config.EmojiConfig
}

type projectStartedMsg struct {
	path string
}

type projectFinishedMsg struct {
	path string
	data ProjectData
	err  error
}

type accountProjectsStartedMsg struct{ accountID string }
type accountProjectsFinishedMsg struct {
	accountID string
	projects  []string
	err       error
}

type accountProjectState struct {
	host     string
	loading  bool
	err      string
	projects []string
}

type projectListRow struct {
	project    string
	label      string
	selectable bool
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

type issueStateFinishedMsg struct {
	iid   int
	state string
	err   error
}

type editIssueFinishedMsg struct {
	iid         int
	title       string
	description string
	err         error
}

type issueAssigneeFinishedMsg struct {
	iid       int
	assignees []string
	err       error
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

type toggleDraftFinishedMsg struct {
	iid  int
	prev bool
	err  error
}

type updateMRLabelsFinishedMsg struct {
	iid    int
	labels []string
	prev   []string
	err    error
}

type refreshStartedMsg struct{}

type refreshFinishedMsg struct {
	items []mr.MergeRequest
	err   error
}

type issuesFinishedMsg struct {
	items []issue.Issue
	err   error
}

type issueDiscussionsFinishedMsg struct {
	iid         int
	discussions []issue.Discussion
	err         error
}

type diffStartedMsg struct{}

type diffFinishedMsg struct {
	iid  int
	rows []mr.DiffRow
	err  error
}

type Model struct {
	items                []mr.MergeRequest
	query                string
	selected             int
	mode                 Mode
	focus                Focus
	width                int
	height               int
	listTop              int
	rightTop             int
	projectPath          string
	recentProjects       []string
	recentProjectOptions []RecentProjectOption
	gitlabProjects       []string
	projectList          []string
	projectRows          []projectListRow
	loadProjects         []AccountProjectLoader
	accountProjectStates map[string]accountProjectState
	section              Section
	sectionCursor        int
	entityID             string
	projectLoaded        bool
	activeTab            DetailTab
	discussions          map[int][]mr.Discussion
	issueDiscussions     map[int][]issue.Discussion
	changedFiles         map[int][]mr.ChangedFile
	selectedFile         int
	fileDiffTop          int
	diffCursor           int
	rangeStart           int
	commentInput         bool
	commentInstant       bool
	commentBuffer        string
	commentError         string
	mrCommentInput       bool
	mrCommentBuffer      string
	mrCommentError       string
	issueCommentInput    bool
	issueCommentBuffer   string
	issueCommentError    string
	mergeConfirmPending  bool
	editInput            bool
	editField            string
	editBuffer           string
	editTitle            string
	actionError          string
	postInlineComment    PostInlineCommentFunc
	postMRComment        PostMRCommentFunc
	postIssueComment     PostIssueCommentFunc
	loadIssueDiscussions LoadIssueDiscussionsFunc
	closeIssue           IssueStateActionFunc
	reopenIssue          IssueStateActionFunc
	editIssue            EditIssueFunc
	assignSelfIssue      IssueStateActionFunc
	unassignSelfIssue    IssueStateActionFunc
	approveMR            ApproveMRFunc
	mergeMR              MergeMRFunc
	editMR               EditMRFunc
	openURL              OpenURLFunc
	openEditor           OpenEditorFunc
	toggleDraftMR        ToggleDraftMRFunc
	updateMRLabels       UpdateMRLabelsFunc
	emoji                config.EmojiConfig
	projectLabels        []mr.Label
	labelCursor          int
	labelPending         []string
	drafts               map[int][]mr.DraftComment
	submitDrafts         SubmitDraftsFunc
	discardDrafts        DiscardDraftsFunc
	discussionCursor     int
	replyInput           bool
	replyDraft           bool
	replyDiscussionID    string
	replyBuffer          string
	replyToDiscussion    ReplyToDiscussionFunc
	draftReply           DraftReplyFunc
	resolveDiscussion    ResolveDiscussionFunc
	unresolveDiscussion  UnresolveDiscussionFunc
	discussionsLoading   bool
	filesLoading         bool
	discussionsError     string
	filesError           string
	loadDiscussions      LoadDiscussionsFunc
	loadFiles            LoadFilesFunc
	projectInput         string
	projectFilterActive  bool
	refresh              RefreshFunc
	issueItems           []issue.Issue
	issueState           string
	loadIssues           LoadIssuesFunc
	loadDiff             DiffFunc
	loadProject          ProjectLoadFunc
	loading              bool
	projectLoading       bool
	projectError         bool
	diffLoading          bool
	errorMessage         string
	keyBarExpanded       bool
	globals              GlobalKeys
	projectListKeys      ProjectListKeys
}

func NewFakeModel() Model {
	return NewModel(FakeMergeRequests())
}

func NewModel(items []mr.MergeRequest) Model {
	return NewModelWithProject(items, ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
}

func NewModelWithProject(items []mr.MergeRequest, options ProjectOptions) Model {
	projectListRecents := options.Recents
	if len(options.RecentProjects) > 0 {
		projectListRecents = nil
	}
	model := Model{
		items:                items,
		focus:                FocusDetail,
		width:                100,
		height:               30,
		projectPath:          options.Path,
		recentProjects:       options.Recents,
		recentProjectOptions: buildRecentProjectOptions(options.Recents, options.RecentProjects),
		gitlabProjects:       options.Projects,
		projectList:          buildProjectList(options.Path, projectListRecents, options.Projects),
		loadProjects:         options.LoadProjects,
		accountProjectStates: initialAccountProjectStates(options.LoadProjects),
		section:              options.Section,
		entityID:             options.EntityID,
		refresh:              options.Refresh,
		issueItems:           options.Issues,
		issueState:           "opened",
		loadIssues:           options.LoadIssues,
		loadDiff:             options.LoadDiff,
		loadProject:          options.LoadProject,
		loadDiscussions:      options.LoadDiscussions,
		loadFiles:            options.LoadFiles,
		discussions:          map[int][]mr.Discussion{},
		issueDiscussions:     map[int][]issue.Discussion{},
		changedFiles:         map[int][]mr.ChangedFile{},
		drafts:               map[int][]mr.DraftComment{},
		rangeStart:           -1,
		submitDrafts:         options.SubmitDrafts,
		discardDrafts:        options.DiscardDrafts,
		replyToDiscussion:    options.ReplyToDiscussion,
		draftReply:           options.DraftReply,
		resolveDiscussion:    options.ResolveDiscussion,
		unresolveDiscussion:  options.UnresolveDiscussion,
		postInlineComment:    options.PostInlineComment,
		postMRComment:        options.PostMRComment,
		postIssueComment:     options.PostIssueComment,
		loadIssueDiscussions: options.LoadIssueDiscussions,
		closeIssue:           options.CloseIssue,
		reopenIssue:          options.ReopenIssue,
		editIssue:            options.EditIssue,
		assignSelfIssue:      options.AssignSelfIssue,
		unassignSelfIssue:    options.UnassignSelfIssue,
		approveMR:            options.ApproveMR,
		mergeMR:              options.MergeMR,
		editMR:               options.EditMR,
		openURL:              options.OpenURL,
		openEditor:           options.OpenEditor,
		toggleDraftMR:        options.ToggleDraftMR,
		updateMRLabels:       options.UpdateMRLabels,
		emoji:                options.Emoji,
		globals:              newGlobalKeys(),
		projectListKeys:      newProjectListKeys(),
	}
	model.rebuildProjectRows()
	if model.projectPath == "" {
		if len(model.projectList) > 0 || len(model.recentProjectOptions) > 0 || len(model.loadProjects) > 0 {
			model.mode = ModeProjectSelect
			model.selected = model.nextSelectable(-1, 1)
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
	program := newProgram(NewModelWithProject(nil, options), stdout)
	_, err := program.Run()
	return err
}

func newProgram(model tea.Model, stdout io.Writer) *tea.Program {
	return tea.NewProgram(model, programOptions(stdout)...)
}

func programOptions(stdout io.Writer) []tea.ProgramOption {
	return []tea.ProgramOption{tea.WithMouseCellMotion(), tea.WithAltScreen(), tea.WithOutput(stdout)}
}

func (m Model) Init() tea.Cmd {
	if m.projectPath != "" && m.loadProject != nil && !m.projectLoaded && (m.section == SectionMergeRequests || m.section == SectionIssues) {
		_, cmd := m.openProjectCommand(m.projectPath)
		return cmd
	}
	if m.mode == ModeProjectSelect && len(m.loadProjects) > 0 {
		cmds := make([]tea.Cmd, 0, len(m.loadProjects))
		for _, loader := range m.loadProjects {
			cmds = append(cmds, loadAccountProjectsCommand(loader))
		}
		return tea.Batch(cmds...)
	}
	return nil
}

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
		m.mode = ModeDetail
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
			m.activeTab = (m.activeTab + 1) % (TabFiles + 1)
			return m.onTabEntered()
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
	case msg.String() == "backspace":
		if m.mode == ModeDiff {
			m.mode = ModeDetail
			m.rightTop = 0
		}
	}

	return m, nil
}

func (m Model) updateLabelSelect(msg tea.KeyMsg) (Model, tea.Cmd) {
	count := len(m.projectLabels)
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = ModeDetail
		m.labelPending = nil
		return m, nil
	case tea.KeyEnter:
		item, ok := m.selectedItem()
		if !ok {
			m.mode = ModeDetail
			return m, nil
		}
		selected := append([]string(nil), m.labelPending...)
		prev := append([]string(nil), item.Labels...)
		for i := range m.items {
			if m.items[i].IID == item.IID {
				m.items[i].Labels = selected
				break
			}
		}
		m.mode = ModeDetail
		m.labelPending = nil
		if m.updateMRLabels == nil {
			return m, nil
		}
		fn := m.updateMRLabels
		iid := item.IID
		return m, func() tea.Msg {
			err := fn(iid, selected)
			return updateMRLabelsFinishedMsg{iid: iid, labels: selected, prev: prev, err: err}
		}
	case tea.KeyRunes:
		switch msg.String() {
		case "k", "up":
			m.labelCursor = clamp(m.labelCursor-1, 0, max(0, count-1))
		case "j", "down":
			m.labelCursor = clamp(m.labelCursor+1, 0, max(0, count-1))
		case " ":
			if count > 0 {
				name := m.projectLabels[m.labelCursor].Name
				m.labelPending = toggleStringSlice(m.labelPending, name)
			}
		}
	case tea.KeyUp:
		m.labelCursor = clamp(m.labelCursor-1, 0, max(0, count-1))
	case tea.KeyDown:
		m.labelCursor = clamp(m.labelCursor+1, 0, max(0, count-1))
	case tea.KeySpace:
		if count > 0 {
			name := m.projectLabels[m.labelCursor].Name
			m.labelPending = toggleStringSlice(m.labelPending, name)
		}
	}
	return m, nil
}

func toggleStringSlice(slice []string, val string) []string {
	for i, s := range slice {
		if s == val {
			result := make([]string, 0, len(slice)-1)
			result = append(result, slice[:i]...)
			result = append(result, slice[i+1:]...)
			return result
		}
	}
	return append(slice, val)
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
	case tea.KeyRunes:
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
	case tea.KeyRunes:
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

func (m Model) selectedIssue() (issue.Issue, bool) {
	items := m.filteredIssues()
	if len(items) == 0 {
		return issue.Issue{}, false
	}
	return items[clampSelection(m.selected, len(items))], true
}

func (m Model) focusedIssueDiscussion() (issue.Discussion, bool) {
	item, ok := m.selectedIssue()
	if !ok {
		return issue.Discussion{}, false
	}
	ds := m.issueDiscussions[item.IID]
	if m.discussionCursor < 0 || m.discussionCursor >= len(ds) {
		return issue.Discussion{}, false
	}
	return ds[m.discussionCursor], true
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

func (m Model) updateIssueDiscussionsTab(msg tea.KeyMsg) (Model, tea.Cmd) {
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
			m.replyInput = false
			m.replyBuffer = ""
			m.replyDiscussionID = ""
		}
		return m, nil
	}

	switch {
	case msg.String() == "j" || msg.String() == "down":
		if item, ok := m.selectedIssue(); ok {
			count := len(m.issueDiscussions[item.IID])
			m.discussionCursor = clamp(m.discussionCursor+1, 0, max(0, count-1))
		}
	case msg.String() == "k" || msg.String() == "up":
		m.discussionCursor = clamp(m.discussionCursor-1, 0, max(0, m.discussionCursor))
	case msg.String() == "r":
		if d, ok := m.focusedIssueDiscussion(); ok {
			m.replyInput = true
			m.replyDraft = false
			m.replyDiscussionID = d.ID
			m.replyBuffer = ""
		}
	}
	return m, nil
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

	switch {
	case msg.String() == "j" || msg.String() == "down":
		m.discussionCursor = clamp(m.discussionCursor+1, 0, max(0, count-1))
	case msg.String() == "k" || msg.String() == "up":
		m.discussionCursor = clamp(m.discussionCursor-1, 0, max(0, count-1))
	case msg.String() == "r":
		if d, ok := m.focusedDiscussion(); ok {
			m.replyInput = true
			m.replyDraft = false
			m.replyDiscussionID = d.ID
			m.replyBuffer = ""
		}
	case msg.String() == "d":
		if d, ok := m.focusedDiscussion(); ok {
			m.replyInput = true
			m.replyDraft = true
			m.replyDiscussionID = d.ID
			m.replyBuffer = ""
		}
	case msg.String() == "x":
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
	case msg.String() == "tab":
		m.activeTab = (m.activeTab + 1) % (TabFiles + 1)
		return m.onTabEntered()
	case key.Matches(msg, m.globals.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) returnToProjectPicker() {
	m.projectPath = ""
	if len(m.projectList) > 0 {
		m.mode = ModeProjectSelect
		m.focus = FocusDetail
		return
	}
	m.mode = ModeProjectInput
	m.focus = FocusFilter
}

func (m Model) selectProject(path string) (Model, tea.Cmd) {
	m.projectPath = path
	m.mode = ModeSections
	m.focus = FocusDetail
	m.selected = 0
	m.listTop = 0
	m.rightTop = 0
	m.projectLoaded = false
	m.items = nil
	return m, nil
}

func (m Model) openProjectCommand(path string) (Model, tea.Cmd) {
	m.projectPath = path
	m.mode = ModeEntityList
	m.focus = FocusDetail
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
	if m.section == SectionIssues {
		return m.loadIssuesCommand()
	}
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

func (m Model) loadIssuesCommand() tea.Cmd {
	if m.loadIssues == nil || m.loading {
		return nil
	}
	loadIssues := m.loadIssues
	state := m.issueState
	return tea.Sequence(
		func() tea.Msg { return refreshStartedMsg{} },
		func() tea.Msg {
			items, err := loadIssues(state, "")
			return issuesFinishedMsg{items: items, err: err}
		},
	)
}

func (m Model) loadIssueDiscussionsCommand() tea.Cmd {
	if m.activeTab != TabDiscussions || m.loadIssueDiscussions == nil {
		return nil
	}
	item, ok := m.selectedIssue()
	if !ok {
		return nil
	}
	load := m.loadIssueDiscussions
	iid := item.IID
	return func() tea.Msg {
		discussions, err := load(iid)
		return issueDiscussionsFinishedMsg{iid: iid, discussions: discussions, err: err}
	}
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

func (m Model) View() string {
	var body string
	if m.mode == ModeProjectSelect || m.mode == ModeProjectInput {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderAppContextPane(), m.renderProjectPicker())
	} else if m.mode == ModeSections {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderProjectList(), m.renderSections())
	} else if m.mode == ModeEntityList {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderSectionsContext(), m.renderEntityListPane())
	} else if m.mode == ModeFileDiff {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderChangedFilesPane(), m.renderFileDiffPane())
	} else if m.mode == ModeLabelSelect {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderList(), m.renderLabelSelector())
	} else {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderList(), m.renderRight())
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, m.renderKeyBar())
}

func initialAccountProjectStates(loaders []AccountProjectLoader) map[string]accountProjectState {
	states := map[string]accountProjectState{}
	for _, loader := range loaders {
		states[loader.ID] = accountProjectState{host: loader.Host, loading: true}
	}
	return states
}

func loadAccountProjectsCommand(loader AccountProjectLoader) tea.Cmd {
	return func() tea.Msg {
		projects, err := loader.Load()
		return accountProjectsFinishedMsg{accountID: loader.ID, projects: projects, err: err}
	}
}

func buildRecentProjectOptions(recents []string, recentProjects []RecentProjectOption) []RecentProjectOption {
	if len(recentProjects) > 0 {
		return recentProjects
	}
	options := make([]RecentProjectOption, 0, len(recents))
	for _, recent := range recents {
		options = append(options, RecentProjectOption{Path: recent})
	}
	return options
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
	style := paneStyle(width, m.paneHeight(), false)
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
	style := paneStyle(width, m.paneHeight(), true)
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
	if !tuiSections[m.sectionCursor].available {
		lines = append(lines, "", "Not yet implemented")
	}
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderSectionsContext() string {
	width := m.leftWidth()
	style := paneStyle(width, m.paneHeight(), false)
	lines := []string{"Sections", ""}
	for _, sec := range tuiSections {
		prefix := "  "
		if sec.id == m.section {
			prefix = "> "
		}
		lines = append(lines, prefix+sec.label)
	}
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderEntityListPane() string {
	width := max(20, m.width-m.leftWidth())
	height := m.paneHeight()
	style := paneStyle(width, height, true)
	if m.section == SectionIssues {
		return style.Render(strings.Join(m.issueListLines(height), "\n"))
	}
	lines := []string{"Project: " + m.projectPath, "Merge Requests", "Filter: " + m.query}
	if m.projectLoading {
		lines = append(lines, "Loading project…")
	} else if m.loading {
		lines = append(lines, "Refreshing…")
	}
	if m.errorMessage != "" {
		lines = append(lines, "Error: "+m.errorMessage)
	}
	items := m.filtered()
	if len(items) == 0 {
		lines = append(lines, "No opened MRs")
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
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) issueListLines(height int) []string {
	lines := []string{"Project: " + m.projectPath, "Issues [" + m.issueStateLabel() + "]", "Filter: " + m.query}
	if m.loading {
		lines = append(lines, "Refreshing…")
	}
	if m.errorMessage != "" {
		lines = append(lines, "Error: "+m.errorMessage)
	}
	items := m.filteredIssues()
	if len(items) == 0 {
		lines = append(lines, "No issues")
		return lines
	}
	visible := max(1, (height-5)/2)
	end := min(len(items), m.listTop+visible)
	for i := m.listTop; i < end; i++ {
		prefix := "  "
		if i == m.selected {
			prefix = "> "
		}
		item := items[i]
		lines = append(lines, fmt.Sprintf("%s#%d %s", prefix, item.IID, item.Title))
		lines = append(lines, "  "+formatIssueMeta(item))
	}
	return lines
}

func (m Model) inputActive() bool {
	return m.projectFilterActive || m.mode == ModeProjectInput || m.commentInput || m.mrCommentInput || m.issueCommentInput || m.editInput || m.replyInput || m.focus == FocusFilter
}

func (m *Model) syncGlobalKeys() {
	active := m.inputActive()
	m.globals.Quit.SetEnabled(!active)
	m.globals.Back.SetEnabled(!active)
}

func bindingHelp(binding key.Binding) string {
	if !binding.Enabled() {
		return ""
	}
	k, d := binding.Help().Key, binding.Help().Desc
	if k == "" || d == "" {
		return ""
	}
	return k + " " + d
}

func joinBindingHelp(bindings []key.Binding) string {
	parts := []string{}
	for _, binding := range bindings {
		if help := bindingHelp(binding); help != "" {
			parts = append(parts, help)
		}
	}
	return strings.Join(parts, "  ")
}

func truncateLine(line string, width int) string {
	if width <= 0 || len(line) <= width {
		return line
	}
	if width == 1 {
		return "…"
	}
	return line[:width-1] + "…"
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

func (m Model) renderKeyBar() string {
	width := max(20, m.width)
	inner := max(1, width-4)
	lines := []string{}
	if m.keyBarExpanded {
		locals := m.localKeys()
		mid := (len(locals) + 1) / 2
		for i := 0; i < mid; i++ {
			left := bindingHelp(locals[i])
			right := ""
			if i+mid < len(locals) {
				right = bindingHelp(locals[i+mid])
			}
			lines = append(lines, fmt.Sprintf("%-24s %s", left, right))
		}
		lines = append(lines, strings.Repeat("─", min(inner, 24)))
		lines = append(lines, "Global: "+joinBindingHelp(m.globalKeys()))
	} else {
		lines = append(lines, truncateLine("Local: "+joinBindingHelp(m.localKeys()), inner))
		lines = append(lines, truncateLine("Global: "+joinBindingHelp(m.globalKeys()), inner))
	}
	style := lipgloss.NewStyle().Width(width - 2).Border(lipgloss.RoundedBorder())
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderAppContextPane() string {
	width := m.leftWidth()
	style := paneStyle(width, m.paneHeight(), false)
	return style.Render("gitlab-tui")
}

func (m *Model) rebuildProjectRows() {
	m.projectRows = nil
	if len(m.filteredRecentProjects()) > 0 {
		m.projectRows = append(m.projectRows, projectListRow{label: "Recent"}, projectListRow{})
		for _, recent := range m.filteredRecentProjects() {
			label := recent.Path
			if recent.Account != "" {
				label += " (" + recent.Account + ")"
			}
			m.projectRows = append(m.projectRows, projectListRow{project: recent.Path, label: label, selectable: true})
		}
	}
	for _, project := range m.projectList {
		if m.matchesProjectFilter(project) {
			m.projectRows = append(m.projectRows, projectListRow{project: project, label: project, selectable: true})
		}
	}
	if len(m.projectRows) > 0 && len(m.loadProjects) > 0 {
		m.projectRows = append(m.projectRows, projectListRow{})
	}
	for _, loader := range m.loadProjects {
		state := m.accountProjectStates[loader.ID]
		projects := filteredProjectPaths(state.projects[:min(len(state.projects), 15)], m.query)
		showStatus := !m.projectFilterActive && len(projects) == 0
		if len(projects) == 0 && !showStatus {
			continue
		}
		header := fmt.Sprintf("[%s]  %s", loader.ID, state.host)
		m.projectRows = append(m.projectRows, projectListRow{label: header})
		if state.loading && showStatus {
			m.projectRows = append(m.projectRows, projectListRow{label: "Loading…"})
			continue
		}
		if state.err != "" && showStatus {
			m.projectRows = append(m.projectRows, projectListRow{label: "Error: " + state.err + "  r: retry"})
			continue
		}
		for _, project := range projects {
			m.projectRows = append(m.projectRows, projectListRow{project: project, label: project, selectable: true})
		}
	}
}

func (m Model) filteredRecentProjects() []RecentProjectOption {
	projects := make([]RecentProjectOption, 0, len(m.recentProjectOptions))
	for _, recent := range m.recentProjectOptions {
		if m.matchesProjectFilter(recent.Path) {
			projects = append(projects, recent)
		}
	}
	return projects
}

func filteredProjectPaths(projects []string, query string) []string {
	if strings.TrimSpace(query) == "" {
		return projects
	}
	filtered := make([]string, 0, len(projects))
	needle := strings.ToLower(query)
	for _, project := range projects {
		if strings.Contains(strings.ToLower(project), needle) {
			filtered = append(filtered, project)
		}
	}
	return filtered
}

func (m Model) matchesProjectFilter(project string) bool {
	if strings.TrimSpace(m.query) == "" {
		return true
	}
	return strings.Contains(strings.ToLower(project), strings.ToLower(m.query))
}

func (m Model) nextSelectable(from int, delta int) int {
	if len(m.projectRows) == 0 {
		return 0
	}
	for i := clamp(from+delta, 0, len(m.projectRows)-1); i >= 0 && i < len(m.projectRows); i += delta {
		if m.projectRows[i].selectable {
			return i
		}
		if i == 0 && delta < 0 || i == len(m.projectRows)-1 && delta > 0 {
			break
		}
	}
	return from
}

func (m Model) nearestSelectable(index int) int {
	if len(m.projectRows) == 0 {
		return 0
	}
	if index >= 0 && index < len(m.projectRows) && m.projectRows[index].selectable {
		return index
	}
	if next := m.nextSelectable(index, 1); next != index {
		return next
	}
	return m.nextSelectable(index, -1)
}

func (m Model) selectedProject() (string, bool) {
	if m.selected < 0 || m.selected >= len(m.projectRows) || !m.projectRows[m.selected].selectable {
		return "", false
	}
	return m.projectRows[m.selected].project, true
}

func (m Model) retryFailedProjectLoads() tea.Cmd {
	cmds := []tea.Cmd{}
	for _, loader := range m.loadProjects {
		if state := m.accountProjectStates[loader.ID]; state.err != "" {
			cmds = append(cmds, loadAccountProjectsCommand(loader))
		}
	}
	if len(cmds) == 1 {
		return cmds[0]
	}
	return tea.Batch(cmds...)
}

func (m Model) renderProjectPicker() string {
	width := max(20, m.width-m.leftWidth())
	style := paneStyle(width, m.paneHeight(), true)
	if m.mode == ModeProjectInput {
		return style.Render(strings.Join([]string{
			"Open GitLab project",
			"",
			"Project path:",
			m.projectInput,
		}, "\n"))
	}

	lines := []string{"Projects", ""}
	if m.projectFilterActive || m.query != "" {
		lines = append(lines, "Filter: "+m.query, "")
	}
	if len(m.projectRows) == 0 {
		lines = append(lines, "No matching projects")
	}
	for i, row := range m.projectRows {
		prefix := "  "
		if i == m.selected && row.selectable {
			prefix = "> "
		}
		lines = append(lines, prefix+row.label)
	}
	lines = append(lines, "", "Enter/click: open  i: manual input  r: retry")
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderList() string {
	width := m.leftWidth()
	height := m.paneHeight()
	style := paneStyle(width, height, m.focus == FocusList || m.focus == FocusFilter)
	lines := []string{"Project: " + m.projectPath, "Merge Requests", "Filter: " + m.query}
	if m.projectLoading {
		lines = append(lines, "Loading project…")
	} else if m.loading {
		lines = append(lines, "Refreshing…")
	}
	if m.errorMessage != "" {
		lines = append(lines, "Error: "+m.errorMessage)
	}
	if m.diffLoading {
		lines = append(lines, "Loading diff…")
	}
	items := m.filtered()
	if len(items) == 0 {
		lines = append(lines, "No opened MRs")
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
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderRight() string {
	width := max(20, m.width-m.leftWidth())
	height := m.paneHeight()
	style := paneStyle(width, height, m.focus == FocusDetail)
	if m.section == SectionIssues {
		return style.Render(m.renderIssueDetail())
	}
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
	icons := m.emoji.Resolve()
	header := fmt.Sprintf("%s\n%s", mrTitleLine(item, icons), tabs)

	switch m.activeTab {
	case TabDiscussions:
		return style.Render(header + "\n\n" + m.renderDiscussions(item))
	case TabFiles:
		return style.Render(header + "\n\n" + m.renderFiles(item))
	default:
		authorPart := iconPrefix(icons.Author) + formatAuthor(item)
		reviewerPart := ""
		if len(item.Reviewers) > 0 {
			reviewerPart = iconPrefix(icons.Reviewers) + strings.Join(item.Reviewers, ", ")
		}
		assigneePart := ""
		if len(item.Assignees) > 0 {
			assigneePart = iconPrefix(icons.Assignees) + strings.Join(item.Assignees, ", ")
		}
		peopleParts := []string{authorPart}
		if reviewerPart != "" {
			peopleParts = append(peopleParts, reviewerPart)
		}
		if assigneePart != "" {
			peopleParts = append(peopleParts, assigneePart)
		}

		statePart := stateEmoji(icons, item.State) + " " + item.State
		if icons.State == "" {
			statePart = item.State
		}
		pipelinePart := iconPrefix(icons.Pipeline) + pipelineIcon(item.Pipeline) + " " + item.Pipeline
		approvalsPart := iconPrefix(icons.Approvals) + item.Approvals

		lines := []string{
			header,
			"",
			strings.Join(peopleParts, "  ·  "),
			iconPrefix(icons.Branch) + item.SourceBranch + " → " + item.TargetBranch,
			strings.Join([]string{statePart, pipelinePart, approvalsPart}, "  ·  "),
		}
		if len(item.Labels) > 0 {
			pills := make([]string, 0, len(item.Labels))
			for _, name := range item.Labels {
				color := m.labelColor(name)
				pills = append(pills, renderLabelPill(name, color))
			}
			lines = append(lines, iconPrefix(icons.Labels)+strings.Join(pills, " "))
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
			lines = append(lines, "", fmt.Sprintf("Edit %s: %s█", m.editField, m.editBuffer))
		} else if m.mrCommentInput {
			lines = append(lines, "", "MR comment: "+m.mrCommentBuffer+"█")
		} else {
			lines = append(lines, "", item.Description)
		}
		return style.Render(strings.Join(lines, "\n"))
	}
}

func (m Model) renderIssueDetail() string {
	items := m.filteredIssues()
	if len(items) == 0 {
		return "No issue selected"
	}
	item := items[clampSelection(m.selected, len(items))]
	tabs := "[>Summary<] [Discussions]"
	if m.activeTab == TabDiscussions {
		tabs = "[Summary] [>Discussions<]"
	}
	header := fmt.Sprintf("#%d %s\n%s", item.IID, item.Title, tabs)
	if m.activeTab == TabDiscussions {
		return header + "\n\n" + m.renderIssueDiscussions(item)
	}
	lines := []string{
		header,
		"",
		"👤 " + item.Author + " · assigned: " + strings.Join(item.Assignees, ", "),
		issueStateIcon(item.State) + " " + item.State + fmt.Sprintf(" · 💬 %d", item.CommentCount),
		"🏷️ " + formatIssueLabels(item.Labels),
		"📅 Due: " + item.DueDate + " · 🏁 " + item.Milestone,
	}
	if item.Weight > 0 {
		lines = append(lines, fmt.Sprintf("⚖️ Weight: %d", item.Weight))
	}
	if m.editInput {
		lines = append(lines, "", fmt.Sprintf("Edit %s: %s█", m.editField, m.editBuffer))
	} else if m.issueCommentInput {
		lines = append(lines, "", "Issue comment: "+m.issueCommentBuffer+"█")
	} else {
		lines = append(lines, "", item.Description)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderIssueDiscussions(item issue.Issue) string {
	discussions, loaded := m.issueDiscussions[item.IID]
	if !loaded {
		return "No discussions"
	}
	if len(discussions) == 0 {
		return "No discussions"
	}
	lines := []string{}
	for i, d := range discussions {
		cursor := "  "
		if i == m.discussionCursor {
			cursor = "> "
		}
		firstAuthor := ""
		if len(d.Notes) > 0 {
			firstAuthor = d.Notes[0].Author
		}
		lines = append(lines, cursor+firstAuthor)
		for j, note := range d.Notes {
			if j == 0 {
				lines = append(lines, "  "+note.Body)
			} else {
				lines = append(lines, "  ↳ "+note.Author+": "+note.Body)
			}
		}
	}
	if m.replyInput {
		lines = append(lines, "", "Reply: "+m.replyBuffer+"█")
	}
	return strings.Join(lines, "\n")
}

func issueStateIcon(state string) string {
	if state == "closed" {
		return "🔴"
	}
	return "🟢"
}

func formatIssueLabels(labels []string) string {
	parts := make([]string, 0, len(labels))
	for _, label := range labels {
		parts = append(parts, "["+label+"]")
	}
	return strings.Join(parts, " ")
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
	sep := "─────────────────────────────────────────"
	lines := []string{}
	for i, d := range discussions {
		if i > 0 {
			lines = append(lines, sep)
		}
		status := "open"
		if d.Resolved {
			status = "resolved"
		}
		cursor := "  "
		if i == m.discussionCursor {
			cursor = "> "
		}
		firstAuthor := ""
		if len(d.Notes) > 0 {
			firstAuthor = d.Notes[0].Author
		}
		lines = append(lines, fmt.Sprintf("%s[%s] %s", cursor, status, firstAuthor))
		for j, note := range d.Notes {
			if j == 0 {
				lines = append(lines, "  "+note.Body)
			} else {
				lines = append(lines, "  ↳ "+note.Author+": "+note.Body)
			}
		}
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
	height := m.paneHeight()
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
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderFileDiffPane() string {
	width := max(20, m.width-m.leftWidth())
	height := m.paneHeight()
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
	addStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	delStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	ctxStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	rowFmt := fmt.Sprintf("%%s │ %%-%ds │ %%s │ %%s", colWidth)
	for i, arow := range annotated {
		cursor := "  "
		if i == m.diffCursor {
			cursor = "> "
		}
		draftMarker := m.draftGutterMarker(file.Path, arow.NewLine, draftsForMR)
		if draftMarker == " " && m.isActiveDraftRangeRow(i) {
			draftMarker = "·"
		}
		discussionMarker := m.discussionGutterMarker(arow.Discussions)
		var oldNum, newNum, oldContent, newContent string
		var rowStyle lipgloss.Style
		switch {
		case arow.OldLine == 0 && arow.NewLine > 0: // addition
			oldNum = "    "
			newNum = fmt.Sprintf("%4d", arow.NewLine)
			oldContent = strings.Repeat(" ", colWidth)
			newContent = "+ " + arow.NewText
			rowStyle = addStyle
		case arow.NewLine == 0 && arow.OldLine > 0: // deletion
			oldNum = fmt.Sprintf("%4d", arow.OldLine)
			newNum = "    "
			oldContent = "- " + arow.OldText
			newContent = ""
			rowStyle = delStyle
		default: // context
			oldNum = fmt.Sprintf("%4d", arow.OldLine)
			newNum = fmt.Sprintf("%4d", arow.NewLine)
			oldContent = "  " + arow.OldText
			newContent = "  " + arow.NewText
			rowStyle = ctxStyle
		}
		lineContent := fmt.Sprintf(rowFmt, oldNum, oldContent, newNum, newContent)
		line := cursor + draftMarker + discussionMarker + " " + rowStyle.Render(lineContent)
		lines = append(lines, line)
	}
	if m.commentError != "" {
		lines = append(lines, "", "Error: "+m.commentError)
	}
	if m.commentInput {
		prompt := "Comment"
		if m.commentInstant {
			prompt = "Instant comment"
		}
		lines = append(lines, "", prompt+": "+m.commentBuffer+"█")
	}
	if m.fileDiffTop >= len(lines) {
		m.fileDiffTop = max(0, len(lines)-1)
	}
	visible := max(1, height-2)
	end := min(len(lines), m.fileDiffTop+visible)
	return style.Render(strings.Join(lines[m.fileDiffTop:end], "\n"))
}

func (m Model) draftGutterMarker(path string, newLine int, drafts []mr.DraftComment) string {
	if newLine == 0 {
		return " "
	}
	for _, dr := range drafts {
		if dr.Position == nil || dr.Position.NewPath != path {
			continue
		}
		startLine := dr.Position.NewLine
		endLine := dr.EndLine
		if endLine == 0 {
			endLine = startLine
		}
		if newLine >= startLine && newLine <= endLine {
			if m.emoji.Enabled {
				icon := m.emoji.Resolve().Draft
				if icon != "" {
					return icon
				}
			}
			return "●"
		}
	}
	return " "
}

func (m Model) discussionGutterMarker(discussions []mr.Discussion) string {
	if len(discussions) == 0 {
		return " "
	}
	if m.emoji.Enabled {
		return "💬"
	}
	return "○"
}

func (m Model) isActiveDraftRangeRow(index int) bool {
	if m.rangeStart < 0 {
		return false
	}
	start, end := m.rangeStart, m.diffCursor
	if start > end {
		start, end = end, start
	}
	return index >= start && index <= end
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

func (m Model) renderLabelSelector() string {
	width := max(20, m.width-m.leftWidth())
	height := m.paneHeight()
	style := paneStyle(width, height, true)
	lines := []string{"Labels  Space toggle  Enter save  Esc cancel", ""}
	for i, l := range m.projectLabels {
		marker := "○"
		for _, sel := range m.labelPending {
			if sel == l.Name {
				marker = "●"
				break
			}
		}
		cursor := "  "
		if i == m.labelCursor {
			cursor = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%s %s", cursor, marker, renderLabelPill(l.Name, l.Color)))
	}
	if len(m.projectLabels) == 0 {
		lines = append(lines, "No project labels")
	}
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) labelColor(name string) string {
	for _, l := range m.projectLabels {
		if l.Name == name {
			return l.Color
		}
	}
	return ""
}

func mrTitleLine(item mr.MergeRequest, icons config.EmojiMap) string {
	prefix := ""
	if icons.Draft != "" {
		prefix = icons.Draft + " "
	}
	title := item.Title
	if item.Draft {
		title = "Draft: " + title
	}
	return fmt.Sprintf("%s!%d %s", prefix, item.IID, title)
}

func stateEmoji(icons config.EmojiMap, state string) string {
	if icons.State == "" {
		return ""
	}
	switch state {
	case "opened":
		return "🟢"
	case "merged":
		return "🟣"
	case "closed":
		return "🔴"
	default:
		return icons.State
	}
}

func iconPrefix(icon string) string {
	if icon == "" {
		return ""
	}
	return icon + " "
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

func (m Model) filteredIssues() []issue.Issue {
	query := strings.ToLower(strings.TrimSpace(m.query))
	if query == "" {
		return m.issueItems
	}
	filtered := make([]issue.Issue, 0, len(m.issueItems))
	for _, item := range m.issueItems {
		text := strings.ToLower(item.Title + " " + item.Author)
		if strings.Contains(text, query) {
			filtered = append(filtered, item)
		}
	}
	return filtered
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

func formatIssueMeta(item issue.Issue) string {
	parts := []string{item.Author}
	labels := item.Labels
	if len(labels) > 2 {
		labels = labels[:2]
	}
	labelParts := make([]string, 0, len(labels))
	for _, label := range labels {
		labelParts = append(labelParts, "["+label+"]")
	}
	if len(labelParts) > 0 {
		parts = append(parts, strings.Join(labelParts, " "))
	}
	if item.CommentCount > 0 {
		parts = append(parts, fmt.Sprintf("💬 %d", item.CommentCount))
	}
	return strings.Join(parts, " · ")
}

func (m Model) clampEntitySelection(selected int) int {
	if m.section == SectionIssues {
		return clampSelection(selected, len(m.filteredIssues()))
	}
	return clampSelection(selected, len(m.filtered()))
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
