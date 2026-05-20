// Package debuglog writes diagnostic entries to a file when GITLAB_TUI_DEBUG is set.
// Usage: GITLAB_TUI_DEBUG=1 ./gitlab-tui  → logs appear in /tmp/gitlab-tui.log
package debuglog

import (
	"fmt"
	"log"
	"os"
	"time"
)

var logger *log.Logger

func init() {
	if os.Getenv("GITLAB_TUI_DEBUG") == "" {
		return
	}

	f, err := os.OpenFile("/tmp/gitlab-tui.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}

	logger = log.New(f, "", 0)
	Log("--- session started %s ---", time.Now().Format(time.RFC3339))
}

// Log writes a formatted line to the debug log file. No-op when GITLAB_TUI_DEBUG is unset.
func Log(format string, args ...any) {
	if logger == nil {
		return
	}

	logger.Output(2, fmt.Sprintf(format, args...)) //nolint:errcheck
}
