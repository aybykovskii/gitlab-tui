package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := New("test-version").Run([]string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := strings.TrimSpace(stdout.String()); got != "test-version" {
		t.Fatalf("expected version output, got %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunInitCreatesConfig(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	configPath := filepath.Join(t.TempDir(), "config.yaml")

	code := NewWithEnv("test-version", []string{"GITLAB_TUI_CONFIG_FILE=" + configPath}).Run([]string{"init"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "created config: "+configPath) {
		t.Fatalf("expected created config output, got %q", stdout.String())
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := New("test-version").Run([]string{"wat"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown command: wat") {
		t.Fatalf("expected unknown command error, got %q", stderr.String())
	}
}
