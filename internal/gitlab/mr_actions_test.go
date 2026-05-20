package gitlab

import (
	"context"
	"testing"

	glab "gitlab.com/gitlab-org/api/client-go"
)

func TestApproveMergeRequestCallsAPI(t *testing.T) {
	t.Parallel()

	approvals := &fakeApprovals{}
	client := NewClientWithServices(&fakeMergeRequests{pages: [][]*glab.BasicMergeRequest{{}}}, approvals)

	if err := client.ApproveMergeRequest(context.Background(), "group/project", 42); err != nil {
		t.Fatalf("ApproveMergeRequest: %v", err)
	}
	if approvals.approveIID != 42 {
		t.Fatalf("expected approve iid 42, got %d", approvals.approveIID)
	}
}

func TestAcceptMergeRequestCallsAPI(t *testing.T) {
	t.Parallel()

	mrs := &fakeMergeRequests{pages: [][]*glab.BasicMergeRequest{{}}}
	client := NewClientWithMergeRequests(mrs)

	if err := client.AcceptMergeRequest(context.Background(), "group/project", 42); err != nil {
		t.Fatalf("AcceptMergeRequest: %v", err)
	}
	if mrs.acceptIID != 42 {
		t.Fatalf("expected accept iid 42, got %d", mrs.acceptIID)
	}
}

func TestUpdateMergeRequestCallsAPI(t *testing.T) {
	t.Parallel()

	fake := &fakeMergeRequestEdit{}
	client := NewClientWithMergeRequestEdit(fake)

	if err := client.UpdateMergeRequest(context.Background(), "group/project", 42, "New title", "New description"); err != nil {
		t.Fatalf("UpdateMergeRequest: %v", err)
	}
	if fake.lastIID != 42 || fake.lastOpts == nil || *fake.lastOpts.Title != "New title" || *fake.lastOpts.Description != "New description" {
		t.Fatalf("unexpected update call: iid=%d opts=%+v", fake.lastIID, fake.lastOpts)
	}
}
