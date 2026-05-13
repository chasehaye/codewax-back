package workspace

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"codewax/internal/models"
	"codewax/internal/dtos"
)

func InitWorkspace(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(uint)

	workspace := models.Workspace{
		UserID: userID,
	}

	if err := db.Create(&workspace).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "failed to create workspace"})
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

func DeleteWorkspace(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(uint)
	workspaceID := c.Param("id")

	result := db.Where("id = ? AND user_id = ?", workspaceID, userID).Delete(&models.Workspace{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{Error: "failed to delete workspace"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dtos.NotFoundErrorResponse{Error: "workspace not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workspace deleted"})
}