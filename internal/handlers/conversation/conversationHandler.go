package conversation


import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"errors"

	"codewax/internal/models"
	"codewax/internal/dtos"
)


func CreateConversation(c *gin.Context, db *gorm.DB) {
    userID := c.MustGet("userID").(uint)

    var input CreateConversationRequest
    c.ShouldBindJSON(&input)

    conversation := models.Conversation{
        UserID: userID,
    }

    if len(input.RepositoryIDs) > 0 {
        var repos []models.Repository
        if err := db.Where("id IN ? AND user_id = ?", input.RepositoryIDs, userID).Find(&repos).Error; err != nil {
            c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
                Error: "failed to fetch repositories",
            })
            return
        }
        conversation.Repositories = repos
    }

    if err := db.Create(&conversation).Error; err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "failed to create conversation",
        })
        return
    }

    c.JSON(http.StatusCreated, NewConversationResponse(conversation))
}

func ListConversations(c *gin.Context, db *gorm.DB) {
    userID := c.MustGet("userID").(uint)

    var conversations []models.Conversation
    if err := db.Where("user_id = ?", userID).
        Preload("Repositories").
        Order("updated_at DESC").
        Find(&conversations).Error; err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "failed to fetch conversations",
        })
        return
    }

    response := make([]ConversationResponse, len(conversations))
    for i, conv := range conversations {
        response[i] = NewConversationResponse(conv)
    }

    c.JSON(http.StatusOK, response)
}

func DeleteConversation(c *gin.Context, db *gorm.DB) {
    userID := c.MustGet("userID").(uint)
    conversationID := c.Param("id")

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

    if err := db.Delete(&conversation).Error; err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "failed to delete conversation",
        })
        return
    }

    c.Status(http.StatusNoContent)
}