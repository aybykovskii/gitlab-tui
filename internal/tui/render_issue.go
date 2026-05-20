//nolint:mnd,gocritic // UI layout branching and dimensions are deliberately explicit.
package tui

import (
	"fmt"
	"strings"

	"github.com/aybykovskii/gitlab-tui/internal/issue"
)

func (m Model) renderIssueDetail() string {
	item, ok := m.selectedIssue()
	if !ok {
		return "No issue selected"
	}

	if !m.editInput && !m.issueCommentInput && !m.replyInput {
		return m.IssueDetailState.View(LayoutState{Width: m.width, Height: m.height, Focus: m.focus, Mode: m.mode}, item)
	}

	tabs := TabsComponent{
		Labels: []string{"Summary", "Discussions"},
		Active: int(m.IssueDetailState.activeTab),
	}.View()

	header := fmt.Sprintf("#%d %s\n%s", item.IID, item.Title, tabs)
	if m.IssueDetailState.activeTab == TabDiscussions {
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
		lines = append(lines, "", fmt.Sprintf("Edit %s: %s█", m.editField, m.Value()))
	} else if m.issueCommentInput {
		lines = append(lines, "", "Issue comment: "+m.Value()+"█")
	} else {
		lines = append(lines, "", item.Description)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderIssueDiscussions(item issue.Issue) string {
	discussions, loaded := m.IssueDetailState.discussions[item.IID]
	if !loaded {
		return "No discussions"
	}

	output := renderDiscussionList(discussions, m.IssueDetailState.discussionCursor, DiscussionListOptions{})
	if m.replyInput {
		output += "\n\nReply: " + m.Value() + "█"
	}

	return output
}
