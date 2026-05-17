package main

import (
	"os"

	"github.com/aybykovskii/gitlab-tui/internal/app"
	"github.com/aybykovskii/gitlab-tui/internal/version"
)

func main() {
	os.Exit(app.New(version.Value).Run(os.Args[1:], os.Stdout, os.Stderr))
}
