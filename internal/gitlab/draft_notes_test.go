package gitlab

import (
	"context"
	"testing"

	glab "gitlab.com/gitlab-org/api/client-go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type fakeDraftNotes struct {
	createIID     int64
	createOpt     *glab.CreateDraftNoteOptions
	publishAllIID int64
	listed        []*glab.DraftNote
	deleted       []int64
}

func (f *fakeDraftNotes) CreateDraftNote(pid any, mergeRequest int64, opt *glab.CreateDraftNoteOptions, options ...glab.RequestOptionFunc) (*glab.DraftNote, *glab.Response, error) {
	f.createIID = mergeRequest
	f.createOpt = opt
	return &glab.DraftNote{ID: 123}, &glab.Response{}, nil
}

func (f *fakeDraftNotes) PublishAllDraftNotes(pid any, mergeRequest int64, options ...glab.RequestOptionFunc) (*glab.Response, error) {
	f.publishAllIID = mergeRequest
	return &glab.Response{}, nil
}

func (f *fakeDraftNotes) ListDraftNotes(pid any, mergeRequest int64, opt *glab.ListDraftNotesOptions, options ...glab.RequestOptionFunc) ([]*glab.DraftNote, *glab.Response, error) {
	return f.listed, &glab.Response{}, nil
}

func (f *fakeDraftNotes) DeleteDraftNote(pid any, mergeRequest int64, note int64, options ...glab.RequestOptionFunc) (*glab.Response, error) {
	f.deleted = append(f.deleted, note)
	return &glab.Response{}, nil
}

func TestCreateDraftNote(t *testing.T) {
	t.Run("creates inline draft", func(t *testing.T) {
		t.Parallel()

		fake := &fakeDraftNotes{}
		client := NewClientWithDraftNotes(fake)
		id, err := client.CreateDraftNote(context.Background(), "group/project", 42, "", "Check this", &mr.DiffPosition{NewPath: "main.go", NewLine: 7})
		require.NoError(t, err)
		assert.Equal(t, 123, id)
		assert.Equal(t, int64(42), fake.createIID)
		require.NotNil(t, fake.createOpt)
		assert.Equal(t, "Check this", *fake.createOpt.Note)
		assert.NotNil(t, fake.createOpt.Position)
		assert.Nil(t, fake.createOpt.Position.OldLine)
	})

	t.Run("creates reply draft", func(t *testing.T) {
		t.Parallel()

		fake := &fakeDraftNotes{}
		client := NewClientWithDraftNotes(fake)
		_, err := client.CreateDraftNote(context.Background(), "group/project", 42, "disc-1", "Reply", nil)
		require.NoError(t, err)
		require.NotNil(t, fake.createOpt)
		require.NotNil(t, fake.createOpt.InReplyToDiscussionID)
		assert.Equal(t, "disc-1", *fake.createOpt.InReplyToDiscussionID)
	})
}

func TestBulkPublishDraftNotes(t *testing.T) {
	t.Run("noops without IDs", func(t *testing.T) {
		t.Parallel()

		fake := &fakeDraftNotes{}
		client := NewClientWithDraftNotes(fake)
		require.NoError(t, client.BulkPublishDraftNotes(context.Background(), "group/project", 42, nil))
		assert.Equal(t, int64(0), fake.publishAllIID)
	})

	t.Run("publishes all drafts once", func(t *testing.T) {
		t.Parallel()

		fake := &fakeDraftNotes{}
		client := NewClientWithDraftNotes(fake)
		require.NoError(t, client.BulkPublishDraftNotes(context.Background(), "group/project", 42, []int{1, 2}))
		assert.Equal(t, int64(42), fake.publishAllIID)
	})
}

func TestDeleteAllDraftNotesDeletesListedDrafts(t *testing.T) {
	t.Parallel()

	fake := &fakeDraftNotes{listed: []*glab.DraftNote{{ID: 10}, {ID: 11}}}
	client := NewClientWithDraftNotes(fake)
	require.NoError(t, client.DeleteAllDraftNotes(context.Background(), "group/project", 42))
	assert.Equal(t, []int64{10, 11}, fake.deleted)
}
