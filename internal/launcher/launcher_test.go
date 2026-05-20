package launcher

import (
	"errors"
	"runtime"
	"testing"
)

func TestOpenURLRunsPlatformCommand(t *testing.T) {
	var name string
	var args []string
	old := commandRunner
	commandRunner = func(gotName string, gotArgs ...string) error {
		name = gotName
		args = gotArgs
		return nil
	}
	t.Cleanup(func() { commandRunner = old })

	if err := OpenURL("https://gitlab.example.com/group/project"); err != nil {
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			t.Fatalf("OpenURL: %v", err)
		}
		return
	}

	expected := "xdg-open"
	if runtime.GOOS == "darwin" {
		expected = "open"
	}
	if name != expected || len(args) != 1 || args[0] != "https://gitlab.example.com/group/project" {
		t.Fatalf("unexpected command %q args %+v", name, args)
	}
}

func TestOpenURLReturnsCommandError(t *testing.T) {
	old := commandRunner
	commandRunner = func(string, ...string) error { return errors.New("boom") }
	t.Cleanup(func() { commandRunner = old })

	err := OpenURL("https://gitlab.example.com")
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if err == nil {
			t.Fatal("expected command error")
		}
	}
}

func TestOpenEditorRequiresEditor(t *testing.T) {
	t.Setenv("EDITOR", "")
	if err := OpenEditor("file.go", 12); err == nil {
		t.Fatal("expected unset EDITOR error")
	}
}

func TestOpenEditorBuildsLineFlags(t *testing.T) {
	tests := []struct {
		editor string
		want   []string
	}{
		{editor: "vim", want: []string{"+12", "file.go"}},
		{editor: "nvim", want: []string{"+12", "file.go"}},
		{editor: "nano", want: []string{"+12", "file.go"}},
		{editor: "code", want: []string{"--goto", "file.go:12"}},
		{editor: "emacs", want: []string{"file.go"}},
	}

	for _, tc := range tests {
		t.Run(tc.editor, func(t *testing.T) {
			var gotName string
			var gotArgs []string
			old := commandRunner
			commandRunner = func(name string, args ...string) error {
				gotName = name
				gotArgs = args
				return nil
			}
			t.Cleanup(func() { commandRunner = old })
			t.Setenv("EDITOR", tc.editor)

			if err := OpenEditor("file.go", 12); err != nil {
				t.Fatalf("OpenEditor: %v", err)
			}
			if gotName != tc.editor {
				t.Fatalf("expected editor %q, got %q", tc.editor, gotName)
			}
			if len(gotArgs) != len(tc.want) {
				t.Fatalf("expected args %+v, got %+v", tc.want, gotArgs)
			}
			for i := range tc.want {
				if gotArgs[i] != tc.want[i] {
					t.Fatalf("expected args %+v, got %+v", tc.want, gotArgs)
				}
			}
		})
	}
}
