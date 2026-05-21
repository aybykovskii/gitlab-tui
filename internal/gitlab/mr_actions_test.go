package gitlab

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	glab "gitlab.com/gitlab-org/api/client-go"
)

func TestApproveMergeRequestCallsAPI(t *testing.T) {
	t.Parallel()

	approvals := &fakeApprovals{}
	client := NewClientWithServices(&fakeMergeRequests{pages: [][]*glab.BasicMergeRequest{{}}}, approvals)

	require.NoError(t, client.ApproveMergeRequest(context.Background(), "group/project", 42))
	assert.Equal(t, int64(42), approvals.approveIID)
}

func TestAcceptMergeRequestCallsAPI(t *testing.T) {
	t.Parallel()

	mrs := &fakeMergeRequests{pages: [][]*glab.BasicMergeRequest{{}}}
	client := NewClientWithMergeRequests(mrs)

	require.NoError(t, client.AcceptMergeRequest(context.Background(), "group/project", 42))
	assert.Equal(t, int64(42), mrs.acceptIID)
}

func TestUpdateMergeRequestCallsAPI(t *testing.T) {
	t.Parallel()

	fake := &fakeMergeRequestEdit{}
	client := NewClientWithMergeRequestEdit(fake)

	require.NoError(t, client.UpdateMergeRequest(context.Background(), "group/project", 42, "New title", "New description"))
	assert.Equal(t, int64(42), fake.lastIID)
	require.NotNil(t, fake.lastOpts)
	assert.Equal(t, "New title", *fake.lastOpts.Title)
	assert.Equal(t, "New description", *fake.lastOpts.Description)
}
