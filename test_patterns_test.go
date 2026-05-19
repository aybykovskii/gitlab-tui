package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestFunctionsCallParallelFirst(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("**/*_test.go")
	require.NoError(t, err)

	for _, path := range append([]string{"lint_config_test.go"}, files...) {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			require.NoError(t, err)

			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || !strings.HasPrefix(fn.Name.Name, "Test") {
					continue
				}

				require.NotEmpty(t, fn.Body.List, "%s has empty body", fn.Name.Name)
				assert.True(t, isTParallelCall(fn.Body.List[0]), "%s must call t.Parallel() first", fn.Name.Name)
				assertSubtestsCallParallelFirst(t, fn)
			}
		})
	}
}

func assertSubtestsCallParallelFirst(t *testing.T, fn *ast.FuncDecl) {
	t.Helper()

	ast.Inspect(fn, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isTRunCall(call) || len(call.Args) < 2 {
			return true
		}

		callback, ok := call.Args[1].(*ast.FuncLit)
		require.True(t, ok, "%s has t.Run without func literal", fn.Name.Name)
		require.NotEmpty(t, callback.Body.List, "%s has empty subtest", fn.Name.Name)
		assert.True(t, isTParallelCall(callback.Body.List[0]), "%s subtest must call t.Parallel() first", fn.Name.Name)

		return true
	})
}

func isTRunCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Run" {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)

	return ok && ident.Name == "t"
}

func isTParallelCall(stmt ast.Stmt) bool {
	expr, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	call, ok := expr.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Parallel" {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)

	return ok && ident.Name == "t"
}
