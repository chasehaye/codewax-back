package message

import (
	"time"
	"codewax/internal/models"
)
type CreateMessageRequest struct {
    Content string `json:"content" binding:"required"`
}

type MessageResponse struct {
    ID             uint      `json:"id"`
    Role           string    `json:"role"`
    Content        string    `json:"content"`
    ConversationID uint      `json:"conversation_id"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}

func NewMessageResponse(m models.Message) MessageResponse {
    return MessageResponse{
        ID:             m.ID,
        Role:           m.Role,
        Content:        m.Content,
        ConversationID: m.ConversationID,
        CreatedAt:      m.CreatedAt,
        UpdatedAt:      m.UpdatedAt,
    }
}