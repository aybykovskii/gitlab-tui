package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
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
