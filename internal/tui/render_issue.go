package tui

import (
	"fmt"
	"strings"

	"github.com/aybykovskii/gitlab-tui/internal/issue"
)

func (m Model) renderIssueDetail() string {
	items := m.filteredIssues()
	if len(items) == 0 {
		return "No issue selected"
	}

	item := items[clampSelection(m.selected, len(items))]

	tabs := "[>Summary<] [Discussions]"
	if m.activeTab == TabDiscussions {
		tabs = "[Summary] [>Discussions<]"
	}

	header := fmt.Sprintf("#%d %s\n%s", item.IID, item.Title, tabs)
	if m.activeTab == TabDiscussions {
		return header + "\n\n" + m.renderIssueDiscussions(item)
	}

	lines := []string{
		header,
		"",
		"👤 " + item.Author + " · assigned: " + strings.Join(item.Assignees, ", "),
		issueStateIcon(item.State) + " " + item.State + fmt.Sprintf(" · 💬 %d", item.CommentCount),
		"🏷️ " + formatIssueLabels(item.Labels),
		"📅 Due: " + item.DueDate + " · 🏁 " + item.Milestone,
	}
	if item.Weight > 0 {
		lines = append(lines, fmt.Sprintf("⚖️ Weight: %d", item.Weight))
	}

	if m.editInput {
		lines = append(lines, "", fmt.Sprintf("Edit %s: %s█", m.editField, m.editBuffer))
	} else if m.issueCommentInput {
		lines = append(lines, "", "Issue comment: "+m.issueCommentBuffer+"█")
	} else {
		lines = append(lines, "", item.Description)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderIssueDiscussions(item issue.Issue) string {
	discussions, loaded := m.issueDiscussions[item.IID]
	if !loaded {
		return "No discussions"
	}

	output := renderDiscussionList(discussions, m.discussionCursor, DiscussionListOptions{})
	if m.replyInput {
		output += "\n\nReply: " + m.replyBuffer + "█"
	}

	return output
}

func (m Model) issueListLines(height int) []string {
	lines := []string{"Project: " + m.projectPath, "Issues [" + m.issueStateLabel() + "]", "Filter: " + m.query}
	if m.loading {
		lines = append(lines, "Refreshing…")
	}

	if m.errorMessage != "" {
		lines = append(lines, "Error: "+m.errorMessage)
	}

	items := m.filteredIssues()
	if len(items) == 0 {
		lines = append(lines, "No issues")
		return lines
	}

	visible := max(1, (height-5)/2)

	end := min(len(items), m.listTop+visible)
	for i := m.listTop; i < end; i++ {
		prefix := "  "
		if i == m.selected {
			prefix = "> "
		}

		item := items[i]
		lines = append(lines, fmt.Sprintf("%s#%d %s", prefix, item.IID, item.Title))
		lines = append(lines, "  "+formatIssueMeta(item))
	}

	return lines
}
