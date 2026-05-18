package issue

import "github.com/aybykovskii/gitlab-tui/internal/mr"

type Issue struct {
	IID            int
	Title          string
	Author         string
	AuthorUsername string
	State          string
	Labels         []string
	Assignees      []string
	Description    string
	WebURL         string
	CommentCount   int
	Milestone      string
	DueDate        string
	Weight         int
	Confidential   bool
}

type Discussion = mr.Discussion
