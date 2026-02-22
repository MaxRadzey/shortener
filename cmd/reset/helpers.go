package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const generateComment = "generate:reset"

type structInfo struct {
	name     string
	receiver string
	fields   []fieldInfo
}

type fieldInfo struct {
	name string
	typ  ast.Expr
}

func fileContaining(pass *analysis.Pass, pos token.Pos) *ast.File {
	for _, f := range pass.Files {
		if f.Pos() <= pos && pos <= f.End() {
			return f
		}
	}
	return nil
}

func structInfosFromGenDecl(g *ast.GenDecl, file *ast.File) []*structInfo {
	var result []*structInfo
	for _, spec := range g.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}
		name := typeSpec.Name.Name
		receiver := receiverName(name)
		var fields []fieldInfo
		for _, f := range structType.Fields.List {
			for _, n := range f.Names {
				fields = append(fields, fieldInfo{name: n.Name, typ: f.Type})
			}
		}
		result = append(result, &structInfo{
			name:     name,
			receiver: receiver,
			fields:   fields,
		})
	}
	return result
}

func hasGenerateReset(comments []*ast.CommentGroup) bool {
	for _, cg := range comments {
		for _, c := range cg.List {
			if strings.TrimSpace(strings.TrimPrefix(c.Text, "//")) == generateComment {
				return true
			}
		}
	}
	return false
}

func receiverName(typeName string) string {
	if typeName == "" {
		return "r"
	}
	var b strings.Builder
	for i, r := range typeName {
		if i == 0 {
			if r >= 'A' && r <= 'Z' {
				b.WriteRune(r - 'A' + 'a')
			} else if r >= 'a' && r <= 'z' {
				b.WriteRune(r)
			}
			continue
		}
		if r >= 'A' && r <= 'Z' {
			b.WriteRune(r - 'A' + 'a')
			break
		}
	}
	if b.Len() == 0 {
		return "r"
	}
	return b.String()
}

func generateResetBody(s *structInfo, resetterTypes map[string]bool) string {
	var lines []string
	for _, f := range s.fields {
		line := generateResetLine(s.receiver, f.name, f.typ, resetterTypes)
		if line != "" {
			lines = append(lines, "\t"+line)
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func generateResetLine(receiver, fieldName string, typ ast.Expr, resetterTypes map[string]bool) string {
	base := receiver + "." + fieldName

	switch t := typ.(type) {
	case *ast.Ident:
		name := t.Name
		if zero, ok := zeroValueForBasic(name); ok {
			return base + " = " + zero
		}
		if resetterTypes[name] {
			return "(&" + base + ").Reset()"
		}
		return ""
	case *ast.StarExpr:
		inner, ok := t.X.(*ast.Ident)
		if !ok {
			return ""
		}
		name := inner.Name
		if zero, ok := zeroValueForBasic(name); ok {
			return fmt.Sprintf("if %s != nil {\n\t\t*%s = %s\n\t}", base, base, zero)
		}
		if resetterTypes[name] {
			return fmt.Sprintf("if %s != nil {\n\t\t%s.Reset()\n\t}", base, base)
		}
		return ""
	case *ast.ArrayType:
		if t.Len == nil {
			return base + " = " + base + "[:0]"
		}
		return ""
	case *ast.MapType:
		return "clear(" + base + ")"
	default:
		return ""
	}
}

func zeroValueForBasic(name string) (string, bool) {
	switch name {
	case "bool":
		return "false", true
	case "string":
		return `""`, true
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"byte", "rune":
		return "0", true
	case "float32", "float64":
		return "0", true
	case "complex64", "complex128":
		return "0", true
	default:
		return "", false
	}
}
