//nolint:mnd // Header line counts are intentional UI page-size constants.
package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"

	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type EntityListState struct {
	mrList      list.Model
	issueList   list.Model
	items       []mr.MergeRequest
	issueItems  []issue.Issue
	query       string
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
	s := EntityListState{
		mrList:     newFancyList("Merge Requests"),
		issueList:  newFancyList("Issues"),
		items:      items,
		issueItems: issues,
	}
	s.syncMRList()
	s.syncIssueList()

	return s
}

func (s *EntityListState) syncMRList() {
	filtered := s.filteredMRs()
	items := make([]list.Item, len(filtered))

	for i, m := range filtered {
		items[i] = mrListItem{m}
	}

	_ = s.mrList.SetItems(items)
}

func (s *EntityListState) syncIssueList() {
	filtered := s.filteredIssues()
	items := make([]list.Item, len(filtered))

	for i, issue := range filtered {
		items[i] = issueListItem{issue}
	}

	_ = s.issueList.SetItems(items)
}

func (s EntityListState) View(layout LayoutState, data EntityListViewData) string {
	if data.Section == SectionIssues {
		return s.issueView(layout, data)
	}

	return s.mrView(layout, data)
}

func (s EntityListState) mrView(layout LayoutState, data EntityListViewData) string {
	// Sync ensures the list reflects the current items+query, even when state
	// is mutated directly outside the Update path (e.g. in unit tests).
	s.syncMRList()

	header := s.mrHeader(data)

	if len(s.mrList.Items()) == 0 {
		if msg := s.mrNoItemsMsg(data); msg != "" {
			return strings.Join(append(header, "Merge Requests", msg), "\n")
		}

		return strings.Join(header, "\n")
	}

	listH := max(1, layout.Height-len(header))
	s.mrList.SetSize(layout.Width, listH)

	return strings.Join(header, "\n") + "\n" + s.mrList.View()
}

func (s EntityListState) issueView(layout LayoutState, data EntityListViewData) string {
	// Sync ensures the list reflects the current items+query, even when state
	// is mutated directly outside the Update path (e.g. in unit tests).
	s.syncIssueList()

	title := "Issues [" + data.IssueStateLabel + "]"
	s.issueList.Title = title

	header := s.issueHeader(data)

	if len(s.issueList.Items()) == 0 {
		return strings.Join(append(header, title, s.issueNoItemsMsg()), "\n")
	}

	listH := max(1, layout.Height-len(header))
	s.issueList.SetSize(layout.Width, listH)

	return strings.Join(header, "\n") + "\n" + s.issueList.View()
}

func (s EntityListState) mrHeader(data EntityListViewData) []string {
	lines := []string{"Project: " + s.projectPath, "Filter: " + s.query}

	if data.ProjectLoading {
		lines = append(lines, "Loading project…")
	} else if data.Loading {
		lines = append(lines, "Refreshing…")
	}

	if data.ErrorMessage != "" {
		lines = append(lines, "Error: "+data.ErrorMessage)
	}

	return lines
}

func (s EntityListState) issueHeader(data EntityListViewData) []string {
	lines := []string{"Project: " + s.projectPath, "Filter: " + s.query}

	if data.Loading {
		lines = append(lines, "Refreshing…")
	}

	if data.ErrorMessage != "" {
		lines = append(lines, "Error: "+data.ErrorMessage)
	}

	return lines
}

func (s EntityListState) mrNoItemsMsg(data EntityListViewData) string {
	if data.ProjectLoading {
		return ""
	}

	return "No opened MRs"
}

func (s EntityListState) issueNoItemsMsg() string { return "No issues" }

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
