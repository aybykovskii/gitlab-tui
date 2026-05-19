package mr

import (
	"strings"

	"github.com/aybykovskii/gitlab-tui/pkg/diff"
)

type MergeRequest struct {
	IID            int
	Title          string
	Author         string
	AuthorUsername string
	SourceBranch   string
	TargetBranch   string
	State          string
	Pipeline       string
	Approvals      string
	Description    string
	WebURL         string
	Labels         []string
	Draft          bool
	Reviewers      []string
	Assignees      []string
}

type Label struct {
	Name  string
	Color string
}

type DiffRow = diff.Row

type Note = diff.Note

type DiffPosition = diff.Position

type Discussion = diff.Discussion

type DraftComment = diff.DraftComment

type ChangedFile struct {
	Path         string
	OldPath      string
	IsNew        bool
	IsDeleted    bool
	IsRenamed    bool
	AddedLines   int
	RemovedLines int
	Diff         []DiffRow
}

func Filter(list []MergeRequest, query string) []MergeRequest {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return list
	}

	filtered := make([]MergeRequest, 0, len(list))

	for _, item := range list {
		text := strings.ToLower(item.Title + " " + item.SourceBranch + " " + item.TargetBranch + " " + item.Author)
		if strings.Contains(text, query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}
