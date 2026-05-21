package launcher

import (
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenURL(t *testing.T) {
	t.Run("runs platform command", func(t *testing.T) {
		var name string
		var args []string
		old := commandRunner
		commandRunner = func(gotName string, gotArgs ...string) error {
			name = gotName
			args = gotArgs
			return nil
		}
		t.Cleanup(func() { commandRunner = old })

		err := OpenURL("https://gitlab.example.com/group/project")
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			require.NoError(t, err)
			expected := "xdg-open"
			if runtime.GOOS == "darwin" {
				expected = "open"
			}
			assert.Equal(t, expected, name)
			assert.Equal(t, []string{"https://gitlab.example.com/group/project"}, args)
		}
	})

	t.Run("returns command error", func(t *testing.T) {
		old := commandRunner
		commandRunner = func(string, ...string) error { return errors.New("boom") }
		t.Cleanup(func() { commandRunner = old })

		err := OpenURL("https://gitlab.example.com")
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			assert.Error(t, err)
		}
	})
}

func TestOpenEditorRequiresEditor(t *testing.T) {
	t.Setenv("EDITOR", "")
	assert.Error(t, OpenEditor("file.go", 12))
}

func TestOpenEditorBuildsLineFlags(t *testing.T) {
	tests := []struct {
		editor string
		want   []string
		name   string
	}{
		{editor: "vim", name: "vim", want: []string{"+12", "file.go"}},
		{editor: "nvim", name: "nvim", want: []string{"+12", "file.go"}},
		{editor: "nano", name: "nano", want: []string{"+12", "file.go"}},
		{editor: "code", name: "code", want: []string{"--goto", "file.go:12"}},
		{editor: "code --wait", name: "code", want: []string{"--wait", "--goto", "file.go:12"}},
		{editor: "emacs", name: "emacs", want: []string{"file.go"}},
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

			require.NoError(t, OpenEditor("file.go", 12))
			assert.Equal(t, tc.name, gotName)
			assert.Equal(t, tc.want, gotArgs)
		})
	}
}
