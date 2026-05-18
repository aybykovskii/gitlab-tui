package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPathUsesOverride(t *testing.T) {
	path, err := (Paths{Env: []string{FileOverrideEnv + "=/tmp/custom.yaml", "XDG_CONFIG_HOME=/tmp/xdg"}}).Path()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != "/tmp/custom.yaml" {
		t.Fatalf("expected override path, got %q", path)
	}
}

func TestPathUsesXDGConfigHome(t *testing.T) {
	path, err := (Paths{Env: []string{"XDG_CONFIG_HOME=/tmp/xdg"}}).Path()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := filepath.Join("/tmp/xdg", "gitlab-tui", "config.yaml")
	if path != want {
		t.Fatalf("expected %q, got %q", want, path)
	}
}

func TestPathFallsBackToHome(t *testing.T) {
	path, err := (Paths{Env: []string{"HOME=/tmp/home"}}).Path()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := filepath.Join("/tmp/home", ".config", "gitlab-tui", "config.yaml")
	if path != want {
		t.Fatalf("expected %q, got %q", want, path)
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := Default()

	if err := Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.DefaultAccount != "default" {
		t.Fatalf("expected default account, got %q", loaded.DefaultAccount)
	}
	account, ok := loaded.Account("default")
	if !ok {
		t.Fatal("expected default account to exist")
	}
	if account.Host != "https://gitlab.com" || account.TokenEnv != DefaultTokenEnv {
		t.Fatalf("unexpected account: %+v", account)
	}
	if loaded.RecentProjectsLimit != 10 {
		t.Fatalf("expected default recent projects limit 10, got %d", loaded.RecentProjectsLimit)
	}
}

func TestLoadDefaultsRecentProjectsLimitWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`default_account: default
accounts:
  - id: default
    host: https://gitlab.com
    token_env: GITLAB_TOKEN
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.RecentProjectsLimit != 10 {
		t.Fatalf("expected default recent projects limit 10, got %d", cfg.RecentProjectsLimit)
	}
}

func TestSavePersistsRecentProjectsLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := Default()
	cfg.RecentProjectsLimit = 0

	if err := Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.RecentProjectsLimit != 0 {
		t.Fatalf("expected recent projects limit 0, got %d", loaded.RecentProjectsLimit)
	}
}

func TestInitDoesNotOverwriteExistingConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := Save(path, Default()); err != nil {
		t.Fatalf("save config: %v", err)
	}

	err := Init(path, Default())
	if err == nil {
		t.Fatal("expected init to fail for existing config")
	}
	if errors.Is(err, os.ErrExist) {
		t.Fatalf("expected contextual error, got %v", err)
	}
}

func TestValidateRequiresDefaultAccountToExist(t *testing.T) {
	cfg := Default()
	cfg.DefaultAccount = "missing"

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRecentProjectsReturnsMixedAccountsSortedAndLimited(t *testing.T) {
	cfg := Default()
	oldest := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	middle := oldest.Add(time.Hour)
	newest := middle.Add(time.Hour)
	cfg.RecentProjectsLimit = 2
	cfg.RecentProjectHistory = []RecentProject{
		{Account: "default", Path: "group/oldest", LastUsedAt: oldest},
		{Account: "work", Path: "group/newest", LastUsedAt: newest},
		{Account: "default", Path: "group/middle", LastUsedAt: middle},
	}

	projects := cfg.RecentProjects()

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].Path != "group/newest" || projects[1].Path != "group/middle" {
		t.Fatalf("unexpected recent projects: %+v", projects)
	}
}

func TestRecentProjectsReturnsEmptyWhenLimitIsZero(t *testing.T) {
	cfg := Default()
	cfg.RecentProjectsLimit = 0
	cfg.RecentProjectHistory = []RecentProject{{Account: "default", Path: "group/project", LastUsedAt: time.Now()}}

	projects := cfg.RecentProjects()

	if len(projects) != 0 {
		t.Fatalf("expected no recent projects, got %+v", projects)
	}
}

func TestRecentProjectsReturnsAllWhenLimitExceedsList(t *testing.T) {
	cfg := Default()
	cfg.RecentProjectsLimit = 10
	cfg.RecentProjectHistory = []RecentProject{
		{Account: "default", Path: "group/one", LastUsedAt: time.Now()},
		{Account: "work", Path: "group/two", LastUsedAt: time.Now().Add(time.Hour)},
	}

	projects := cfg.RecentProjects()

	if len(projects) != 2 {
		t.Fatalf("expected all projects, got %+v", projects)
	}
}

func TestRecentProjectsForAccountSortsByLastUsed(t *testing.T) {
	cfg := Default()
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)
	cfg.RecentProjectHistory = []RecentProject{
		{Account: "default", Path: "group/old", LastUsedAt: older},
		{Account: "other", Path: "group/other", LastUsedAt: newer},
		{Account: "default", Path: "group/new", LastUsedAt: newer},
	}

	projects := cfg.RecentProjectsForAccount("default")

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].Path != "group/new" || projects[1].Path != "group/old" {
		t.Fatalf("unexpected order: %+v", projects)
	}
}

func TestRememberProjectUpdatesExistingEntry(t *testing.T) {
	cfg := Default()
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)
	cfg.RememberProject(RecentProject{Account: "default", Path: "group/project", LastUsedAt: older})
	cfg.RememberProject(RecentProject{Account: "default", Path: "group/project", LastUsedAt: newer})

	if len(cfg.RecentProjectHistory) != 1 {
		t.Fatalf("expected 1 recent project, got %d", len(cfg.RecentProjectHistory))
	}
	if !cfg.RecentProjectHistory[0].LastUsedAt.Equal(newer) {
		t.Fatalf("expected newer timestamp, got %s", cfg.RecentProjectHistory[0].LastUsedAt)
	}
}

func TestAccountTokenReadsConfiguredEnv(t *testing.T) {
	account := Account{ID: "work", Host: "https://gitlab.example.com", TokenEnv: "GITLAB_WORK_TOKEN"}
	token, err := account.Token([]string{"GITLAB_WORK_TOKEN=secret"})
	if err != nil {
		t.Fatalf("expected token, got %v", err)
	}
	if token != "secret" {
		t.Fatalf("expected token, got %q", token)
	}
}

func TestAccountTokenErrorsWhenEnvMissing(t *testing.T) {
	account := Account{ID: "work", Host: "https://gitlab.example.com", TokenEnv: "GITLAB_WORK_TOKEN"}
	if _, err := account.Token([]string{"OTHER=secret"}); err == nil {
		t.Fatal("expected missing env error")
	}
}
