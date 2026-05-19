package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfoIncludesInjectedBuildMetadata(t *testing.T) {
	t.Parallel()

	oldVersion := Version
	oldGitCommit := GitCommit
	oldSourceDateEpoch := SourceDateEpoch
	oldGoVersion := GoVersion
	oldPlatform := Platform

	t.Cleanup(func() {
		Version = oldVersion
		GitCommit = oldGitCommit
		SourceDateEpoch = oldSourceDateEpoch
		GoVersion = oldGoVersion
		Platform = oldPlatform
	})

	Version = "main"
	GitCommit = "0123456789abcdef"
	SourceDateEpoch = "1716200000"
	GoVersion = "go-test"
	Platform = "test/os"

	info := Info()

	assert.Contains(t, info, "version=main")
	assert.Contains(t, info, "commit=0123456789abcdef")
	assert.Contains(t, info, "date=2024-05-20T10:13:20Z")
	assert.Contains(t, info, "go-test")
	assert.Contains(t, info, "test/os")
}
