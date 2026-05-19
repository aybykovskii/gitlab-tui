package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReusableDiffPackageLivesUnderPkg(t *testing.T) {
	t.Parallel()

	_, err := os.Stat("pkg/diff")
	require.NoError(t, err)

	files, err := filepath.Glob("pkg/diff/*.go")
	require.NoError(t, err)
	require.NotEmpty(t, files)

	for _, path := range files {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			require.NoError(t, err)

			for _, imp := range file.Imports {
				assert.NotContains(t, imp.Path.Value, "/internal/", "pkg/diff must not import internal packages")
			}
		})
	}
}

func TestEveryPackageHasDocumentedDocGo(t *testing.T) {
	t.Parallel()

	packages := map[string]struct{}{}
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		require.NoError(t, err)

		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == ".claude" {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			packages[filepath.Dir(path)] = struct{}{}
		}

		return nil
	})
	require.NoError(t, err)

	for dir := range packages {
		dir := dir
		t.Run(dir, func(t *testing.T) {
			t.Parallel()

			docPath := filepath.Join(dir, "doc.go")
			file, err := parser.ParseFile(token.NewFileSet(), docPath, nil, parser.ParseComments)
			require.NoError(t, err)
			require.NotNil(t, file.Doc, "%s needs a package comment", docPath)
			assert.NotEmpty(t, strings.TrimSpace(file.Doc.Text()))
			assert.NotEmpty(t, packageName(file))
		})
	}
}

func packageName(file *ast.File) string {
	if file == nil || file.Name == nil {
		return ""
	}

	return file.Name.Name
}
