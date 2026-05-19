package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type MRDetailState struct {
	viewport.Model
	activeTab          DetailTab
	discussions        map[int][]mr.Discussion
	changedFiles       map[int][]mr.ChangedFile
	reviewCursor       int
	reviewSummaryInput bool
	reviewSummary      string
	drafts             map[int][]mr.DraftComment
	discussionsLoading bool
	discussionsError   string
	filesLoading       bool
	filesError         string
	discussionCursor   int
}

func NewMRDetailState() MRDetailState {
	return MRDetailState{
		Model:        viewport.New(0, 0),
		discussions:  map[int][]mr.Discussion{},
		changedFiles: map[int][]mr.ChangedFile{},
		drafts:       map[int][]mr.DraftComment{},
	}
}

func (s *MRDetailState) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case discussionsStartedMsg:
		s.discussionsLoading = true
		s.discussionsError = ""

		return nil
	case discussionsFinishedMsg:
		s.discussionsLoading = false
		if msg.err != nil {
			s.discussionsError = msg.err.Error()

			return nil
		}

		s.discussions[msg.iid] = msg.discussions

		return nil
	case filesStartedMsg:
		s.filesLoading = true
		s.filesError = ""

		return nil
	case filesFinishedMsg:
		s.filesLoading = false
		if msg.err != nil {
			s.filesError = msg.err.Error()

			return nil
		}

		s.changedFiles[msg.iid] = msg.files

		return nil
	}

	// Key messages are handled by update_detail.go via direct YOffset manipulation.
	// Delegating them to the viewport would cause double-scroll.
	if _, ok := msg.(tea.KeyMsg); ok {
		return nil
	}

	var cmd tea.Cmd
	s.Model, cmd = s.Model.Update(msg)

	return cmd
}

func (s *MRDetailState) View(layout LayoutState, item mr.MergeRequest) string {
	s.Width = layout.Width
	s.Height = max(1, layout.Height)
	s.SetContent(s.content(item))

	return s.Model.View()
}

func (s MRDetailState) content(item mr.MergeRequest) string {
	header := fmt.Sprintf("!%d %s\n%s", item.IID, item.Title, s.tabs(item))

	switch s.activeTab {
	case TabDiscussions:
		return header + "\n\n" + s.discussionsContent(item)
	case TabFiles:
		return header + "\n\n" + s.filesContent(item)
	case TabReview:
		return header + "\n\n" + s.reviewContent(item)
	default:
		return strings.Join([]string{header, "", item.Author, item.Description}, "\n")
	}
}

func (s MRDetailState) tabs(item mr.MergeRequest) string {
	return TabsComponent{
		Labels: []string{"Summary", "Discussions", "Files", s.reviewTabLabel(item)},
		Active: int(s.activeTab),
	}.View()
}

func (s MRDetailState) reviewTabLabel(item mr.MergeRequest) string {
	count := len(s.drafts[item.IID])
	if count == 0 {
		return "Review"
	}

	return fmt.Sprintf("Review (%d)", count)
}

func (s MRDetailState) discussionsContent(item mr.MergeRequest) string {
	discussions := s.discussions[item.IID]
	if len(discussions) == 0 {
		return "No discussions"
	}

	return renderDiscussionList(discussions, 0, DiscussionListOptions{})
}

func (s MRDetailState) filesContent(item mr.MergeRequest) string {
	files, loaded := s.changedFiles[item.IID]
	if !loaded {
		return "Tab to load files"
	}

	if len(files) == 0 {
		return "No changed files"
	}

	lines := make([]string, 0, len(files))
	for _, file := range files {
		lines = append(lines, file.Path)
	}

	return strings.Join(lines, "\n")
}

func (s MRDetailState) reviewContent(item mr.MergeRequest) string {
	drafts := s.drafts[item.IID]
	if len(drafts) == 0 {
		return "No draft comments\n\nAdd inline comments from Files diff."
	}

	lines := []string{"Draft comments", ""}

	for i, draft := range drafts {
		prefix := "  "
		if i == s.reviewCursor && !s.reviewSummaryInput {
			prefix = "> "
		}

		path := "unknown"
		line := 0

		if draft.Position != nil {
			path = draft.Position.NewPath
			line = draft.Position.NewLine
		}

		lines = append(lines, fmt.Sprintf("%s%s:%d %s", prefix, path, line, oneLinePreview(draft.Body)))
	}

	return strings.Join(lines, "\n")
}
