package auth

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"fmt"


	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"codewax/internal/crypt"
	"codewax/internal/dtos"
	"codewax/internal/models"
    "codewax/internal/config"
	"codewax/internal/handlers/auth/validation"
	"codewax/internal/mailer"
)

// ForgotPassword godoc
// @Summary      Request Password Reset
// @Description  Sends a password reset link to the provided email if the account exists. 
// @Description  Always returns a success message to prevent email enumeration.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        input  body      ForgotPasswordInput  true  "User Email"
// @Success      200    {object}  ForgotPasswordResponse
// @Failure      400    {object}  dtos.ValidationErrorResponse
// @Failure      500    {object}  dtos.ServerErrorResponse
// @Router       /api/user/forgot-password [post]
func ForgotPassword(c *gin.Context, db *gorm.DB){
	var input ForgotPasswordInput
		if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Validation failed, incorrect format follow the specifications",
			Details: map[string]string{
				"email": "Email is required and must be a valid email address",
			},
		})
        return
	}

	cleanEmail, ok := validation.CleanAndValidateEmail(c, input.Email)
    if !ok {
        return 
    }
	var user models.User
    result := db.Where("email = ?", cleanEmail).Limit(1).Find(&user)
    if result.Error != nil {
        log.Printf("Database error during forgot password lookup for email=%s: %v", cleanEmail, result.Error)
        c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link"})
        return
    }
    if result.RowsAffected == 0 {
        c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link"})
        return
    }

	token, err := crypt.GenerateToken()
    if err != nil {
        c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link"})
        return
    }

	db.Model(&models.PasswordReset{}).
        Where("user_id = ? AND used = ?", user.ID, false).
        Update("used", true)
    resetRecord := models.PasswordReset{
        UserID:    user.ID,
        Token:     token,
        ExpiresAt: time.Now().Add(5 * time.Minute),
    }
    if err := db.Create(&resetRecord).Error; err != nil {
        log.Printf("Failed to create password reset record for userID=%d: %v", user.ID, err)
        c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link"})
        return
    }
    frontendURL := strings.TrimSuffix(os.Getenv("FRONTEND_URL"), "/")
    resetLink := fmt.Sprintf("%s/reset-password/%s", frontendURL, token)
    
    if err := mailer.SendResetEmail(cleanEmail, resetLink); err != nil {
        log.Printf("Failed to send reset email to %s: %v", cleanEmail, err)
        db.Delete(&resetRecord)
        c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link"})
        return
    }
    c.JSON(http.StatusOK, ForgotPasswordResponse{Message: "Check your inbox for a reset link",})
}

// ResetPassword godoc
// @Summary      Change Password
// @Description  Validates the reset token and updates the user's password.
// @Description  Automatically logs the user in by setting a session cookie upon success.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        token  path      string         true  "Reset Token"
// @Param        input  body      PasswordInput  true  "New Password"
// @Header       200    {string}  Set-Cookie     "token=jwt_value; HttpOnly; Secure; SameSite=Lax"
// @Success      200    {object}  ResetPasswordResponse
// @Failure      400    {object}  dtos.ValidationErrorResponse
// @Failure      401    {object}  dtos.UnauthorizedResponse
// @Failure      500    {object}  dtos.ServerErrorResponse
// @Router       /api/user/change-password/{token} [post]
func ResetPassword(c *gin.Context, db *gorm.DB) {
    token := c.Param("token")
    var input ResetPasswordInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Invalid format",
			Details: map[string]string{
				"password": "Password is required (min 8, max 72 characters)",
			},
		})
        return
    }
    
    var resetRecord models.PasswordReset
    result := db.Preload("User").Where("token = ? AND used = ?", token, false).Limit(1).Find(&resetRecord)
    if result.Error != nil {
        log.Printf("Database error during password reset lookup for token=%s: %v", token, result.Error)
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "An unexpected error occurred. Please try again later."})
        return
    }
    if result.RowsAffected == 0 {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Invalid or expired token"})
        return
    }

    if time.Now().After(resetRecord.ExpiresAt) {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Expired token"})
        return
    }

    hashedPassword, err := validation.HashPassword(input.Password)
    if err != nil {
        log.Printf("Bcrypt hashing failed: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Failed to process password"})
        return
    }

    err = db.Transaction(func(action *gorm.DB) error {
        if err := action.Model(&models.User{}).Where("id = ?", resetRecord.UserID).Update("password", hashedPassword).Error; err != nil {
            return err
        }
        if err := action.Model(&resetRecord).Update("used", true).Error; err != nil {
            return err
        }
        return nil
    })

    if err != nil {
        log.Printf("Failed to update password for userID=%d: %v", resetRecord.UserID, err)
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Could not complete password reset"})
        return
    }

	jwtToken, err := crypt.GenerateJWT(resetRecord.User.ID, resetRecord.User.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Password updated, please login manually"})
        return
    }

    c.SetCookie(
        "token",
        jwtToken,
        86400,
        "/",
        "",
        config.IsProduction(),
        true,
    )

    c.JSON(http.StatusOK, ResetPasswordResponse{
		Message: "Password updated successfully",
	})
}