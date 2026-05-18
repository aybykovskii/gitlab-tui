package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	gitremote "github.com/aybykovskii/gitlab-tui/internal/git"
	gitlabclient "github.com/aybykovskii/gitlab-tui/internal/gitlab"
	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
	"github.com/aybykovskii/gitlab-tui/internal/tui"
)

type gitLabClient interface {
	ListProjects(ctx context.Context, limit int) ([]string, error)
	OpenMergeRequests(ctx context.Context, projectPath string) ([]mr.MergeRequest, error)
	ListProjectIssues(ctx context.Context, projectPath string, state string, search string) ([]issue.Issue, error)
	ListIssueDiscussions(ctx context.Context, projectPath string, iid int) ([]issue.Discussion, error)
	AddIssueComment(ctx context.Context, projectPath string, iid int, body string) error
	CloseIssue(ctx context.Context, projectPath string, iid int) error
	ReopenIssue(ctx context.Context, projectPath string, iid int) error
	EditIssue(ctx context.Context, projectPath string, iid int, title, description string) error
	UpdateIssueLabels(ctx context.Context, projectPath string, iid int, labels []string) error
	AssignSelfIssue(ctx context.Context, projectPath string, iid int) error
	UnassignSelfIssue(ctx context.Context, projectPath string, iid int) error
	MergeRequestDiff(ctx context.Context, projectPath string, iid int) ([]mr.DiffRow, error)
	MergeRequestDiscussions(ctx context.Context, projectPath string, iid int) ([]mr.Discussion, error)
	MergeRequestChangedFiles(ctx context.Context, projectPath string, iid int) ([]mr.ChangedFile, error)
	ListProjectLabels(ctx context.Context, projectPath string) ([]mr.Label, error)
	UpdateMRLabels(ctx context.Context, projectPath string, iid int, labels []string) error
}

type gitLabClientFactory func(config.Account) (gitLabClient, error)

type App struct {
	version string
	env     []string
}

func New(version string) App {
	return App{version: version}
}

func NewWithEnv(version string, env []string) App {
	return App{version: version, env: env}
}

func (a App) Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 0 {
		switch args[0] {
		case "init":
			return a.runInit(stdout, stderr)
		case "version", "--version", "-v":
			fmt.Fprintln(stdout, a.version)
			return 0
		case "help", "--help", "-h":
			writeUsage(stdout)
			return 0
		}
	}

	intent, err := ParseCLI(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		writeUsage(stderr)
		return 2
	}
	return a.runTUI(stdout, stderr, intent)
}

func (a App) runTUI(stdout io.Writer, stderr io.Writer, intent CLIIntent) int {
	cfg := config.Default()
	configPath, err := (config.Paths{Env: a.env}).Path()
	configLoaded := false
	if err == nil {
		loaded, loadErr := config.Load(configPath)
		if loadErr == nil {
			cfg = loaded
			configLoaded = true
		} else if !errors.Is(loadErr, os.ErrNotExist) {
			fmt.Fprintf(stderr, "load config: %v\n", loadErr)
			return 1
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "resolve cwd: %v\n", err)
		return 1
	}
	resolution := ProjectResolver{Config: cfg, Remotes: gitremote.CommandRunner{Dir: cwd}, ProjectOverride: intent.ProjectOverride}.Resolve()
	options := buildProjectOptions(&cfg, configPath, configLoaded, resolution, intent, func(account config.Account) (gitLabClient, error) {
		client, err := gitlabclient.NewClient(account, a.env)
		if err != nil {
			return nil, err
		}
		return client, nil
	})

	if err := tui.RunWithProject(stdout, options); err != nil {
		fmt.Fprintf(stderr, "run TUI: %v\n", err)
		return 1
	}
	return 0
}

func buildProjectOptions(cfg *config.Config, configPath string, configLoaded bool, resolution ProjectResolution, intent CLIIntent, newClient gitLabClientFactory) tui.ProjectOptions {
	loadProject := func(projectPath string) (tui.ProjectData, error) {
		account, ok := cfg.Account(resolution.Account)
		if !ok {
			return tui.ProjectData{}, fmt.Errorf("account %q not found", resolution.Account)
		}
		client, err := newClient(account)
		if err != nil {
			return tui.ProjectData{}, fmt.Errorf("create GitLab client: %w", err)
		}
		loadMRs := func() ([]mr.MergeRequest, error) {
			return client.OpenMergeRequests(context.Background(), projectPath)
		}
		loadIssues := func(state string, search string) ([]issue.Issue, error) {
			return client.ListProjectIssues(context.Background(), projectPath, state, search)
		}
		postIssueComment := func(iid int, body string) error {
			return client.AddIssueComment(context.Background(), projectPath, iid, body)
		}
		loadIssueDiscussions := func(iid int) ([]issue.Discussion, error) {
			return client.ListIssueDiscussions(context.Background(), projectPath, iid)
		}
		closeIssue := func(iid int) error {
			return client.CloseIssue(context.Background(), projectPath, iid)
		}
		reopenIssue := func(iid int) error {
			return client.ReopenIssue(context.Background(), projectPath, iid)
		}
		editIssue := func(iid int, title, description string) error {
			return client.EditIssue(context.Background(), projectPath, iid, title, description)
		}
		assignSelfIssue := func(iid int) error {
			return client.AssignSelfIssue(context.Background(), projectPath, iid)
		}
		unassignSelfIssue := func(iid int) error {
			return client.UnassignSelfIssue(context.Background(), projectPath, iid)
		}
		loadDiff := func(iid int) ([]mr.DiffRow, error) {
			return client.MergeRequestDiff(context.Background(), projectPath, iid)
		}
		loadDiscussions := func(iid int) ([]mr.Discussion, error) {
			return client.MergeRequestDiscussions(context.Background(), projectPath, iid)
		}
		loadFiles := func(iid int) ([]mr.ChangedFile, error) {
			return client.MergeRequestChangedFiles(context.Background(), projectPath, iid)
		}
		type mrResult struct {
			items []mr.MergeRequest
			err   error
		}
		type labelsResult struct {
			labels []mr.Label
			err    error
		}

		var issues []issue.Issue
		if intent.Section == tui.SectionIssues {
			var issueErr error
			issues, issueErr = loadIssues("opened", "")
			if issueErr != nil {
				return tui.ProjectData{}, fmt.Errorf("load issues: %w", issueErr)
			}
		}

		mrCh := make(chan mrResult, 1)
		labelsCh := make(chan labelsResult, 1)
		go func() {
			items, err := loadMRs()
			mrCh <- mrResult{items, err}
		}()
		go func() {
			labels, err := client.ListProjectLabels(context.Background(), projectPath)
			labelsCh <- labelsResult{labels, err}
		}()

		mrRes := <-mrCh
		if mrRes.err != nil {
			return tui.ProjectData{}, fmt.Errorf("load merge requests: %w", mrRes.err)
		}
		labelsRes := <-labelsCh
		labels := labelsRes.labels

		RememberResolvedProject(cfg, resolution.Account, projectPath, time.Now())
		if configLoaded {
			if err := config.Save(configPath, *cfg); err != nil {
				return tui.ProjectData{}, fmt.Errorf("save recent project: %w", err)
			}
		}

		updateLabels := func(iid int, lbls []string) error {
			return client.UpdateMRLabels(context.Background(), projectPath, iid, lbls)
		}

		return tui.ProjectData{
			Items: mrRes.items, Issues: issues, Labels: labels,
			Refresh: loadMRs, LoadIssues: loadIssues,
			PostIssueComment: postIssueComment, LoadIssueDiscussions: loadIssueDiscussions,
			LoadDiff: loadDiff, LoadDiscussions: loadDiscussions, LoadFiles: loadFiles,
			CloseIssue: closeIssue, ReopenIssue: reopenIssue,
			EditIssue: editIssue, AssignSelfIssue: assignSelfIssue, UnassignSelfIssue: unassignSelfIssue,
			UpdateMRLabels: updateLabels,
		}, nil
	}

	options := tui.ProjectOptions{Path: resolution.Path, Section: intent.Section, EntityID: intent.EntityID, LoadProject: loadProject}
	for _, recent := range cfg.RecentProjects() {
		options.Recents = append(options.Recents, recent.Path)
		options.RecentProjects = append(options.RecentProjects, tui.RecentProjectOption{Path: recent.Path, Account: recent.Account})
	}
	for _, account := range cfg.Accounts {
		account := account
		options.LoadProjects = append(options.LoadProjects, tui.AccountProjectLoader{
			ID:   account.ID,
			Host: account.Host,
			Load: func() ([]string, error) {
				client, err := newClient(account)
				if err != nil {
					return nil, fmt.Errorf("create GitLab client: %w", err)
				}
				return client.ListProjects(context.Background(), 15)
			},
		})
	}
	return options
}

func (a App) runInit(stdout io.Writer, stderr io.Writer) int {
	path, err := (config.Paths{Env: a.env}).Path()
	if err != nil {
		fmt.Fprintf(stderr, "resolve config path: %v\n", err)
		return 1
	}

	if err := config.Init(path, config.Default()); err != nil {
		fmt.Fprintf(stderr, "init config: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "created config: %s\n", path)
	return 0
}

func writeUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: gitlab-tui-go [--project <path>] [mr|issue|pipeline] [entity-id]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  init     Create YAML config")
	fmt.Fprintln(w, "  version  Print version")
}
