package gitlab

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func TestCreateMergeRequestNoteCreatesGeneralDiscussion(t *testing.T) {
	t.Parallel()

	discussions := &fakeDiscussions{}
	client := NewClientWithDiscussions(discussions)

	require.NoError(t, client.CreateMergeRequestNote(context.Background(), "group/project", 42, "Looks good"))
	assert.Equal(t, int64(42), discussions.mrCommentIID)
	assert.Equal(t, "Looks good", discussions.mrCommentBody)
}

func TestCreateMergeRequestDiscussionCreatesInlineThread(t *testing.T) {
	t.Parallel()

	discussions := &fakeDiscussions{}
	client := NewClientWithDiscussions(discussions)

	require.NoError(t, client.CreateMergeRequestDiscussion(context.Background(), "group/project", 42, "Check", &mr.DiffPosition{NewPath: "main.go", OldPath: "main.go", NewLine: 7}))
	require.NotNil(t, discussions.mrCommentPosition)
	require.NotNil(t, discussions.mrCommentPosition.NewLine)
	assert.Equal(t, int64(7), *discussions.mrCommentPosition.NewLine)
	assert.Nil(t, discussions.mrCommentPosition.OldLine)
}

func TestAddMergeRequestDiscussionNoteRepliesToThread(t *testing.T) {
	t.Parallel()

	discussions := &fakeDiscussions{}
	client := NewClientWithDiscussions(discussions)

	require.NoError(t, client.AddMergeRequestDiscussionNote(context.Background(), "group/project", 42, "abc", "Fixed"))
	assert.Equal(t, int64(42), discussions.replyIID)
	assert.Equal(t, "abc", discussions.replyID)
	assert.Equal(t, "Fixed", discussions.replyBody)
}

func TestResolveMergeRequestDiscussionSetsResolved(t *testing.T) {
	t.Parallel()

	discussions := &fakeDiscussions{}
	client := NewClientWithDiscussions(discussions)

	require.NoError(t, client.ResolveMergeRequestDiscussion(context.Background(), "group/project", 42, "abc", true))
	assert.Equal(t, int64(42), discussions.resolveIID)
	assert.Equal(t, "abc", discussions.resolveID)
	assert.True(t, discussions.resolved)
}
