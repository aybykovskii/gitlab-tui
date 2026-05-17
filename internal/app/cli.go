package app

import "fmt"

type Section string

const (
	SectionMergeRequests Section = "mr"
	SectionIssues        Section = "issue"
	SectionPipelines     Section = "pipeline"
)

type CLIIntent struct {
	ProjectOverride string
	Section         Section
	EntityID        string
}

func ParseCLI(args []string) (CLIIntent, error) {
	intent := CLIIntent{}
	remaining := args
	if len(remaining) >= 2 && remaining[0] == "--project" {
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

func parseSection(value string) (Section, bool) {
	switch value {
	case string(SectionMergeRequests):
		return SectionMergeRequests, true
	case string(SectionIssues):
		return SectionIssues, true
	case string(SectionPipelines):
		return SectionPipelines, true
	default:
		return "", false
	}
}
