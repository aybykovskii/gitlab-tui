package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGolangCILintConfig(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(".golangci.yml")
	require.NoError(t, err)

	var cfg struct {
		Version string `yaml:"version"`
		Linters struct {
			Default string   `yaml:"default"`
			Enable  []string `yaml:"enable"`
		} `yaml:"linters"`
		Formatters struct {
			Enable []string `yaml:"enable"`
		} `yaml:"formatters"`
	}

	require.NoError(t, yaml.Unmarshal(data, &cfg))

	assert.Equal(t, "2", cfg.Version)
	assert.Equal(t, "none", cfg.Linters.Default)
	assert.Subset(t, cfg.Linters.Enable, []string{"bodyclose", "dupl", "errcheck", "gocritic", "gocyclo", "godot", "govet", "ineffassign", "mnd", "prealloc", "revive", "staticcheck", "unparam", "unused", "wsl"})
	assert.Subset(t, cfg.Formatters.Enable, []string{"gofmt", "gofumpt", "goimports"})
}
