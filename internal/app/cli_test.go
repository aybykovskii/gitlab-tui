package app

import "testing"

func TestParseCLIProjectOverride(t *testing.T) {
	intent, err := ParseCLI([]string{"--project", "group/project", "pipeline"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if intent.ProjectOverride != "group/project" {
		t.Fatalf("expected project override, got %q", intent.ProjectOverride)
	}
	if intent.Section != SectionPipelines {
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

	if intent.Section != SectionMergeRequests {
		t.Fatalf("expected MR section, got %q", intent.Section)
	}
	if intent.EntityID != "123" {
		t.Fatalf("expected entity id 123, got %q", intent.EntityID)
	}
	if intent.ProjectOverride != "" {
		t.Fatalf("expected no project override, got %q", intent.ProjectOverride)
	}
}
