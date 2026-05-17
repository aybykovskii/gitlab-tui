package tui

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type Mode int

const (
	ModeProjectSelect Mode = iota
	ModeProjectInput
	ModeDetail
	ModeDiff
)

type Focus int

const (
	FocusList Focus = iota
	FocusDetail
	FocusFilter
)

type RefreshFunc func() ([]mr.MergeRequest, error)

type ProjectOptions struct {
	Path    string
	Recents []string
	Items   []mr.MergeRequest
	Refresh RefreshFunc
}

type refreshStartedMsg struct{}

type refreshFinishedMsg struct {
	items []mr.MergeRequest
	err   error
}

type Model struct {
	items          []mr.MergeRequest
	query          string
	selected       int
	mode           Mode
	focus          Focus
	width          int
	height         int
	listTop        int
	rightTop       int
	projectPath    string
	recentProjects []string
	projectInput   string
	refresh        RefreshFunc
	loading        bool
	errorMessage   string
}

func NewFakeModel() Model {
	return NewModel(FakeMergeRequests())
}

func NewModel(items []mr.MergeRequest) Model {
	return NewModelWithProject(items, ProjectOptions{Path: "group/project"})
}

func NewModelWithProject(items []mr.MergeRequest, options ProjectOptions) Model {
	if options.Items != nil {
		items = options.Items
	}
	model := Model{
		items:          items,
		focus:          FocusList,
		width:          100,
		height:         30,
		projectPath:    options.Path,
		recentProjects: options.Recents,
		refresh:        options.Refresh,
	}
	if model.projectPath == "" {
		if len(model.recentProjects) > 0 {
			model.mode = ModeProjectSelect
		} else {
			model.mode = ModeProjectInput
			model.focus = FocusFilter
		}
	} else {
		model.mode = ModeDetail
	}
	return model
}

func Run(stdout io.Writer) error {
	return RunWithProject(stdout, ProjectOptions{Path: "group/project"})
}

func RunWithProject(stdout io.Writer, options ProjectOptions) error {
	program := tea.NewProgram(NewModelWithProject(FakeMergeRequests(), options), tea.WithMouseCellMotion(), tea.WithOutput(stdout))
	_, err := program.Run()
	return err
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		return m.updateKey(msg)
	case tea.MouseMsg:
		return m.updateMouse(msg), nil
	case refreshStartedMsg:
		m.loading = true
		m.errorMessage = ""
		return m, nil
	case refreshFinishedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			return m, nil
		}
		m.items = msg.items
		m.selected = clampSelection(m.selected, len(m.filtered()))
		m.listTop = 0
		return m, nil
	}

	return m, nil
}

func (m Model) updateKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.mode == ModeProjectSelect {
		switch msg.String() {
		case "up", "k":
			m.selected = clamp(m.selected-1, 0, len(m.recentProjects)-1)
		case "down", "j":
			m.selected = clamp(m.selected+1, 0, len(m.recentProjects)-1)
		case "enter":
			if len(m.recentProjects) > 0 {
				m.projectPath = m.recentProjects[m.selected]
				m.mode = ModeDetail
				m.selected = 0
			}
		case "i":
			m.mode = ModeProjectInput
			m.focus = FocusFilter
			m.projectInput = ""
		}
		return m, nil
	}

	if m.mode == ModeProjectInput {
		switch msg.Type {
		case tea.KeyEnter:
			if strings.TrimSpace(m.projectInput) != "" {
				m.projectPath = strings.TrimSpace(m.projectInput)
				m.mode = ModeDetail
				m.focus = FocusList
			}
		case tea.KeyBackspace:
			if len(m.projectInput) > 0 {
				m.projectInput = m.projectInput[:len(m.projectInput)-1]
			}
		case tea.KeyRunes:
			m.projectInput += msg.String()
		}
		return m, nil
	}

	if m.focus == FocusFilter {
		switch msg.Type {
		case tea.KeyEsc, tea.KeyEnter:
			m.focus = FocusList
		case tea.KeyBackspace:
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.selected = clampSelection(m.selected, len(m.filtered()))
			}
		case tea.KeyRunes:
			m.query += msg.String()
			m.selected = clampSelection(m.selected, len(m.filtered()))
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "/":
		m.focus = FocusFilter
	case "r":
		return m, m.refreshCommand()
	case "tab":
		if m.focus == FocusList {
			m.focus = FocusDetail
		} else {
			m.focus = FocusList
		}
	case "up", "k":
		m.moveSelection(-1)
	case "down", "j":
		m.moveSelection(1)
	case "enter":
		if len(m.filtered()) > 0 {
			m.mode = ModeDiff
			m.focus = FocusDetail
			m.rightTop = 0
		}
	case "esc", "backspace":
		if m.mode == ModeDiff {
			m.mode = ModeDetail
			m.rightTop = 0
		}
	}

	return m, nil
}

func (m Model) refreshCommand() tea.Cmd {
	if m.refresh == nil || m.loading {
		return nil
	}
	refresh := m.refresh
	return tea.Sequence(
		func() tea.Msg { return refreshStartedMsg{} },
		func() tea.Msg {
			items, err := refresh()
			return refreshFinishedMsg{items: items, err: err}
		},
	)
}

func (m Model) updateMouse(msg tea.MouseMsg) Model {
	if m.mode == ModeProjectSelect {
		if msg.Button == tea.MouseButtonLeft && msg.Y >= 2 {
			idx := msg.Y - 2
			if idx >= 0 && idx < len(m.recentProjects) {
				m.selected = idx
				m.projectPath = m.recentProjects[idx]
				m.mode = ModeDetail
			}
		}
		if msg.Button == tea.MouseButtonWheelUp {
			m.selected = clamp(m.selected-1, 0, len(m.recentProjects)-1)
		}
		if msg.Button == tea.MouseButtonWheelDown {
			m.selected = clamp(m.selected+1, 0, len(m.recentProjects)-1)
		}
		return m
	}

	leftWidth := m.leftWidth()
	if msg.X < leftWidth {
		m.focus = FocusList
	} else {
		m.focus = FocusDetail
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.scrollFocused(-1)
	case tea.MouseButtonWheelDown:
		m.scrollFocused(1)
	case tea.MouseButtonLeft:
		if msg.X < leftWidth && msg.Y >= 4 {
			idx := m.listTop + msg.Y - 4
			if idx >= 0 && idx < len(m.filtered()) {
				m.selected = idx
				m.mode = ModeDetail
			}
		} else if msg.X >= leftWidth && len(m.filtered()) > 0 {
			m.mode = ModeDiff
		}
	}

	return m
}

func (m *Model) moveSelection(delta int) {
	count := len(m.filtered())
	if count == 0 {
		m.selected = 0
		return
	}

	m.selected = clamp(m.selected+delta, 0, count-1)
	visible := max(1, m.height-4)
	if m.selected < m.listTop {
		m.listTop = m.selected
	}
	if m.selected >= m.listTop+visible {
		m.listTop = m.selected - visible + 1
	}
}

func (m *Model) scrollFocused(delta int) {
	if m.focus == FocusList {
		m.moveSelection(delta)
		return
	}
	m.rightTop = max(0, m.rightTop+delta)
}

func (m Model) View() string {
	if m.mode == ModeProjectSelect || m.mode == ModeProjectInput {
		return m.renderProjectPicker()
	}

	left := m.renderList()
	right := m.renderRight()
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) renderProjectPicker() string {
	style := paneStyle(max(40, m.width), max(8, m.height), true)
	if m.mode == ModeProjectInput {
		return style.Render(strings.Join([]string{
			"Open GitLab project",
			"",
			"Project path:",
			m.projectInput,
			"",
			"Enter: open project",
		}, "\n"))
	}

	lines := []string{"Recent projects", ""}
	for i, project := range m.recentProjects {
		prefix := "  "
		if i == m.selected {
			prefix = "> "
		}
		lines = append(lines, prefix+project)
	}
	lines = append(lines, "", "Enter/click: open  i: manual input")
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderList() string {
	width := m.leftWidth()
	height := max(8, m.height)
	style := paneStyle(width, height, m.focus == FocusList || m.focus == FocusFilter)
	lines := []string{"Project: " + m.projectPath, "Merge Requests", "Filter: " + m.query}
	if m.loading {
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
			lines = append(lines, fmt.Sprintf("%s!%d %s", prefix, items[i].IID, items[i].Title))
		}
	}
	lines = append(lines, "", "↑/↓ select  / filter  r refresh  Enter diff")
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderRight() string {
	width := max(20, m.width-m.leftWidth())
	height := max(8, m.height)
	style := paneStyle(width, height, m.focus == FocusDetail)
	items := m.filtered()
	if len(items) == 0 {
		return style.Render("No MR selected")
	}
	item := items[clampSelection(m.selected, len(items))]
	if m.mode == ModeDiff {
		return style.Render(m.renderDiff(item))
	}
	return style.Render(strings.Join([]string{
		fmt.Sprintf("!%d %s", item.IID, item.Title),
		"",
		"Author: " + item.Author,
		"Branches: " + item.SourceBranch + " → " + item.TargetBranch,
		"State: " + item.State,
		"Pipeline: " + item.Pipeline,
		"Approvals: " + item.Approvals,
		"",
		item.Description,
		"",
		"Enter/click right pane: open fake diff",
	}, "\n"))
}

func (m Model) renderDiff(item mr.MergeRequest) string {
	lines := []string{fmt.Sprintf("Diff !%d %s", item.IID, item.Title), ""}
	for _, row := range item.Diff {
		lines = append(lines, fmt.Sprintf("%4d │ %-36s │ %4d │ %s", row.OldLine, row.OldText, row.NewLine, row.NewText))
	}
	lines = append(lines, "", "Esc/backspace: back to detail")
	if m.rightTop >= len(lines) {
		m.rightTop = max(0, len(lines)-1)
	}
	visible := max(1, m.height-2)
	end := min(len(lines), m.rightTop+visible)
	return strings.Join(lines[m.rightTop:end], "\n")
}

func (m Model) filtered() []mr.MergeRequest {
	return mr.Filter(m.items, m.query)
}

func (m Model) leftWidth() int {
	if m.width <= 0 {
		return 40
	}
	return max(24, m.width*35/100)
}

func paneStyle(width int, height int, focused bool) lipgloss.Style {
	color := lipgloss.Color("240")
	if focused {
		color = lipgloss.Color("63")
	}
	return lipgloss.NewStyle().Width(width-2).Height(height-2).Border(lipgloss.RoundedBorder()).BorderForeground(color).Padding(0, 1)
}

func clampSelection(selected int, count int) int {
	if count <= 0 {
		return 0
	}
	return clamp(selected, 0, count-1)
}

func clamp(v int, minValue int, maxValue int) int {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
