package diff

import "testing"

func TestProjectDiscussionsAttachesToMatchingRow(t *testing.T) {
	t.Parallel()

	rows := []Row{
		{OldLine: 10, NewLine: 10, OldText: "ctx", NewText: "ctx"},
		{OldLine: 0, NewLine: 11, NewText: "new line"},
	}
	discussions := []Discussion{
		{ID: "d1", Notes: []Note{{Author: "alice", Body: "fix this"}}, Position: &Position{NewPath: "main.go", NewLine: 11}},
		{ID: "d2", Notes: []Note{{Author: "bob", Body: "LGTM"}}, Position: &Position{NewPath: "other.go", NewLine: 11}},
		{ID: "d3", Notes: []Note{{Author: "carol", Body: "general"}}, Position: nil},
	}

	annotated := ProjectDiscussions(rows, discussions, "main.go")

	if len(annotated) != 2 {
		t.Fatalf("expected 2 annotated rows, got %d", len(annotated))
	}

	if len(annotated[0].Discussions) != 0 {
		t.Fatalf("expected no discussions on row 0, got %d", len(annotated[0].Discussions))
	}

	if len(annotated[1].Discussions) != 1 {
		t.Fatalf("expected 1 discussion on row 1, got %d", len(annotated[1].Discussions))
	}

	if annotated[1].Discussions[0].ID != "d1" {
		t.Fatalf("expected discussion d1 on row 1, got %q", annotated[1].Discussions[0].ID)
	}
}

func TestParseUnifiedDiffRows(t *testing.T) {
	t.Parallel()

	rows := Parse(`@@ -10,3 +10,4 @@
 context
-old
+new
+added`)

	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}

	if rows[0].OldLine != 10 || rows[0].NewLine != 10 || rows[0].OldText != "context" || rows[0].NewText != "context" {
		t.Fatalf("unexpected context row: %+v", rows[0])
	}

	if rows[1].OldLine != 11 || rows[1].NewLine != 0 || rows[1].OldText != "old" {
		t.Fatalf("unexpected removed row: %+v", rows[1])
	}

	if rows[2].OldLine != 0 || rows[2].NewLine != 11 || rows[2].NewText != "new" {
		t.Fatalf("unexpected added row: %+v", rows[2])
	}

	if rows[3].NewLine != 12 || rows[3].NewText != "added" {
		t.Fatalf("unexpected second added row: %+v", rows[3])
	}
}
