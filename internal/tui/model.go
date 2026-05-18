package tui

import (
	"io"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

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
	fileDiffReturnTab    DetailTab
	rangeStart           int
	reviewCursor         int
	reviewSummaryInput   bool
	reviewSummary        string
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
	loadProject          ProjectLoadFunc
	loading              bool
	projectLoading       bool
	projectError         bool
	errorMessage         string
	keyBarExpanded       bool
	threadPanelVisible   bool
	threadPanelCursor    int
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
		threadPanelVisible:   true,
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
	return []tea.ProgramOption{tea.WithAltScreen(), tea.WithOutput(stdout)}
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

// renderDiscussionBlock returns lines for a single discussion block.
// header is the first line text (cursor will be prepended).
// cursor is "  " or "> ".
// dimResolved dims all lines when the discussion is resolved.
// authorInFirstNote includes the author name in the first note line
// (use when the header does not already identify the author, e.g. Thread Panel).
