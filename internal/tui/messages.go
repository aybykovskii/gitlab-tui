package tui

import (
	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

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
