package gitlab

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	glab "gitlab.com/gitlab-org/api/client-go"
)

func TestMapIssueKeepsIssueFields(t *testing.T) {
	t.Parallel()

	dueDate := glab.ISOTime(time.Date(2026, time.May, 20, 0, 0, 0, 0, time.UTC))

	item := MapIssue(&glab.Issue{
		IID:            78,
		Title:          "Issue domain model",
		Author:         &glab.IssueAuthor{Name: "Alice Doe", Username: "alice"},
		State:          "opened",
		Labels:         glab.Labels{"frontend", "domain"},
		Assignees:      []*glab.IssueAssignee{{Name: "Bob Smith", Username: "bob"}, {Username: "carol"}},
		Description:    "Create internal issue package",
		WebURL:         "https://gitlab.com/group/project/-/issues/78",
		UserNotesCount: 5,
		Milestone:      &glab.Milestone{Title: "v1"},
		DueDate:        &dueDate,
		Weight:         3,
		Confidential:   true,
	})

	assert.Equal(t, 78, item.IID)
	assert.Equal(t, "Issue domain model", item.Title)
	assert.Equal(t, "opened", item.State)
	assert.Equal(t, "Alice Doe", item.Author)
	assert.Equal(t, "alice", item.AuthorUsername)
	assert.Equal(t, []string{"frontend", "domain"}, item.Labels)
	assert.Equal(t, []string{"Bob Smith", "carol"}, item.Assignees)
	assert.Equal(t, "Create internal issue package", item.Description)
	assert.Equal(t, "https://gitlab.com/group/project/-/issues/78", item.WebURL)
	assert.Equal(t, 5, item.CommentCount)
	assert.Equal(t, "v1", item.Milestone)
	assert.Equal(t, "2026-05-20", item.DueDate)
	assert.Equal(t, 3, item.Weight)
	assert.True(t, item.Confidential)
}
