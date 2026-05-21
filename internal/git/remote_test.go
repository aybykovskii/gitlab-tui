package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRunner struct {
	out string
	err error
}

func (r fakeRunner) RunGit(args ...string) ([]byte, error) {
	return []byte(r.out), r.err
}

func TestRemoteURLsDeduplicatesFetchAndPush(t *testing.T) {
	t.Parallel()

	urls, err := RemoteURLs(fakeRunner{out: "origin\tgit@gitlab.com:group/project.git (fetch)\norigin\tgit@gitlab.com:group/project.git (push)\nupstream\thttps://gitlab.com/other/project.git (fetch)\n"})
	require.NoError(t, err)
	assert.Len(t, urls, 2)
}

func TestProjectPathFromRemote(t *testing.T) {
	t.Run("SSH remote", func(t *testing.T) {
		t.Parallel()

		path, ok := ProjectPathFromRemote("git@gitlab.com:group/project.git", "https://gitlab.com")
		require.True(t, ok)
		assert.Equal(t, "group/project", path)
	})

	t.Run("HTTPS remote with subgroups", func(t *testing.T) {
		t.Parallel()

		path, ok := ProjectPathFromRemote("https://gitlab.example.com/group/sub/project.git", "https://gitlab.example.com")
		require.True(t, ok)
		assert.Equal(t, "group/sub/project", path)
	})

	t.Run("rejects different host", func(t *testing.T) {
		t.Parallel()

		_, ok := ProjectPathFromRemote("git@gitlab.com:group/project.git", "https://gitlab.example.com")
		assert.False(t, ok)
	})
}
