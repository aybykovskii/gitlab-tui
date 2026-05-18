package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmojiConfigResolveReturnsDefaultsWhenEnabled(t *testing.T) {
	cfg := EmojiConfig{Enabled: true}
	icons := cfg.Resolve()

	if icons.Author != "👤" {
		t.Fatalf("Author: want 👤, got %q", icons.Author)
	}

	if icons.Branch != "🌿" {
		t.Fatalf("Branch: want 🌿, got %q", icons.Branch)
	}

	if icons.Draft != "📝" {
		t.Fatalf("Draft: want 📝, got %q", icons.Draft)
	}

	if icons.Reviewers != "👥" {
		t.Fatalf("Reviewers: want 👥, got %q", icons.Reviewers)
	}
}

func TestEmojiConfigResolveReturnsEmptyWhenDisabled(t *testing.T) {
	cfg := EmojiConfig{Enabled: false}
	icons := cfg.Resolve()

	if icons.Author != "" || icons.Branch != "" || icons.Draft != "" || icons.Pipeline != "" {
		t.Fatalf("expected all empty when disabled, got %+v", icons)
	}
}

func TestEmojiConfigResolvePartialOverrideMergesDefaults(t *testing.T) {
	cfg := EmojiConfig{
		Enabled: true,
		Icons:   EmojiMap{Author: "X", Draft: "D"},
	}
	icons := cfg.Resolve()

	if icons.Author != "X" {
		t.Fatalf("Author: want X, got %q", icons.Author)
	}

	if icons.Draft != "D" {
		t.Fatalf("Draft: want D, got %q", icons.Draft)
	}

	if icons.Branch != "🌿" {
		t.Fatalf("Branch: want default 🌿, got %q", icons.Branch)
	}

	if icons.Approvals != "✅" {
		t.Fatalf("Approvals: want default ✅, got %q", icons.Approvals)
	}
}

func TestLoadParsesEmojiSection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`default_account: default
accounts:
  - id: default
    host: https://gitlab.com
    token_env: GITLAB_TOKEN
emoji:
  enabled: true
  icons:
    author: "A"
    branch: "B"
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if !cfg.Emoji.Enabled {
		t.Fatal("expected emoji enabled")
	}

	icons := cfg.Emoji.Resolve()
	if icons.Author != "A" {
		t.Fatalf("Author: want A, got %q", icons.Author)
	}

	if icons.Branch != "B" {
		t.Fatalf("Branch: want B, got %q", icons.Branch)
	}

	if icons.Draft != "📝" {
		t.Fatalf("Draft: want default 📝, got %q", icons.Draft)
	}
}

func TestLoadEmojiAbsentDefaultsToDisabled(t *testing.T) {
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
		t.Fatalf("load: %v", err)
	}

	icons := cfg.Emoji.Resolve()
	if icons.Author != "" || icons.Branch != "" {
		t.Fatalf("expected empty icons when emoji absent from config, got %+v", icons)
	}
}
