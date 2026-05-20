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

func TestAddMergeRequestDiscussionNoteRepliesToThread(t *testing.T) {
	t.Parallel()

	discussions := &fakeDiscussions{}
	client := NewClientWithDiscussions(discussions)

	if err := client.AddMergeRequestDiscussionNote(context.Background(), "group/project", 42, "abc", "Fixed"); err != nil {
		t.Fatalf("AddMergeRequestDiscussionNote: %v", err)
	}
	if discussions.replyIID != 42 || discussions.replyID != "abc" || discussions.replyBody != "Fixed" {
		t.Fatalf("unexpected reply call: iid=%d id=%q body=%q", discussions.replyIID, discussions.replyID, discussions.replyBody)
	}
}

func TestResolveMergeRequestDiscussionSetsResolved(t *testing.T) {
	t.Parallel()

	discussions := &fakeDiscussions{}
	client := NewClientWithDiscussions(discussions)

	if err := client.ResolveMergeRequestDiscussion(context.Background(), "group/project", 42, "abc", true); err != nil {
		t.Fatalf("ResolveMergeRequestDiscussion: %v", err)
	}
	if discussions.resolveIID != 42 || discussions.resolveID != "abc" || !discussions.resolved {
		t.Fatalf("unexpected resolve call: iid=%d id=%q resolved=%t", discussions.resolveIID, discussions.resolveID, discussions.resolved)
	}
}
