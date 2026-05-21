package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aybykovskii/gitlab-tui/internal/config"
)

type fakeRemotes struct {
	urls []string
	err  error
}

func (r fakeRemotes) RemoteURLs() ([]string, error) {
	return r.urls, r.err
}

func TestProjectResolver(t *testing.T) {
	t.Run("uses project override first", func(t *testing.T) {
		t.Parallel()

		resolution := ProjectResolver{
			Config:          config.Default(),
			Remotes:         fakeRemotes{urls: []string{"git@gitlab.com:group/project.git"}},
			ProjectOverride: "override/project",
		}.Resolve()

		assert.Equal(t, ProjectSourceOverride, resolution.Source)
		assert.Equal(t, "override/project", resolution.Path)
	})

	t.Run("uses git remote first", func(t *testing.T) {
		t.Parallel()

		cfg := config.Default()
		cfg.RecentProjectHistory = []config.RecentProject{{Account: "default", Path: "recent/project", LastUsedAt: time.Now()}}
		resolution := ProjectResolver{
			Config:  cfg,
			Remotes: fakeRemotes{urls: []string{"git@gitlab.com:group/project.git"}},
		}.Resolve()

		assert.Equal(t, ProjectSourceGitRemote, resolution.Source)
		assert.Equal(t, "group/project", resolution.Path)
	})

	t.Run("falls back to recent projects", func(t *testing.T) {
		t.Parallel()

		cfg := config.Default()
		cfg.RecentProjectHistory = []config.RecentProject{{Account: "default", Path: "recent/project", LastUsedAt: time.Now()}}
		resolution := ProjectResolver{
			Config:  cfg,
			Remotes: fakeRemotes{urls: []string{"git@gitlab.example.com:group/project.git"}},
		}.Resolve()

		assert.Equal(t, ProjectSourceRecentProjects, resolution.Source)
		require.Len(t, resolution.Recents, 1)
		assert.Equal(t, "recent/project", resolution.Recents[0].Path)
	})

	t.Run("falls back to manual input", func(t *testing.T) {
		t.Parallel()

		resolution := ProjectResolver{Config: config.Default(), Remotes: fakeRemotes{}}.Resolve()
		assert.Equal(t, ProjectSourceManualInput, resolution.Source)
	})
}

func TestRememberResolvedProject(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	now := time.Now()
	RememberResolvedProject(&cfg, "default", "group/project", now)

	require.Len(t, cfg.RecentProjectHistory, 1)
	assert.Equal(t, "default", cfg.RecentProjectHistory[0].Account)
	assert.Equal(t, "group/project", cfg.RecentProjectHistory[0].Path)
}
