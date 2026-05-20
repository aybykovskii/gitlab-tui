package gitlab

import (
	"context"

	glab "gitlab.com/gitlab-org/api/client-go"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func (c Client) CreateMergeRequestNote(ctx context.Context, projectPath string, iid int, body string) error {
	return c.CreateMergeRequestDiscussion(ctx, projectPath, iid, body, nil)
}

func (c Client) CreateMergeRequestDiscussion(ctx context.Context, projectPath string, iid int, body string, position *mr.DiffPosition) error {
	if c.discussions == nil {
		return ErrClientNotConfigured
	}

	opt := &glab.CreateMergeRequestDiscussionOptions{Body: &body}
	if position != nil {
		opt.Position = diffPositionOptions(position)
	}

	_, _, err := c.discussions.CreateMergeRequestDiscussion(projectPath, int64(iid), opt, glab.WithContext(ctx))
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

	_, _, err := c.discussions.ResolveMergeRequestDiscussion(projectPath, int64(iid), discussionID, &glab.ResolveMergeRequestDiscussionOptions{Resolved: &resolved}, glab.WithContext(ctx))
	return err
}
