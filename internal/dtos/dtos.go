package dtos

// --- 400 + 500 Response DTOs ------------------------------------------------------------------------------------

// ValidationErrorResponse for 400 errors with field-specific details
type ValidationErrorResponse struct {
	Error   string            `json:"error" example:"Validation failed"`
	Details map[string]string `json:"details"`
}

// UnauthorizedResponse for 401 authentication failures
type UnauthorizedResponse struct {
    Error string `json:"error" example:"Invalid credentials or session"`
}

// ForbiddenResponse for 403 Autheticated but not allowed to perform action
type ForbiddenResponse struct {
	Error string `json:"error" example:"You do not have permission to perform this action"`
}

// NotFoundErrorResponse for 404 resource not found errors
type NotFoundErrorResponse struct {
    Error string `json:"error" example:"Resource not found"`
}

// AlreadyExistsResponse for 409 Conflict errors
type AlreadyExistsResponse struct {
	Error string `json:"error" example:"Already in use"`
}

// ServerErrorResponse for 500 Internal Server errors
type ServerErrorResponse struct {
	Error string `json:"error" example:"An unexpected error occurred"`
}