package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type fakeGitLabClient struct {
	account        string
	listLimit      int
	listCalled     bool
	calls          []string
	createdProject string
	createdIID     int
	createdBody    string
	createdPos     *mr.DiffPosition
	createdDraftID int
	publishedIDs   []int
}

func (f *fakeGitLabClient) ListProjects(ctx context.Context, limit int) ([]string, error) {
	f.listCalled = true
	f.listLimit = limit

	return []string{f.account + "/project"}, nil
}

func (f *fakeGitLabClient) OpenMergeRequests(ctx context.Context, projectPath string) ([]mr.MergeRequest, error) {
	return nil, nil
}

func (f *fakeGitLabClient) ListProjectIssues(ctx context.Context, projectPath string, state string, search string) ([]issue.Issue, error) {
	return nil, nil
}

func (f *fakeGitLabClient) ListIssueDiscussions(ctx context.Context, projectPath string, iid int) ([]issue.Discussion, error) {
	return nil, nil
}

func (f *fakeGitLabClient) AddIssueComment(ctx context.Context, projectPath string, iid int, body string) error {
	return nil
}

func (f *fakeGitLabClient) CloseIssue(ctx context.Context, projectPath string, iid int) error {
	return nil
}

func (f *fakeGitLabClient) ReopenIssue(ctx context.Context, projectPath string, iid int) error {
	return nil
}

func (f *fakeGitLabClient) EditIssue(ctx context.Context, projectPath string, iid int, title, description string) error {
	return nil
}

func (f *fakeGitLabClient) UpdateIssueLabels(ctx context.Context, projectPath string, iid int, labels []string) error {
	return nil
}

func (f *fakeGitLabClient) AssignSelfIssue(ctx context.Context, projectPath string, iid int) error {
	return nil
}

func (f *fakeGitLabClient) UnassignSelfIssue(ctx context.Context, projectPath string, iid int) error {
	return nil
}

func (f *fakeGitLabClient) MergeRequestDiscussions(ctx context.Context, projectPath string, iid int) ([]mr.Discussion, error) {
	return nil, nil
}

func (f *fakeGitLabClient) MergeRequestChangedFiles(ctx context.Context, projectPath string, iid int) ([]mr.ChangedFile, error) {
	return nil, nil
}

func (f *fakeGitLabClient) ListProjectLabels(ctx context.Context, projectPath string) ([]mr.Label, error) {
	return nil, nil
}

func (f *fakeGitLabClient) UpdateMRLabels(ctx context.Context, projectPath string, iid int, labels []string) error {
	return nil
}

func (f *fakeGitLabClient) SearchProjects(ctx context.Context, query string, limit int) ([]string, error) {
	return nil, nil
}

func (f *fakeGitLabClient) ToggleDraftMR(ctx context.Context, projectPath string, iid int, title string, draft bool) error {
	return nil
}

func (f *fakeGitLabClient) ApproveMergeRequest(ctx context.Context, projectPath string, iid int) error {
	f.calls = append(f.calls, "approve")
	return nil
}

func (f *fakeGitLabClient) AcceptMergeRequest(ctx context.Context, projectPath string, iid int) error {
	f.calls = append(f.calls, "merge")
	return nil
}

func (f *fakeGitLabClient) UpdateMergeRequest(ctx context.Context, projectPath string, iid int, title, description string) error {
	f.calls = append(f.calls, "edit")
	return nil
}

func (f *fakeGitLabClient) CreateMergeRequestDiscussion(ctx context.Context, projectPath string, iid int, body string, position *mr.DiffPosition) error {
	if position == nil {
		f.calls = append(f.calls, "mr-comment")
	} else {
		f.calls = append(f.calls, "inline-comment")
	}
	f.createdProject = projectPath
	f.createdIID = iid
	f.createdBody = body
	f.createdPos = position
	return nil
}

func (f *fakeGitLabClient) AddMergeRequestDiscussionNote(ctx context.Context, projectPath string, iid int, discussionID string, body string) error {
	f.calls = append(f.calls, "reply")
	return nil
}

func (f *fakeGitLabClient) ResolveMergeRequestDiscussion(ctx context.Context, projectPath string, iid int, discussionID string, resolved bool) error {
	if resolved {
		f.calls = append(f.calls, "resolve")
	} else {
		f.calls = append(f.calls, "unresolve")
	}
	return nil
}

func (f *fakeGitLabClient) CreateDraftNote(ctx context.Context, projectPath string, iid int, discussionID string, body string, position *mr.DiffPosition) (int, error) {
	if position == nil {
		f.calls = append(f.calls, "draft-reply")
	} else {
		f.calls = append(f.calls, "draft-inline")
	}
	if f.createdDraftID == 0 {
		f.createdDraftID = 321
	}
	return f.createdDraftID, nil
}

func (f *fakeGitLabClient) BulkPublishDraftNotes(ctx context.Context, projectPath string, iid int, draftIDs []int) error {
	f.calls = append(f.calls, "submit-drafts")
	f.publishedIDs = append([]int(nil), draftIDs...)
	return nil
}

func (f *fakeGitLabClient) DeleteAllDraftNotes(ctx context.Context, projectPath string, iid int) error {
	f.calls = append(f.calls, "discard-drafts")
	return nil
}

func TestBuildProjectOptionsUsesAccountsAndLimitedRecentProjects(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.Accounts = append(cfg.Accounts, config.Account{ID: "work", Host: "https://gitlab.example.com", TokenEnv: "WORK_TOKEN"})
	cfg.RecentProjectsLimit = 2
	old := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg.RecentProjectHistory = []config.RecentProject{
		{Account: "default", Path: "group/old", LastUsedAt: old},
		{Account: "work", Path: "group/newest", LastUsedAt: old.Add(2 * time.Hour)},
		{Account: "default", Path: "group/middle", LastUsedAt: old.Add(time.Hour)},
	}
	clients := map[string]*fakeGitLabClient{}

	options := buildProjectOptions(&cfg, "", false, ProjectResolution{Account: "default", Path: "remote/project", Source: ProjectSourceGitRemote}, CLIIntent{}, func(account config.Account) (gitLabClient, error) {
		client := &fakeGitLabClient{account: account.ID}
		clients[account.ID] = client

		return client, nil
	})

	require.Len(t, options.LoadProjects, 2)
	require.Len(t, options.Recents, 2)
	assert.Equal(t, "group/newest", options.Recents[0])
	assert.Equal(t, "group/middle", options.Recents[1])
	require.Len(t, options.RecentProjects, 2)
	assert.Equal(t, "work", options.RecentProjects[0].Account)
	assert.Equal(t, "default", options.RecentProjects[1].Account)
	assert.Equal(t, "remote/project", options.Path)

	for _, loader := range options.LoadProjects {
		projects, err := loader.Load()
		require.NoError(t, err, "load projects for %s", loader.ID)
		require.Len(t, projects, 1)
		assert.Equal(t, loader.ID+"/project", projects[0])
		assert.True(t, clients[loader.ID].listCalled)
		assert.Equal(t, 15, clients[loader.ID].listLimit)
	}
}

func TestRun(t *testing.T) {
	t.Run("version", func(t *testing.T) {
		t.Parallel()

		var stdout, stderr bytes.Buffer
		code := New("test-version").Run([]string{"version"}, &stdout, &stderr)

		assert.Equal(t, 0, code)
		assert.Equal(t, "test-version", strings.TrimSpace(stdout.String()))
		assert.Empty(t, stderr.String())
	})

	t.Run("init creates config", func(t *testing.T) {
		t.Parallel()

		var stdout, stderr bytes.Buffer
		configPath := filepath.Join(t.TempDir(), "config.yaml")
		code := NewWithEnv("test-version", []string{"GITLAB_TUI_CONFIG_FILE=" + configPath}).Run([]string{"init"}, &stdout, &stderr)

		assert.Equal(t, 0, code, "stderr: %s", stderr.String())
		assert.Contains(t, stdout.String(), "created config: "+configPath)
		_, err := os.Stat(configPath)
		assert.NoError(t, err)
	})

	t.Run("section alias is not unknown command", func(t *testing.T) {
		t.Parallel()

		var stdout, stderr bytes.Buffer
		code := NewWithEnv("test-version", []string{"GITLAB_TOKEN="}).Run([]string{"mr"}, &stdout, &stderr)

		assert.NotContains(t, stderr.String(), "unknown command")
		assert.NotEqual(t, 2, code, "stderr: %s", stderr.String())
	})

	t.Run("unknown command", func(t *testing.T) {
		t.Parallel()

		var stdout, stderr bytes.Buffer
		code := New("test-version").Run([]string{"wat"}, &stdout, &stderr)

		assert.Equal(t, 2, code)
		assert.Contains(t, stderr.String(), "unknown command: wat")
	})
}

type fakeGitLabClientWithLabels struct {
	fakeGitLabClient
	labels []mr.Label
}

func (f *fakeGitLabClientWithLabels) ListProjectLabels(_ context.Context, _ string) ([]mr.Label, error) {
	return f.labels, nil
}

func (f *fakeGitLabClientWithLabels) UpdateMRLabels(_ context.Context, _ string, _ int, _ []string) error {
	return nil
}

func TestLoadProject(t *testing.T) {
	t.Run("wires MR write callbacks with draft IDs", func(t *testing.T) {
		t.Parallel()

		cfg := config.Default()
		client := &fakeGitLabClient{account: "default", createdDraftID: 777}
		resolution := ProjectResolution{Account: "default", Path: "group/project", Source: ProjectSourceGitRemote}

		options := buildProjectOptions(&cfg, "", false, resolution, CLIIntent{}, func(account config.Account) (gitLabClient, error) {
			return client, nil
		})

		data, err := options.LoadProject("group/project", "")
		require.NoError(t, err)
		require.NotNil(t, data.PostInlineComment)
		require.NotNil(t, data.DraftInlineComment)
		require.NotNil(t, data.SubmitDrafts)

		pos := mr.DiffPosition{NewPath: "main.go", NewLine: 7}
		require.NoError(t, data.PostInlineComment(42, pos, "instant"))
		assert.Equal(t, "group/project", client.createdProject)
		assert.Equal(t, 42, client.createdIID)
		assert.Equal(t, "instant", client.createdBody)
		assert.NotNil(t, client.createdPos)

		id, err := data.DraftInlineComment(42, pos, "draft")
		require.NoError(t, err)
		assert.Equal(t, 777, id)

		require.NoError(t, data.SubmitDrafts(42, []mr.DraftComment{{ID: 777, LocalID: "d1"}}))
		require.Len(t, client.publishedIDs, 1)
		assert.Equal(t, 777, client.publishedIDs[0])
	})

	t.Run("wires all callbacks to selected account", func(t *testing.T) {
		t.Parallel()

		cfg := config.Default()
		cfg.Accounts = append(cfg.Accounts, config.Account{ID: "work", Host: "https://gitlab.example.com", TokenEnv: "WORK_TOKEN"})
		clients := map[string]*fakeGitLabClient{}
		resolution := ProjectResolution{Account: "default", Path: "group/project", Source: ProjectSourceGitRemote}

		options := buildProjectOptions(&cfg, "", false, resolution, CLIIntent{}, func(account config.Account) (gitLabClient, error) {
			client := &fakeGitLabClient{account: account.ID, createdDraftID: 777}
			clients[account.ID] = client
			return client, nil
		})

		data, err := options.LoadProject("group/project", "work")
		require.NoError(t, err)

		callbacks := map[string]bool{
			"ApproveMR":           data.ApproveMR != nil,
			"MergeMR":             data.MergeMR != nil,
			"EditMR":              data.EditMR != nil,
			"PostMRComment":       data.PostMRComment != nil,
			"ReplyToDiscussion":   data.ReplyToDiscussion != nil,
			"DraftReply":          data.DraftReply != nil,
			"PostInlineComment":   data.PostInlineComment != nil,
			"DraftInlineComment":  data.DraftInlineComment != nil,
			"SubmitDrafts":        data.SubmitDrafts != nil,
			"DiscardDrafts":       data.DiscardDrafts != nil,
			"ResolveDiscussion":   data.ResolveDiscussion != nil,
			"UnresolveDiscussion": data.UnresolveDiscussion != nil,
		}
		for name, ok := range callbacks {
			assert.True(t, ok, "expected %s to be wired", name)
		}

		pos := mr.DiffPosition{BaseSHA: "base", HeadSHA: "head", StartSHA: "start", NewPath: "main.go", OldPath: "main.go", NewLine: 7}
		_ = data.ApproveMR(42)
		_ = data.MergeMR(42)
		_ = data.EditMR(42, "title", "description")
		_ = data.PostMRComment(42, "general")
		_ = data.ReplyToDiscussion(42, "disc-1", "reply")
		_, _ = data.DraftReply(42, "disc-1", "draft reply")
		_ = data.PostInlineComment(42, pos, "inline")
		_, _ = data.DraftInlineComment(42, pos, "draft inline")
		_ = data.SubmitDrafts(42, []mr.DraftComment{{ID: 777, LocalID: "d1"}})
		_ = data.DiscardDrafts(42)
		_ = data.ResolveDiscussion(42, "disc-1")
		_ = data.UnresolveDiscussion(42, "disc-1")

		work := clients["work"]
		require.NotNil(t, work)
		for _, want := range []string{"approve", "merge", "edit", "mr-comment", "reply", "draft-reply", "inline-comment", "draft-inline", "submit-drafts", "discard-drafts", "resolve", "unresolve"} {
			assert.Contains(t, work.calls, want)
		}
		if defaultClient := clients["default"]; defaultClient != nil {
			assert.Empty(t, defaultClient.calls, "expected default account not to receive write calls")
		}
	})

	t.Run("includes labels in project data", func(t *testing.T) {
		t.Parallel()

		expectedLabels := []mr.Label{
			{Name: "bug", Color: "#EE0701"},
			{Name: "feature", Color: "#0075CA"},
		}
		cfg := config.Default()
		resolution := ProjectResolution{Account: "default", Path: "group/project", Source: ProjectSourceGitRemote}

		options := buildProjectOptions(&cfg, "", false, resolution, CLIIntent{}, func(account config.Account) (gitLabClient, error) {
			return &fakeGitLabClientWithLabels{
				fakeGitLabClient: fakeGitLabClient{account: account.ID},
				labels:           expectedLabels,
			}, nil
		})

		data, err := options.LoadProject("group/project", "")
		require.NoError(t, err)
		require.Len(t, data.Labels, 2)
		assert.Equal(t, "bug", data.Labels[0].Name)
	})
}
