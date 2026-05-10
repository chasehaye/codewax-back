package router

import (
	"log"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "codewax/internal/router/routes"
	"codewax/internal/config"
    "codewax/internal/middleware"


    _ "codewax/docs"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(db *gorm.DB) *gin.Engine {
    if config.IsProduction() {
        log.Println("Running in Production mode")
        gin.SetMode(gin.ReleaseMode)
    } else {
        log.Println("Running in Development mode")
        gin.SetMode(gin.DebugMode)
    }

    r := gin.New()
    r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
    r.Use(gin.Recovery())
    if !config.IsProduction() {
        r.Use(gin.Logger())
    }

    r.Use(middleware.CORSMiddleware())

    routes.Auth(r, db)


    if !config.IsProduction() {
        r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }

    return r
}

// public  // open to anyone, API key auth
// guest   // frontend, no token needed  
// auth    // frontend, token required
// admin   // frontend, token + admin