//nolint:mnd // Fake data uses distinct fixture IDs.
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
		},
	}
}
