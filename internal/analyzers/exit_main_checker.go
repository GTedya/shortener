// Package analyzers provides static analysis for Go code.
package analyzers

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ErrExitMainCheckAnalyzer is an analyzer that checks for calls to os.Exit in the main function of the main package.
var ErrExitMainCheckAnalyzer = &analysis.Analyzer{
	Name: "exitmaincheck",
	Doc:  "check call os.Exit in func main() of package main",
	Run:  checkExitInMain,
}

// checkExitInMain is the function that performs the analysis.
// It iterates over the files in the pass, checks if the package
// is the main package, and inspects the main function for calls to os.Exit. If such a call is found, it reports it.
func checkExitInMain(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// tests are generating build cache that has main package, ignoring such files

		if fullPath := pass.Fset.Position(file.Pos()).String(); strings.Contains(fullPath, "go-build") {
			continue
		}

		if pass.Pkg.Name() != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			mainDecl, isFuncDecl := node.(*ast.FuncDecl)
			if !isFuncDecl {
				return true
			}

			if mainDecl.Name.Name != "main" {
				return false
			}

			ast.Inspect(mainDecl, func(node ast.Node) bool {
				callExpr, isCallExpr := node.(*ast.CallExpr)
				if !isCallExpr {
					return true
				}

				s, isSelectorExpr := callExpr.Fun.(*ast.SelectorExpr)
				if !isSelectorExpr {
					return true
				}

				if s.Sel.Name == "Exit" {
					ident, isIdent := s.X.(*ast.Ident)
					if !isIdent {
						return true
					}
					if ident.Name == "os" {
						pass.Reportf(s.Pos(), "exit call in main function")
					}
				}

				return false
			})
			return false
		})
	}
	return nil, nil //nolint:all //Execution of nillint linter is triggered
}
