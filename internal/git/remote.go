package git

import (
	"net/url"
	"os/exec"
	"strings"
)

type Runner interface {
	RunGit(args ...string) ([]byte, error)
}

type CommandRunner struct {
	Dir string
}

func (r CommandRunner) RunGit(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Dir
	return cmd.Output()
}

func (r CommandRunner) RemoteURLs() ([]string, error) {
	return RemoteURLs(r)
}

func RemoteURLs(runner Runner) ([]string, error) {
	out, err := runner.RunGit("remote", "-v")
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	urls := []string{}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		remoteURL := fields[1]
		if !seen[remoteURL] {
			seen[remoteURL] = true
			urls = append(urls, remoteURL)
		}
	}

	return urls, nil
}

func ProjectPathFromRemote(remoteURL string, accountHost string) (string, bool) {
	host := hostName(accountHost)
	if host == "" {
		return "", false
	}

	path, ok := projectPathFromRemote(remoteURL, host)
	if !ok {
		return "", false
	}

	path = strings.TrimSuffix(strings.TrimPrefix(path, "/"), ".git")
	if path == "" || !strings.Contains(path, "/") {
		return "", false
	}

	return path, true
}

func projectPathFromRemote(remoteURL string, host string) (string, bool) {
	if parsed, err := url.Parse(remoteURL); err == nil && parsed.Hostname() != "" {
		if parsed.Hostname() != host {
			return "", false
		}
		return parsed.Path, true
	}

	sshPrefix := "git@" + host + ":"
	if strings.HasPrefix(remoteURL, sshPrefix) {
		return strings.TrimPrefix(remoteURL, sshPrefix), true
	}

	return "", false
}

func hostName(rawHost string) string {
	if parsed, err := url.Parse(rawHost); err == nil && parsed.Hostname() != "" {
		return parsed.Hostname()
	}
	return strings.TrimSuffix(rawHost, "/")
}
