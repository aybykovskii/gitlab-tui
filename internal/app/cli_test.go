package app

import (
	"testing"

	"github.com/aybykovskii/gitlab-tui/internal/tui"
)

func TestParseCLIProjectOverride(t *testing.T) {
	intent, err := ParseCLI([]string{"--project", "group/project", "pipeline"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if intent.ProjectOverride != "group/project" {
		t.Fatalf("expected project override, got %q", intent.ProjectOverride)
	}
	if intent.Section != tui.SectionPipelines {
		t.Fatalf("expected pipeline section, got %q", intent.Section)
	}
}

func TestParseCLIRejectsPositionalProjectPath(t *testing.T) {
	_, err := ParseCLI([]string{"group/project"})
	if err == nil {
		t.Fatal("expected positional project path to be rejected")
	}
}

func TestParseCLISectionEntityIntent(t *testing.T) {
	intent, err := ParseCLI([]string{"mr", "123"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if intent.Section != tui.SectionMergeRequests {
		t.Fatalf("expected MR section, got %q", intent.Section)
	}
	if intent.EntityID != "123" {
		t.Fatalf("expected entity id 123, got %q", intent.EntityID)
	}
	if intent.ProjectOverride != "" {
		t.Fatalf("expected no project override, got %q", intent.ProjectOverride)
	}
}

func TestParseCLIProjectWithoutValueErrors(t *testing.T) {
	_, err := ParseCLI([]string{"--project"})
	if err == nil {
		t.Fatal("expected error for --project without value")
	}
	if err.Error() != "--project requires a value" {
		t.Fatalf("expected informative error, got %q", err.Error())
	}
}
