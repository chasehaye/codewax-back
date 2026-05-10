package routes

import (
	"github.com/gin-gonic/gin"
    "gorm.io/gorm"

	"codewax/internal/middleware"
    "codewax/internal/handlers/auth"
)

func Auth(r *gin.Engine, db *gorm.DB){

	guest := r.Group("/api")
    {
        guest.POST("/user/register", func(c *gin.Context) {auth.RegisterUser(c, db)})
        guest.POST("/user/login", func(c *gin.Context) {auth.LoginUser(c, db)})
    }




	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
    {
        protected.GET("/user/me", func(c *gin.Context) {auth.GetMe(c, db)})
        protected.POST("/user/logout", func(c *gin.Context) {auth.LogOut(c, db)})
    }
}
