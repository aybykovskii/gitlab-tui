package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathResolution(t *testing.T) {
	t.Run("uses override env", func(t *testing.T) {
		t.Parallel()

		path, err := (Paths{Env: []string{FileOverrideEnv + "=/tmp/custom.yaml", "XDG_CONFIG_HOME=/tmp/xdg"}}).Path()
		require.NoError(t, err)
		assert.Equal(t, "/tmp/custom.yaml", path)
	})

	t.Run("uses XDG_CONFIG_HOME", func(t *testing.T) {
		t.Parallel()

		path, err := (Paths{Env: []string{"XDG_CONFIG_HOME=/tmp/xdg"}}).Path()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/tmp/xdg", "gitlab-tui", "config.yaml"), path)
	})

	t.Run("falls back to HOME", func(t *testing.T) {
		t.Parallel()

		path, err := (Paths{Env: []string{"HOME=/tmp/home"}}).Path()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/tmp/home", ".config", "gitlab-tui", "config.yaml"), path)
	})
}

func TestSaveAndLoad(t *testing.T) {
	t.Run("roundtrip preserves defaults", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "config.yaml")
		require.NoError(t, Save(path, Default()))

		loaded, err := Load(path)
		require.NoError(t, err)

		assert.Equal(t, "default", loaded.DefaultAccount)
		assert.Equal(t, 10, loaded.RecentProjectsLimit)

		account, ok := loaded.Account("default")
		require.True(t, ok)
		assert.Equal(t, "https://gitlab.com", account.Host)
		assert.Equal(t, DefaultTokenEnv, account.TokenEnv)
	})

	t.Run("defaults recent projects limit when missing from file", func(t *testing.T) {
		t.Parallel()

		cfg, err := Load("testdata/config.test.yaml")
		require.NoError(t, err)
		assert.Equal(t, 10, cfg.RecentProjectsLimit)
	})

	t.Run("persists zero recent projects limit", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "config.yaml")
		cfg := Default()
		cfg.RecentProjectsLimit = 0
		require.NoError(t, Save(path, cfg))

		loaded, err := Load(path)
		require.NoError(t, err)
		assert.Equal(t, 0, loaded.RecentProjectsLimit)
	})
}

func TestInitDoesNotOverwriteExistingConfig(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, Save(path, Default()))

	err := Init(path, Default())
	require.Error(t, err)
	assert.False(t, errors.Is(err, os.ErrExist), "expected contextual error, got %v", err)
}

func TestValidateRequiresDefaultAccountToExist(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.DefaultAccount = "missing"
	assert.Error(t, cfg.Validate())
}

func TestRecentProjects(t *testing.T) {
	t.Run("returns mixed accounts sorted and limited", func(t *testing.T) {
		t.Parallel()

		oldest := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		middle := oldest.Add(time.Hour)
		newest := middle.Add(time.Hour)

		cfg := Default()
		cfg.RecentProjectsLimit = 2
		cfg.RecentProjectHistory = []RecentProject{
			{Account: "default", Path: "group/oldest", LastUsedAt: oldest},
			{Account: "work", Path: "group/newest", LastUsedAt: newest},
			{Account: "default", Path: "group/middle", LastUsedAt: middle},
		}

		projects := cfg.RecentProjects()
		require.Len(t, projects, 2)
		assert.Equal(t, "group/newest", projects[0].Path)
		assert.Equal(t, "group/middle", projects[1].Path)
	})

	t.Run("returns empty when limit is zero", func(t *testing.T) {
		t.Parallel()

		cfg := Default()
		cfg.RecentProjectsLimit = 0
		cfg.RecentProjectHistory = []RecentProject{{Account: "default", Path: "group/project", LastUsedAt: time.Now()}}

		assert.Empty(t, cfg.RecentProjects())
	})

	t.Run("returns all when limit exceeds list", func(t *testing.T) {
		t.Parallel()

		cfg := Default()
		cfg.RecentProjectsLimit = 10
		cfg.RecentProjectHistory = []RecentProject{
			{Account: "default", Path: "group/one", LastUsedAt: time.Now()},
			{Account: "work", Path: "group/two", LastUsedAt: time.Now().Add(time.Hour)},
		}

		assert.Len(t, cfg.RecentProjects(), 2)
	})
}

func TestRecentProjectsForAccountSortsByLastUsed(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)

	cfg := Default()
	cfg.RecentProjectHistory = []RecentProject{
		{Account: "default", Path: "group/old", LastUsedAt: older},
		{Account: "other", Path: "group/other", LastUsedAt: newer},
		{Account: "default", Path: "group/new", LastUsedAt: newer},
	}

	projects := cfg.RecentProjectsForAccount("default")
	require.Len(t, projects, 2)
	assert.Equal(t, "group/new", projects[0].Path)
	assert.Equal(t, "group/old", projects[1].Path)
}

func TestRememberProjectUpdatesExistingEntry(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)

	cfg := Default()
	cfg.RememberProject(RecentProject{Account: "default", Path: "group/project", LastUsedAt: older})
	cfg.RememberProject(RecentProject{Account: "default", Path: "group/project", LastUsedAt: newer})

	require.Len(t, cfg.RecentProjectHistory, 1)
	assert.True(t, cfg.RecentProjectHistory[0].LastUsedAt.Equal(newer))
}

func TestAccountToken(t *testing.T) {
	t.Run("reads configured env", func(t *testing.T) {
		t.Parallel()

		account := Account{ID: "work", Host: "https://gitlab.example.com", TokenEnv: "GITLAB_WORK_TOKEN"}
		token, err := account.Token([]string{"GITLAB_WORK_TOKEN=secret"})
		require.NoError(t, err)
		assert.Equal(t, "secret", token)
	})

	t.Run("errors when env missing", func(t *testing.T) {
		t.Parallel()

		account := Account{ID: "work", Host: "https://gitlab.example.com", TokenEnv: "GITLAB_WORK_TOKEN"}
		_, err := account.Token([]string{"OTHER=secret"})
		assert.Error(t, err)
	})
}
