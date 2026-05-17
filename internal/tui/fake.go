package tui

import "github.com/aybykovskii/gitlab-tui/internal/mr"

func FakeMergeRequests() []mr.MergeRequest {
	return []mr.MergeRequest{
		{
			IID:          42,
			Title:        "Port TUI shell to Bubble Tea",
			Author:       "alice",
			SourceBranch: "go/tui-shell",
			TargetBranch: "main",
			State:        "opened",
			Pipeline:     "success",
			Approvals:    "1/2",
			Description:  "Tracer bullet for the Go port using fake data.",
			Diff: []mr.DiffRow{
				{OldLine: 1, NewLine: 1, OldText: "import React from 'react'", NewText: "package tui"},
				{OldLine: 2, NewLine: 2, OldText: "export function App() {", NewText: "type Model struct {"},
				{OldLine: 3, NewLine: 3, OldText: "  return <Box />", NewText: "\tselected int"},
				{OldLine: 4, NewLine: 4, OldText: "}", NewText: "}"},
			},
		},
		{
			IID:          43,
			Title:        "Add YAML config init command",
			Author:       "bob",
			SourceBranch: "go/config",
			TargetBranch: "main",
			State:        "opened",
			Pipeline:     "running",
			Approvals:    "0/1",
			Description:  "Config bootstrap for the Go binary.",
			Diff: []mr.DiffRow{
				{OldLine: 10, NewLine: 10, OldText: "config.json", NewText: "config.yaml"},
				{OldLine: 11, NewLine: 11, OldText: "token: secret", NewText: "token_env: GITLAB_TOKEN"},
			},
		},
		{
			IID:          44,
			Title:        "Render side-by-side diff rows",
			Author:       "carol",
			SourceBranch: "go/diff-view",
			TargetBranch: "main",
			State:        "opened",
			Pipeline:     "pending",
			Approvals:    "0/2",
			Description:  "Read-only side-by-side diff rendering for MVP.",
			Diff: []mr.DiffRow{
				{OldLine: 20, NewLine: 20, OldText: "left pane", NewText: "left pane | right pane"},
				{OldLine: 21, NewLine: 21, OldText: "", NewText: "mouse wheel scroll"},
			},
		},
	}
}
