package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmojiConfigResolveReturnsDefaultsWhenEnabled(t *testing.T) {
	t.Run("returns default values when enabled", func(t *testing.T) {
		t.Parallel()

		cfg := EmojiConfig{Enabled: true}
		icons := cfg.Resolve()

		assert.Equal(t, defaultEmojiMap(), icons)
	})

	t.Run("returns empty values when disabled", func(t *testing.T) {
		t.Parallel()

		cfg := EmojiConfig{Enabled: false}
		icons := cfg.Resolve()

		assert.Equal(t, EmojiMap{}, icons)
	})

	t.Run("partial override merges defaults", func(t *testing.T) {
		t.Parallel()

		cfg := EmojiConfig{
			Enabled: true,
			Icons:   EmojiMap{Author: "X", Draft: "D"},
		}
		icons := cfg.Resolve()

		assert.Equal(t, "X", icons.Author)
		assert.Equal(t, "D", icons.Draft)
	})
}
