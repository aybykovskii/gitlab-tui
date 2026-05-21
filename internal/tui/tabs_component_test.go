package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTabsComponentView(t *testing.T) {
	t.Run("marks active tab", func(t *testing.T) {
		t.Parallel()

		component := TabsComponent{Labels: []string{"Summary", "Discussions"}, Active: 1}
		assert.Equal(t, "[Summary] [>Discussions<]", component.View())
	})

	t.Run("all inactive when index is out of range", func(t *testing.T) {
		t.Parallel()

		component := TabsComponent{Labels: []string{"Summary", "Discussions"}, Active: -1}
		assert.Equal(t, "[Summary] [Discussions]", component.View())
	})

	t.Run("single tab", func(t *testing.T) {
		t.Parallel()

		component := TabsComponent{Labels: []string{"Summary"}, Active: 0}
		assert.Equal(t, "[>Summary<]", component.View())
	})

	t.Run("joins labels with single space", func(t *testing.T) {
		t.Parallel()

		component := TabsComponent{Labels: []string{"Summary", "Discussions", "Files"}, Active: 2}
		assert.Equal(t, "[Summary] [Discussions] [>Files<]", component.View())
	})
}
