package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aybykovskii/gitlab-tui/internal/tui"
)

func TestParseCLI(t *testing.T) {
	t.Run("project override with section", func(t *testing.T) {
		t.Parallel()

		intent, err := ParseCLI([]string{"--project", "group/project", "pipeline"})
		require.NoError(t, err)
		assert.Equal(t, "group/project", intent.ProjectOverride)
		assert.Equal(t, tui.SectionPipelines, intent.Section)
	})

	t.Run("rejects positional project path", func(t *testing.T) {
		t.Parallel()

		_, err := ParseCLI([]string{"group/project"})
		assert.Error(t, err)
	})

	t.Run("section and entity intent", func(t *testing.T) {
		t.Parallel()

		intent, err := ParseCLI([]string{"mr", "123"})
		require.NoError(t, err)
		assert.Equal(t, tui.SectionMergeRequests, intent.Section)
		assert.Equal(t, "123", intent.EntityID)
		assert.Empty(t, intent.ProjectOverride)
	})

	t.Run("errors when --project has no value", func(t *testing.T) {
		t.Parallel()

		_, err := ParseCLI([]string{"--project"})
		require.Error(t, err)
		assert.Equal(t, "--project requires a value", err.Error())
	})
}
