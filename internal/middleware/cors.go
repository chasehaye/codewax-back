package middleware

import (
    "net/http"
    "os"
    "strings"

    "github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")

        if strings.HasPrefix(c.Request.URL.Path, "/api/public/") {
            c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
            c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
            c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        } else {
            allowedOrigin := os.Getenv("FRONTEND_URL")
            if allowedOrigin != "" && allowedOrigin == origin {
                c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
                c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
            }
            c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
            c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        }

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}