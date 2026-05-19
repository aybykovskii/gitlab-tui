package app

import (
	"errors"
	"net/http"
	"testing"

	gitlabclient "github.com/aybykovskii/gitlab-tui/internal/gitlab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitLabUserErrorPreservesTypedErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		err       error
		assertion func(*testing.T, error)
	}{
		{
			name: "client not configured sentinel",
			err:  gitlabclient.ErrClientNotConfigured,
			assertion: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, gitlabclient.ErrClientNotConfigured)
			},
		},
		{
			name: "unexpected response typed error",
			err:  &gitlabclient.ErrUnexpectedResponse{Status: "418 I'm a teapot", StatusCode: http.StatusTeapot, Body: "short and stout"},
			assertion: func(t *testing.T, err error) {
				t.Helper()

				var responseErr *gitlabclient.ErrUnexpectedResponse
				require.True(t, errors.As(err, &responseErr))
				assert.Equal(t, http.StatusTeapot, responseErr.StatusCode)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := gitLabUserError(tc.err)

			require.Error(t, err)
			tc.assertion(t, err)
		})
	}
}
