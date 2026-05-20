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
			assert.NotContains(t, payload, "old_line\":0")
			_, err = w.Write([]byte(`{}`))
			require.NoError(t, err)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	err := client.CreateMergeRequestDiscussion(context.Background(), "group/project", 42, "Check this", &mr.DiffPosition{NewPath: "main.go", OldPath: "main.go", NewLine: 7})

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
