// Package main содержит статический анализатор для проверки использования panic,
// log.Fatal и os.Exit вне функции main пакета main.
package main

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer — анализатор, сообщающий о:
// - любом использовании встроенной функции panic;
// - вызовах log.Fatal / log.Fatalf / log.Fatalln и os.Exit вне функции main пакета main.
// Диагностику можно отключить для строки комментарием // linter:ignore на той же строке.
var Analyzer = &analysis.Analyzer{
	Name:     "linter",
	Doc:      "reports use of panic, and log.Fatal/os.Exit outside main.main; use // linter:ignore on same line to suppress",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspectResult := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	pkgName := pass.Pkg.Name()
	isMainPkg := pkgName == "main"

	inspectResult.Nodes([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node, push bool) bool {
		if !push {
			return true
		}
		call := n.(*ast.CallExpr)

		// 1) Встроенная panic
		if isBuiltinPanic(pass, call) {
			if !hasLinterIgnore(pass, call.Pos()) {
				pass.Reportf(call.Pos(), "use of builtin panic is not allowed")
			}
			return true
		}

		// 2) log.Fatal / os.Exit — только в main пакете в функции main
		enclosingFunc := findEnclosingFunc(pass, call)
		allowedInMain := isMainPkg && enclosingFunc == "main"

		if isLogFatal(pass, call) {
			if !allowedInMain && !hasLinterIgnore(pass, call.Pos()) {
				pass.Reportf(call.Pos(), "log.Fatal/Fatalf/Fatalln must not be used outside main.main")
			}
			return true
		}
		if isOsExit(pass, call) {
			if !allowedInMain && !hasLinterIgnore(pass, call.Pos()) {
				pass.Reportf(call.Pos(), "os.Exit must not be used outside main.main")
			}
			return true
		}
		return true
	})
	return nil, nil
}

func findEnclosingFunc(pass *analysis.Pass, call ast.Node) string {
	callPos := call.Pos()
	var innermost *ast.FuncDecl
	for _, f := range pass.Files {
		if f.Pos() <= callPos && callPos <= f.End() {
			ast.Inspect(f, func(n ast.Node) bool {
				if n == nil {
					return true
				}
				decl, ok := n.(*ast.FuncDecl)
				if !ok || decl.Pos() > callPos || callPos > decl.End() {
					return true
				}
				// выбираем самую вложенную функцию (наименьший диапазон)
				if innermost == nil || (decl.End()-decl.Pos() < innermost.End()-innermost.Pos()) {
					innermost = decl
				}
				return true
			})
			break
		}
	}
	if innermost != nil && innermost.Name != nil {
		return innermost.Name.Name
	}
	return ""
}

// hasLinterIgnore возвращает true, если для позиции pos в том же файле есть комментарий
// «// linter:ignore» на той же строке.
func hasLinterIgnore(pass *analysis.Pass, pos token.Pos) bool {
	nodeLine := pass.Fset.Position(pos).Line
	for _, f := range pass.Files {
		if pos < f.Pos() || pos > f.End() {
			continue
		}
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				commentLine := pass.Fset.Position(c.Slash).Line
				if commentLine == nodeLine && strings.Contains(c.Text, "linter:ignore") {
					return true
				}
			}
		}
		break
	}
	return false
}

func isBuiltinPanic(pass *analysis.Pass, call *ast.CallExpr) bool {
	id, ok := call.Fun.(*ast.Ident)
	if !ok || id.Name != "panic" {
		return false
	}
	obj, ok := pass.TypesInfo.Uses[id]
	if !ok {
		return false
	}
	_, isBuiltin := obj.(*types.Builtin)
	return isBuiltin
}

func isLogFatal(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	name := sel.Sel.Name
	if name != "Fatal" && name != "Fatalf" && name != "Fatalln" {
		return false
	}
	obj := pass.TypesInfo.Uses[sel.Sel]
	if obj == nil {
		return false
	}
	pkg := obj.Pkg()
	if pkg == nil {
		return false
	}
	return pkg.Name() == "log" || pkg.Path() == "log"
}

func isOsExit(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "Exit" {
		return false
	}
	obj := pass.TypesInfo.Uses[sel.Sel]
	if obj == nil {
		return false
	}
	pkg := obj.Pkg()
	if pkg == nil {
		return false
	}
	return pkg.Name() == "os" || pkg.Path() == "os"
}
