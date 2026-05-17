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
	"github.com/aybykovskii/gitlab-tui/internal/mr"
	"github.com/aybykovskii/gitlab-tui/internal/tui"
)

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
	if len(args) == 0 {
		return a.runTUI(stdout, stderr)
	}

	switch args[0] {
	case "init":
		return a.runInit(stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, a.version)
		return 0
	case "help", "--help", "-h":
		writeUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		writeUsage(stderr)
		return 2
	}
}

func (a App) runTUI(stdout io.Writer, stderr io.Writer) int {
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
	resolution := ProjectResolver{Config: cfg, Remotes: gitremote.CommandRunner{Dir: cwd}}.Resolve()
	options := tui.ProjectOptions{Path: resolution.Path}
	for _, recent := range resolution.Recents {
		options.Recents = append(options.Recents, recent.Path)
	}

	if resolution.Path != "" {
		account, ok := cfg.Account(resolution.Account)
		if !ok {
			fmt.Fprintf(stderr, "account %q not found\n", resolution.Account)
			return 1
		}
		client, err := gitlabclient.NewClient(account, a.env)
		if err != nil {
			fmt.Fprintf(stderr, "create GitLab client: %v\n", err)
			return 1
		}
		loadMRs := func() ([]mr.MergeRequest, error) {
			return client.OpenMergeRequests(context.Background(), resolution.Path)
		}
		loadDiff := func(iid int) ([]mr.DiffRow, error) {
			return client.MergeRequestDiff(context.Background(), resolution.Path, iid)
		}
		items, err := loadMRs()
		if err != nil {
			fmt.Fprintf(stderr, "load merge requests: %v\n", err)
			return 1
		}
		options.Items = items
		options.Refresh = loadMRs
		options.LoadDiff = loadDiff
		RememberResolvedProject(&cfg, resolution.Account, resolution.Path, time.Now())
		if configLoaded {
			if err := config.Save(configPath, cfg); err != nil {
				fmt.Fprintf(stderr, "save recent project: %v\n", err)
				return 1
			}
		}
	}

	if err := tui.RunWithProject(stdout, options); err != nil {
		fmt.Fprintf(stderr, "run TUI: %v\n", err)
		return 1
	}
	return 0
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
	fmt.Fprintln(w, "Usage: gitlab-tui-go [command]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  init     Create YAML config")
	fmt.Fprintln(w, "  version  Print version")
}
