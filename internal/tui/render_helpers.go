//nolint:mnd // UI layout constants are deliberate minimum dimensions.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

func (m Model) renderEntityListPane() string {
	width := max(20, m.width-m.leftWidth())
	height := m.paneHeight()
	style := paneStyle(width, height, true)

	if m.section == SectionIssues {
		return style.Render(strings.Join(m.issueListLines(height), "\n"))
	}

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

func bindingHelp(binding key.Binding) string {
	if !binding.Enabled() {
		return ""
	}

	keyText, description := binding.Help().Key, binding.Help().Desc
	if keyText == "" || description == "" {
		return ""
	}

	return keyText + " " + description
}

func joinBindingHelp(bindings []key.Binding) string {
	parts := []string{}

	for _, binding := range bindings {
		if help := bindingHelp(binding); help != "" {
			parts = append(parts, help)
		}
	}

	return strings.Join(parts, "  ")
}

func truncateLine(line string, width int) string {
	if width <= 0 || len(line) <= width {
		return line
	}

	if width == 1 {
		return "…"
	}

	return line[:width-1] + "…"
}

func (m Model) renderKeyBar() string {
	width := max(20, m.width)
	inner := max(1, width-4)
	lines := []string{}

	if m.keyBarExpanded {
		locals := m.localKeys()
		mid := (len(locals) + 1) / 2

		for i := 0; i < mid; i++ {
			left := bindingHelp(locals[i])

			right := ""
			if i+mid < len(locals) {
				right = bindingHelp(locals[i+mid])
			}

			lines = append(lines, fmt.Sprintf("%-24s %s", left, right))
		}

		lines = append(lines, strings.Repeat("─", min(inner, 24)))
		lines = append(lines, "Global: "+joinBindingHelp(m.globalKeys()))
	} else {
		lines = append(lines, truncateLine("Local: "+joinBindingHelp(m.localKeys()), inner))
		lines = append(lines, truncateLine("Global: "+joinBindingHelp(m.globalKeys()), inner))
	}

	style := lipgloss.NewStyle().Width(width - 2).Border(lipgloss.RoundedBorder())

	return style.Render(strings.Join(lines, "\n"))
}

func issueStateIcon(state string) string {
	if state == "closed" {
		return "🔴"
	}

	return "🟢"
}

func formatIssueLabels(labels []string) string {
	parts := make([]string, 0, len(labels))
	for _, label := range labels {
		parts = append(parts, "["+label+"]")
	}

	return strings.Join(parts, " ")
}

func mrTitleLine(item mr.MergeRequest, icons config.EmojiMap) string {
	prefix := ""
	if icons.Draft != "" {
		prefix = icons.Draft + " "
	}

	title := item.Title
	if item.Draft {
		title = "Draft: " + title
	}

	return fmt.Sprintf("%s!%d %s", prefix, item.IID, title)
}

func stateEmoji(icons config.EmojiMap, state string) string {
	if icons.State == "" {
		return ""
	}

	switch state {
	case "opened":
		return "🟢"
	case "merged":
		return "🟣"
	case "closed":
		return "🔴"
	default:
		return icons.State
	}
}

func iconPrefix(icon string) string {
	if icon == "" {
		return ""
	}

	return icon + " "
}

func formatAuthor(item mr.MergeRequest) string {
	if item.AuthorUsername == "" || item.AuthorUsername == item.Author {
		return item.Author
	}

	return item.Author + " @" + item.AuthorUsername
}

func pipelineIcon(status string) string {
	switch status {
	case "success":
		return "✓"
	case "failed":
		return "✗"
	case "running":
		return "●"
	case "pending":
		return "○"
	default:
		return "–"
	}
}

func formatIssueMeta(item issue.Issue) string {
	parts := []string{item.Author}

	labels := item.Labels
	if len(labels) > 2 {
		labels = labels[:2]
	}

	labelParts := make([]string, 0, len(labels))
	for _, label := range labels {
		labelParts = append(labelParts, "["+label+"]")
	}

	if len(labelParts) > 0 {
		parts = append(parts, strings.Join(labelParts, " "))
	}

	if item.CommentCount > 0 {
		parts = append(parts, fmt.Sprintf("💬 %d", item.CommentCount))
	}

	return strings.Join(parts, " · ")
}

func paneStyle(width int, height int, focused bool) lipgloss.Style {
	color := lipgloss.Color("240")
	if focused {
		color = lipgloss.Color("63")
	}

	return lipgloss.NewStyle().Width(width-2).Height(height-2).Border(lipgloss.RoundedBorder()).BorderForeground(color).Padding(0, 1)
}
