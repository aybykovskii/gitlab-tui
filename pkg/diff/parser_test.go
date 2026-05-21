package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectDiscussionsAttachesToMatchingRow(t *testing.T) {
	t.Parallel()

	rows := []Row{
		{OldLine: 10, NewLine: 10, OldText: "ctx", NewText: "ctx"},
		{OldLine: 0, NewLine: 11, NewText: "new line"},
	}
	discussions := []Discussion{
		{ID: "d1", Notes: []Note{{Author: "alice", Body: "fix this"}}, Position: &Position{NewPath: "main.go", NewLine: 11}},
		{ID: "d2", Notes: []Note{{Author: "bob", Body: "LGTM"}}, Position: &Position{NewPath: "other.go", NewLine: 11}},
		{ID: "d3", Notes: []Note{{Author: "carol", Body: "general"}}, Position: nil},
	}

	annotated := ProjectDiscussions(rows, discussions, "main.go")

	require.Len(t, annotated, 2)
	assert.Empty(t, annotated[0].Discussions)
	require.Len(t, annotated[1].Discussions, 1)
	assert.Equal(t, "d1", annotated[1].Discussions[0].ID)
}

func TestParseUnifiedDiffRows(t *testing.T) {
	t.Parallel()

	rows := Parse(`@@ -10,3 +10,4 @@
 context
-old
+new
+added`)

	require.Len(t, rows, 4)

	assert.Equal(t, 10, rows[0].OldLine)
	assert.Equal(t, 10, rows[0].NewLine)
	assert.Equal(t, "context", rows[0].OldText)
	assert.Equal(t, "context", rows[0].NewText)

	assert.Equal(t, 11, rows[1].OldLine)
	assert.Equal(t, 0, rows[1].NewLine)
	assert.Equal(t, "old", rows[1].OldText)

	assert.Equal(t, 0, rows[2].OldLine)
	assert.Equal(t, 11, rows[2].NewLine)
	assert.Equal(t, "new", rows[2].NewText)

	assert.Equal(t, 12, rows[3].NewLine)
	assert.Equal(t, "added", rows[3].NewText)
}
