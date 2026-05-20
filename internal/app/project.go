package app

import (
	"time"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	gitremote "github.com/aybykovskii/gitlab-tui/internal/git"
)

type ProjectSource int

const (
	ProjectSourceOverride ProjectSource = iota
	ProjectSourceGitRemote
	ProjectSourceRecentProjects
	ProjectSourceManualInput
)

type ProjectResolution struct {
	Account string
	Path    string
	Source  ProjectSource
	Recents []config.RecentProject
}

type RemoteReader interface {
	RemoteURLs() ([]string, error)
}

type ProjectResolver struct {
	Config          config.Config
	Remotes         RemoteReader
	ProjectOverride string
}

func (r ProjectResolver) Resolve() ProjectResolution {
	defaultAccountID := r.Config.DefaultAccount

	if r.ProjectOverride != "" {
		return ProjectResolution{Account: defaultAccountID, Path: r.ProjectOverride, Source: ProjectSourceOverride}
	}

	if r.Remotes != nil {
		if urls, err := r.Remotes.RemoteURLs(); err == nil {
			for _, account := range r.Config.Accounts {
				for _, remoteURL := range urls {
					if path, ok := gitremote.ProjectPathFromRemote(remoteURL, account.Host); ok {
						return ProjectResolution{Account: account.ID, Path: path, Source: ProjectSourceGitRemote}
					}
				}
			}
		}
	}

	recents := r.Config.RecentProjectsForAccount(defaultAccountID)
	if len(recents) > 0 {
		return ProjectResolution{Account: defaultAccountID, Source: ProjectSourceRecentProjects, Recents: recents}
	}

	return ProjectResolution{Account: defaultAccountID, Source: ProjectSourceManualInput}
}

func RememberResolvedProject(cfg *config.Config, account string, path string, now time.Time) {
	if path == "" {
		return
	}

	cfg.RememberProject(config.RecentProject{Account: account, Path: path, LastUsedAt: now})
}
