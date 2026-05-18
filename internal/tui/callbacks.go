package tui

import (
	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
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
