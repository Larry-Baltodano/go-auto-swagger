package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Larry-Baltodano/go-auto-swagger/internal"
	"github.com/Larry-Baltodano/go-auto-swagger/internal/handler"
)

type OpenAPIGenerator struct {
	coordinator *internal.EnhancedCoordinator
}

// NewOpenAPIGenerator crea un nuevo generador
func NewOpenAPIGenerator(coordinator *internal.EnhancedCoordinator) *OpenAPIGenerator {
	return &OpenAPIGenerator{
		coordinator: coordinator,
	}
}

type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       Info                   `json:"info"`
	Paths      map[string]PathItem    `json:"paths"`
	Components *Components            `json:"components,omitempty"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
}

type Operation struct {
	Summary     string               `json:"summary,omitempty"`
	Description string               `json:"description,omitempty"`
	Tags        []string             `json:"tags,omitempty"`
	Parameters  []Parameter          `json:"parameters,omitempty"`
	RequestBody *RequestBody         `json:"requestBody,omitempty"`
	Responses   map[string]Response  `json:"responses"`
}

type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"` // path, query, header, cookie
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type MediaType struct {
	Schema *Schema `json:"schema,omitempty"`
}

type Schema struct {
	Type       string            `json:"type,omitempty"`
	Format     string            `json:"format,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Required   []string          `json:"required,omitempty"`
	Example    interface{}       `json:"example,omitempty"`
}

type Components struct {
	Schemas map[string]Schema `json:"schemas,omitempty"`
}

func (g *OpenAPIGenerator) Generate(apiDesc *internal.APIDescription, title, version string) *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:   title,
			Version: version,
		},
		Paths:      make(map[string]PathItem),
		Components: &Components{},
	}

	// Generar paths y operaciones
	g.generatePaths(spec, apiDesc)

	// Generar schemas de componentes si es necesario
	g.generateSchemas(spec, apiDesc)

	return spec
}

func (g *OpenAPIGenerator) generatePaths(spec *OpenAPISpec, apiDesc *internal.APIDescription) {
	for _, route := range apiDesc.Routes {
		// Crear o obtener el path item
		pathItem, exists := spec.Paths[route.Path]
		if !exists {
			pathItem = PathItem{}
		}

		// Crear operación
		operation := g.routeToOperation(route)

		// Asignar operación al método HTTP correspondiente
		switch route.Method {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "PATCH":
			pathItem.Patch = operation
		}

		spec.Paths[route.Path] = pathItem
	}
}

func (g *OpenAPIGenerator) routeToOperation(route internal.RouteDescription) *Operation {
	operation := &Operation{
		Summary:     g.generateSummary(route),
		Description: g.generateDescription(route),
		Tags:        g.generateTags(route),
		Responses:   g.generateResponses(route),
	}

	// Generar parámetros y request body
	if route.HandlerInfo != nil {
		operation.Parameters = g.generateParameters(route.HandlerInfo)
		operation.RequestBody = g.generateRequestBody(route.HandlerInfo)
	}

	return operation
}

func (g *OpenAPIGenerator) generateSummary(route internal.RouteDescription) string {
	handlerName := route.Handler
	parts := strings.Split(handlerName, ".")
	if len(parts) > 0 {
		handlerName = parts[len(parts)-1]
	}

	// Inferir basado en el nombre del handler y método HTTP
	switch {
	case strings.Contains(strings.ToLower(handlerName), "get"):
		return fmt.Sprintf("Get %s", g.extractResourceName(route.Path))
	case strings.Contains(strings.ToLower(handlerName), "create"):
		return fmt.Sprintf("Create %s", g.extractResourceName(route.Path))
	case strings.Contains(strings.ToLower(handlerName), "update"):
		return fmt.Sprintf("Update %s", g.extractResourceName(route.Path))
	case strings.Contains(strings.ToLower(handlerName), "delete"):
		return fmt.Sprintf("Delete %s", g.extractResourceName(route.Path))
	case strings.Contains(strings.ToLower(handlerName), "list"):
		return fmt.Sprintf("List %s", g.extractResourceName(route.Path))
	default:
		return fmt.Sprintf("%s operation", route.Method)
	}
}

func (g *OpenAPIGenerator) generateDescription(route internal.RouteDescription) string {
	return fmt.Sprintf("Automatically generated endpoint for %s %s", route.Method, route.Path)
}

func (g *OpenAPIGenerator) generateTags(route internal.RouteDescription) []string {
	// Extraer el primer segmento del path como tag
	path := strings.Trim(route.Path, "/")
	segments := strings.Split(path, "/")
	if len(segments) > 0 {
		return []string{segments[0]}
	}
	return []string{"api"}
}

func (g *OpenAPIGenerator) generateParameters(handlerInfo *handler.HandlerInfo) []Parameter {
	var parameters []Parameter

	for _, param := range handlerInfo.Params {
		// Solo generar parámetros para locations que van en parameters (no body)
		if param.Location == "path" || param.Location == "query" || param.Location == "header" {
			paramSchema := g.paramToSchema(param)

			parameter := Parameter{
				Name:        g.getParameterName(param),
				In:          param.Location,
				Description: g.generateParamDescription(param),
				Required:    param.Required,
				Schema:      paramSchema,
			}

			parameters = append(parameters, parameter)
		}

		// Procesar sub-parámetros (campos de struct)
		for _, subParam := range param.SubParams {
			if subParam.Location == "path" || subParam.Location == "query" || subParam.Location == "header" {
				subParamSchema := g.paramToSchema(subParam)

				parameter := Parameter{
					Name:        g.getParameterName(subParam),
					In:          subParam.Location,
					Description: g.generateParamDescription(subParam),
					Required:    subParam.Required,
					Schema:      subParamSchema,
				}

				parameters = append(parameters, parameter)
			}
		}
	}

	return parameters
}

func (g *OpenAPIGenerator) generateRequestBody(handlerInfo *handler.HandlerInfo) *RequestBody {
	var bodyParams []handler.ParamInfo

	// Encontrar todos los parámetros que son body
	for _, param := range handlerInfo.Params {
		if param.Location == "body" {
			bodyParams = append(bodyParams, param)
		}
	}

	if len(bodyParams) == 0 {
		return nil
	}

	// Por simplicidad, tomamos el primer parámetro body
	// En una implementación completa, podríamos combinar múltiples body params
	bodyParam := bodyParams[0]

	// Crear schema para el body
	schema := g.paramToSchema(bodyParam)

	return &RequestBody{
		Description: fmt.Sprintf("%s data", bodyParam.Name),
		Required:    bodyParam.Required,
		Content: map[string]MediaType{
			"application/json": {
				Schema: schema,
			},
		},
	}
}

func (g *OpenAPIGenerator) generateResponses(route internal.RouteDescription) map[string]Response {
	responses := make(map[string]Response)

	// Respuesta exitosa basada en el método HTTP
	successCode := "200"
	switch route.Method {
	case "POST":
		successCode = "201"
	case "DELETE":
		successCode = "204"
	}

	// Crear schema de respuesta si tenemos información del handler
	var responseSchema *Schema
	if route.HandlerInfo != nil && route.HandlerInfo.ReturnType != "" {
		responseSchema = g.returnTypeToSchema(route.HandlerInfo.ReturnType)
	}

	// Respuesta exitosa
	successResponse := Response{
		Description: "Success",
	}

	if responseSchema != nil {
		successResponse.Content = map[string]MediaType{
			"application/json": {
				Schema: responseSchema,
			},
		}
	}

	responses[successCode] = successResponse

	// Respuestas de error comunes
	errorResponses := map[string]string{
		"400": "Bad Request",
		"401": "Unauthorized",
		"403": "Forbidden",
		"404": "Not Found",
		"500": "Internal Server Error",
	}

	for code, description := range errorResponses {
		responses[code] = Response{
			Description: description,
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{
						Type: "object",
						Properties: map[string]Schema{
							"error": {
								Type: "string",
							},
							"message": {
								Type: "string",
							},
						},
					},
				},
			},
		}
	}

	return responses
}

func (g *OpenAPIGenerator) paramToSchema(param handler.ParamInfo) *Schema {
	openAPIType, format := g.coordinator.HandlerAnalyzer.GetOpenAPIType(param.Type)

	schema := &Schema{
		Type:   openAPIType,
		Format: format,
	}

	// Manejar arrays
	if strings.HasPrefix(param.Type, "[]") {
		schema.Type = "array"
		itemType := strings.TrimPrefix(param.Type, "[]")
		itemOpenAPIType, itemFormat := g.coordinator.HandlerAnalyzer.GetOpenAPIType(itemType)
		schema.Items = &Schema{
			Type:   itemOpenAPIType,
			Format: itemFormat,
		}
	}

	return schema
}

func (g *OpenAPIGenerator) returnTypeToSchema(returnType string) *Schema {
	openAPIType, format := g.coordinator.HandlerAnalyzer.GetOpenAPIType(returnType)

	schema := &Schema{
		Type:   openAPIType,
		Format: format,
	}

	// Para arrays, crear schema de array
	if strings.HasPrefix(returnType, "[]") {
		schema.Type = "array"
		itemType := strings.TrimPrefix(returnType, "[]")
		itemOpenAPIType, itemFormat := g.coordinator.HandlerAnalyzer.GetOpenAPIType(itemType)
		schema.Items = &Schema{
			Type:   itemOpenAPIType,
			Format: itemFormat,
		}
	}

	return schema
}

func (g *OpenAPIGenerator) generateSchemas(spec *OpenAPISpec, apiDesc *internal.APIDescription) {
	// Por ahora, esta es una implementación básica
	// En una implementación completa, analizaríamos todos los structs usados
	if spec.Components.Schemas == nil {
		spec.Components.Schemas = make(map[string]Schema)
	}

	// Agregar schema de error básico
	spec.Components.Schemas["Error"] = Schema{
		Type: "object",
		Properties: map[string]Schema{
			"error": {
				Type: "string",
			},
			"message": {
				Type: "string",
			},
			"code": {
				Type: "integer",
			},
		},
	}
}

func (g *OpenAPIGenerator) getParameterName(param handler.ParamInfo) string {
	if param.JSONName != "" {
		return param.JSONName
	}
	return param.Name
}

func (g *OpenAPIGenerator) generateParamDescription(param handler.ParamInfo) string {
	return fmt.Sprintf("%s parameter", param.Name)
}

func (g *OpenAPIGenerator) extractResourceName(path string) string {
	path = strings.Trim(path, "/")
	segments := strings.Split(path, "/")
	
	if len(segments) > 0 {
		// Remover parámetros {id}
		resource := segments[len(segments)-1]
		resource = strings.Trim(resource, "{}")
		return strings.Title(resource)
	}
	
	return "Resource"
}

func (g *OpenAPIGenerator) GenerateJSON(apiDesc *internal.APIDescription, title, version string) ([]byte, error) {
	spec := g.Generate(apiDesc, title, version)
	return json.MarshalIndent(spec, "", "  ")
}

func (g *OpenAPIGenerator) SaveToFile(apiDesc *internal.APIDescription, filename, title, version string) error {
	data, err := g.GenerateJSON(apiDesc, title, version)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}