package middleware

import (
	"codewax/internal/crypt"
    "codewax/internal/dtos"
    "net/http"
	"github.com/gin-gonic/gin"

    

)


func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString, err := c.Cookie("token")
        
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
				Error: "Authentication required",
			})
            return
        }
        token, err := crypt.ValidateJWT(tokenString)
        if err != nil || token == nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
				Error: "Invalid or expired session",
			})
            return
        }
        uid, err := crypt.GetUserIDFromJWT(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{
				Error: "Failed to identify user from session",
			})
			return
		}
        c.Set("userID", uid)
        c.Next()
    }
}