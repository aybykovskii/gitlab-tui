package gitlab

import (
	"context"
	"testing"

	glab "gitlab.com/gitlab-org/api/client-go"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type fakeDraftNotes struct {
	createIID     int64
	createOpt     *glab.CreateDraftNoteOptions
	publishAllIID int64
	published     []int64
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

func (f *fakeDraftNotes) PublishDraftNote(pid any, mergeRequest int64, note int64, options ...glab.RequestOptionFunc) (*glab.Response, error) {
	f.published = append(f.published, note)
	return &glab.Response{}, nil
}

func (f *fakeDraftNotes) ListDraftNotes(pid any, mergeRequest int64, opt *glab.ListDraftNotesOptions, options ...glab.RequestOptionFunc) ([]*glab.DraftNote, *glab.Response, error) {
	return f.listed, &glab.Response{}, nil
}

func (f *fakeDraftNotes) DeleteDraftNote(pid any, mergeRequest int64, note int64, options ...glab.RequestOptionFunc) (*glab.Response, error) {
	f.deleted = append(f.deleted, note)
	return &glab.Response{}, nil
}

func TestCreateDraftNoteCreatesInlineDraft(t *testing.T) {
	t.Parallel()

	fake := &fakeDraftNotes{}
	client := NewClientWithDraftNotes(fake)
	id, err := client.CreateDraftNote(context.Background(), "group/project", 42, "", "Check this", &mr.DiffPosition{NewPath: "main.go", NewLine: 7})
	if err != nil {
		t.Fatalf("CreateDraftNote: %v", err)
	}
	if id != 123 || fake.createIID != 42 || fake.createOpt == nil || *fake.createOpt.Note != "Check this" || fake.createOpt.Position == nil {
		t.Fatalf("unexpected create draft call: id=%d iid=%d opt=%+v", id, fake.createIID, fake.createOpt)
	}
}

func TestCreateDraftNoteCreatesReplyDraft(t *testing.T) {
	t.Parallel()

	fake := &fakeDraftNotes{}
	client := NewClientWithDraftNotes(fake)
	_, err := client.CreateDraftNote(context.Background(), "group/project", 42, "disc-1", "Reply", nil)
	if err != nil {
		t.Fatalf("CreateDraftNote reply: %v", err)
	}
	if fake.createOpt.InReplyToDiscussionID == nil || *fake.createOpt.InReplyToDiscussionID != "disc-1" {
		t.Fatalf("expected reply discussion id, got %+v", fake.createOpt)
	}
}

func TestBulkPublishDraftNotesPublishesProvidedIDs(t *testing.T) {
	t.Parallel()

	fake := &fakeDraftNotes{}
	client := NewClientWithDraftNotes(fake)
	if err := client.BulkPublishDraftNotes(context.Background(), "group/project", 42, []int{1, 2}); err != nil {
		t.Fatalf("BulkPublishDraftNotes: %v", err)
	}
	if len(fake.published) != 2 || fake.published[0] != 1 || fake.published[1] != 2 {
		t.Fatalf("unexpected published ids: %+v", fake.published)
	}
}

func TestDeleteAllDraftNotesDeletesListedDrafts(t *testing.T) {
	t.Parallel()

	fake := &fakeDraftNotes{listed: []*glab.DraftNote{{ID: 10}, {ID: 11}}}
	client := NewClientWithDraftNotes(fake)
	if err := client.DeleteAllDraftNotes(context.Background(), "group/project", 42); err != nil {
		t.Fatalf("DeleteAllDraftNotes: %v", err)
	}
	if len(fake.deleted) != 2 || fake.deleted[0] != 10 || fake.deleted[1] != 11 {
		t.Fatalf("unexpected deleted ids: %+v", fake.deleted)
	}
}
