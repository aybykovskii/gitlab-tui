package gitlab

import (
	"context"

	glab "gitlab.com/gitlab-org/api/client-go"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func (c Client) CreateDraftNote(ctx context.Context, projectPath string, iid int, discussionID string, body string, position *mr.DiffPosition) (int, error) {
	if c.draftNotes == nil {
		return 0, ErrClientNotConfigured
	}

	opt := &glab.CreateDraftNoteOptions{Note: &body}
	if discussionID != "" {
		opt.InReplyToDiscussionID = &discussionID
	}
	if position != nil {
		opt.Position = diffPositionOptions(position)
	}

	draft, _, err := c.draftNotes.CreateDraftNote(projectPath, int64(iid), opt, glab.WithContext(ctx))
	if err != nil {
		return 0, err
	}
	if draft == nil {
		return 0, nil
	}
	return int(draft.ID), nil
}

func (c Client) BulkPublishDraftNotes(ctx context.Context, projectPath string, iid int, draftIDs []int) error {
	if c.draftNotes == nil {
		return ErrClientNotConfigured
	}
	if len(draftIDs) == 0 {
		return nil
	}

	_, err := c.draftNotes.PublishAllDraftNotes(projectPath, int64(iid), glab.WithContext(ctx))
	return err
}

func (c Client) DeleteAllDraftNotes(ctx context.Context, projectPath string, iid int) error {
	if c.draftNotes == nil {
		return ErrClientNotConfigured
	}

	drafts, _, err := c.draftNotes.ListDraftNotes(projectPath, int64(iid), &glab.ListDraftNotesOptions{}, glab.WithContext(ctx))
	if err != nil {
		return err
	}
	for _, draft := range drafts {
		if draft == nil {
			continue
		}
		if _, err := c.draftNotes.DeleteDraftNote(projectPath, int64(iid), draft.ID, glab.WithContext(ctx)); err != nil {
			return err
		}
	}
	return nil
}

func diffPositionOptions(position *mr.DiffPosition) *glab.PositionOptions {
	positionType := "text"
	opt := &glab.PositionOptions{PositionType: &positionType}

	if position.BaseSHA != "" {
		opt.BaseSHA = &position.BaseSHA
	}
	if position.HeadSHA != "" {
		opt.HeadSHA = &position.HeadSHA
	}
	if position.StartSHA != "" {
		opt.StartSHA = &position.StartSHA
	}
	if position.NewPath != "" {
		opt.NewPath = &position.NewPath
	}
	if position.OldPath != "" {
		opt.OldPath = &position.OldPath
	}
	if position.NewLine > 0 {
		newLine := int64(position.NewLine)
		opt.NewLine = &newLine
	}
	if position.OldLine > 0 {
		oldLine := int64(position.OldLine)
		opt.OldLine = &oldLine
	}

	return opt
}
