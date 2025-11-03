package handler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// EnhancedHandlerAnalyzer extiende el analizador con inferencia avanzada
type EnhancedHandlerAnalyzer struct {
	*HandlerAnalyzer
}

// NewEnhancedHandlerAnalyzer crea un analizador mejorado
func NewEnhancedHandlerAnalyzer() *EnhancedHandlerAnalyzer {
	return &EnhancedHandlerAnalyzer{
		HandlerAnalyzer: NewHandlerAnalyzer(),
	}
}

// AnalyzeFunction mejorado con inferencia de tipos
func (a *EnhancedHandlerAnalyzer) AnalyzeFunction(packageName, filePath, functionName string) (*HandlerInfo, error) {
	info, err := a.HandlerAnalyzer.AnalyzeFunction(packageName, filePath, functionName)
	if err != nil {
		return nil, err
	}

	if info != nil {
		a.enhanceHandlerInfo(info, filePath)
	}

	return info, nil
}

// AnalyzeMethod mejorado con inferencia de tipos
func (a *EnhancedHandlerAnalyzer) AnalyzeMethod(packageName, filePath, structName, methodName string) (*HandlerInfo, error) {
	info, err := a.HandlerAnalyzer.AnalyzeMethod(packageName, filePath, structName, methodName)
	if err != nil {
		return nil, err
	}

	if info != nil {
		a.enhanceHandlerInfo(info, filePath)
	}

	return info, nil
}

// enhanceHandlerInfo mejora la información del handler con inferencia avanzada
func (a *EnhancedHandlerAnalyzer) enhanceHandlerInfo(info *HandlerInfo, filePath string) {
	// Mejorar análisis de parámetros
	for i := range info.Params {
		a.enhanceParameterInfo(&info.Params[i], filePath)
	}

	// Inferir tipo de retorno si es desconocido
	if info.ReturnType == "unknown" || info.ReturnType == "" {
		info.ReturnType = a.inferReturnType(info, filePath)
	}
}

// enhanceParameterInfo mejora la información de un parámetro
func (a *EnhancedHandlerAnalyzer) enhanceParameterInfo(param *ParamInfo, filePath string) {
	// Si es un struct, analizar sus campos para extraer información detallada
	if strings.Contains(param.Type, "struct") || !a.isBasicType(param.Type) {
		a.analyzeStructParameter(param, filePath)
	}

	// Para parámetros básicos, inferir más información basado en el nombre
	if a.isBasicType(param.Type) {
		a.inferBasicParameterInfo(param)
	}
}

// analyzeStructParameter analiza un parámetro struct en detalle
func (a *EnhancedHandlerAnalyzer) analyzeStructParameter(param *ParamInfo, filePath string) {
	// Buscar la definición del struct en el archivo
	file, err := parser.ParseFile(a.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return
	}

	// Buscar el struct por nombre
	structName := strings.TrimPrefix(param.Type, "*")
	ast.Inspect(file, func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.Name == structName {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						a.extractStructFieldsInfo(param, structType)
						return false
					}
				}
			}
		}
		return true
	})
}

// extractStructFieldsInfo extrae información de los campos de un struct
func (a *EnhancedHandlerAnalyzer) extractStructFieldsInfo(param *ParamInfo, structType *ast.StructType) {
	if structType.Fields == nil {
		return
	}

	// Resetear información del parámetro ya que ahora tendrá campos detallados
	param.Location = "body" // Los structs generalmente son body

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue // Campos anónimos
		}

		fieldName := field.Names[0].Name
		if !ast.IsExported(fieldName) {
			continue // Ignorar campos no exportados
		}

		fieldType := a.getTypeName(field.Type)
		fieldInfo := ParamInfo{
			Name: fieldName,
			Type: fieldType,
		}

		// Analizar tags del campo
		if field.Tag != nil {
			tags := a.parseStructTags(field.Tag.Value)
			fieldInfo.Location = a.determineFieldLocation(tags)
			fieldInfo.JSONName = a.getJSONName(tags, fieldName)
			fieldInfo.Required = a.isFieldRequired(tags)
		}

		// Si no tenemos JSON name, usar el nombre del campo
		if fieldInfo.JSONName == "" {
			fieldInfo.JSONName = fieldName
		}

		// Agregar como sub-parámetro (en una implementación real, podríamos tener una estructura jerárquica)
		if fieldInfo.Location != "" && fieldInfo.Location != "body" {
			// Este campo es un parámetro individual (query, path, etc.)
			param.SubParams = append(param.SubParams, fieldInfo)
		}
	}
}

// parseStructTags parsea los tags de un campo de struct
func (a *EnhancedHandlerAnalyzer) parseStructTags(tagLiteral string) map[string]string {
	tags := make(map[string]string)
	
	// Remover backticks
	tagStr := strings.Trim(tagLiteral, "`")
	
	// Parsear cada tag
	for _, tagPart := range strings.Split(tagStr, " ") {
		tagPart = strings.TrimSpace(tagPart)
		if tagPart == "" {
			continue
		}
		
		parts := strings.SplitN(tagPart, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		tagName := parts[0]
		tagValue := strings.Trim(parts[1], `"`)
		tags[tagName] = tagValue
	}
	
	return tags
}

// determineFieldLocation determina la ubicación de un campo basado en sus tags
func (a *EnhancedHandlerAnalyzer) determineFieldLocation(tags map[string]string) string {
	if tags["uri"] != "" {
		return "path"
	}
	if tags["form"] != "" {
		return "form"
	}
	if tags["header"] != "" {
		return "header"
	}
	if tags["query"] != "" {
		return "query"
	}
	if tags["json"] != "" {
		return "body"
	}
	return ""
}

// getJSONName obtiene el nombre JSON de un campo
func (a *EnhancedHandlerAnalyzer) getJSONName(tags map[string]string, fieldName string) string {
	if jsonTag, exists := tags["json"]; exists {
		// json:"name,omitempty" → tomar solo "name"
		if commaPos := strings.Index(jsonTag, ","); commaPos != -1 {
			return jsonTag[:commaPos]
		}
		return jsonTag
	}
	return fieldName
}

// isFieldRequired determina si un campo es requerido
func (a *EnhancedHandlerAnalyzer) isFieldRequired(tags map[string]string) bool {
	if validateTag, exists := tags["validate"]; exists {
		return strings.Contains(validateTag, "required")
	}
	if bindingTag, exists := tags["binding"]; exists {
		return strings.Contains(bindingTag, "required")
	}
	return false
}

// inferBasicParameterInfo infiere información para parámetros básicos
func (a *EnhancedHandlerAnalyzer) inferBasicParameterInfo(param *ParamInfo) {
	// Basado en el nombre del parámetro, inferir ubicación
	switch strings.ToLower(param.Name) {
	case "id", "userid", "productid":
		param.Location = "path"
		param.Required = true
	case "limit", "offset", "page", "sort":
		param.Location = "query"
	default:
		param.Location = "query" // Por defecto para tipos básicos
	}
}

// inferReturnType infiere el tipo de retorno del handler
func (a *EnhancedHandlerAnalyzer) inferReturnType(info *HandlerInfo, filePath string) string {
	// Por ahora, inferimos basado en el nombre del handler
	// En una implementación completa, analizaríamos el cuerpo de la función
	
	handlerName := strings.ToLower(info.Name)
	
	switch {
	case strings.Contains(handlerName, "get") || strings.Contains(handlerName, "list"):
		return "array" // Probablemente retorna una lista
	case strings.Contains(handlerName, "create") || strings.Contains(handlerName, "update"):
		return "object" // Probablemente retorna un objeto creado/actualizado
	case strings.Contains(handlerName, "delete"):
		return "boolean" // Probablemente retorna éxito/fracaso
	default:
		return "object" // Por defecto
	}
}

// GetOpenAPIType mapea tipos de Go a tipos OpenAPI
func (a *EnhancedHandlerAnalyzer) GetOpenAPIType(goType string) (string, string) {
	switch goType {
	case "string":
		return "string", ""
	case "int", "int8", "int16", "int32", "int64", 
	     "uint", "uint8", "uint16", "uint32", "uint64":
		return "integer", a.getIntegerFormat(goType)
	case "float32", "float64":
		return "number", a.getNumberFormat(goType)
	case "bool":
		return "boolean", ""
	case "byte":
		return "string", "byte"
	case "rune":
		return "string", ""
	case "time.Time":
		return "string", "date-time"
	default:
		if strings.HasPrefix(goType, "[]") {
			return "array", ""
		}
		if strings.HasPrefix(goType, "map[") {
			return "object", ""
		}
		return "object", "" // Structs personalizados
	}
}

func (a *EnhancedHandlerAnalyzer) getIntegerFormat(goType string) string {
	switch goType {
	case "int8", "uint8", "byte":
		return "int32"
	case "int16", "uint16":
		return "int32"
	case "int32", "uint32", "rune":
		return "int32"
	case "int64", "uint64":
		return "int64"
	case "int", "uint":
		return "int64"
	default:
		return "int64"
	}
}

func (a *EnhancedHandlerAnalyzer) getNumberFormat(goType string) string {
	switch goType {
	case "float32":
		return "float"
	case "float64":
		return "double"
	default:
		return "double"
	}
}