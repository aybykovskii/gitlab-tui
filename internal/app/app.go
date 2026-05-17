package app

import (
	"fmt"
	"io"

	"github.com/aybykovskii/gitlab-tui/internal/config"
	"github.com/aybykovskii/gitlab-tui/internal/tui"
)

type App struct {
	version string
	env     []string
}

func New(version string) App {
	return App{version: version}
}

func NewWithEnv(version string, env []string) App {
	return App{version: version, env: env}
}

func (a App) Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		if err := tui.Run(stdout); err != nil {
			fmt.Fprintf(stderr, "run TUI: %v\n", err)
			return 1
		}
		return 0
	}

	switch args[0] {
	case "init":
		return a.runInit(stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, a.version)
		return 0
	case "help", "--help", "-h":
		writeUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		writeUsage(stderr)
		return 2
	}
}

func (a App) runInit(stdout io.Writer, stderr io.Writer) int {
	path, err := (config.Paths{Env: a.env}).Path()
	if err != nil {
		fmt.Fprintf(stderr, "resolve config path: %v\n", err)
		return 1
	}

	if err := config.Init(path, config.Default()); err != nil {
		fmt.Fprintf(stderr, "init config: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "created config: %s\n", path)
	return 0
}

func writeUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: gitlab-tui-go [command]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  init     Create YAML config")
	fmt.Fprintln(w, "  version  Print version")
}
