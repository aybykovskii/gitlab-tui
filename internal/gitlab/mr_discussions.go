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
