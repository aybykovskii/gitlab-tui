package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type fakeGitLabClient struct {
	account    string
	listLimit  int
	listCalled bool
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

func (f *fakeGitLabClient) MergeRequestDiff(ctx context.Context, projectPath string, iid int) ([]mr.DiffRow, error) {
	return nil, nil
}

func (f *fakeGitLabClient) MergeRequestDiscussions(ctx context.Context, projectPath string, iid int) ([]mr.Discussion, error) {
	return nil, nil
}

func (f *fakeGitLabClient) MergeRequestChangedFiles(ctx context.Context, projectPath string, iid int) ([]mr.ChangedFile, error) {
	return nil, nil
}

func TestBuildProjectOptionsUsesAccountsAndLimitedRecentProjects(t *testing.T) {
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

	if len(options.LoadProjects) != 2 {
		t.Fatalf("expected 2 account loaders, got %d", len(options.LoadProjects))
	}
	if len(options.Recents) != 2 || options.Recents[0] != "group/newest" || options.Recents[1] != "group/middle" {
		t.Fatalf("expected limited recent paths in newest order, got %+v", options.Recents)
	}
	if len(options.RecentProjects) != 2 || options.RecentProjects[0].Account != "work" || options.RecentProjects[1].Account != "default" {
		t.Fatalf("expected account-tagged recent projects, got %+v", options.RecentProjects)
	}
	if options.Path != "remote/project" {
		t.Fatalf("expected resolved project path, got %q", options.Path)
	}

	for _, loader := range options.LoadProjects {
		projects, err := loader.Load()
		if err != nil {
			t.Fatalf("load projects for %s: %v", loader.ID, err)
		}
		if len(projects) != 1 || projects[0] != loader.ID+"/project" {
			t.Fatalf("unexpected projects for %s: %+v", loader.ID, projects)
		}
		if clients[loader.ID].listLimit != 15 || !clients[loader.ID].listCalled {
			t.Fatalf("expected loader %s to call ListProjects(15), got called=%t limit=%d", loader.ID, clients[loader.ID].listCalled, clients[loader.ID].listLimit)
		}
	}
}

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := New("test-version").Run([]string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := strings.TrimSpace(stdout.String()); got != "test-version" {
		t.Fatalf("expected version output, got %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunInitCreatesConfig(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	configPath := filepath.Join(t.TempDir(), "config.yaml")

	code := NewWithEnv("test-version", []string{"GITLAB_TUI_CONFIG_FILE=" + configPath}).Run([]string{"init"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "created config: "+configPath) {
		t.Fatalf("expected created config output, got %q", stdout.String())
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}

func TestRunSectionAliasIsNotUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := NewWithEnv("test-version", []string{"GITLAB_TOKEN="}).Run([]string{"mr"}, &stdout, &stderr)

	if strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("section alias 'mr' produced unknown command error: %q", stderr.String())
	}
	if code == 2 {
		t.Fatalf("section alias 'mr' produced CLI parse error (exit 2), stderr %q", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := New("test-version").Run([]string{"wat"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown command: wat") {
		t.Fatalf("expected unknown command error, got %q", stderr.String())
	}
}
