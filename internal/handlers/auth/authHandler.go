package auth

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"codewax/internal/crypt"
	"codewax/internal/dtos"
	"codewax/internal/models"
    "codewax/internal/config"
	"codewax/internal/handlers/auth/validation"
)

var (
    adminEmail string
	cookieDomain string
)

func init() {
    _ = godotenv.Load()
    adminEmail = strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL")))

}










// CreateUser godoc
// @Summary      Register New User
// @Description  Creates a user account, hashes the password
// @Description  Returns and sets an HttpOnly 'token' cookie.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterInput  true  "User Registration Data"
// @Header       201   {string}  Set-Cookie     "Contains the JWT session token (HttpOnly, Secure)"
// @Success      201   {object}  RegisterResponse
// @Failure      400   {object}  dtos.ValidationErrorResponse "Validation failed"
// @Failure      409   {object}  dtos.AlreadyExistsResponse   "User already exists with email"
// @Failure      500   {object}  dtos.ServerErrorResponse     "Server error"
// @Router       /api/user/register [post]
func RegisterUser(c *gin.Context, db *gorm.DB) {
    var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Validation failed, incorrect format follow the specifications",
			Details: map[string]string{
				"email":    "Email is required and must be a valid email",
				"password": "Password is required (min 8, max 72 characters)",
				"name":     "Name is optional but must be less than 255 characters",
			},
		})
        return
	}

	cleanEmail, ok := validation.CleanAndValidateEmail(c, input.Email)
    if !ok {
        return 
    }

    var existingUser models.User
    result := db.Where("email = ?", cleanEmail).Limit(1).Find(&existingUser)
    if result.Error != nil {
        log.Printf("Database lookup error: %v", result.Error)
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "An unexpected error occurred - Please try again later"})
        return
    }
    if result.RowsAffected > 0 {
        c.JSON(http.StatusConflict, dtos.AlreadyExistsResponse{Error: "Email is already in use - try again"})
        return
    }

    displayName := strings.TrimSpace(input.Name)
    if displayName == "" {
		displayName = "User" 
	}

    if err := validation.ValidatePassword(input.Password); err != nil {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{Error: "Invalid password",})
        return
    }
	hashedPassword, err := validation.HashPassword(input.Password)
	if err != nil {
        log.Printf("Password hashing error: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "An unexpected error occurred - Please try again later"})
		return
	}
	isAdmin := false
	if cleanEmail == adminEmail {
		isAdmin = true
	}

    user := models.User{
        Name:     displayName,
        Email:    cleanEmail,
        Password: string(hashedPassword),
        IsAdmin:  isAdmin,
    }
    if err := db.Create(&user).Error; err != nil {
        log.Printf("Failed to create user in database: %v", err)
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Could not complete registration"})
        return
    }

    jwtToken, err := crypt.GenerateJWT(user.ID, user.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Failed to start session"})
        return
    }

    c.SetCookie(
        "token",                // Name
        jwtToken,               // Value
        86400,                  // MaxAge (24 hours in seconds)
        "/",                    // Path
        "",           // Domain (leave empty for current domain)
        config.IsProduction(),  // Secure (SET TO TRUE IN PRODUCTION/HTTPS)
        true,                   // HttpOnly (CRITICAL: prevents JS access)
    )

    c.JSON(http.StatusCreated, RegisterResponse{
		Message:   "success",
		IsAdmin:   isAdmin,
		UserEmail: cleanEmail,
		UserName:  displayName,
	})
}


// LoginUser godoc
// @Summary     Login Existing User
// @Description Login for exisitng user using email and passwrod (compares the stored hash to the input).
// @Description Returns a success message and sets an HttpOnly 'token' cookie.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        user  body      LoginInput  true  "User Login Credentials"
// @Header       201   {string}  Set-Cookie     "Contains the JWT session token (HttpOnly, Secure)"
// @Success      201   {object}  LoginResponse
// @Failure      400   {object}  dtos.ValidationErrorResponse "Invalid input or malformed email"
// @Failure      401   {object}  dtos.UnauthorizedResponse "Validation failed"
// @Failure      500   {object}  dtos.ServerErrorResponse     "Server error"
// @Router       /api/user/login [post]
func LoginUser(c *gin.Context, db *gorm.DB) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Validation failed, incorrect format follow the specifications",
			Details: map[string]string{
				"email":    "Email is required and must be a valid email",
				"password": "Password is required (min 8, max 72 characters)",
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
	if result.Error != nil || result.RowsAffected == 0 {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Invalid credentials"})
        return
    }

    if err := validation.ComparePassword(user.Password, input.Password); err != nil {
        c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Invalid credentials"})
        return
    }

	jwtToken, err := crypt.GenerateJWT(user.ID, user.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Failed to create session"})
        return
    }
    log.Printf("IsProduction: %v", config.IsProduction())
    c.SetCookie(
        "token",
        jwtToken,
        86400,
        "/",
        "",
        config.IsProduction(),
        true,
    )

	c.JSON(http.StatusOK, LoginResponse{
		Message:   "success",
		IsAdmin:   user.IsAdmin,
		UserEmail: user.Email,
		UserName:  user.Name,
	})
}


// LogOut godoc
// @Summary      User Logout
// @Description  Invalidates the session by clearing the 'token' cookie.
// @Tags         accounts
// @Produce      json
// @Success      200  {object}  LogOutResponse
// @Header       200  {string}  Set-Cookie  "token=; Max-Age=0; Path=/; HttpOnly; Secure"
// @Failure      401  {object}  dtos.UnauthorizedResponse "Session expired or invalid"
// @Failure      500  {object}  dtos.ServerErrorResponse
// @Router       /api/user/logout [post]
func LogOut(c *gin.Context, db *gorm.DB) {
    c.SetCookie("token", "", -1, "/", "", config.IsProduction(), true)

    c.JSON(http.StatusOK, LogOutResponse{Message: "Successfully logged out",})
}


// GetMe godoc
// @Summary      Get Current User Info
// @Description  Returns the details of the authenticated user based on the session cookie.
// @Tags         accounts
// @Produce      json
// @Success      200  {object}  LoginResponse
// @Failure      401  {object}  dtos.UnauthorizedResponse
// @Failure      500  {object}  dtos.ServerErrorResponse
// @Router       /api/user/me [get]
func GetMe(c *gin.Context, db *gorm.DB) {
	uidInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.UnauthorizedResponse{Error: "Session context missing"})
		return
	}
	uid, ok := uidInterface.(uint)
	if !ok {
		log.Printf("Type assertion failed: userID in context is %T, expected uint", uidInterface)
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "Internal configuration error"})
		return
	}
	var user models.User
	if err := db.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{Error: "User not found"})
		return
	}

    c.JSON(http.StatusOK, GetMeResponse{
        ID:        user.ID,
        UserName:  user.Name,
        UserEmail: user.Email,
        IsAdmin:   user.IsAdmin,
    })
}