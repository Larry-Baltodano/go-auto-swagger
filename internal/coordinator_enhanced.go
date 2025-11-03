package internal

import (
	"path/filepath"
	"strings"

	"github.com/Larry-Baltodano/go-auto-swagger/internal/handler"
	"github.com/Larry-Baltodano/go-auto-swagger/internal/router"
)

type EnhancedCoordinator struct {
	RouterAnalyzer  *router.GinAnalyzer
	HandlerAnalyzer *handler.EnhancedHandlerAnalyzer
}

func NewEnhancedCoordinator() *EnhancedCoordinator {
	return &EnhancedCoordinator{
		RouterAnalyzer:  router.NewGinAnalyzer(),
		HandlerAnalyzer: handler.NewEnhancedHandlerAnalyzer(),
	}
}

func (c *EnhancedCoordinator) AnalyzeAPI(sourceDir string) (*APIDescription, error) {
	routes, err := c.RouterAnalyzer.AnalyzeDirectory(sourceDir)
	if err != nil {
		return nil, err
	}

	apiDesc := &APIDescription{
		Routes: make([]RouteDescription, 0, len(routes)),
	}

	for _, route := range routes {
		routeDesc := RouteDescription{
			Method:  route.Method,
			Path:    route.Path,
			Handler: route.Handler,
			File:    route.File,
		}

		handlerInfo, err := c.analyzeHandlerEnhanced(route.Handler, route.File)
		if err == nil && handlerInfo != nil {
			routeDesc.HandlerInfo = handlerInfo
		}

		apiDesc.Routes = append(apiDesc.Routes, routeDesc)
	}

	return apiDesc, nil
}

func (c *EnhancedCoordinator) analyzeHandlerEnhanced(handlerName, filePath string) (*handler.HandlerInfo, error) {
	packageName := filepath.Base(filepath.Dir(filePath))
	parts := strings.Split(handlerName, ".")
	
	if len(parts) == 1 {
		return c.HandlerAnalyzer.AnalyzeFunction(packageName, filePath, parts[0])
	} else if len(parts) == 2 {
		structName := c.inferStructName(parts[0])
		return c.HandlerAnalyzer.AnalyzeMethod(packageName, filePath, structName, parts[1])
	}

	return nil, nil
}

func (c *EnhancedCoordinator) inferStructName(handlerVar string) string {
	switch handlerVar {
	case "handler", "h", "userHandler":
		return "UserHandler"
	case "productHandler":
		return "ProductHandler"
	default:
		if len(handlerVar) > 0 {
			return strings.Title(handlerVar)
		}
		return "Handler"
	}
}