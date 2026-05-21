package gitlab

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	glab "gitlab.com/gitlab-org/api/client-go"
)

type fakeMergeRequests struct {
	calls     int
	pages     [][]*glab.BasicMergeRequest
	acceptIID int64
	diffRefs  glab.MergeRequestDiffRefs
}

type fakeIssues struct {
	calls       int
	state       string
	search      string
	limit       int64
	page        int64
	updateIID   int64
	stateEvent  string
	title       string
	description string
	labels      string
	assigneeID  int64
	items       []*glab.Issue
}

type fakeDiscussions struct {
	issueIID          int64
	commentIID        int64
	commentBody       string
	mrCommentIID      int64
	mrCommentBody     string
	mrCommentPosition *glab.PositionOptions
	replyIID          int64
	replyID           string
	replyBody         string
	resolveIID        int64
	resolveID         string
	resolved          bool
	items             []*glab.Discussion
}

type fakeApprovals struct {
	configs    map[int64]*glab.MergeRequestApprovals
	approveIID int64
}

type fakeProjects struct {
	limit      int64
	page       int64
	membership bool
	orderBy    string
	projects   []*glab.Project
	err        error
}

func (f *fakeMergeRequests) ListProjectMergeRequests(pid any, opt *glab.ListProjectMergeRequestsOptions, options ...glab.RequestOptionFunc) ([]*glab.BasicMergeRequest, *glab.Response, error) {
	f.calls++

	page := int(opt.Page)
	if page == 0 {
		page = 1
	}

	response := &glab.Response{}
	if page < len(f.pages) {
		response.NextPage = int64(page + 1)
	}

	return f.pages[page-1], response, nil
}

func (f *fakeMergeRequests) GetMergeRequest(pid any, mergeRequest int64, opt *glab.GetMergeRequestsOptions, options ...glab.RequestOptionFunc) (*glab.MergeRequest, *glab.Response, error) {
	refs := f.diffRefs
	if refs == (glab.MergeRequestDiffRefs{}) {
		refs = glab.MergeRequestDiffRefs{BaseSha: "base", HeadSha: "head", StartSha: "start"}
	}

	return &glab.MergeRequest{DiffRefs: refs}, &glab.Response{}, nil
}

func (f *fakeMergeRequests) ListMergeRequestDiffs(pid any, mergeRequest int64, opt *glab.ListMergeRequestDiffsOptions, options ...glab.RequestOptionFunc) ([]*glab.MergeRequestDiff, *glab.Response, error) {
	return []*glab.MergeRequestDiff{{OldPath: "main.go", NewPath: "main.go", Diff: "@@ -1 +1 @@\n-old\n+new"}}, &glab.Response{}, nil
}

func (f *fakeMergeRequests) AcceptMergeRequest(pid any, mergeRequest int64, opt *glab.AcceptMergeRequestOptions, options ...glab.RequestOptionFunc) (*glab.MergeRequest, *glab.Response, error) {
	f.acceptIID = mergeRequest
	return &glab.MergeRequest{}, &glab.Response{}, nil
}

func (f *fakeIssues) ListProjectIssues(pid any, opt *glab.ListProjectIssuesOptions, options ...glab.RequestOptionFunc) ([]*glab.Issue, *glab.Response, error) {
	f.calls++

	if opt != nil {
		if opt.State != nil {
			f.state = *opt.State
		}

		if opt.Search != nil {
			f.search = *opt.Search
		}

		f.limit = opt.PerPage
		f.page = opt.Page
	}

	return f.items, &glab.Response{}, nil
}

func (f *fakeIssues) UpdateIssue(pid any, issue int64, opt *glab.UpdateIssueOptions, options ...glab.RequestOptionFunc) (*glab.Issue, *glab.Response, error) {
	f.updateIID = issue
	if opt != nil && opt.StateEvent != nil {
		f.stateEvent = *opt.StateEvent
	}

	if opt != nil && opt.Title != nil {
		f.title = *opt.Title
	}

	if opt != nil && opt.Description != nil {
		f.description = *opt.Description
	}

	if opt != nil && opt.Labels != nil {
		f.labels = strings.Join(*opt.Labels, ",")
	}

	if opt != nil && opt.AssigneeID != nil {
		f.assigneeID = *opt.AssigneeID
	}

	return &glab.Issue{}, &glab.Response{}, nil
}

func (f *fakeDiscussions) ListMergeRequestDiscussions(pid any, mergeRequest int64, opt *glab.ListMergeRequestDiscussionsOptions, options ...glab.RequestOptionFunc) ([]*glab.Discussion, *glab.Response, error) {
	return f.items, &glab.Response{}, nil
}

func (f *fakeDiscussions) ListIssueDiscussions(pid any, issue int64, opt *glab.ListIssueDiscussionsOptions, options ...glab.RequestOptionFunc) ([]*glab.Discussion, *glab.Response, error) {
	f.issueIID = issue
	return f.items, &glab.Response{}, nil
}

func (f *fakeDiscussions) CreateIssueDiscussion(pid any, issue int64, opt *glab.CreateIssueDiscussionOptions, options ...glab.RequestOptionFunc) (*glab.Discussion, *glab.Response, error) {
	f.commentIID = issue
	if opt != nil && opt.Body != nil {
		f.commentBody = *opt.Body
	}

	return &glab.Discussion{}, &glab.Response{}, nil
}

func (f *fakeDiscussions) CreateMergeRequestDiscussion(pid any, mergeRequest int64, opt *glab.CreateMergeRequestDiscussionOptions, options ...glab.RequestOptionFunc) (*glab.Discussion, *glab.Response, error) {
	f.mrCommentIID = mergeRequest
	if opt != nil && opt.Body != nil {
		f.mrCommentBody = *opt.Body
	}
	if opt != nil {
		f.mrCommentPosition = opt.Position
	}

	return &glab.Discussion{}, &glab.Response{}, nil
}

func (f *fakeDiscussions) AddMergeRequestDiscussionNote(pid any, mergeRequest int64, discussion string, opt *glab.AddMergeRequestDiscussionNoteOptions, options ...glab.RequestOptionFunc) (*glab.Note, *glab.Response, error) {
	f.replyIID = mergeRequest
	f.replyID = discussion
	if opt != nil && opt.Body != nil {
		f.replyBody = *opt.Body
	}
	return &glab.Note{}, &glab.Response{}, nil
}

func (f *fakeDiscussions) ResolveMergeRequestDiscussion(pid any, mergeRequest int64, discussion string, opt *glab.ResolveMergeRequestDiscussionOptions, options ...glab.RequestOptionFunc) (*glab.Discussion, *glab.Response, error) {
	f.resolveIID = mergeRequest
	f.resolveID = discussion
	if opt != nil && opt.Resolved != nil {
		f.resolved = *opt.Resolved
	}
	return &glab.Discussion{}, &glab.Response{}, nil
}

func (f *fakeApprovals) GetConfiguration(pid any, mergeRequest int64, options ...glab.RequestOptionFunc) (*glab.MergeRequestApprovals, *glab.Response, error) {
	if f.configs == nil {
		return nil, &glab.Response{}, nil
	}

	return f.configs[mergeRequest], &glab.Response{}, nil
}

func (f *fakeApprovals) ApproveMergeRequest(pid any, mergeRequest int64, opt *glab.ApproveMergeRequestOptions, options ...glab.RequestOptionFunc) (*glab.MergeRequestApprovals, *glab.Response, error) {
	f.approveIID = mergeRequest
	return &glab.MergeRequestApprovals{}, &glab.Response{}, nil
}

func (f *fakeProjects) ListProjects(opt *glab.ListProjectsOptions, options ...glab.RequestOptionFunc) ([]*glab.Project, *glab.Response, error) {
	if opt != nil {
		f.limit = opt.PerPage
		f.page = opt.Page

		if opt.Membership != nil {
			f.membership = *opt.Membership
		}

		if opt.OrderBy != nil {
			f.orderBy = *opt.OrderBy
		}
	}

	return f.projects, &glab.Response{}, f.err
}

func TestListProjects(t *testing.T) {
	t.Run("returns project paths", func(t *testing.T) {
		t.Parallel()

		projects := &fakeProjects{projects: []*glab.Project{
			{PathWithNamespace: "group/new"},
			{PathWithNamespace: "team/old"},
		}}
		client := NewClientWithProjects(projects)

		paths, err := client.ListProjects(context.Background(), 5)
		require.NoError(t, err)
		assert.Equal(t, []string{"group/new", "team/old"}, paths)
		assert.Equal(t, int64(5), projects.limit)
		assert.Equal(t, int64(1), projects.page)
		assert.True(t, projects.membership)
		assert.Empty(t, projects.orderBy)
	})

	t.Run("returns empty list", func(t *testing.T) {
		t.Parallel()

		client := NewClientWithProjects(&fakeProjects{})

		paths, err := client.ListProjects(context.Background(), 10)
		require.NoError(t, err)
		assert.Empty(t, paths)
	})

	t.Run("returns API error", func(t *testing.T) {
		t.Parallel()

		apiErr := errors.New("api failed")
		client := NewClientWithProjects(&fakeProjects{err: apiErr})

		_, err := client.ListProjects(context.Background(), 10)
		assert.ErrorIs(t, err, apiErr)
	})
}

func TestOpenMergeRequests(t *testing.T) {
	t.Run("maps all pages", func(t *testing.T) {
		t.Parallel()

		fake := &fakeMergeRequests{pages: [][]*glab.BasicMergeRequest{
			{{IID: 1, Title: "First", State: "opened", SourceBranch: "feature/a", TargetBranch: "main", Author: &glab.BasicUser{Username: "alice"}}},
			{{IID: 2, Title: "Second", State: "opened", SourceBranch: "feature/b", TargetBranch: "main", Author: &glab.BasicUser{Name: "Bob"}}},
		}}
		client := NewClientWithMergeRequests(fake)

		items, err := client.OpenMergeRequests(context.Background(), "group/project")
		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, 1, items[0].IID)
		assert.Equal(t, "alice", items[0].Author)
		assert.Equal(t, 2, items[1].IID)
		assert.Equal(t, "Bob", items[1].Author)
		assert.Equal(t, 2, fake.calls)
	})

	t.Run("adds approval counts", func(t *testing.T) {
		t.Parallel()

		client := NewClientWithServices(&fakeMergeRequests{pages: [][]*glab.BasicMergeRequest{{{IID: 3, Title: "MR"}}}}, &fakeApprovals{
			configs: map[int64]*glab.MergeRequestApprovals{3: {ApprovalsRequired: 2, ApprovalsLeft: 1}},
		})

		items, err := client.OpenMergeRequests(context.Background(), "group/project")
		require.NoError(t, err)
		assert.Equal(t, "1/2", items[0].Approvals)
	})
}

func TestListProjectIssues(t *testing.T) {
	t.Run("passes state and maps items", func(t *testing.T) {
		t.Parallel()

		issues := &fakeIssues{items: []*glab.Issue{{
			IID:            79,
			Title:          "Issues API",
			State:          "opened",
			Author:         &glab.IssueAuthor{Name: "Alice", Username: "alice"},
			UserNotesCount: 2,
		}}}
		client := NewClientWithIssues(issues)

		items, err := client.ListProjectIssues(context.Background(), "group/project", "opened", "api")
		require.NoError(t, err)
		assert.Equal(t, "opened", issues.state)
		assert.Equal(t, "api", issues.search)
		assert.Equal(t, int64(50), issues.limit)
		assert.Equal(t, int64(1), issues.page)
		require.Len(t, items, 1)
		assert.Equal(t, 79, items[0].IID)
		assert.Equal(t, "Issues API", items[0].Title)
		assert.Equal(t, "Alice", items[0].Author)
		assert.Equal(t, 2, items[0].CommentCount)
	})

	t.Run("returns empty list", func(t *testing.T) {
		t.Parallel()

		client := NewClientWithIssues(&fakeIssues{})

		items, err := client.ListProjectIssues(context.Background(), "group/project", "closed", "")
		require.NoError(t, err)
		assert.Empty(t, items)
	})
}

func TestListIssueDiscussionsMapsComments(t *testing.T) {
	t.Parallel()

	discussions := &fakeDiscussions{items: []*glab.Discussion{{
		ID:    "issue-discussion-1",
		Notes: []*glab.Note{{Author: glab.NoteAuthor{Name: "Alice", Username: "alice"}, Body: "Looks good"}},
	}}}
	client := NewClientWithDiscussions(discussions)

	items, err := client.ListIssueDiscussions(context.Background(), "group/project", 79)
	require.NoError(t, err)
	assert.Equal(t, int64(79), discussions.issueIID)
	require.Len(t, items, 1)
	assert.Equal(t, "issue-discussion-1", items[0].ID)
	require.Len(t, items[0].Notes, 1)
	assert.Equal(t, "Alice", items[0].Notes[0].Author)
	assert.Equal(t, "Looks good", items[0].Notes[0].Body)
}

func TestIssueUpdateActions(t *testing.T) {
	t.Run("edit maps title and description", func(t *testing.T) {
		t.Parallel()

		issues := &fakeIssues{}
		client := NewClientWithIssues(issues)

		require.NoError(t, client.EditIssue(context.Background(), "group/project", 84, "New title", "New description"))
		assert.Equal(t, int64(84), issues.updateIID)
		assert.Equal(t, "New title", issues.title)
		assert.Equal(t, "New description", issues.description)
	})

	t.Run("update labels sets labels string", func(t *testing.T) {
		t.Parallel()

		issues := &fakeIssues{}
		client := NewClientWithIssues(issues)

		require.NoError(t, client.UpdateIssueLabels(context.Background(), "group/project", 84, []string{"bug", "tui"}))
		assert.Equal(t, "bug,tui", issues.labels)
	})

	t.Run("assign self sets assignee", func(t *testing.T) {
		t.Parallel()

		issues := &fakeIssues{}
		client := NewClientWithIssues(issues)

		require.NoError(t, client.AssignSelfIssue(context.Background(), "group/project", 84))
		assert.Equal(t, int64(0), issues.assigneeID)
	})

	t.Run("unassign self noops", func(t *testing.T) {
		t.Parallel()

		issues := &fakeIssues{}
		client := NewClientWithIssues(issues)

		require.NoError(t, client.UnassignSelfIssue(context.Background(), "group/project", 84))
	})
}

func TestCloseAndReopenIssue(t *testing.T) {
	t.Run("close sets state event", func(t *testing.T) {
		t.Parallel()

		issues := &fakeIssues{}
		client := NewClientWithIssues(issues)

		require.NoError(t, client.CloseIssue(context.Background(), "group/project", 83))
		assert.Equal(t, int64(83), issues.updateIID)
		assert.Equal(t, "close", issues.stateEvent)
	})

	t.Run("reopen sets state event", func(t *testing.T) {
		t.Parallel()

		issues := &fakeIssues{}
		client := NewClientWithIssues(issues)

		require.NoError(t, client.ReopenIssue(context.Background(), "group/project", 83))
		assert.Equal(t, int64(83), issues.updateIID)
		assert.Equal(t, "reopen", issues.stateEvent)
	})
}

func TestAddIssueCommentCreatesIssueDiscussion(t *testing.T) {
	t.Parallel()

	discussions := &fakeDiscussions{}
	client := NewClientWithDiscussions(discussions)

	require.NoError(t, client.AddIssueComment(context.Background(), "group/project", 82, "General comment"))
	assert.Equal(t, int64(82), discussions.commentIID)
	assert.Equal(t, "General comment", discussions.commentBody)
}

func TestMapDiscussion(t *testing.T) {
	t.Run("maps notes and resolution", func(t *testing.T) {
		t.Parallel()

		item := MapDiscussion(&glab.Discussion{
			ID: "abc123",
			Notes: []*glab.Note{
				{Author: glab.NoteAuthor{Name: "Alice", Username: "alice"}, Body: "Needs a fix", Resolved: false},
				{Author: glab.NoteAuthor{Name: "Bob", Username: "bob"}, Body: "Fixed", Resolved: true, Resolvable: true},
			},
		})

		assert.Equal(t, "abc123", item.ID)
		assert.False(t, item.Resolved)
		require.Len(t, item.Notes, 2)
		assert.Equal(t, "Alice", item.Notes[0].Author)
		assert.Equal(t, "Needs a fix", item.Notes[0].Body)
	})

	t.Run("excludes system notes", func(t *testing.T) {
		t.Parallel()

		item := MapDiscussion(&glab.Discussion{
			ID: "sys1",
			Notes: []*glab.Note{
				{Author: glab.NoteAuthor{Name: "GitLab"}, Body: "changed milestone", System: true},
				{Author: glab.NoteAuthor{Name: "Alice"}, Body: "Real comment"},
			},
		})

		require.Len(t, item.Notes, 1)
		assert.Equal(t, "Alice", item.Notes[0].Author)
	})
}

func TestMapChangedFile(t *testing.T) {
	t.Run("maps path markers and line counts", func(t *testing.T) {
		t.Parallel()

		item := MapChangedFile(&glab.MergeRequestDiff{
			NewPath:     "internal/tui/model.go",
			OldPath:     "internal/tui/model.go",
			NewFile:     false,
			DeletedFile: false,
			RenamedFile: false,
			Diff:        "@@ -10,3 +10,4 @@\n context\n-old\n+new\n+added\n",
		})

		assert.Equal(t, "internal/tui/model.go", item.Path)
		assert.False(t, item.IsNew)
		assert.False(t, item.IsDeleted)
		assert.False(t, item.IsRenamed)
		assert.Equal(t, 2, item.AddedLines)
		assert.Equal(t, 1, item.RemovedLines)
	})

	t.Run("marks new and deleted files", func(t *testing.T) {
		t.Parallel()

		newFile := MapChangedFile(&glab.MergeRequestDiff{NewPath: "new.go", NewFile: true, Diff: "@@ -0,0 +1 @@\n+hello\n"})
		assert.True(t, newFile.IsNew)
		assert.Equal(t, 1, newFile.AddedLines)

		deleted := MapChangedFile(&glab.MergeRequestDiff{OldPath: "old.go", DeletedFile: true, Diff: "@@ -1 +0,0 @@\n-bye\n"})
		assert.True(t, deleted.IsDeleted)
		assert.Equal(t, "old.go", deleted.Path)
	})
}

func TestMergeRequestChangedFilesCarriesDiffRefs(t *testing.T) {
	t.Parallel()

	fake := &fakeMergeRequests{diffRefs: glab.MergeRequestDiffRefs{BaseSha: "base-sha", HeadSha: "head-sha", StartSha: "start-sha"}}
	client := NewClientWithMergeRequests(fake)

	files, err := client.MergeRequestChangedFiles(context.Background(), "group/project", 42)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "base-sha", files[0].BaseSHA)
	assert.Equal(t, "head-sha", files[0].HeadSHA)
	assert.Equal(t, "start-sha", files[0].StartSHA)
}

func TestMapMergeRequest(t *testing.T) {
	t.Run("fills labels and draft", func(t *testing.T) {
		t.Parallel()

		item := MapMergeRequest(&glab.BasicMergeRequest{
			IID:    10,
			Title:  "Draft: My MR",
			Draft:  true,
			Labels: glab.Labels{"backend", "performance"},
			Assignees: []*glab.BasicUser{
				{Name: "Alice", Username: "alice"},
				{Name: "", Username: "bob"},
			},
			Reviewers: []*glab.BasicUser{
				{Name: "Carol", Username: "carol"},
			},
		})

		assert.True(t, item.Draft)
		assert.Equal(t, []string{"backend", "performance"}, item.Labels)
		assert.Equal(t, []string{"Alice", "bob"}, item.Assignees)
		assert.Equal(t, []string{"Carol"}, item.Reviewers)
	})

	t.Run("keeps previous MR info", func(t *testing.T) {
		t.Parallel()

		item := MapMergeRequest(&glab.BasicMergeRequest{
			IID:                 3,
			Title:               "MR",
			DetailedMergeStatus: "success",
			WebURL:              "https://gitlab.com/group/project/-/merge_requests/3",
			Author:              &glab.BasicUser{Name: "Alice Doe", Username: "alice"},
		})

		assert.Equal(t, "success", item.Pipeline)
		assert.Equal(t, "Alice Doe", item.Author)
		assert.Equal(t, "alice", item.AuthorUsername)
		assert.Equal(t, "https://gitlab.com/group/project/-/merge_requests/3", item.WebURL)
	})
}

// --- #67: Labels, Draft, Reviewers, Assignees ---

type fakeLabels struct {
	labels []*glab.Label
	err    error
}

func (f *fakeLabels) ListLabels(pid any, opt *glab.ListLabelsOptions, options ...glab.RequestOptionFunc) ([]*glab.Label, *glab.Response, error) {
	return f.labels, &glab.Response{}, f.err
}

type fakeMergeRequestEdit struct {
	lastIID  int64
	lastOpts *glab.UpdateMergeRequestOptions
	err      error
}

func (f *fakeMergeRequestEdit) UpdateMergeRequest(pid any, mergeRequest int64, opt *glab.UpdateMergeRequestOptions, options ...glab.RequestOptionFunc) (*glab.MergeRequest, *glab.Response, error) {
	f.lastIID = mergeRequest
	f.lastOpts = opt

	return &glab.MergeRequest{}, &glab.Response{}, f.err
}

func TestListProjectLabels(t *testing.T) {
	t.Run("returns mapped labels", func(t *testing.T) {
		t.Parallel()

		fake := &fakeLabels{labels: []*glab.Label{
			{Name: "backend", Color: "#e11d48"},
			{Name: "bug", Color: "#dc2626"},
		}}
		client := NewClientWithLabels(fake)

		labels, err := client.ListProjectLabels(context.Background(), "group/project")
		require.NoError(t, err)
		require.Len(t, labels, 2)
		assert.Equal(t, "backend", labels[0].Name)
		assert.Equal(t, "#e11d48", labels[0].Color)
		assert.Equal(t, "bug", labels[1].Name)
		assert.Equal(t, "#dc2626", labels[1].Color)
	})

	t.Run("returns API error", func(t *testing.T) {
		t.Parallel()

		apiErr := errors.New("labels api failed")
		client := NewClientWithLabels(&fakeLabels{err: apiErr})

		_, err := client.ListProjectLabels(context.Background(), "group/project")
		assert.ErrorIs(t, err, apiErr)
	})
}

func TestUpdateMRLabels(t *testing.T) {
	t.Run("sets labels on MR", func(t *testing.T) {
		t.Parallel()

		fake := &fakeMergeRequestEdit{}
		client := NewClientWithMergeRequestEdit(fake)

		require.NoError(t, client.UpdateMRLabels(context.Background(), "group/project", 42, []string{"backend", "bug"}))
		assert.Equal(t, int64(42), fake.lastIID)
		require.NotNil(t, fake.lastOpts)
		require.NotNil(t, fake.lastOpts.Labels)
		assert.Equal(t, []string{"backend", "bug"}, []string(*fake.lastOpts.Labels))
	})

	t.Run("returns API error", func(t *testing.T) {
		t.Parallel()

		apiErr := errors.New("update failed")
		client := NewClientWithMergeRequestEdit(&fakeMergeRequestEdit{err: apiErr})

		err := client.UpdateMRLabels(context.Background(), "group/project", 42, []string{"backend"})
		assert.ErrorIs(t, err, apiErr)
	})
}

func TestToggleDraftMR(t *testing.T) {
	t.Run("adds draft prefix", func(t *testing.T) {
		t.Parallel()

		fake := &fakeMergeRequestEdit{}
		client := NewClientWithMergeRequestEdit(fake)

		require.NoError(t, client.ToggleDraftMR(context.Background(), "group/project", 42, "My MR", true))
		require.NotNil(t, fake.lastOpts)
		require.NotNil(t, fake.lastOpts.Title)
		assert.Equal(t, "Draft: My MR", *fake.lastOpts.Title)
	})

	t.Run("removes draft prefix", func(t *testing.T) {
		t.Parallel()

		fake := &fakeMergeRequestEdit{}
		client := NewClientWithMergeRequestEdit(fake)

		require.NoError(t, client.ToggleDraftMR(context.Background(), "group/project", 42, "Draft: My MR", false))
		require.NotNil(t, fake.lastOpts)
		require.NotNil(t, fake.lastOpts.Title)
		assert.Equal(t, "My MR", *fake.lastOpts.Title)
	})

	t.Run("does not double prefix already draft", func(t *testing.T) {
		t.Parallel()

		fake := &fakeMergeRequestEdit{}
		client := NewClientWithMergeRequestEdit(fake)

		require.NoError(t, client.ToggleDraftMR(context.Background(), "group/project", 42, "Draft: My MR", true))
		require.NotNil(t, fake.lastOpts)
		require.NotNil(t, fake.lastOpts.Title)
		assert.Equal(t, "Draft: My MR", *fake.lastOpts.Title)
	})

	t.Run("returns API error", func(t *testing.T) {
		t.Parallel()

		apiErr := errors.New("edit failed")
		client := NewClientWithMergeRequestEdit(&fakeMergeRequestEdit{err: apiErr})

		err := client.ToggleDraftMR(context.Background(), "group/project", 42, "My MR", true)
		assert.ErrorIs(t, err, apiErr)
	})
}
