package gitlab

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/mr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	_testdataPathProjects = "./testdata/projects.json"
	_testdataPathIssues   = "./testdata/issues.json"
)

func TestHTTPListProjectsReturnsProjectPaths(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "true", r.URL.Query().Get("membership"))
			assert.Equal(t, "5", r.URL.Query().Get("per_page"))
			assert.Equal(t, "1", r.URL.Query().Get("page"))
			writeFixture(t, w, _testdataPathProjects)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	paths, err := client.ListProjects(context.Background(), 5)

	require.NoError(t, err)
	assert.Equal(t, []string{"group/new", "team/old"}, paths)
}

func TestHTTPListProjectIssuesPassesStateAndMapsItems(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects/group%2Fproject/issues": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "opened", r.URL.Query().Get("state"))
			assert.Equal(t, "api", r.URL.Query().Get("search"))
			assert.Equal(t, "50", r.URL.Query().Get("per_page"))
			assert.Equal(t, "1", r.URL.Query().Get("page"))
			writeFixture(t, w, _testdataPathIssues)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	items, err := client.ListProjectIssues(context.Background(), "group/project", "opened", "api")

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, 79, items[0].IID)
	assert.Equal(t, "Issues API", items[0].Title)
	assert.Equal(t, "Alice", items[0].Author)
	assert.Equal(t, 2, items[0].CommentCount)
}

func TestHTTPResolveMergeRequestDiscussionUsesDiscussionEndpoint(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects/group%2Fproject/merge_requests/42/discussions/disc-1": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.Contains(t, string(body), "resolved")
			assert.NotContains(t, r.URL.Path, "/notes/0")
			_, err = w.Write([]byte(`{}`))
			require.NoError(t, err)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	err := client.ResolveMergeRequestDiscussion(context.Background(), "group/project", 42, "disc-1", true)

	require.NoError(t, err)
}

func TestHTTPApproveMergeRequestUsesApproveEndpoint(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects/group%2Fproject/merge_requests/42/approve": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			_, err := w.Write([]byte(`{}`))
			require.NoError(t, err)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	err := client.ApproveMergeRequest(context.Background(), "group/project", 42)

	require.NoError(t, err)
}

func TestHTTPAcceptMergeRequestUsesMergeEndpoint(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects/group%2Fproject/merge_requests/42/merge": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			_, err := w.Write([]byte(`{}`))
			require.NoError(t, err)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	err := client.AcceptMergeRequest(context.Background(), "group/project", 42)

	require.NoError(t, err)
}

func TestHTTPUpdateMergeRequestSendsTitleAndDescription(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects/group%2Fproject/merge_requests/42": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			payload := string(body)
			assert.Contains(t, payload, "title")
			assert.Contains(t, payload, "New")
			assert.Contains(t, payload, "description")
			assert.Contains(t, payload, "Details")
			_, err = w.Write([]byte(`{}`))
			require.NoError(t, err)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	err := client.UpdateMergeRequest(context.Background(), "group/project", 42, "New", "Details")

	require.NoError(t, err)
}

func TestHTTPAddMergeRequestDiscussionNoteUsesReplyEndpoint(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects/group%2Fproject/merge_requests/42/discussions/disc-1/notes": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.Contains(t, string(body), "Reply")
			_, err = w.Write([]byte(`{}`))
			require.NoError(t, err)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	err := client.AddMergeRequestDiscussionNote(context.Background(), "group/project", 42, "disc-1", "Reply")

	require.NoError(t, err)
}

func TestHTTPCreateMergeRequestDiscussionSendsInlinePosition(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects/group%2Fproject/merge_requests/42/discussions": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			payload := string(body)
			assert.True(t, strings.Contains(payload, "new_path") || strings.Contains(payload, "new_path%"), payload)
			assert.Contains(t, payload, "main.go")
			assert.Contains(t, payload, "new_line")
			assert.Contains(t, payload, "base_sha")
			assert.Contains(t, payload, "head_sha")
			assert.Contains(t, payload, "start_sha")
			assert.NotContains(t, payload, "old_line\":0")
			_, err = w.Write([]byte(`{}`))
			require.NoError(t, err)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	err := client.CreateMergeRequestDiscussion(context.Background(), "group/project", 42, "Check this", &mr.DiffPosition{BaseSHA: "base", HeadSHA: "head", StartSHA: "start", NewPath: "main.go", OldPath: "main.go", NewLine: 7})

	require.NoError(t, err)
}

func TestHTTPCreateDraftNoteSendsInlinePositionAndReturnsID(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects/group%2Fproject/merge_requests/42/draft_notes": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			payload := string(body)
			assert.Contains(t, payload, "Draft")
			assert.Contains(t, payload, "base_sha")
			assert.Contains(t, payload, "head_sha")
			assert.Contains(t, payload, "start_sha")
			_, err = w.Write([]byte(`{"id":123}`))
			require.NoError(t, err)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	id, err := client.CreateDraftNote(context.Background(), "group/project", 42, "", "Draft", &mr.DiffPosition{BaseSHA: "base", HeadSHA: "head", StartSHA: "start", NewPath: "main.go", OldPath: "main.go", NewLine: 7})

	require.NoError(t, err)
	assert.Equal(t, 123, id)
}

func TestHTTPBulkPublishDraftNotesUsesBulkEndpoint(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects/group%2Fproject/merge_requests/42/draft_notes/bulk_publish": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			_, err := w.Write([]byte(`{}`))
			require.NoError(t, err)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	err := client.BulkPublishDraftNotes(context.Background(), "group/project", 42, []int{123})

	require.NoError(t, err)
}

func TestHTTPDeleteAllDraftNotesListsThenDeletes(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects/group%2Fproject/merge_requests/42/draft_notes": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			_, err := w.Write([]byte(`[{"id":10},{"id":11}]`))
			require.NoError(t, err)
		},
		"/api/v4/projects/group%2Fproject/merge_requests/42/draft_notes/10": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			w.WriteHeader(http.StatusNoContent)
		},
		"/api/v4/projects/group%2Fproject/merge_requests/42/draft_notes/11": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			w.WriteHeader(http.StatusNoContent)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	err := client.DeleteAllDraftNotes(context.Background(), "group/project", 42)

	require.NoError(t, err)
}

func gitlabTestServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	for pattern, handler := range handlers {
		mux.HandleFunc(pattern, handler)
	}

	return httptest.NewServer(mux)
}

func newHTTPTestClient(t *testing.T, host string) Client {
	t.Helper()

	client, err := NewClient(config.Account{Host: host, TokenEnv: "GITLAB_TOKEN"}, []string{"GITLAB_TOKEN=test-token"})
	require.NoError(t, err)

	return client
}

func writeFixture(t *testing.T, w http.ResponseWriter, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	require.NoError(t, err)
}
