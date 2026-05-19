package routes

import (
	"github.com/gin-gonic/gin"
    "gorm.io/gorm"

	"codewax/internal/middleware"
	"codewax/internal/handlers/repository"
)

func Repository(r *gin.Engine, db *gorm.DB){
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
    {
        protected.POST("/repositories", func(c *gin.Context) { repository.CreateRepository(c, db) })
        protected.GET("/repositories/:id", func(c *gin.Context) { repository.GetRepository(c, db) })
    }
}
