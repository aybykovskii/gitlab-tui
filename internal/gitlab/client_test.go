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

func TestMapMergeRequestUsesDetailedMergeStatusAsPipeline(t *testing.T) {
	item := MapMergeRequest(&glab.BasicMergeRequest{IID: 3, Title: "MR", DetailedMergeStatus: "mergeable"})
	if item.Pipeline != "mergeable" {
		t.Fatalf("expected mergeable pipeline, got %q", item.Pipeline)
	}
}
