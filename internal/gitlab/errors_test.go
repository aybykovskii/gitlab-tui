package gitlab

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMissingServiceReturnsSentinelError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		call func(Client) error
	}{
		{
			name: "projects",
			call: func(client Client) error {
				_, err := client.ListProjects(context.Background(), 5)
				return err
			},
		},
		{
			name: "merge requests",
			call: func(client Client) error {
				_, err := client.OpenMergeRequests(context.Background(), "group/project")
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.call(Client{})

			require.Error(t, err)
			assert.ErrorIs(t, err, ErrClientNotConfigured)
		})
	}
}

func TestUnexpectedHTTPResponseReturnsTypedError(t *testing.T) {
	t.Parallel()

	server := gitlabTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/projects": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
			_, err := w.Write([]byte("short and stout"))
			require.NoError(t, err)
		},
	})

	client := newHTTPTestClient(t, server.URL)
	_, err := client.ListProjects(context.Background(), 5)

	require.Error(t, err)

	var responseErr *ErrUnexpectedResponse

	require.True(t, errors.As(err, &responseErr))
	assert.Equal(t, http.StatusTeapot, responseErr.StatusCode)
	assert.Equal(t, "418 I'm a teapot", responseErr.Status)
	assert.Contains(t, responseErr.Body, "short and stout")
}
