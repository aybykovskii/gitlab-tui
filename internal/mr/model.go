package mr

import "strings"

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
	Diff           []DiffRow
	Labels         []string
	Draft          bool
	Reviewers      []string
	Assignees      []string
}

type Label struct {
	Name  string
	Color string
}

type DiffRow struct {
	OldLine int
	NewLine int
	OldText string
	NewText string
}

type Note struct {
	Author   string
	Body     string
	Resolved bool
}

type DiffPosition struct {
	NewPath string
	NewLine int
	OldPath string
	OldLine int
}

type Discussion struct {
	ID       string
	Resolved bool
	Notes    []Note
	Position *DiffPosition
}

type DraftComment struct {
	LocalID  string
	Body     string
	Position *DiffPosition
	EndLine  int
}

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
