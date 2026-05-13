package conversation


import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"codewax/internal/models"
	"codewax/internal/dtos"
)


func InitConversation(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(uint)

	var workspace models.Workspace
	if err := db.Where("user_id = ?", userID).First(&workspace).Error; err != nil {
		c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{Error: "workspace not found"})
		return
	}

	conversation := models.Conversation{
		WorkspaceID: workspace.ID,
	}

	if err := db.Create(&conversation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "failed to create conversation"})
		return
	}

	c.JSON(http.StatusCreated, conversation)
}