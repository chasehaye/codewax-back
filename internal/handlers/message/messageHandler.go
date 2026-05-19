package message

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"errors"

	"codewax/internal/models"
	"codewax/internal/dtos"
)

func CreateMessage(c *gin.Context, db *gorm.DB) {
    userID := c.MustGet("userID").(uint)
    conversationID := c.Param("id")

    var input CreateMessageRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
            Error: "Validation failed",
            Details: map[string]string{
                "content": "content is required",
            },
        })
        return
    }

    var conversation models.Conversation
    if err := db.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, dtos.ServerErrorResponse{
                Error: "conversation not found",
            })
            return
        }
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "failed to fetch conversation",
        })
        return
    }

    message := models.Message{
        ConversationID: conversation.ID,
        Role:           "user",
        Content:        input.Content,
    }

    if err := db.Create(&message).Error; err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "failed to create message",
        })
        return
    }

	if conversation.Title == "" {
        title := input.Content
        if len(title) > 255 {
            title = title[:255]
        }
        db.Model(&conversation).Update("title", title)
    }

    c.JSON(http.StatusCreated, NewMessageResponse(message))
}