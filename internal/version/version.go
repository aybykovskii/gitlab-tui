package version

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

var (
	Version         = "v0.0.0-dev"
	GitCommit       = ""
	SourceDateEpoch = "-1"
	GoVersion       = runtime.Version()
	Platform        = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

func Info() string {
	i, _ := strconv.ParseInt(SourceDateEpoch, 10, 64)
	commitDate := ""

	if i >= 0 {
		commitDate = time.Unix(i, 0).UTC().Format(time.RFC3339)
	}

	return fmt.Sprintf("version=%s commit=%s date=%s %s %s", Version, GitCommit, commitDate, GoVersion, Platform)
}
