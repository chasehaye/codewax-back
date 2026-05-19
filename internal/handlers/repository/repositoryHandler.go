package repository

import (
	"codewax/internal/dtos"
	"net/http"
	"errors"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
	"codewax/internal/models"
	"codewax/internal/services"
)


func CreateRepository(c *gin.Context, db *gorm.DB){
	userID := c.MustGet("userID").(uint)

	var input CreateRepositoryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ValidationErrorResponse{
			Error: "Validation failed missing fields",
			Details: map[string]string{
				"name": "Name is required and must be less than 255 characters",
				"github_url": "GithubURL is required",
				"branch": "github url branch must be specified",
			},
		})
		return
	}

	repo := models.Repository{
        Name:      input.Name,
        GithubURL: input.GithubURL,
        Branch:    input.Branch,
        UserID:    userID,
        Status:    "pending",
    }

    if err := db.Create(&repo).Error; err != nil {
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "failed to create repository",
        })
        return
    }

    c.JSON(http.StatusCreated, NewRepositoryResponse(repo))

    go services.ProcessRepository(db, repo)


}

func GetRepository(c *gin.Context, db *gorm.DB) {
    userID := c.MustGet("userID").(uint)
    repoID := c.Param("id")

    var repo models.Repository
    if err := db.Where("id = ? AND user_id = ?", repoID, userID).First(&repo).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, dtos.ServerErrorResponse{
                Error: "repository not found",
            })
            return
        }
        c.JSON(http.StatusInternalServerError, dtos.ServerErrorResponse{
            Error: "failed to fetch repository",
        })
        return
    }

    c.JSON(http.StatusOK, NewRepositoryResponse(repo))
}