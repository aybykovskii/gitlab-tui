package diff

import "testing"

func TestParseUnifiedDiffRows(t *testing.T) {
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
