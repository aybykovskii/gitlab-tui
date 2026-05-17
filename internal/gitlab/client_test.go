package gitlab

import (
	"context"
	"testing"

	glab "gitlab.com/gitlab-org/api/client-go"
)

type fakeMergeRequests struct {
	calls int
	pages [][]*glab.BasicMergeRequest
}

type fakeApprovals struct {
	configs map[int64]*glab.MergeRequestApprovals
}

func (f *fakeMergeRequests) ListProjectMergeRequests(pid any, opt *glab.ListProjectMergeRequestsOptions, options ...glab.RequestOptionFunc) ([]*glab.BasicMergeRequest, *glab.Response, error) {
	f.calls++
	page := int(opt.Page)
	if page == 0 {
		page = 1
	}
	response := &glab.Response{}
	if page < len(f.pages) {
		response.NextPage = int64(page + 1)
	}
	return f.pages[page-1], response, nil
}

func (f *fakeMergeRequests) ListMergeRequestDiffs(pid any, mergeRequest int64, opt *glab.ListMergeRequestDiffsOptions, options ...glab.RequestOptionFunc) ([]*glab.MergeRequestDiff, *glab.Response, error) {
	return []*glab.MergeRequestDiff{{Diff: "@@ -1 +1 @@\n-old\n+new"}}, &glab.Response{}, nil
}

func (f fakeApprovals) GetConfiguration(pid any, mergeRequest int64, options ...glab.RequestOptionFunc) (*glab.MergeRequestApprovals, *glab.Response, error) {
	if f.configs == nil {
		return nil, &glab.Response{}, nil
	}
	return f.configs[mergeRequest], &glab.Response{}, nil
}

func TestOpenMergeRequestsMapsAllPages(t *testing.T) {
	fake := &fakeMergeRequests{pages: [][]*glab.BasicMergeRequest{
		{{IID: 1, Title: "First", State: "opened", SourceBranch: "feature/a", TargetBranch: "main", Author: &glab.BasicUser{Username: "alice"}}},
		{{IID: 2, Title: "Second", State: "opened", SourceBranch: "feature/b", TargetBranch: "main", Author: &glab.BasicUser{Name: "Bob"}}},
	}}
	client := NewClientWithMergeRequests(fake)

	items, err := client.OpenMergeRequests(context.Background(), "group/project")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].IID != 1 || items[0].Author != "alice" {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[1].IID != 2 || items[1].Author != "Bob" {
		t.Fatalf("unexpected second item: %+v", items[1])
	}
	if fake.calls != 2 {
		t.Fatalf("expected 2 calls, got %d", fake.calls)
	}
}

func TestMergeRequestDiffParsesRows(t *testing.T) {
	client := NewClientWithMergeRequests(&fakeMergeRequests{})

	rows, err := client.MergeRequestDiff(context.Background(), "group/project", 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].OldLine != 1 || rows[0].OldText != "old" {
		t.Fatalf("unexpected removed row: %+v", rows[0])
	}
	if rows[1].NewLine != 1 || rows[1].NewText != "new" {
		t.Fatalf("unexpected added row: %+v", rows[1])
	}
}

func TestOpenMergeRequestsAddsApprovalCounts(t *testing.T) {
	client := NewClientWithServices(&fakeMergeRequests{pages: [][]*glab.BasicMergeRequest{{{IID: 3, Title: "MR"}}}}, fakeApprovals{
		configs: map[int64]*glab.MergeRequestApprovals{3: {ApprovalsRequired: 2, ApprovalsLeft: 1}},
	})

	items, err := client.OpenMergeRequests(context.Background(), "group/project")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if items[0].Approvals != "1/2" {
		t.Fatalf("expected approval counts, got %q", items[0].Approvals)
	}
}

func TestMapDiscussionMapsNotesAndResolution(t *testing.T) {
	resolved := true
	item := MapDiscussion(&glab.Discussion{
		ID: "abc123",
		Notes: []*glab.Note{
			{Author: glab.NoteAuthor{Name: "Alice", Username: "alice"}, Body: "Needs a fix", Resolved: false},
			{Author: glab.NoteAuthor{Name: "Bob", Username: "bob"}, Body: "Fixed", Resolved: true, Resolvable: true},
		},
	})
	_ = resolved
	if item.ID != "abc123" {
		t.Fatalf("expected ID abc123, got %q", item.ID)
	}
	if item.Resolved {
		t.Fatal("expected discussion to be unresolved (first note is unresolved)")
	}
	if len(item.Notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(item.Notes))
	}
	if item.Notes[0].Author != "Alice" || item.Notes[0].Body != "Needs a fix" {
		t.Fatalf("unexpected first note: %+v", item.Notes[0])
	}
}

func TestMapDiscussionExcludesSystemNotes(t *testing.T) {
	item := MapDiscussion(&glab.Discussion{
		ID: "sys1",
		Notes: []*glab.Note{
			{Author: glab.NoteAuthor{Name: "GitLab"}, Body: "changed milestone", System: true},
			{Author: glab.NoteAuthor{Name: "Alice"}, Body: "Real comment"},
		},
	})
	if len(item.Notes) != 1 {
		t.Fatalf("expected 1 non-system note, got %d", len(item.Notes))
	}
	if item.Notes[0].Author != "Alice" {
		t.Fatalf("expected Alice, got %q", item.Notes[0].Author)
	}
}

func TestMapChangedFileMapsPathMarkersAndLineCounts(t *testing.T) {
	item := MapChangedFile(&glab.MergeRequestDiff{
		NewPath:     "internal/tui/model.go",
		OldPath:     "internal/tui/model.go",
		NewFile:     false,
		DeletedFile: false,
		RenamedFile: false,
		Diff: "@@ -10,3 +10,4 @@\n context\n-old\n+new\n+added\n",
	})
	if item.Path != "internal/tui/model.go" {
		t.Fatalf("expected path, got %q", item.Path)
	}
	if item.IsNew || item.IsDeleted || item.IsRenamed {
		t.Fatalf("unexpected markers: %+v", item)
	}
	if item.AddedLines != 2 {
		t.Fatalf("expected 2 added lines, got %d", item.AddedLines)
	}
	if item.RemovedLines != 1 {
		t.Fatalf("expected 1 removed line, got %d", item.RemovedLines)
	}
}

func TestMapChangedFileMarksNewAndDeletedFiles(t *testing.T) {
	newFile := MapChangedFile(&glab.MergeRequestDiff{NewPath: "new.go", NewFile: true, Diff: "@@ -0,0 +1 @@\n+hello\n"})
	if !newFile.IsNew {
		t.Fatal("expected IsNew=true")
	}
	if newFile.AddedLines != 1 {
		t.Fatalf("expected 1 added line, got %d", newFile.AddedLines)
	}

	deleted := MapChangedFile(&glab.MergeRequestDiff{OldPath: "old.go", DeletedFile: true, Diff: "@@ -1 +0,0 @@\n-bye\n"})
	if !deleted.IsDeleted {
		t.Fatal("expected IsDeleted=true")
	}
	if deleted.Path != "old.go" {
		t.Fatalf("expected old.go path for deleted file, got %q", deleted.Path)
	}
}

func TestMapMergeRequestKeepsPreviousMRInfo(t *testing.T) {
	item := MapMergeRequest(&glab.BasicMergeRequest{
		IID:                 3,
		Title:               "MR",
		DetailedMergeStatus: "success",
		WebURL:              "https://gitlab.com/group/project/-/merge_requests/3",
		Author:              &glab.BasicUser{Name: "Alice Doe", Username: "alice"},
	})
	if item.Pipeline != "success" {
		t.Fatalf("expected success pipeline, got %q", item.Pipeline)
	}
	if item.Author != "Alice Doe" || item.AuthorUsername != "alice" {
		t.Fatalf("unexpected author: %+v", item)
	}
	if item.WebURL != "https://gitlab.com/group/project/-/merge_requests/3" {
		t.Fatalf("unexpected web URL: %q", item.WebURL)
	}
}
