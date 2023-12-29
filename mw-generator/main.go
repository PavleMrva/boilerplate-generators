package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

type MethodInfo struct {
	Name       string
	Parameters []string
	Returns    []string
	Content    string
}

// getFieldList returns a slice of strings describing the fields in a FieldList (parameters or return values).
func getFieldList(fl *ast.FieldList) []string {
	var fields []string
	if fl != nil {
		for _, field := range fl.List {
			typeName := exprToString(field.Type)
			// If the field has names, prefix each name with the type (e.g., "int x, int y").
			// Otherwise, just add the type (e.g., "int").
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					fields = append(fields, fmt.Sprintf("%s %s", typeName, name.Name))
				}
			} else {
				fields = append(fields, typeName)
			}
		}
	}
	return fields
}

// exprToString gets the string representation of an expression (like a type).
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.ArrayType:
		return "[]" + exprToString(e.Elt)
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// Extracts methods from the given interface type.
func extractMethodsFromInterface(itype *ast.InterfaceType) []MethodInfo {
	var methods []MethodInfo

	for _, field := range itype.Methods.List {
		var mInfo MethodInfo

		mInfo.Name = field.Names[0].Name

		if ftype, ok := field.Type.(*ast.FuncType); ok {
			mInfo.Parameters = getFieldList(ftype.Params)
			mInfo.Returns = getFieldList(ftype.Results)
			methods = append(methods, mInfo)
		}
	}

	return methods
}

func main() {
	interfaceName := flag.String("interface", "", "Name of the interface")
	middlewareType := flag.String("type", "", "Middleware type (logger, tracer)")

	flag.Parse()

	if *interfaceName == "" {
		fmt.Println("Please provide an interface name")
		os.Exit(1)
	}

	if *middlewareType == "" {
		fmt.Println("Please provide a middleware type")
		os.Exit(1)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, 0)
	if err != nil {
		fmt.Println("Failed to parse package:", err)
		os.Exit(1)
	}

	var methods []MethodInfo
	found := false

	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.Name == *interfaceName {
							if itype, ok := typeSpec.Type.(*ast.InterfaceType); ok {
								methods = extractMethodsFromInterface(itype)
								found = true
								break
							}
						}
					}
				}
				if found {
					break
				}
			}
			if found {
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		fmt.Println("Interface not found:", *interfaceName)
		os.Exit(1)
	}

	var str string

	switch *middlewareType {
	case "logger":
		str = generateLoggerMiddleware(*interfaceName, methods)
	case "tracer":
		str = generateTracerMiddleware(*interfaceName, methods)
	default:
		fmt.Println("Please provide a valid middleware type")
		os.Exit(1)
	}

	if err := os.MkdirAll("middleware", 0755); err != nil {
		panic(err)
	}

	var filename string

	switch *middlewareType {
	case "logger":
		filename = "middleware/log.go"
	case "tracer":
		filename = "middleware/trace.go"
	}

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Fprint(file, str)
}
