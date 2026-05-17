package app

import (
	"testing"
	"time"

	"github.com/aybykovskii/gitlab-tui/internal/config"
)

type fakeRemotes struct {
	urls []string
	err  error
}

func (r fakeRemotes) RemoteURLs() ([]string, error) {
	return r.urls, r.err
}

func TestProjectResolverUsesProjectOverrideFirst(t *testing.T) {
	cfg := config.Default()
	resolution := ProjectResolver{
		Config:          cfg,
		Remotes:         fakeRemotes{urls: []string{"git@gitlab.com:group/project.git"}},
		ProjectOverride: "override/project",
	}.Resolve()

	if resolution.Source != ProjectSourceOverride {
		t.Fatalf("expected override source, got %v", resolution.Source)
	}
	if resolution.Path != "override/project" {
		t.Fatalf("expected override/project, got %q", resolution.Path)
	}
}

func TestProjectResolverUsesGitRemoteFirst(t *testing.T) {
	cfg := config.Default()
	cfg.RecentProjects = []config.RecentProject{{Account: "default", Path: "recent/project", LastUsedAt: time.Now()}}
	resolution := ProjectResolver{
		Config:  cfg,
		Remotes: fakeRemotes{urls: []string{"git@gitlab.com:group/project.git"}},
	}.Resolve()

	if resolution.Source != ProjectSourceGitRemote {
		t.Fatalf("expected git remote source, got %v", resolution.Source)
	}
	if resolution.Path != "group/project" {
		t.Fatalf("expected group/project, got %q", resolution.Path)
	}
}

func TestProjectResolverFallsBackToRecentProjects(t *testing.T) {
	cfg := config.Default()
	now := time.Now()
	cfg.RecentProjects = []config.RecentProject{{Account: "default", Path: "recent/project", LastUsedAt: now}}
	resolution := ProjectResolver{
		Config:  cfg,
		Remotes: fakeRemotes{urls: []string{"git@gitlab.example.com:group/project.git"}},
	}.Resolve()

	if resolution.Source != ProjectSourceRecentProjects {
		t.Fatalf("expected recent projects source, got %v", resolution.Source)
	}
	if len(resolution.Recents) != 1 || resolution.Recents[0].Path != "recent/project" {
		t.Fatalf("unexpected recents: %+v", resolution.Recents)
	}
}

func TestProjectResolverFallsBackToManualInput(t *testing.T) {
	resolution := ProjectResolver{Config: config.Default(), Remotes: fakeRemotes{}}.Resolve()

	if resolution.Source != ProjectSourceManualInput {
		t.Fatalf("expected manual input source, got %v", resolution.Source)
	}
}

func TestRememberResolvedProject(t *testing.T) {
	cfg := config.Default()
	now := time.Now()
	RememberResolvedProject(&cfg, "default", "group/project", now)

	if len(cfg.RecentProjects) != 1 {
		t.Fatalf("expected 1 recent project, got %d", len(cfg.RecentProjects))
	}
	if cfg.RecentProjects[0].Account != "default" || cfg.RecentProjects[0].Path != "group/project" {
		t.Fatalf("unexpected recent project: %+v", cfg.RecentProjects[0])
	}
}
