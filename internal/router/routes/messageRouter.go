package routes

import (
	"github.com/gin-gonic/gin"
    "gorm.io/gorm"

	"codewax/internal/middleware"
	"codewax/internal/handlers/message"
)

func Message(r *gin.Engine, db *gorm.DB){
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
    {
        protected.POST("/conversations/:id/messages", func(c *gin.Context) {message.CreateMessage(c, db)})
    }
}
