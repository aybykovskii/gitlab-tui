package tui

import (
	"fmt"
	"strings"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type LabelSelectorState struct {
	labels  []mr.Label
	pending []string
	cursor  int
}

func NewLabelSelectorState() LabelSelectorState {
	return LabelSelectorState{}
}

func (s LabelSelectorState) View(layout LayoutState) string {
	lines := []string{"Labels  Space toggle  Enter save  Esc cancel", ""}

	for i, label := range s.labels {
		marker := "○"

		for _, sel := range s.pending {
			if sel == label.Name {
				marker = "●"
				break
			}
		}

		cursor := "  "
		if i == s.cursor {
			cursor = "> "
		}

		lines = append(lines, fmt.Sprintf("%s%s %s", cursor, marker, renderLabelPill(label.Name, label.Color)))
	}

	if len(s.labels) == 0 {
		lines = append(lines, "No project labels")
	}

	return strings.Join(lines, "\n")
}
