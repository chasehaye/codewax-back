package routes

import (
	"github.com/gin-gonic/gin"
    "gorm.io/gorm"

	"codewax/internal/middleware"
    "codewax/internal/handlers/conversation"
)

func Conversation(r *gin.Engine, db *gorm.DB){
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
    {
        protected.POST("/conversations", func(c *gin.Context) {conversation.CreateConversation(c, db)})
        protected.GET("/conversations", func(c *gin.Context) {conversation.ListConversations(c, db)})
        protected.DELETE("/conversations/:id", func(c *gin.Context) {conversation.DeleteConversation(c, db)})
    }
}
