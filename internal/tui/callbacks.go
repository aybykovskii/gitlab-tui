package tui

import (
	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type (
	RefreshFunc              func() ([]mr.MergeRequest, error)
	LoadIssuesFunc           func(state string, search string) ([]issue.Issue, error)
	LoadIssueDiscussionsFunc func(iid int) ([]issue.Discussion, error)
	ProjectLoadFunc          func(path string) (ProjectData, error)
	AccountProjectsLoadFunc  func() ([]string, error)
	LoadDiscussionsFunc      func(iid int) ([]mr.Discussion, error)
	LoadFilesFunc            func(iid int) ([]mr.ChangedFile, error)
	SubmitDraftsFunc         func(iid int, drafts []mr.DraftComment) error
	DiscardDraftsFunc        func(iid int) error
	ReplyToDiscussionFunc    func(iid int, discussionID string, body string) error
	DraftReplyFunc           func(iid int, discussionID string, body string) error
	ResolveDiscussionFunc    func(iid int, discussionID string) error
	UnresolveDiscussionFunc  func(iid int, discussionID string) error
	PostInlineCommentFunc    func(iid int, position mr.DiffPosition, body string) error
	PostMRCommentFunc        func(iid int, body string) error
	PostIssueCommentFunc     func(iid int, body string) error
	IssueStateActionFunc     func(iid int) error
	EditIssueFunc            func(iid int, title, description string) error
	ApproveMRFunc            func(iid int) error
	MergeMRFunc              func(iid int) error
	EditMRFunc               func(iid int, title, description string) error
	OpenURLFunc              func(url string) error
	OpenEditorFunc           func(path string, line int) error
	ToggleDraftMRFunc        func(iid int) error
	UpdateMRLabelsFunc       func(iid int, labels []string) error
)
