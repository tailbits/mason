package openapi

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"sync"
)

type structDocKey struct {
	pkgPath  string
	typeName string
}

var queryParamDocCache sync.Map // map[structDocKey]map[string]string

func QueryParamDescriptions(t reflect.Type) map[string]string {
	if t == nil {
		return nil
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	key := structDocKey{pkgPath: t.PkgPath(), typeName: t.Name()}
	if cached, ok := queryParamDocCache.Load(key); ok {
		return cached.(map[string]string)
	}
	desc := parseStructFieldDocs(key.pkgPath, key.typeName)
	queryParamDocCache.Store(key, desc)
	return desc
}

func parseStructFieldDocs(pkgPath, typeName string) map[string]string {
	result := make(map[string]string)
	if pkgPath == "" || typeName == "" {
		return result
	}
	pkg, err := importPackage(pkgPath)
	if err != nil {
		return result
	}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkg.Dir, nil, parser.ParseComments)
	if err != nil {
		return result
	}
	for _, astPkg := range pkgs {
		for _, file := range astPkg.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range gen.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok || typeSpec.Name.Name != typeName {
						continue
					}
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					collectFieldDescriptions(structType, result)
					return result
				}
			}
		}
	}
	return result
}

func collectFieldDescriptions(st *ast.StructType, docs map[string]string) {
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		name := field.Names[0].Name
		desc := extractDescription(name, field)
		if desc != "" {
			docs[name] = desc
		}
	}
}

func importPackage(pkgPath string) (*build.Package, error) {
	pkg, err := build.Default.Import(pkgPath, ".", build.FindOnly)
	if err == nil {
		return pkg, nil
	}
	trimmed := trimTestPackagePath(pkgPath)
	if trimmed != pkgPath {
		if pkg, err2 := build.Default.Import(trimmed, ".", build.FindOnly); err2 == nil {
			return pkg, nil
		}
	}
	return nil, err
}

func trimTestPackagePath(pkgPath string) string {
	if !strings.HasSuffix(pkgPath, "_test") {
		return pkgPath
	}
	idx := strings.LastIndex(pkgPath, "/")
	if idx == -1 {
		return strings.TrimSuffix(pkgPath, "_test")
	}
	return pkgPath[:idx+1] + strings.TrimSuffix(pkgPath[idx+1:], "_test")
}

func extractDescription(name string, field *ast.Field) string {
	var cg *ast.CommentGroup
	if field.Doc != nil {
		cg = field.Doc
	} else if field.Comment != nil {
		cg = field.Comment
	}
	if cg == nil {
		return ""
	}
	text := strings.TrimSpace(cg.Text())
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	first := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(first, name) {
		return ""
	}
	first = strings.TrimSpace(strings.TrimLeft(strings.TrimPrefix(first, name), ":-., \t"))
	lines[0] = first
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			parts = append(parts, line)
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}
