package router

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
)

type RouteInfo struct {
	Method      string
	Path        string
	Handler     string
	HandlerType string
	File        string
	Line        int
}

type GinAnalyzer struct {
	fset *token.FileSet
}

func NewGinAnalyzer() *GinAnalyzer {
	return &GinAnalyzer{fset: token.NewFileSet()}
}

func (a *GinAnalyzer) AnalyzeDirectory(dirPath string) ([]RouteInfo, error) {
	var routes []RouteInfo

	files, err := filepath.Glob(filepath.Join(dirPath, "*.go"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		fileRoutes, err := a.AnalyzeFile(file)
		if err != nil {
			return nil, err
		}

		routes = append(routes, fileRoutes...)
	}

	return routes, nil
}

func (a *GinAnalyzer) AnalyzeFile(filePath string) ([]RouteInfo, error) {
	var routes []RouteInfo

	file, err := parser.ParseFile(a.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ast.Inspect(file, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if route := a.analyzeCallExpression(callExpr, filePath); route != nil {
				routes = append(routes, *route)
			}
		}
		return true
	})

	return routes, nil
}

func (a *GinAnalyzer) analyzeCallExpression(call *ast.CallExpr, filePath string) *RouteInfo {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	methodName := selector.Sel.Name
	if !a.isGinHTTPMethod(methodName) {
		return nil
	}

	if len(call.Args) < 2 {
		return nil
	}

	path, ok := a.extractStringLiteral(call.Args[0])
	if !ok {
		return nil
	}

	handler := a.extractHandlerName(call.Args[1])

	pos := a.fset.Position(call.Pos())

	openAPIPath := a.convertGinPathToOpenAPI(path)

	return &RouteInfo{
		Method: strings.ToUpper(methodName),
		Path: openAPIPath,
		Handler: handler,
		HandlerType: a.determineHandlerType(call.Args[1]),
		File: filePath,
		Line: pos.Line,
	}
}

func (a *GinAnalyzer) isGinHTTPMethod(method string) bool {
	ginMethod := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD", "ANY"}
	for _, m := range ginMethod {
		if strings.EqualFold(method, m) {
			return true
		}
	}
	return false
}

func (a *GinAnalyzer) extractStringLiteral(expr ast.Expr) (string, bool) {
	switch lit := expr.(type) {
	case *ast.BasicLit:
		if lit.Kind == token.STRING {
			return strings.Trim(lit.Value, `"`), true
		}
	case *ast.Ident:
		return lit.Name, true
	}
	return "", false
}

func (a *GinAnalyzer) extractHandlerName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return a.extractHandlerName(x.X) + "." + x.Sel.Name
	case *ast.CallExpr:
		return a.extractHandlerName(x.Fun)
	case *ast.FuncLit:
		return "anonymous"
	default:
		return "unknown"
	}
}

func (a *GinAnalyzer) determineHandlerType(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return "function"
	case *ast.SelectorExpr:
		return "method"
	case *ast.FuncLit:
		return "anonymous"
	case *ast.CallExpr:
		return a.determineHandlerType(x.Fun)
	default:
		return "unknown"
	}
}

func (a *GinAnalyzer) convertGinPathToOpenAPI(ginPath string) string {
	re := regexp.MustCompile(`:(\w+)`)
	return re.ReplaceAllString(ginPath, "{$1}")
}