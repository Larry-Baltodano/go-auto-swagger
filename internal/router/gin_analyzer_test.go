package router

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGinAnalyzer(t *testing.T) {
	testCode := `
package main

import "github.com/gin-gonic/gin"

type UserHandler struct{}

func (h *UserHandler) GetUser(c *gin.Context) {}

func CreateUser(c *gin.Context) {}

func main() {
	r := gin.Default()
	handler := &UserHandler{}
	
	// Rutas de prueba
	r.GET("/users/:id", handler.GetUser)
	r.POST("/users", CreateUser)
	r.PUT("/users/:id", handler.GetUser)
	r.DELETE("/users/:id", func(c *gin.Context) {})
}
	`

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "main.go")

	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	analyzer := NewGinAnalyzer()
	routes, err := analyzer.AnalyzeFile(testFile)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	if len(routes) != 4 {
		t.Errorf("Expected 4 routes, got %d", len(routes))
	}

	expectedRoutes := map[string]string{
		"GET": "/users/{id}",
		"POST": "/users",
		"PUT": "/users/{id}",
		"DELETE": "/users/{id}",
	}

	for _, route := range routes {
		expectedPath, exists := expectedRoutes[route.Method]
		if !exists {
			t.Errorf("Unexpected method: %s", route.Method)
			continue
		}

		if route.Path != expectedPath {
			t.Errorf("For method %s, expected path %s, got %s", route.Method, expectedPath, route.Path)
		}
	}

	for _, route := range routes {
		switch route.Method {
		case "GET", "PUT":
			if route.HandlerType != "method" {
				t.Errorf("Expected method handler for %s, got %s", route.Method, route.HandlerType)
			}
		case "POST":
			if route.HandlerType != "function" {
				t.Errorf("Expected function handler for POST, got %s", route.HandlerType)
			}
		case "DELETE":
			if route.HandlerType != "anonymous" {
				t.Errorf("Expected anonymous handler for DELETE, got %s", route.HandlerType)
			}
		}
	}
}