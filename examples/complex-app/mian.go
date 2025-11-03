package main

import "github.com/gin-gonic/gin"

// User representa un usuario completo
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	IsActive bool   `json:"is_active"`
}

// Product representa un producto
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// CreateUserRequest para crear usuario
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"min=0"`
}

// UpdateProductRequest para actualizar producto
type UpdateProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price" binding:"min=0"`
	Stock int     `json:"stock" binding:"min=0"`
}

// QueryFilters para filtros de búsqueda
type QueryFilters struct {
	Limit   int    `form:"limit" binding:"min=1,max=100"`
	Offset  int    `form:"offset" binding:"min=0"`
	Search  string `form:"search"`
	SortBy  string `form:"sort_by"`
	SortDir string `form:"sort_dir" binding:"oneof=asc desc"`
}

type APIHandler struct{}

// GetUserByID obtiene usuario por ID
func (h *APIHandler) GetUserByID(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required,min=1"`
	}
	
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	user := User{ID: req.ID, Name: "John Doe", Email: "john@example.com", Age: 30, IsActive: true}
	c.JSON(200, user)
}

// CreateUser crea nuevo usuario
func (h *APIHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	user := User{ID: 1, Name: req.Name, Email: req.Email, Age: req.Age, IsActive: true}
	c.JSON(201, user)
}

// ListUsers lista usuarios con filtros
func (h *APIHandler) ListUsers(c *gin.Context) {
	var query QueryFilters
	
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	users := []User{
		{ID: 1, Name: "John Doe", Email: "john@example.com", Age: 30, IsActive: true},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Age: 25, IsActive: true},
	}
	c.JSON(200, users)
}

// UpdateProduct actualiza producto
func (h *APIHandler) UpdateProduct(c *gin.Context) {
	var pathParams struct {
		ID int `uri:"id" binding:"required,min=1"`
	}
	
	var bodyParams UpdateProductRequest
	
	if err := c.ShouldBindUri(&pathParams); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	product := Product{
		ID:    pathParams.ID,
		Name:  bodyParams.Name,
		Price: bodyParams.Price,
		Stock: bodyParams.Stock,
	}
	c.JSON(200, product)
}

// GetProductByID obtiene producto por ID
func (h *APIHandler) GetProductByID(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required,min=1"`
	}
	
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	product := Product{ID: req.ID, Name: "Laptop", Price: 999.99, Stock: 10}
	c.JSON(200, product)
}

func main() {
	r := gin.Default()
	handler := &APIHandler{}
	
	// User routes
	r.GET("/users/:id", handler.GetUserByID)
	r.POST("/users", handler.CreateUser)
	r.GET("/users", handler.ListUsers)
	
	// Product routes  
	r.GET("/products/:id", handler.GetProductByID)
	r.PUT("/products/:id", handler.UpdateProduct)
	
	r.Run(":8080")
}