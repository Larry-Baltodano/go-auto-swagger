package main

import "github.com/gin-gonic/gin"

// User representa un usuario
type User struct {
	ID    int    `json:"id" example:"1"`
	Name  string `json:"name" example:"John Doe"`
	Email string `json:"email" example:"john@example.com"`
}

// CreateUserRequest define el request para crear usuario
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required" example:"John Doe"`
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
	Age   int    `json:"age" example:"30"`
}

// UpdateUserRequest define el request para actualizar usuario
type UpdateUserRequest struct {
	ID    int    `uri:"id" binding:"required"`
	Name  string `json:"name" example:"John Doe Updated"`
	Email string `json:"email" example:"john.updated@example.com"`
}

// QueryParams define parámetros de query
type QueryParams struct {
	Limit  int    `form:"limit" example:"10"`
	Offset int    `form:"offset" example:"0"`
	Search string `form:"search" example:"john"`
}

type UserHandler struct{}

// GetUser obtiene un usuario por ID - CON STRUCT EXPLÍCITO PARA PATH
func (h *UserHandler) GetUser(c *gin.Context) {
	var req struct {
		ID int `uri:"id" binding:"required"`
	}
	
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	user := User{ID: req.ID, Name: "John", Email: "john@example.com"}
	c.JSON(200, user)
}

// CreateUser crea un nuevo usuario - CON STRUCT COMPLEJO
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	user := User{ID: 1, Name: req.Name, Email: req.Email}
	c.JSON(201, user)
}

// UpdateUser actualiza un usuario - COMBINA PATH Y BODY
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var pathParams struct {
		ID int `uri:"id" binding:"required"`
	}
	
	var bodyParams UpdateUserRequest
	
	if err := c.ShouldBindUri(&pathParams); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	user := User{
		ID:    pathParams.ID,
		Name:  bodyParams.Name,
		Email: bodyParams.Email,
	}
	c.JSON(200, user)
}

// ListUsers lista usuarios con filtros - USA QUERY PARAMS
func (h *UserHandler) ListUsers(c *gin.Context) {
	var query QueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	users := []User{
		{ID: 1, Name: "John", Email: "john@example.com"},
		{ID: 2, Name: "Jane", Email: "jane@example.com"},
	}
	c.JSON(200, users)
}

func main() {
	r := gin.Default()
	handler := &UserHandler{}
	
	// Registrar rutas
	r.GET("/users/:id", handler.GetUser)
	r.POST("/users", handler.CreateUser)
	r.PUT("/users/:id", handler.UpdateUser)
	r.GET("/users", handler.ListUsers)
	
	r.Run(":8080")
}