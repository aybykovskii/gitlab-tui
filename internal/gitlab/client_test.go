package gitlab

import (
	"context"
	"errors"
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

type fakeProjects struct {
	limit      int64
	page       int64
	membership bool
	orderBy    string
	projects   []*glab.Project
	err        error
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

func (f *fakeProjects) ListProjects(opt *glab.ListProjectsOptions, options ...glab.RequestOptionFunc) ([]*glab.Project, *glab.Response, error) {
	if opt != nil {
		f.limit = opt.PerPage
		f.page = opt.Page
		if opt.Membership != nil {
			f.membership = *opt.Membership
		}
		if opt.OrderBy != nil {
			f.orderBy = *opt.OrderBy
		}
	}
	return f.projects, &glab.Response{}, f.err
}

func TestListProjectsReturnsProjectPaths(t *testing.T) {
	projects := &fakeProjects{projects: []*glab.Project{
		{PathWithNamespace: "group/new"},
		{PathWithNamespace: "team/old"},
	}}
	client := NewClientWithProjects(projects)

	paths, err := client.ListProjects(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(paths) != 2 || paths[0] != "group/new" || paths[1] != "team/old" {
		t.Fatalf("unexpected project paths: %+v", paths)
	}
	if projects.limit != 5 || projects.page != 1 || !projects.membership || projects.orderBy != "" {
		t.Fatalf("unexpected list options: limit=%d page=%d membership=%t orderBy=%q", projects.limit, projects.page, projects.membership, projects.orderBy)
	}
}

func TestListProjectsReturnsEmptyList(t *testing.T) {
	client := NewClientWithProjects(&fakeProjects{})

	paths, err := client.ListProjects(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected empty project paths, got %+v", paths)
	}
}

func TestListProjectsReturnsAPIError(t *testing.T) {
	apiErr := errors.New("api failed")
	client := NewClientWithProjects(&fakeProjects{err: apiErr})

	_, err := client.ListProjects(context.Background(), 10)

	if !errors.Is(err, apiErr) {
		t.Fatalf("expected API error, got %v", err)
	}
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
		Diff:        "@@ -10,3 +10,4 @@\n context\n-old\n+new\n+added\n",
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

// --- #67: Labels, Draft, Reviewers, Assignees ---

type fakeLabels struct {
	labels []*glab.Label
	err    error
}

func (f *fakeLabels) ListLabels(pid any, opt *glab.ListLabelsOptions, options ...glab.RequestOptionFunc) ([]*glab.Label, *glab.Response, error) {
	return f.labels, &glab.Response{}, f.err
}

type fakeMergeRequestEdit struct {
	lastIID  int64
	lastOpts *glab.UpdateMergeRequestOptions
	err      error
}

func (f *fakeMergeRequestEdit) UpdateMergeRequest(pid any, mergeRequest int64, opt *glab.UpdateMergeRequestOptions, options ...glab.RequestOptionFunc) (*glab.MergeRequest, *glab.Response, error) {
	f.lastIID = mergeRequest
	f.lastOpts = opt
	return &glab.MergeRequest{}, &glab.Response{}, f.err
}

func TestMapMergeRequestFillsLabelsAndDraft(t *testing.T) {
	item := MapMergeRequest(&glab.BasicMergeRequest{
		IID:   10,
		Title: "Draft: My MR",
		Draft: true,
		Labels: glab.Labels{"backend", "performance"},
		Assignees: []*glab.BasicUser{
			{Name: "Alice", Username: "alice"},
			{Name: "", Username: "bob"},
		},
		Reviewers: []*glab.BasicUser{
			{Name: "Carol", Username: "carol"},
		},
	})

	if !item.Draft {
		t.Fatal("expected Draft=true")
	}
	if len(item.Labels) != 2 || item.Labels[0] != "backend" || item.Labels[1] != "performance" {
		t.Fatalf("expected labels [backend performance], got %v", item.Labels)
	}
	if len(item.Assignees) != 2 || item.Assignees[0] != "Alice" || item.Assignees[1] != "bob" {
		t.Fatalf("expected assignees [Alice bob], got %v", item.Assignees)
	}
	if len(item.Reviewers) != 1 || item.Reviewers[0] != "Carol" {
		t.Fatalf("expected reviewers [Carol], got %v", item.Reviewers)
	}
}

func TestListProjectLabelsReturnsMappedLabels(t *testing.T) {
	fake := &fakeLabels{labels: []*glab.Label{
		{Name: "backend", Color: "#e11d48"},
		{Name: "bug", Color: "#dc2626"},
	}}
	client := NewClientWithLabels(fake)

	labels, err := client.ListProjectLabels(context.Background(), "group/project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(labels))
	}
	if labels[0].Name != "backend" || labels[0].Color != "#e11d48" {
		t.Fatalf("unexpected first label: %+v", labels[0])
	}
	if labels[1].Name != "bug" || labels[1].Color != "#dc2626" {
		t.Fatalf("unexpected second label: %+v", labels[1])
	}
}

func TestListProjectLabelsReturnsAPIError(t *testing.T) {
	apiErr := errors.New("labels api failed")
	client := NewClientWithLabels(&fakeLabels{err: apiErr})

	_, err := client.ListProjectLabels(context.Background(), "group/project")
	if !errors.Is(err, apiErr) {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestUpdateMRLabelsSetsLabelsOnMR(t *testing.T) {
	fake := &fakeMergeRequestEdit{}
	client := NewClientWithMergeRequestEdit(fake)

	err := client.UpdateMRLabels(context.Background(), "group/project", 42, []string{"backend", "bug"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.lastIID != 42 {
		t.Fatalf("expected iid=42, got %d", fake.lastIID)
	}
	if fake.lastOpts == nil || fake.lastOpts.Labels == nil {
		t.Fatal("expected UpdateMergeRequest to be called with labels")
	}
	got := []string(*fake.lastOpts.Labels)
	if len(got) != 2 || got[0] != "backend" || got[1] != "bug" {
		t.Fatalf("unexpected labels: %v", got)
	}
}

func TestUpdateMRLabelsReturnsAPIError(t *testing.T) {
	apiErr := errors.New("update failed")
	client := NewClientWithMergeRequestEdit(&fakeMergeRequestEdit{err: apiErr})

	err := client.UpdateMRLabels(context.Background(), "group/project", 42, []string{"backend"})
	if !errors.Is(err, apiErr) {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestToggleDraftMRAddsDraftPrefix(t *testing.T) {
	fake := &fakeMergeRequestEdit{}
	client := NewClientWithMergeRequestEdit(fake)

	err := client.ToggleDraftMR(context.Background(), "group/project", 42, "My MR", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.lastOpts == nil || fake.lastOpts.Title == nil {
		t.Fatal("expected UpdateMergeRequest to be called with title")
	}
	if *fake.lastOpts.Title != "Draft: My MR" {
		t.Fatalf("expected title 'Draft: My MR', got %q", *fake.lastOpts.Title)
	}
}

func TestToggleDraftMRRemovesDraftPrefix(t *testing.T) {
	fake := &fakeMergeRequestEdit{}
	client := NewClientWithMergeRequestEdit(fake)

	err := client.ToggleDraftMR(context.Background(), "group/project", 42, "Draft: My MR", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.lastOpts == nil || fake.lastOpts.Title == nil {
		t.Fatal("expected UpdateMergeRequest to be called with title")
	}
	if *fake.lastOpts.Title != "My MR" {
		t.Fatalf("expected title 'My MR' after removing draft prefix, got %q", *fake.lastOpts.Title)
	}
}

func TestToggleDraftMRDoesNotDoublePrefixAlreadyDraft(t *testing.T) {
	fake := &fakeMergeRequestEdit{}
	client := NewClientWithMergeRequestEdit(fake)

	err := client.ToggleDraftMR(context.Background(), "group/project", 42, "Draft: My MR", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *fake.lastOpts.Title != "Draft: My MR" {
		t.Fatalf("expected no double prefix, got %q", *fake.lastOpts.Title)
	}
}

func TestToggleDraftMRReturnsAPIError(t *testing.T) {
	apiErr := errors.New("edit failed")
	client := NewClientWithMergeRequestEdit(&fakeMergeRequestEdit{err: apiErr})

	err := client.ToggleDraftMR(context.Background(), "group/project", 42, "My MR", true)
	if !errors.Is(err, apiErr) {
		t.Fatalf("expected API error, got %v", err)
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
