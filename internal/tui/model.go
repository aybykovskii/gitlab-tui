package tui

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

type Mode int

const (
	ModeProjectSelect Mode = iota
	ModeProjectInput
	ModeSections
	ModeDetail
	ModeDiff
)

const SectionMergeRequests = "mr"

type Focus int

const (
	FocusList Focus = iota
	FocusDetail
	FocusFilter
)

type RefreshFunc func() ([]mr.MergeRequest, error)
type DiffFunc func(iid int) ([]mr.DiffRow, error)
type ProjectLoadFunc func(path string) (ProjectData, error)

type ProjectData struct {
	Items    []mr.MergeRequest
	Refresh  RefreshFunc
	LoadDiff DiffFunc
}

type ProjectOptions struct {
	Path        string
	Recents     []string
	Projects    []string
	Section     string
	EntityID    string
	Items       []mr.MergeRequest
	Refresh     RefreshFunc
	LoadDiff    DiffFunc
	LoadProject ProjectLoadFunc
}

type projectStartedMsg struct {
	path string
}

type projectFinishedMsg struct {
	path string
	data ProjectData
	err  error
}

type refreshStartedMsg struct{}

type refreshFinishedMsg struct {
	items []mr.MergeRequest
	err   error
}

type diffStartedMsg struct{}

type diffFinishedMsg struct {
	iid  int
	rows []mr.DiffRow
	err  error
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
	gitlabProjects []string
	projectList    []string
	section        string
	sectionCursor  int
	entityID       string
	projectInput   string
	refresh        RefreshFunc
	loadDiff       DiffFunc
	loadProject    ProjectLoadFunc
	loading        bool
	projectLoading bool
	projectError   bool
	diffLoading    bool
	errorMessage   string
}

func NewFakeModel() Model {
	return NewModel(FakeMergeRequests())
}

func NewModel(items []mr.MergeRequest) Model {
	return NewModelWithProject(items, ProjectOptions{Path: "group/project", Section: SectionMergeRequests})
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
		gitlabProjects: options.Projects,
		projectList:    buildProjectList(options.Path, options.Recents, options.Projects),
		section:        options.Section,
		entityID:       options.EntityID,
		refresh:        options.Refresh,
		loadDiff:       options.LoadDiff,
		loadProject:    options.LoadProject,
	}
	if model.projectPath == "" {
		if len(model.projectList) > 0 {
			model.mode = ModeProjectSelect
		} else {
			model.mode = ModeProjectInput
			model.focus = FocusFilter
		}
	} else if model.section == SectionMergeRequests {
		model.selectEntity()
		model.mode = ModeDetail
	} else {
		model.mode = ModeSections
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
		return m.updateMouse(msg)
	case projectStartedMsg:
		m.projectPath = msg.path
		m.mode = ModeDetail
		m.focus = FocusList
		m.loading = true
		m.projectLoading = true
		m.projectError = false
		m.items = nil
		m.errorMessage = ""
		return m, nil
	case projectFinishedMsg:
		m.loading = false
		m.projectLoading = false
		if msg.err != nil {
			m.projectError = true
			m.items = nil
			m.errorMessage = msg.err.Error()
			return m, nil
		}
		m.projectError = false
		m.projectPath = msg.path
		m.items = msg.data.Items
		m.refresh = msg.data.Refresh
		m.loadDiff = msg.data.LoadDiff
		m.selected = clampSelection(0, len(m.filtered()))
		m.selectEntity()
		m.listTop = 0
		m.rightTop = 0
		if m.section == SectionMergeRequests {
			m.mode = ModeDetail
		} else {
			m.mode = ModeSections
		}
		m.focus = FocusList
		return m, nil
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
	case diffStartedMsg:
		m.diffLoading = true
		m.errorMessage = ""
		return m, nil
	case diffFinishedMsg:
		m.diffLoading = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			return m, nil
		}
		m.setDiffRows(msg.iid, msg.rows)
		m.mode = ModeDiff
		m.focus = FocusDetail
		m.rightTop = 0
		return m, nil
	}

	return m, nil
}

func (m Model) updateKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.mode == ModeProjectSelect {
		switch msg.String() {
		case "up", "k":
			m.selected = clamp(m.selected-1, 0, len(m.projectList)-1)
		case "down", "j":
			m.selected = clamp(m.selected+1, 0, len(m.projectList)-1)
		case "enter":
			if len(m.projectList) > 0 {
				return m.openProjectCommand(m.projectList[m.selected])
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
				return m.openProjectCommand(strings.TrimSpace(m.projectInput))
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

	if m.mode == ModeSections {
		switch msg.String() {
		case "up", "k":
			m.sectionCursor = clamp(m.sectionCursor-1, 0, 2)
		case "down", "j":
			m.sectionCursor = clamp(m.sectionCursor+1, 0, 2)
		case "enter":
			if m.sectionCursor == 0 {
				m.section = SectionMergeRequests
				m.mode = ModeDetail
				m.focus = FocusList
			}
		case "esc", "backspace":
			m.mode = ModeProjectSelect
			m.focus = FocusList
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
	case "esc":
		if m.projectError || (m.projectPath != "" && len(m.items) == 0) {
			m.errorMessage = ""
			m.returnToProjectPicker()
			return m, nil
		}
		if m.mode == ModeDiff {
			m.mode = ModeDetail
			m.rightTop = 0
		}
	case "/":
		m.focus = FocusFilter
	case "r":
		if m.projectError && m.projectPath != "" {
			return m.openProjectCommand(m.projectPath)
		}
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
		if item, ok := m.selectedItem(); ok {
			return m.openDiffCommand(item)
		}
	case "backspace":
		if m.mode == ModeDiff {
			m.mode = ModeDetail
			m.rightTop = 0
		}
	}

	return m, nil
}

func (m *Model) returnToProjectPicker() {
	m.projectPath = ""
	if len(m.projectList) > 0 {
		m.mode = ModeProjectSelect
		m.focus = FocusList
		return
	}
	m.mode = ModeProjectInput
	m.focus = FocusFilter
}

func (m Model) openProjectCommand(path string) (Model, tea.Cmd) {
	m.projectPath = path
	m.mode = ModeDetail
	m.focus = FocusList
	m.selected = 0
	m.listTop = 0
	m.rightTop = 0
	if m.loadProject == nil {
		return m, nil
	}
	loadProject := m.loadProject
	return m, tea.Sequence(
		func() tea.Msg { return projectStartedMsg{path: path} },
		func() tea.Msg {
			data, err := loadProject(path)
			return projectFinishedMsg{path: path, data: data, err: err}
		},
	)
}

func (m Model) openDiffCommand(item mr.MergeRequest) (Model, tea.Cmd) {
	if m.loadDiff == nil || len(item.Diff) > 0 {
		m.mode = ModeDiff
		m.focus = FocusDetail
		m.rightTop = 0
		return m, nil
	}
	loadDiff := m.loadDiff
	return m, tea.Sequence(
		func() tea.Msg { return diffStartedMsg{} },
		func() tea.Msg {
			rows, err := loadDiff(item.IID)
			return diffFinishedMsg{iid: item.IID, rows: rows, err: err}
		},
	)
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

func (m Model) updateMouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	if m.mode == ModeProjectSelect {
		if msg.Button == tea.MouseButtonLeft && msg.Y >= 2 {
			idx := msg.Y - 2
			if idx >= 0 && idx < len(m.projectList) {
				m.selected = idx
				return m.openProjectCommand(m.projectList[idx])
			}
		}
		if msg.Button == tea.MouseButtonWheelUp {
			m.selected = clamp(m.selected-1, 0, len(m.projectList)-1)
		}
		if msg.Button == tea.MouseButtonWheelDown {
			m.selected = clamp(m.selected+1, 0, len(m.projectList)-1)
		}
		return m, nil
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
		} else if msg.X >= leftWidth {
			if item, ok := m.selectedItem(); ok {
				return m.openDiffCommand(item)
			}
		}
	}

	return m, nil
}

func (m *Model) selectEntity() {
	if m.entityID == "" {
		return
	}
	iid, err := strconv.Atoi(m.entityID)
	if err != nil {
		return
	}
	for i, item := range m.filtered() {
		if item.IID == iid {
			m.selected = i
			return
		}
	}
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
	if m.mode == ModeSections {
		return lipgloss.JoinHorizontal(lipgloss.Top, m.renderProjectList(), m.renderSections())
	}

	left := m.renderList()
	right := m.renderRight()
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func buildProjectList(opened string, recents []string, projects []string) []string {
	seen := map[string]bool{}
	list := []string{}
	candidates := []string{}
	if opened != "" {
		candidates = append(candidates, opened)
	}
	candidates = append(candidates, recents...)
	candidates = append(candidates, projects...)
	for _, project := range candidates {
		if project == "" || seen[project] {
			continue
		}
		seen[project] = true
		list = append(list, project)
	}
	return list
}

func (m Model) renderProjectList() string {
	width := m.leftWidth()
	style := paneStyle(width, max(8, m.height), false)
	lines := []string{"Projects", ""}
	for _, project := range m.projectList {
		prefix := "  "
		if project == m.projectPath {
			prefix = "> "
		}
		lines = append(lines, prefix+project)
	}
	if len(m.projectList) == 0 {
		lines = append(lines, "No projects")
	}
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderSections() string {
	width := max(20, m.width-m.leftWidth())
	style := paneStyle(width, max(8, m.height), true)
	sections := []string{"Merge Requests", "Issues", "Pipelines"}
	lines := []string{"Sections", ""}
	for i, section := range sections {
		prefix := "  "
		if i == m.sectionCursor {
			prefix = "> "
		}
		lines = append(lines, prefix+section)
	}
	lines = append(lines, "", "Enter: open  Esc: projects")
	return style.Render(strings.Join(lines, "\n"))
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

	lines := []string{"Projects", ""}
	for i, project := range m.projectList {
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
	if m.projectLoading {
		lines = append(lines, "Loading project…")
	} else if m.loading {
		lines = append(lines, "Refreshing…")
	}
	if m.errorMessage != "" {
		lines = append(lines, "Error: "+m.errorMessage)
		if m.projectError {
			lines = append(lines, "Esc: choose project")
		}
	}
	if m.diffLoading {
		lines = append(lines, "Loading diff…")
	}
	items := m.filtered()
	if len(items) == 0 {
		lines = append(lines, "No opened MRs")
		if len(m.items) == 0 && m.projectPath != "" {
			lines = append(lines, "r refresh  Esc: choose project")
		}
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
	lines := []string{
		fmt.Sprintf("!%d %s", item.IID, item.Title),
		"",
		"Author: " + formatAuthor(item),
		"Branches: " + item.SourceBranch + " → " + item.TargetBranch,
		"State: " + item.State,
		"Pipeline: " + pipelineIcon(item.Pipeline) + " " + item.Pipeline,
		"Approvals: " + item.Approvals,
	}
	if item.WebURL != "" {
		lines = append(lines, "URL: "+item.WebURL)
	}
	lines = append(lines, "", item.Description, "", "Enter/click right pane: open diff")
	return style.Render(strings.Join(lines, "\n"))
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

func (m Model) filtered() []mr.MergeRequest {
	return mr.Filter(m.items, m.query)
}

func (m Model) selectedItem() (mr.MergeRequest, bool) {
	items := m.filtered()
	if len(items) == 0 {
		return mr.MergeRequest{}, false
	}
	return items[clampSelection(m.selected, len(items))], true
}

func (m *Model) setDiffRows(iid int, rows []mr.DiffRow) {
	for i := range m.items {
		if m.items[i].IID == iid {
			m.items[i].Diff = rows
			return
		}
	}
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
