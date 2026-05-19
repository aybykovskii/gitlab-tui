package tui

import tea "github.com/charmbracelet/bubbletea"

func batchCommands(cmds ...tea.Cmd) tea.Cmd {
	present := make([]tea.Cmd, 0, len(cmds))

	for _, cmd := range cmds {
		if cmd != nil {
			present = append(present, cmd)
		}
	}

	switch len(present) {
	case 0:
		return nil
	case 1:
		return present[0]
	default:
		return tea.Batch(present...)
	}
}
