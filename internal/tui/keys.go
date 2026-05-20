package tui

import "github.com/charmbracelet/bubbles/key"

type GlobalKeys struct {
	Quit         key.Binding
	Back         key.Binding
	ToggleKeyBar key.Binding
}

type ProjectListKeys struct {
	Up     key.Binding
	Down   key.Binding
	Open   key.Binding
	Filter key.Binding
	Input  key.Binding
	Retry  key.Binding
}

type SectionsKeys struct {
	Up   key.Binding
	Down key.Binding
	Open key.Binding
}

type EntityListKeys struct {
	Up      key.Binding
	Down    key.Binding
	Open    key.Binding
	Filter  key.Binding
	Refresh key.Binding
}

type MRDetailKeys struct {
	Approve     key.Binding
	Merge       key.Binding
	Edit        key.Binding
	OpenURL     key.Binding
	Comment     key.Binding
	NextTab     key.Binding
	ToggleDraft key.Binding
	LabelSelect key.Binding
}

type IssueDetailKeys struct {
	NextTab key.Binding
}

type DiffViewKeys struct {
	Up      key.Binding
	Down    key.Binding
	Comment key.Binding
	Range   key.Binding
	Publish key.Binding
}

type FileDiffKeys struct {
	PrevFile          key.Binding
	NextFile          key.Binding
	Up                key.Binding
	Down              key.Binding
	Comment           key.Binding
	Reply             key.Binding
	ToggleThreadPanel key.Binding
	PrevThread        key.Binding
	NextThread        key.Binding
	ScrollLeft        key.Binding
	ScrollRight       key.Binding
}

func newGlobalKeys() GlobalKeys {
	return GlobalKeys{
		Quit:         key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Back:         key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "back")),
		ToggleKeyBar: key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "keys")),
	}
}

func newProjectListKeys() ProjectListKeys {
	return ProjectListKeys{
		Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Open:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "open")),
		Filter: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Input:  key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "manual")),
		Retry:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
	}
}

func (k ProjectListKeys) LocalKeys() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Open, k.Filter, k.Input, k.Retry}
}

func newSectionsKeys() SectionsKeys {
	return SectionsKeys{
		Up:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Open: key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "open")),
	}
}

func (k SectionsKeys) LocalKeys() []key.Binding { return []key.Binding{k.Up, k.Down, k.Open} }

func newEntityListKeys() EntityListKeys {
	return EntityListKeys{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Open:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "open")),
		Filter:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	}
}

func (k EntityListKeys) LocalKeys() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Open, k.Filter, k.Refresh}
}

func newMRDetailKeys() MRDetailKeys {
	return MRDetailKeys{
		Approve:     key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "approve")),
		Merge:       key.NewBinding(key.WithKeys("M"), key.WithHelp("M", "merge")),
		Edit:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		OpenURL:     key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open")),
		Comment:     key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "comment")),
		NextTab:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "next tab")),
		ToggleDraft: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "toggle draft")),
		LabelSelect: key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "labels")),
	}
}

func (k MRDetailKeys) LocalKeys() []key.Binding {
	return []key.Binding{k.Approve, k.Merge, k.Edit, k.OpenURL, k.Comment, k.NextTab, k.ToggleDraft, k.LabelSelect}
}

func newIssueDetailKeys() IssueDetailKeys {
	return IssueDetailKeys{NextTab: key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "next tab"))}
}

func (k IssueDetailKeys) LocalKeys() []key.Binding { return []key.Binding{k.NextTab} }

func newDiffViewKeys() DiffViewKeys {
	return DiffViewKeys{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Comment: key.NewBinding(key.WithKeys("i", "c"), key.WithHelp("i/c", "comment")),
		Range:   key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "range")),
		Publish: key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "publish")),
	}
}

func (k DiffViewKeys) LocalKeys() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Comment, k.Range, k.Publish}
}

func newFileDiffKeys() FileDiffKeys {
	return FileDiffKeys{
		PrevFile:          key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev file")),
		NextFile:          key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next file")),
		Up:                key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:              key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Comment:           key.NewBinding(key.WithKeys("i", "c"), key.WithHelp("i/c", "comment")),
		Reply:             key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reply")),
		ToggleThreadPanel: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "thread")),
		PrevThread:        key.NewBinding(key.WithKeys("["), key.WithHelp("[", "prev thread")),
		NextThread:        key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "next thread")),
		ScrollLeft:        key.NewBinding(key.WithKeys("shift+left"), key.WithHelp("Shift+←", "scroll left")),
		ScrollRight:       key.NewBinding(key.WithKeys("shift+right"), key.WithHelp("Shift+→", "scroll right")),
	}
}

func (k FileDiffKeys) LocalKeys() []key.Binding {
	return []key.Binding{k.PrevFile, k.NextFile, k.Up, k.Down, k.Comment, k.Reply, k.ToggleThreadPanel, k.PrevThread, k.NextThread, k.ScrollLeft, k.ScrollRight}
}
