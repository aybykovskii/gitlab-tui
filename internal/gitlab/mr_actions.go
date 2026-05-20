package gitlab

import (
	"context"

	glab "gitlab.com/gitlab-org/api/client-go"
)

func (c Client) ApproveMergeRequest(ctx context.Context, projectPath string, iid int) error {
	if c.approvals == nil {
		return ErrClientNotConfigured
	}

	_, _, err := c.approvals.ApproveMergeRequest(projectPath, int64(iid), &glab.ApproveMergeRequestOptions{}, glab.WithContext(ctx))
	return err
}

func (c Client) AcceptMergeRequest(ctx context.Context, projectPath string, iid int) error {
	if c.mergeRequests == nil {
		return ErrClientNotConfigured
	}

	_, _, err := c.mergeRequests.AcceptMergeRequest(projectPath, int64(iid), &glab.AcceptMergeRequestOptions{}, glab.WithContext(ctx))
	return err
}

func (c Client) UpdateMergeRequest(ctx context.Context, projectPath string, iid int, title, description string) error {
	if c.mrEdit == nil {
		return ErrClientNotConfigured
	}

	_, _, err := c.mrEdit.UpdateMergeRequest(projectPath, int64(iid), &glab.UpdateMergeRequestOptions{
		Title:       &title,
		Description: &description,
	}, glab.WithContext(ctx))
	return err
}
