package mr

import "testing"

func TestFilterMatchesTitleAndBranches(t *testing.T) {
	t.Parallel()

	items := []MergeRequest{
		{Title: "Add config", SourceBranch: "go/config", TargetBranch: "main", Author: "alice"},
		{Title: "Render diff", SourceBranch: "go/diff", TargetBranch: "main", Author: "bob"},
	}

	filtered := Filter(items, "diff")

	if len(filtered) != 1 {
		t.Fatalf("expected 1 result, got %d", len(filtered))
	}

	if filtered[0].Title != "Render diff" {
		t.Fatalf("unexpected result: %+v", filtered[0])
	}
}

func TestFilterReturnsAllForEmptyQuery(t *testing.T) {
	t.Parallel()

	items := []MergeRequest{{Title: "A"}, {Title: "B"}}

	filtered := Filter(items, "   ")

	if len(filtered) != len(items) {
		t.Fatalf("expected all items, got %d", len(filtered))
	}
}
