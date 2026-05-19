package main

import (
	"os"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGolangCILintConfigUsesV2WithExpandedLintersAndFormatters(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(".golangci.yml")
	if err != nil {
		t.Fatalf("read .golangci.yml: %v", err)
	}

	var config struct {
		Version string `yaml:"version"`
		Linters struct {
			Default string   `yaml:"default"`
			Enable  []string `yaml:"enable"`
		} `yaml:"linters"`
		Formatters struct {
			Enable []string `yaml:"enable"`
		} `yaml:"formatters"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("parse .golangci.yml: %v", err)
	}

	if config.Version != "2" {
		t.Fatalf("version = %q, want 2", config.Version)
	}

	if config.Linters.Default != "none" {
		t.Fatalf("linters.default = %q, want none", config.Linters.Default)
	}

	wantLinters := []string{"bodyclose", "dupl", "errcheck", "gocritic", "gocyclo", "godot", "govet", "ineffassign", "mnd", "prealloc", "revive", "staticcheck", "unparam", "unused", "wsl"}
	for _, linter := range wantLinters {
		if !slices.Contains(config.Linters.Enable, linter) {
			t.Fatalf("linters.enable missing %q", linter)
		}
	}

	wantFormatters := []string{"gofmt", "gofumpt", "goimports"}
	for _, formatter := range wantFormatters {
		if !slices.Contains(config.Formatters.Enable, formatter) {
			t.Fatalf("formatters.enable missing %q", formatter)
		}
	}
}
