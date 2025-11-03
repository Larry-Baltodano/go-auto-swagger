package internal

import (
	"path/filepath"
	"strings"

	"github.com/Larry-Baltodano/go-auto-swagger/internal/handler"
	"github.com/Larry-Baltodano/go-auto-swagger/internal/router"
)

type APIDescription struct {
	Routes []RouteDescription
}

type RouteDescription struct {
	Method      string
	Path        string
	Handler     string
	HandlerInfo *handler.HandlerInfo
	File        string
}

type Coordinator struct {
	routerAnalyzer  *router.GinAnalyzer
	handlerAnalyzer *handler.HandlerAnalyzer
}

func NewCoordinator() *Coordinator {
	return &Coordinator{
		routerAnalyzer:  router.NewGinAnalyzer(),
		handlerAnalyzer: handler.NewHandlerAnalyzer(),
	}
}

func (c *Coordinator) AnalyzeAPI(sourceDir string) (*APIDescription, error) {
	routes, err := c.routerAnalyzer.AnalyzeDirectory(sourceDir)
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

		handlerInfo, err := c.analyzeHandler(route.Handler, route.File)
		if err == nil && handlerInfo != nil {
			routeDesc.HandlerInfo = handlerInfo
		}

		apiDesc.Routes = append(apiDesc.Routes, routeDesc)
	}

	return apiDesc, nil
}

func (c *Coordinator) analyzeHandler(handlerName, filePath string) (*handler.HandlerInfo, error) {
	packageName := filepath.Base(filepath.Dir(filePath))

	parts := strings.Split(handlerName, ".")
	
	if len(parts) == 1 {
		return c.handlerAnalyzer.AnalyzeFunction(packageName, filePath, parts[0])
	} else if len(parts) == 2 {
		return c.handlerAnalyzer.AnalyzeMethod(packageName, filePath, "UserHandler", parts[1])
	}

	return nil, nil
}