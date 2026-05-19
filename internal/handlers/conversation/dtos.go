package conversation

import (
    "time"
    "codewax/internal/models"
)

type CreateConversationRequest struct {
    RepositoryIDs []uint `json:"repository_ids"`
}

type RepositoryResponse struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    GithubURL string    `json:"github_url"`
    Branch    string    `json:"branch"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type ConversationResponse struct {
    ID           uint                 `json:"id"`
    Title        string               `json:"title"`
    UserID       uint                 `json:"user_id"`
    Repositories []RepositoryResponse `json:"repositories"`
    CreatedAt    time.Time            `json:"created_at"`
    UpdatedAt    time.Time            `json:"updated_at"`
}

func NewRepositoryResponse(r models.Repository) RepositoryResponse {
    return RepositoryResponse{
        ID:        r.ID,
        Name:      r.Name,
        GithubURL: r.GithubURL,
        Branch:    r.Branch,
        CreatedAt: r.CreatedAt,
        UpdatedAt: r.UpdatedAt,
    }
}

func NewConversationResponse(c models.Conversation) ConversationResponse {
    repos := make([]RepositoryResponse, len(c.Repositories))
    for i, r := range c.Repositories {
        repos[i] = NewRepositoryResponse(r)
    }

    return ConversationResponse{
        ID:           c.ID,
        Title:        c.Title,
        UserID:       c.UserID,
        Repositories: repos,
        CreatedAt:    c.CreatedAt,
        UpdatedAt:    c.UpdatedAt,
    }
}