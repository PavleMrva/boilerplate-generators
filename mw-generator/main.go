package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type LoggerMethodInfo struct {
	Name       string
	Parameters []string
	Returns    []string
	Content    string
}

var loggerTemplate string = `package middleware

// TODO: Change <service-package> to the package name of the service

import (
	"context"
	"time"
)

func NewLogMiddleware(service <service-package>.Service) <service-package>.Service {
	return &logMiddleware{
		next:        service,
		serviceName: "%s",
	}
}

type logMiddleware struct {
	next        <service-package>.Service
	serviceName string
}

`

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

func generateMethodCallCode(method LoggerMethodInfo) string {
	var sb strings.Builder

	before := "\tlog.WithContext(ctx).WithFields(log.Fields{\n\t\t"
	before += fmt.Sprintf(`"service": m.serviceName,
		"method":  "%s",
		"layer":   "service",
	}).Info("service-request")
	
	`, method.Name)

	sb.WriteString(before)

	// Handling for different number of return values
	switch len(method.Returns) {
	case 1:
		// Single return value, could be an error or another type
		if method.Returns[0] == "error" {
			sb.WriteString(fmt.Sprintf("err := m.next.%s(ctx)\n\n\t", method.Name))

			rest := "log.WithContext(ctx).WithFields(log.Fields{\n\t\t"
			rest += fmt.Sprintf(`"service": m.serviceName,
		"method":  "%s",
		"layer":   "service",
		"err": 	 err,
	}).Info("service-response")

	return err`, method.Name)

			sb.WriteString(rest)

		} else {
			sb.WriteString(fmt.Sprintf("res := m.next.%s(ctx)\n\n\t", method.Name))

			rest := "log.WithContext(ctx).WithFields(log.Fields{\n\t\t"
			rest += fmt.Sprintf(`"service": m.serviceName,
		"method":  "%s",
		"layer":   "service",
		"res": 	 res,
	}).Info("service-response")

	return res`, method.Name)

			sb.WriteString(rest)
		}
	case 2:
		sb.WriteString(fmt.Sprintf("res, err := m.next.%s(ctx)\n\n\t", method.Name))

		rest := "log.WithContext(ctx).WithFields(log.Fields{\n\t\t"
		rest += fmt.Sprintf(`"service": m.serviceName,
		"method":  "%s",
		"layer":   "service",
		"res": 	 res,
		"err": 	 err,
	}).Info("service-response")

	return res, err`, method.Name)

		sb.WriteString(rest)
	default:
		sb.WriteString(fmt.Sprintf("m.next.%s(ctx)\n", method.Name))
	}

	sb.WriteString("\n")
	return sb.String()
}

func generateLoggerMiddleware(serviceName string, methods []LoggerMethodInfo) string {
	initialStr := fmt.Sprintf(loggerTemplate, serviceName)

	for _, method := range methods {
		contentStr := fmt.Sprintf(`func (m *logMiddleware) %s(`, method.Name)

		for i, param := range method.Parameters {
			p := strings.Split(param, " ")

			if i == len(method.Parameters)-1 {
				contentStr += fmt.Sprintf("%s %s) ", p[1], p[0])
			} else {
				contentStr += fmt.Sprintf("%s %s, ", p[1], p[0])
			}
		}

		if len(method.Returns) > 0 {
			if len(method.Returns) == 1 {
				contentStr += fmt.Sprintf("%s {\n", method.Returns[0])
			} else {
				contentStr += "("

				for i, returns := range method.Returns {
					if i == len(method.Returns)-1 {
						contentStr += fmt.Sprintf("%s) {\n", returns)
					} else {
						contentStr += fmt.Sprintf("%s, ", returns)
					}
				}
			}
		}

		contentStr += generateMethodCallCode(method)

		contentStr += "}\n\n"
		initialStr += contentStr
	}

	return initialStr
}

// Extracts methods from the given interface type.
func extractMethodsFromInterface(itype *ast.InterfaceType) []LoggerMethodInfo {
	var methods []LoggerMethodInfo

	for _, field := range itype.Methods.List {
		var mInfo LoggerMethodInfo

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
	flag.Parse()

	if *interfaceName == "" {
		fmt.Println("Please provide an interface name")
		os.Exit(1)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, 0)
	if err != nil {
		fmt.Println("Failed to parse package:", err)
		os.Exit(1)
	}

	var methods []LoggerMethodInfo
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

	if err := os.MkdirAll("middleware", 0755); err != nil {
		panic(err)
	}

	file, err := os.Create("middleware/log.go")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	str := generateLoggerMiddleware(*interfaceName, methods)

	fmt.Fprint(file, str)
}
