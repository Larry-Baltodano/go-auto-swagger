package handler

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type ParamInfo struct {
	Name     string
	Type     string
	Location string
	Required bool
	JSONName string
	SubParams []ParamInfo
}

type HandlerInfo struct {
	Name        string
	Package     string
	File        string
	Params      []ParamInfo
	ReturnType  string
	ErrorReturn bool
}

type HandlerAnalyzer struct {
	fset *token.FileSet
}

func NewHandlerAnalyzer() *HandlerAnalyzer {
	return &HandlerAnalyzer{
		fset: token.NewFileSet(),
	}
}

func (a *HandlerAnalyzer) AnalyzeFunction(packageName, filePath, functionName string) (*HandlerInfo, error) {
	file, err := parser.ParseFile(a.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var handlerInfo *HandlerInfo

	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == functionName {
				handlerInfo = a.analyzeFunctionDeclaration(packageName, filePath, funcDecl)
				return false
			}
		}
		return true
	})

	return handlerInfo, nil
}

func (a *HandlerAnalyzer) AnalyzeMethod(packageName, filePath, structName, methodName string) (*HandlerInfo, error) {
	file, err := parser.ParseFile(a.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var handlerInfo *HandlerInfo

	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == methodName && a.isMethodOfStruct(funcDecl, structName) {
				handlerInfo = a.analyzeFunctionDeclaration(packageName, filePath, funcDecl)
				return false
			}
		}
		return true
	})

	return handlerInfo, nil
}

func (a *HandlerAnalyzer) analyzeFunctionDeclaration(packageName, filePath string, funcDecl *ast.FuncDecl) *HandlerInfo {
	info := &HandlerInfo{
		Name:    funcDecl.Name.Name,
		Package: packageName,
		File:    filePath,
		Params:  []ParamInfo{},
	}

	if funcDecl.Type.Params != nil {
		for _, param := range funcDecl.Type.Params.List {
			paramInfo := a.analyzeParameter(param)
			info.Params = append(info.Params, paramInfo...)
		}
	}

	if funcDecl.Type.Results != nil {
		if len(funcDecl.Type.Results.List) > 0 {
			info.ReturnType = a.getTypeName(funcDecl.Type.Results.List[0].Type)
		}
		
		for _, result := range funcDecl.Type.Results.List {
			if a.getTypeName(result.Type) == "error" {
				info.ErrorReturn = true
				break
			}
		}
	}

	return info
}

func (a *HandlerAnalyzer) analyzeParameter(param *ast.Field) []ParamInfo {
	var params []ParamInfo

	paramNames := a.getParamNames(param)

	for _, paramName := range paramNames {
		paramInfo := ParamInfo{
			Name: paramName,
			Type: a.getTypeName(param.Type),
		}

		paramInfo.Location = a.determineParameterLocation(paramInfo.Name, paramInfo.Type)

		if paramInfo.Location == "body" {
			a.analyzeStructFields(param.Type, &paramInfo)
		}

		params = append(params, paramInfo)
	}

	return params
}

func (a *HandlerAnalyzer) getParamNames(param *ast.Field) []string {
	var names []string

	if param.Names != nil {
		for _, name := range param.Names {
			names = append(names, name.Name)
		}
	} else {
		names = append(names, "anonymous")
	}

	return names
}

func (a *HandlerAnalyzer) getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + a.getTypeName(t.X)
	case *ast.SelectorExpr:
		return a.getTypeName(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + a.getTypeName(t.Elt)
	case *ast.StructType:
		return "struct{}"
	default:
		return "unknown"
	}
}

func (a *HandlerAnalyzer) determineParameterLocation(name, paramType string) string {
	if paramType == "*gin.Context" || paramType == "gin.Context" {
		return "context"
	}

	if strings.Contains(paramType, "struct") || !a.isBasicType(paramType) {
		return "body"
	}

	return "query"
}

func (a *HandlerAnalyzer) isBasicType(typeName string) bool {
	basicTypes := []string{
		"string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "byte", "rune",
	}

	for _, basicType := range basicTypes {
		if typeName == basicType {
			return true
		}
	}

	return false
}

func (a *HandlerAnalyzer) analyzeStructFields(expr ast.Expr, paramInfo *ParamInfo) {
	switch t := expr.(type) {
	case *ast.StructType:
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				if field.Tag != nil {
					tag := field.Tag.Value			
					paramInfo.Required = true
					fmt.Println(tag) // -> eliminar luego									
				}
			}
		}
	}
}

func (a *HandlerAnalyzer) isMethodOfStruct(funcDecl *ast.FuncDecl, structName string) bool {
	if funcDecl.Recv == nil {
		return false
	}

	for _, recv := range funcDecl.Recv.List {
		recvType := a.getTypeName(recv.Type)
		recvType = strings.TrimPrefix(recvType, "*")
		if recvType == structName {
			return true
		}
	}

	return false
}