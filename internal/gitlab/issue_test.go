package gitlab

import (
	"testing"
	"time"

	glab "gitlab.com/gitlab-org/api/client-go"
)

func TestMapIssueKeepsIssueFields(t *testing.T) {
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

	if item.IID != 78 || item.Title != "Issue domain model" || item.State != "opened" {
		t.Fatalf("unexpected basic fields: %+v", item)
	}
	if item.Author != "Alice Doe" || item.AuthorUsername != "alice" {
		t.Fatalf("unexpected author: %+v", item)
	}
	if len(item.Labels) != 2 || item.Labels[0] != "frontend" || item.Labels[1] != "domain" {
		t.Fatalf("unexpected labels: %+v", item.Labels)
	}
	if len(item.Assignees) != 2 || item.Assignees[0] != "Bob Smith" || item.Assignees[1] != "carol" {
		t.Fatalf("unexpected assignees: %+v", item.Assignees)
	}
	if item.Description != "Create internal issue package" || item.WebURL != "https://gitlab.com/group/project/-/issues/78" {
		t.Fatalf("unexpected text fields: %+v", item)
	}
	if item.CommentCount != 5 || item.Milestone != "v1" || item.DueDate != "2026-05-20" {
		t.Fatalf("unexpected metadata: %+v", item)
	}
	if item.Weight != 3 || !item.Confidential {
		t.Fatalf("unexpected weight/confidential: %+v", item)
	}
}
