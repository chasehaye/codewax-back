package validation

import (
	"regexp"
	"strings"
	"net/http"
	"github.com/gin-gonic/gin"
	"codewax/internal/dtos"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func CleanAndValidateEmail(c *gin.Context, rawEmail string) (string, bool) {
	cleanEmail := strings.ToLower(strings.TrimSpace(rawEmail))
	if len(cleanEmail) < 3 || len(cleanEmail) > 254 || !emailRegex.MatchString(cleanEmail) {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
            Error: "Validation failed",
            Details: map[string]string{
                "email": "Invalid email format or length",
            },
        })
        return "", false
    }

    return cleanEmail, true
}