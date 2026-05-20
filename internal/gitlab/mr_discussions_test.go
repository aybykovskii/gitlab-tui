package gitlab

import (
	"context"
	"testing"
)

func TestCreateMergeRequestNoteCreatesGeneralDiscussion(t *testing.T) {
	t.Parallel()

	discussions := &fakeDiscussions{}
	client := NewClientWithDiscussions(discussions)

	if err := client.CreateMergeRequestNote(context.Background(), "group/project", 42, "Looks good"); err != nil {
		t.Fatalf("CreateMergeRequestNote: %v", err)
	}
	if discussions.mrCommentIID != 42 || discussions.mrCommentBody != "Looks good" {
		t.Fatalf("unexpected MR discussion call: iid=%d body=%q", discussions.mrCommentIID, discussions.mrCommentBody)
	}
}
