package gitlab

import (
	"context"

	glab "gitlab.com/gitlab-org/api/client-go"
)

func (c Client) CreateMergeRequestNote(ctx context.Context, projectPath string, iid int, body string) error {
	if c.discussions == nil {
		return ErrClientNotConfigured
	}

	_, _, err := c.discussions.CreateMergeRequestDiscussion(projectPath, int64(iid), &glab.CreateMergeRequestDiscussionOptions{Body: &body}, glab.WithContext(ctx))
	return err
}

func (c Client) AddMergeRequestDiscussionNote(ctx context.Context, projectPath string, iid int, discussionID string, body string) error {
	if c.discussions == nil {
		return ErrClientNotConfigured
	}

	_, _, err := c.discussions.AddMergeRequestDiscussionNote(projectPath, int64(iid), discussionID, &glab.AddMergeRequestDiscussionNoteOptions{Body: &body}, glab.WithContext(ctx))
	return err
}

func (c Client) ResolveMergeRequestDiscussion(ctx context.Context, projectPath string, iid int, discussionID string, resolved bool) error {
	if c.discussions == nil {
		return ErrClientNotConfigured
	}

	_, _, err := c.discussions.UpdateMergeRequestDiscussionNote(projectPath, int64(iid), discussionID, 0, &glab.UpdateMergeRequestDiscussionNoteOptions{Resolved: &resolved}, glab.WithContext(ctx))
	return err
}
