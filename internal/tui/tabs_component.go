package tui

import (
	"fmt"
	"strings"
)

type LayoutState struct {
	Width  int
	Height int
	Focus  Focus
	Mode   Mode
}

type TabsComponent struct {
	Labels []string
	Active int
}

func (t TabsComponent) View() string {
	parts := make([]string, 0, len(t.Labels))

	for i, label := range t.Labels {
		if i == t.Active {
			parts = append(parts, fmt.Sprintf("[>%s<]", label))
			continue
		}

		parts = append(parts, fmt.Sprintf("[%s]", label))
	}

	return strings.Join(parts, " ")
}
