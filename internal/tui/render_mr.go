//nolint:mnd,gocritic // UI layout branching and dimensions are deliberately explicit.
package tui

import (
	"fmt"
	"strings"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func (m Model) renderList() string {
	width := m.leftWidth()
	height := m.paneHeight()
	style := paneStyle(width, height, m.focus == FocusList || m.focus == FocusFilter)

	lines := []string{"Project: " + m.projectPath, "Merge Requests", "Filter: " + m.query}
	if m.projectLoading {
		lines = append(lines, "Loading project…")
	} else if m.loading {
		lines = append(lines, "Refreshing…")
	}

	if m.errorMessage != "" {
		lines = append(lines, "Error: "+m.errorMessage)
	}

	items := m.filtered()
	if len(items) == 0 {
		lines = append(lines, "No opened MRs")
	} else {
		visible := max(1, height-5)

		end := min(len(items), m.listTop+visible)
		for i := m.listTop; i < end; i++ {
			prefix := "  "
			if i == m.selected {
				prefix = "> "
			}

			item := items[i]
			lines = append(lines, fmt.Sprintf("%s%s !%d %s", prefix, pipelineIcon(item.Pipeline), item.IID, item.Title))
			lines = append(lines, fmt.Sprintf("  %s %s → %s", item.Author, item.SourceBranch, item.TargetBranch))
		}
	}

	return style.Render(strings.Join(lines, "\n"))
}

//nolint:gocyclo // MR rendering handles several UI modes in one view.
func (m Model) renderRight() string {
	width := max(20, m.width-m.leftWidth())
	height := m.paneHeight()

	style := paneStyle(width, height, m.focus == FocusDetail)
	if m.section == SectionIssues {
		return style.Render(m.renderIssueDetail())
	}

	items := m.filtered()
	if len(items) == 0 {
		return style.Render("No MR selected")
	}

	item := items[clampSelection(m.selected, len(items))]
	tabs := TabsComponent{
		Labels: []string{"Summary", "Discussions", "Files", m.reviewTabLabel(item)},
		Active: int(m.activeTab),
	}.View()

	icons := m.emoji.Resolve()
	header := fmt.Sprintf("%s\n%s", mrTitleLine(item, icons), tabs)

	switch m.activeTab {
	case TabDiscussions:
		return style.Render(header + "\n\n" + m.renderDiscussions(item))
	case TabFiles:
		return style.Render(header + "\n\n" + m.renderFiles(item))
	case TabReview:
		return style.Render(header + "\n\n" + m.renderReview(item))
	default:
		authorPart := iconPrefix(icons.Author) + formatAuthor(item)

		reviewerPart := ""
		if len(item.Reviewers) > 0 {
			reviewerPart = iconPrefix(icons.Reviewers) + strings.Join(item.Reviewers, ", ")
		}

		assigneePart := ""
		if len(item.Assignees) > 0 {
			assigneePart = iconPrefix(icons.Assignees) + strings.Join(item.Assignees, ", ")
		}

		peopleParts := []string{authorPart}
		if reviewerPart != "" {
			peopleParts = append(peopleParts, reviewerPart)
		}

		if assigneePart != "" {
			peopleParts = append(peopleParts, assigneePart)
		}

		statePart := stateEmoji(icons, item.State) + " " + item.State
		if icons.State == "" {
			statePart = item.State
		}

		pipelinePart := iconPrefix(icons.Pipeline) + pipelineIcon(item.Pipeline) + " " + item.Pipeline
		approvalsPart := iconPrefix(icons.Approvals) + item.Approvals

		lines := []string{
			header,
			"",
			strings.Join(peopleParts, "  ·  "),
			iconPrefix(icons.Branch) + item.SourceBranch + " → " + item.TargetBranch,
			strings.Join([]string{statePart, pipelinePart, approvalsPart}, "  ·  "),
		}

		if len(item.Labels) > 0 {
			pills := make([]string, 0, len(item.Labels))

			for _, name := range item.Labels {
				color := m.labelColor(name)
				pills = append(pills, renderLabelPill(name, color))
			}

			lines = append(lines, iconPrefix(icons.Labels)+strings.Join(pills, " "))
		}

		if item.WebURL != "" {
			lines = append(lines, "URL: "+item.WebURL)
		}

		if m.actionError != "" {
			lines = append(lines, "", "Action: "+m.actionError)
		}

		if m.mergeConfirmPending {
			lines = append(lines, "", "Press M again to confirm merge  (any other key cancels)")
		}

		if m.mrCommentError != "" {
			lines = append(lines, "", "Comment error: "+m.mrCommentError)
		}

		if m.editInput {
			lines = append(lines, "", fmt.Sprintf("Edit %s: %s█", m.editField, m.editBuffer))
		} else if m.mrCommentInput {
			lines = append(lines, "", "MR comment: "+m.mrCommentBuffer+"█")
		} else {
			lines = append(lines, "", item.Description)
		}

		return style.Render(strings.Join(lines, "\n"))
	}
}

func (m Model) renderFiles(item mr.MergeRequest) string {
	if m.filesLoading {
		return "Loading files…"
	}

	if m.filesError != "" {
		return "Error: " + m.filesError + "\n\nr retry"
	}

	files, loaded := m.changedFiles[item.IID]
	if !loaded {
		return "Tab to load files"
	}

	if len(files) == 0 {
		return "No changed files"
	}

	lines := []string{}

	for _, file := range files {
		marker := " "
		if file.IsNew {
			marker = "A"
		} else if file.IsDeleted {
			marker = "D"
		} else if file.IsRenamed {
			marker = "R"
		}

		lines = append(lines, fmt.Sprintf("%s %s  +%d -%d", marker, file.Path, file.AddedLines, file.RemovedLines))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderReview(item mr.MergeRequest) string {
	drafts := m.drafts[item.IID]
	if len(drafts) == 0 {
		return "No draft comments\n\nAdd inline comments from Files diff."
	}

	lines := []string{"Draft comments", ""}

	for i, draft := range drafts {
		prefix := "  "
		if i == m.reviewCursor && !m.reviewSummaryInput {
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

	lines = append(lines, "")
	summaryPrefix := "  "
	cursor := ""

	if m.reviewSummaryInput {
		summaryPrefix = "> "
		cursor = "█"
	}

	lines = append(lines, summaryPrefix+"Summary: "+m.reviewSummary+cursor)
	lines = append(lines, "", "Enter open draft  p publish  D discard")

	return strings.Join(lines, "\n")
}

func (m Model) reviewTabLabel(item mr.MergeRequest) string {
	count := len(m.drafts[item.IID])
	if count == 0 {
		return "Review"
	}

	return fmt.Sprintf("Review (%d)", count)
}

func (m Model) renderLabelSelector() string {
	width := max(20, m.width-m.leftWidth())
	height := m.paneHeight()
	style := paneStyle(width, height, true)
	lines := []string{"Labels  Space toggle  Enter save  Esc cancel", ""}

	for i, label := range m.projectLabels {
		marker := "○"

		for _, selected := range m.labelPending {
			if selected == label.Name {
				marker = "●"
				break
			}
		}

		cursor := "  "
		if i == m.labelCursor {
			cursor = "> "
		}

		lines = append(lines, fmt.Sprintf("%s%s %s", cursor, marker, renderLabelPill(label.Name, label.Color)))
	}

	if len(m.projectLabels) == 0 {
		lines = append(lines, "No project labels")
	}

	return style.Render(strings.Join(lines, "\n"))
}
