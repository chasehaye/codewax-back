package repository

import (
	"time"
	"codewax/internal/models"
)

type CreateRepositoryRequest struct {
	Name      string `json:"name"`
	GithubURL string `json:"github_url"`
	Branch    string `json:"branch"`
}

type RepositoryResponse struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    GithubURL string    `json:"github_url"`
    Branch    string    `json:"branch"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func NewRepositoryResponse(r models.Repository) RepositoryResponse {
    return RepositoryResponse{
        ID:        r.ID,
        Name:      r.Name,
        GithubURL: r.GithubURL,
        Branch:    r.Branch,
        Status:    r.Status,
        CreatedAt: r.CreatedAt,
        UpdatedAt: r.UpdatedAt,
    }
}