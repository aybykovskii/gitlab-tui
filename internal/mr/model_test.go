package mr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	t.Run("matches title and branches", func(t *testing.T) {
		t.Parallel()

		items := []MergeRequest{
			{Title: "Add config", SourceBranch: "go/config", TargetBranch: "main", Author: "alice"},
			{Title: "Render diff", SourceBranch: "go/diff", TargetBranch: "main", Author: "bob"},
		}

		filtered := Filter(items, "diff")

		require.Len(t, filtered, 1)
		assert.Equal(t, "Render diff", filtered[0].Title)
	})

	t.Run("returns all for empty query", func(t *testing.T) {
		t.Parallel()

		items := []MergeRequest{{Title: "A"}, {Title: "B"}}

		assert.Len(t, Filter(items, "   "), len(items))
	})
}
