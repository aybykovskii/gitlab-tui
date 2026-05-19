//nolint:mnd // Visible row count heuristics are intentional UI page-size constants.
package tui

import (
	"fmt"
	"strings"

	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type EntityListState struct {
	items       []mr.MergeRequest
	issueItems  []issue.Issue
	query       string
	selected    int
	listTop     int
	projectPath string
}

type EntityListViewData struct {
	Section         Section
	IssueStateLabel string
	ProjectLoading  bool
	Loading         bool
	ErrorMessage    string
}

func NewEntityListState(items []mr.MergeRequest, issues []issue.Issue) EntityListState {
	return EntityListState{items: items, issueItems: issues}
}

func (s EntityListState) View(layout LayoutState, data EntityListViewData) string {
	if data.Section == SectionIssues {
		return strings.Join(s.issueLines(layout.Height, data), "\n")
	}

	return strings.Join(s.mrLines(layout.Height, data), "\n")
}

func (s EntityListState) filteredMRs() []mr.MergeRequest {
	return mr.Filter(s.items, s.query)
}

func (s EntityListState) filteredIssues() []issue.Issue {
	query := strings.ToLower(strings.TrimSpace(s.query))
	if query == "" {
		return s.issueItems
	}

	filtered := make([]issue.Issue, 0, len(s.issueItems))

	for _, item := range s.issueItems {
		text := strings.ToLower(item.Title + " " + item.Author)
		if strings.Contains(text, query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func (s EntityListState) mrLines(height int, data EntityListViewData) []string {
	lines := []string{"Project: " + s.projectPath, "Merge Requests", "Filter: " + s.query}
	if data.ProjectLoading {
		lines = append(lines, "Loading project…")
	} else if data.Loading {
		lines = append(lines, "Refreshing…")
	}

	if data.ErrorMessage != "" {
		lines = append(lines, "Error: "+data.ErrorMessage)
	}

	items := s.filteredMRs()
	if len(items) == 0 {
		return append(lines, "No opened MRs")
	}

	visible := max(1, height-5)
	end := min(len(items), s.listTop+visible)

	for i := s.listTop; i < end; i++ {
		prefix := "  "
		if i == s.selected {
			prefix = "> "
		}

		item := items[i]
		lines = append(lines, fmt.Sprintf("%s%s !%d %s", prefix, pipelineIcon(item.Pipeline), item.IID, item.Title))
		lines = append(lines, fmt.Sprintf("  %s %s → %s", item.Author, item.SourceBranch, item.TargetBranch))
	}

	return lines
}

func (s EntityListState) issueLines(height int, data EntityListViewData) []string {
	lines := []string{"Project: " + s.projectPath, "Issues [" + data.IssueStateLabel + "]", "Filter: " + s.query}
	if data.Loading {
		lines = append(lines, "Refreshing…")
	}

	if data.ErrorMessage != "" {
		lines = append(lines, "Error: "+data.ErrorMessage)
	}

	items := s.filteredIssues()
	if len(items) == 0 {
		return append(lines, "No issues")
	}

	visible := max(1, (height-5)/2)
	end := min(len(items), s.listTop+visible)

	for i := s.listTop; i < end; i++ {
		prefix := "  "
		if i == s.selected {
			prefix = "> "
		}

		item := items[i]
		lines = append(lines, fmt.Sprintf("%s#%d %s", prefix, item.IID, item.Title))
		lines = append(lines, "  "+formatIssueMeta(item))
	}

	return lines
}
