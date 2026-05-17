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
	accountID := r.Config.DefaultAccount
	if r.ProjectOverride != "" {
		return ProjectResolution{Account: accountID, Path: r.ProjectOverride, Source: ProjectSourceOverride}
	}

	account, ok := r.Config.Account(accountID)
	if ok && r.Remotes != nil {
		if urls, err := r.Remotes.RemoteURLs(); err == nil {
			for _, remoteURL := range urls {
				if path, ok := gitremote.ProjectPathFromRemote(remoteURL, account.Host); ok {
					return ProjectResolution{Account: accountID, Path: path, Source: ProjectSourceGitRemote}
				}
			}
		}
	}

	recents := r.Config.RecentProjectsForAccount(accountID)
	if len(recents) > 0 {
		return ProjectResolution{Account: accountID, Source: ProjectSourceRecentProjects, Recents: recents}
	}

	return ProjectResolution{Account: accountID, Source: ProjectSourceManualInput}
}

func RememberResolvedProject(cfg *config.Config, account string, path string, now time.Time) {
	if path == "" {
		return
	}
	cfg.RememberProject(config.RecentProject{Account: account, Path: path, LastUsedAt: now})
}
