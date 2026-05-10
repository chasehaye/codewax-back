package auth


type RegisterInput struct {
	Name     string `json:"name" binding:"max=255" example:"User Name"`
	Email    string `json:"email" binding:"required,email,max=255" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8,max=72" example:"SecurePass123!"`
}

type RegisterResponse struct {
	Message   string `json:"message" example:"success"`
	IsAdmin   bool   `json:"is_admin" example:"false"`
	UserEmail string `json:"user_email" example:"user@example.com"`
	UserName  string `json:"user_name" example:"User Name"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email,max=255" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8,max=72" example:"SecurePass123!"`
}

type LoginResponse struct {
    Message   string `json:"message" example:"success"`
    IsAdmin   bool   `json:"is_admin" example:"false"`
    UserEmail string `json:"user_email" example:"user@example.com"`
    UserName  string `json:"user_name" example:"User Name"`
}

type LogOutResponse struct {
	Message  string `json:"message" example:"success"`
}

type GetMeResponse struct {
    ID        uint   `json:"id" example:"1"`
    UserName  string `json:"user_name" example:"User Name"`
    UserEmail string `json:"user_email" example:"user@example.com"`
    IsAdmin   bool   `json:"is_admin" example:"false"`
}

type ForgotPasswordInput struct {
	Email    string `json:"email" binding:"required,email,max=255" example:"user@example.com"`
}

type ForgotPasswordResponse struct {
	Message  string `json:"message" example:"Check your inbox for a reset link"`
}

type ResetPasswordInput struct {
	Password string `json:"password" binding:"required,min=8,max=72" example:"SecurePass123!"`
}

type ResetPasswordResponse struct {
	Message  string `json:"message" example:"Password updated successfully"`
}