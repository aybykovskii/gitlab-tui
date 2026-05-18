package tui

import (
	"github.com/aybykovskii/gitlab-tui/internal/config"
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
	TabReview
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
