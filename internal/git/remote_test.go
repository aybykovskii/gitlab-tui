package git

import "testing"

type fakeRunner struct {
	out string
	err error
}

func (r fakeRunner) RunGit(args ...string) ([]byte, error) {
	return []byte(r.out), r.err
}

func TestRemoteURLsDeduplicatesFetchAndPush(t *testing.T) {
	urls, err := RemoteURLs(fakeRunner{out: "origin\tgit@gitlab.com:group/project.git (fetch)\norigin\tgit@gitlab.com:group/project.git (push)\nupstream\thttps://gitlab.com/other/project.git (fetch)\n"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(urls) != 2 {
		t.Fatalf("expected 2 unique urls, got %d", len(urls))
	}
}

func TestProjectPathFromSSHRemote(t *testing.T) {
	path, ok := ProjectPathFromRemote("git@gitlab.com:group/project.git", "https://gitlab.com")
	if !ok {
		t.Fatal("expected remote to match")
	}

	if path != "group/project" {
		t.Fatalf("expected group/project, got %q", path)
	}
}

func TestProjectPathFromHTTPSRemote(t *testing.T) {
	path, ok := ProjectPathFromRemote("https://gitlab.example.com/group/sub/project.git", "https://gitlab.example.com")
	if !ok {
		t.Fatal("expected remote to match")
	}

	if path != "group/sub/project" {
		t.Fatalf("expected group/sub/project, got %q", path)
	}
}

func TestProjectPathRejectsDifferentHost(t *testing.T) {
	if _, ok := ProjectPathFromRemote("git@gitlab.com:group/project.git", "https://gitlab.example.com"); ok {
		t.Fatal("expected different host to be rejected")
	}
}
