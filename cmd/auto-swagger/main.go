package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Larry-Baltodano/go-auto-swagger/internal"
	"github.com/Larry-Baltodano/go-auto-swagger/internal/generator"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	sourceDir := os.Args[1]
	outputFile := "openapi.json"
	title := "Auto-Generated API"
	version := "1.0.0"

	// Parsear argumentos opcionales
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-o", "--output":
			if i+1 < len(os.Args) {
				outputFile = os.Args[i+1]
				i++
			}
		case "-t", "--title":
			if i+1 < len(os.Args) {
				title = os.Args[i+1]
				i++
			}
		case "-v", "--version":
			if i+1 < len(os.Args) {
				version = os.Args[i+1]
				i++
			}
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		}
	}

	// Asegurar extensión .json
	if !strings.HasSuffix(outputFile, ".json") {
		outputFile += ".json"
	}

	fmt.Printf("🔍 Auto-Swagger analyzing: %s\n", sourceDir)
	fmt.Printf("📄 Output: %s\n", outputFile)
	fmt.Printf("📝 Title: %s\n", title)
	fmt.Printf("🔢 Version: %s\n\n", version)

	// Analizar API completa
	coordinator := internal.NewEnhancedCoordinator()
	apiDesc, err := coordinator.AnalyzeAPI(sourceDir)
	if err != nil {
		fmt.Printf("❌ Error analyzing API: %v\n", err)
		os.Exit(1)
	}

	if len(apiDesc.Routes) == 0 {
		fmt.Println("❌ No Gin routes found!")
		os.Exit(1)
	}

	// Mostrar resumen del análisis
	fmt.Printf("✅ Found %d routes\n", len(apiDesc.Routes))
	
	handlersAnalyzed := 0
	for _, route := range apiDesc.Routes {
		if route.HandlerInfo != nil {
			handlersAnalyzed++
		}
	}
	fmt.Printf("🔧 Handlers analyzed: %d\n", handlersAnalyzed)

	// Generar OpenAPI spec
	fmt.Println("\n🚀 Generating OpenAPI specification...")
	
	openapiGenerator := generator.NewOpenAPIGenerator(coordinator)
	
	// Guardar archivo
	if err := openapiGenerator.SaveToFile(apiDesc, outputFile, title, version); err != nil {
		fmt.Printf("❌ Error generating OpenAPI: %v\n", err)
		os.Exit(1)
	}

	// Mostrar información del spec generado
	specData, _ := openapiGenerator.GenerateJSON(apiDesc, title, version)
	var spec map[string]interface{}
	json.Unmarshal(specData, &spec)

	fmt.Printf("🎉 OpenAPI specification generated successfully!\n")
	fmt.Printf("📊 Generated Spec:\n")
	fmt.Printf("   • Paths: %d\n", len(spec["paths"].(map[string]interface{})))
	
	if components, exists := spec["components"]; exists {
		if schemas, exists := components.(map[string]interface{})["schemas"]; exists {
			fmt.Printf("   • Schemas: %d\n", len(schemas.(map[string]interface{})))
		}
	}

	// Mostrar endpoints generados
	fmt.Printf("\n📋 Generated Endpoints:\n")
	paths := spec["paths"].(map[string]interface{})
	for path, pathItem := range paths {
		operations := pathItem.(map[string]interface{})
		for method := range operations {
			fmt.Printf("   • %s %s\n", strings.ToUpper(method), path)
		}
	}

	fmt.Printf("\n📁 File: %s\n", outputFile)
	fmt.Printf("🎯 OpenAPI 3.0.3 compliant\n")
	fmt.Printf("🚀 You can now use this with Swagger UI or other OpenAPI tools\n")
}

func printUsage() {
	fmt.Println("Usage: auto-swagger <source-directory> [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -o, --output FILE    Output file (default: openapi.json)")
	fmt.Println("  -t, --title TITLE    API title (default: 'Auto-Generated API')")
	fmt.Println("  -v, --version VER    API version (default: '1.0.0')")
	fmt.Println("  -h, --help           Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  auto-swagger ./examples/basic-app")
	fmt.Println("  auto-swagger . -o myapi.json -t \"My API\" -v 2.0.0")
	fmt.Println("  auto-swagger ./internal/api --output docs/openapi.json")
}