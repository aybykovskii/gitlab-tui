package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/issue"
)

type IssueDetailState struct {
	viewport.Model
	activeTab        DetailTab
	discussions      map[int][]issue.Discussion
	discussionCursor int
	discussionsError string
}

func NewIssueDetailState() IssueDetailState {
	return IssueDetailState{
		Model:       viewport.New(0, 0),
		discussions: map[int][]issue.Discussion{},
	}
}

func (s *IssueDetailState) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(issueDiscussionsFinishedMsg); ok {
		if msg.err != nil {
			s.discussionsError = msg.err.Error()

			return nil
		}

		s.discussions[msg.iid] = msg.discussions

		return nil
	}

	var cmd tea.Cmd
	s.Model, cmd = s.Model.Update(msg)

	return cmd
}

func (s *IssueDetailState) View(layout LayoutState, item issue.Issue) string {
	s.Width = layout.Width
	s.Height = max(1, layout.Height)
	s.SetContent(s.content(item))

	return s.Model.View()
}

func (s IssueDetailState) content(item issue.Issue) string {
	header := fmt.Sprintf("#%d %s\n%s", item.IID, item.Title, s.tabs())
	if s.activeTab == TabDiscussions {
		return header + "\n\n" + s.discussionsContent(item)
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

	lines = append(lines, "", item.Description)

	return strings.Join(lines, "\n")
}

func (s IssueDetailState) tabs() string {
	return TabsComponent{Labels: []string{"Summary", "Discussions"}, Active: int(s.activeTab)}.View()
}

func (s IssueDetailState) discussionsContent(item issue.Issue) string {
	discussions, loaded := s.discussions[item.IID]
	if !loaded {
		return "No discussions"
	}

	return renderDiscussionList(discussions, 0, DiscussionListOptions{})
}
