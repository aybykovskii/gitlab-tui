package app

import (
	"fmt"

	"github.com/aybykovskii/gitlab-tui/internal/tui"
)

type CLIIntent struct {
	ProjectOverride string
	Section         tui.Section
	EntityID        string
}

func ParseCLI(args []string) (CLIIntent, error) {
	intent := CLIIntent{}

	remaining := args
	if len(remaining) > 0 && remaining[0] == "--project" {
		if len(remaining) < 2 {
			return CLIIntent{}, fmt.Errorf("--project requires a value")
		}

		intent.ProjectOverride = remaining[1]
		remaining = remaining[2:]
	}

	if len(remaining) == 0 {
		return intent, nil
	}

	section, ok := parseSection(remaining[0])
	if !ok {
		return CLIIntent{}, fmt.Errorf("unknown command: %s", remaining[0])
	}

	intent.Section = section
	if len(remaining) > 1 {
		intent.EntityID = remaining[1]
	}

	if len(remaining) > 2 {
		return CLIIntent{}, fmt.Errorf("too many arguments")
	}

	return intent, nil
}

func parseSection(value string) (tui.Section, bool) {
	switch tui.Section(value) {
	case tui.SectionMergeRequests:
		return tui.SectionMergeRequests, true
	case tui.SectionIssues:
		return tui.SectionIssues, true
	case tui.SectionPipelines:
		return tui.SectionPipelines, true
	default:
		return "", false
	}
}
