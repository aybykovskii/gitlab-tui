package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateLabelSelect(msg tea.KeyMsg) (Model, tea.Cmd) {
	count := len(m.projectLabels)
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = ModeDetail
		m.labelPending = nil
		return m, nil
	case tea.KeyEnter:
		item, ok := m.selectedItem()
		if !ok {
			m.mode = ModeDetail
			return m, nil
		}
		selected := append([]string(nil), m.labelPending...)
		prev := append([]string(nil), item.Labels...)
		for i := range m.items {
			if m.items[i].IID == item.IID {
				m.items[i].Labels = selected
				break
			}
		}
		m.mode = ModeDetail
		m.labelPending = nil
		if m.updateMRLabels == nil {
			return m, nil
		}
		fn := m.updateMRLabels
		iid := item.IID
		return m, func() tea.Msg {
			err := fn(iid, selected)
			return updateMRLabelsFinishedMsg{iid: iid, labels: selected, prev: prev, err: err}
		}
	case tea.KeyRunes:
		switch msg.String() {
		case "k", "up":
			m.labelCursor = clamp(m.labelCursor-1, 0, max(0, count-1))
		case "j", "down":
			m.labelCursor = clamp(m.labelCursor+1, 0, max(0, count-1))
		case " ":
			if count > 0 {
				name := m.projectLabels[m.labelCursor].Name
				m.labelPending = toggleStringSlice(m.labelPending, name)
			}
		}
	case tea.KeyUp:
		m.labelCursor = clamp(m.labelCursor-1, 0, max(0, count-1))
	case tea.KeyDown:
		m.labelCursor = clamp(m.labelCursor+1, 0, max(0, count-1))
	case tea.KeySpace:
		if count > 0 {
			name := m.projectLabels[m.labelCursor].Name
			m.labelPending = toggleStringSlice(m.labelPending, name)
		}
	}
	return m, nil
}

func (m Model) labelColor(name string) string {
	for _, l := range m.projectLabels {
		if l.Name == name {
			return l.Color
		}
	}
	return ""
}
