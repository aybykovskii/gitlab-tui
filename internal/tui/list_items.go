//nolint:mnd,errcheck,wsl // List delegates use fixed sizes and write to io.Writer without error handling.
package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aybykovskii/gitlab-tui/internal/issue"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

// --- MR list ---

type mrListItem struct{ mr.MergeRequest }

func (i mrListItem) Title() string {
	if i.Pipeline == "" {
		return fmt.Sprintf("!%d %s", i.IID, i.MergeRequest.Title)
	}

	return fmt.Sprintf("%s !%d %s", pipelineIcon(i.Pipeline), i.IID, i.MergeRequest.Title)
}
func (i mrListItem) Description() string { return fmt.Sprintf("%s %s → %s", i.Author, i.SourceBranch, i.TargetBranch) }
func (i mrListItem) FilterValue() string {
	return i.MergeRequest.Title + " " + i.Author + " " + i.SourceBranch + " " + i.TargetBranch
}

// --- Issue list ---

type issueListItem struct{ issue.Issue }

func (i issueListItem) Title() string       { return fmt.Sprintf("#%d %s", i.IID, i.Issue.Title) }
func (i issueListItem) Description() string { return formatIssueMeta(i.Issue) }
func (i issueListItem) FilterValue() string { return i.Issue.Title + " " + i.Author }

// --- Section list ---

type sectionListItem struct{ sectionDef }

func (i sectionListItem) FilterValue() string { return i.label }

type sectionListDelegate struct {
	styles list.DefaultItemStyles
}

func newSectionListDelegate() sectionListDelegate {
	return sectionListDelegate{styles: list.NewDefaultItemStyles()}
}

func (d sectionListDelegate) Height() int                             { return 1 }
func (d sectionListDelegate) Spacing() int                            { return 0 }
func (d sectionListDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d sectionListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	sec := item.(sectionListItem)
	label := sec.label
	if !sec.available {
		label += " (soon)"
	}

	switch {
	case index == m.Index():
		fmt.Fprint(w, d.styles.SelectedTitle.Render(label))
	case !sec.available:
		fmt.Fprint(w, d.styles.DimmedTitle.Render(label))
	default:
		fmt.Fprint(w, d.styles.NormalTitle.Render(label))
	}
}

// --- Project picker delegate ---
// projectListRow is defined in messages.go and implements list.Item via FilterValue.

type projectPickerDelegate struct {
	styles list.DefaultItemStyles
}

func newProjectPickerDelegate() projectPickerDelegate {
	return projectPickerDelegate{styles: list.NewDefaultItemStyles()}
}

func (d projectPickerDelegate) Height() int                             { return 1 }
func (d projectPickerDelegate) Spacing() int                            { return 0 }
func (d projectPickerDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d projectPickerDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	row := item.(projectListRow)
	if !row.selectable {
		fmt.Fprint(w, d.styles.DimmedTitle.Render(row.label))
		return
	}

	if index == m.Index() {
		fmt.Fprint(w, d.styles.SelectedTitle.Render(row.label))
	} else {
		fmt.Fprint(w, d.styles.NormalTitle.Render(row.label))
	}
}

// newFancyList creates a styled list.Model with title, status bar, help and pagination enabled.
// Built-in filtering is disabled because the TUI uses its own query-based filter.
func newFancyList(title string) list.Model {
	d := list.NewDefaultDelegate()
	l := list.New(nil, d, 80, 20)
	l.Title = title
	l.SetFilteringEnabled(false)

	return l
}

// newCompactFancyList creates a single-line fancylist with a title bar but without status/help chrome.
// Intended for compact lists like sections and project picker that use custom delegates.
func newCompactFancyList(title string, delegate list.ItemDelegate) list.Model {
	l := list.New(nil, delegate, 80, 20)
	l.Title = title
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)

	return l
}
