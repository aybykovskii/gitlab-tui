package launcher

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var commandRunner = func(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}

func OpenURL(url string) error {
	cmd := ""
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		return fmt.Errorf("open url unsupported on %s", runtime.GOOS)
	}

	if err := commandRunner(cmd, url); err != nil {
		return fmt.Errorf("open url: %w", err)
	}

	return nil
}

func OpenEditor(filePath string, line int) error {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		return errors.New("EDITOR is not set")
	}

	parts := strings.Fields(editor)
	name := filepath.Base(parts[0])
	args := append([]string{}, parts[1:]...)
	args = append(args, editorArgs(name, filePath, line)...)
	if err := commandRunner(parts[0], args...); err != nil {
		return fmt.Errorf("open editor: %w", err)
	}

	return nil
}

func editorArgs(editorName string, filePath string, line int) []string {
	if line < 1 {
		line = 1
	}

	switch editorName {
	case "vim", "nvim", "nano":
		return []string{fmt.Sprintf("+%d", line), filePath}
	case "code":
		return []string{"--goto", fmt.Sprintf("%s:%d", filePath, line)}
	default:
		return []string{filePath}
	}
}
