//nolint:mnd // Default terminal dimensions are intentional UI fallbacks.
package tui

import (
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type Model struct {
	EntityListState
	mode   Mode
	focus  Focus
	width  int
	height int
	MRDetailState
	IssueDetailState IssueDetailState
	DiffViewState
	ProjectPickerState
	LabelSelectorState
	InputState
	projectList          []string
	section              Section
	sectionList          list.Model
	entityID             string
	projectLoaded        bool
	mergeConfirmPending  bool
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
	submitDrafts         SubmitDraftsFunc
	discardDrafts        DiscardDraftsFunc
	replyToDiscussion    ReplyToDiscussionFunc
	draftReply           DraftReplyFunc
	resolveDiscussion    ResolveDiscussionFunc
	unresolveDiscussion  UnresolveDiscussionFunc
	loadDiscussions      LoadDiscussionsFunc
	loadFiles            LoadFilesFunc
	refresh              RefreshFunc
	issueState           string
	loadIssues           LoadIssuesFunc
	loadProject          ProjectLoadFunc
	loading              bool
	projectLoading       bool
	projectError         bool
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

//nolint:gocritic // Model startup branches are clearer as explicit UI flow.
func NewModelWithProject(items []mr.MergeRequest, options ProjectOptions) Model {
	projectListRecents := options.Recents
	if len(options.RecentProjects) > 0 {
		projectListRecents = nil
	}

	sectionItems := make([]list.Item, len(tuiSections))
	for i, sec := range tuiSections {
		sectionItems[i] = sectionListItem{sec}
	}

	sectionList := newCompactFancyList("Sections", newSectionListDelegate())
	_ = sectionList.SetItems(sectionItems)

	model := Model{
		EntityListState: NewEntityListState(items, options.Issues),
		InputState:      NewInputState(),
		sectionList:     sectionList,
		ProjectPickerState: NewProjectPickerState(
			buildRecentProjectOptions(options.Recents, options.RecentProjects),
			options.LoadProjects,
			buildProjectList(options.Path, projectListRecents, options.Projects)...,
		),
		focus:                FocusDetail,
		width:                100,
		height:               30,
		projectList:          buildProjectList(options.Path, projectListRecents, options.Projects),
		section:              options.Section,
		entityID:             options.EntityID,
		refresh:              options.Refresh,
		issueState:           "opened",
		loadIssues:           options.LoadIssues,
		loadProject:          options.LoadProject,
		loadDiscussions:      options.LoadDiscussions,
		loadFiles:            options.LoadFiles,
		IssueDetailState:     NewIssueDetailState(),
		DiffViewState:        NewDiffViewState(),
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
		MRDetailState:        NewMRDetailState(),
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
	model.EntityListState.projectPath = options.Path
	model.rebuildProjectRows()

	if model.projectPath == "" {
		if len(model.projectList) > 0 || len(model.ProjectPickerState.recentProjectOptions) > 0 || len(model.ProjectPickerState.loadProjects) > 0 {
			model.mode = ModeProjectSelect
			model.ProjectPickerState.selected = model.ProjectPickerState.nextSelectable(-1, 1)
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

	if m.mode == ModeProjectSelect && len(m.ProjectPickerState.loadProjects) > 0 {
		cmds := make([]tea.Cmd, 0, len(m.ProjectPickerState.loadProjects))
		for _, loader := range m.ProjectPickerState.loadProjects {
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
