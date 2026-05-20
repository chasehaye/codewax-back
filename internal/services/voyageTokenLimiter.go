package services

import (
	"fmt"
	"time"

	"codewax/internal/models"
	"gorm.io/gorm"
)

const MonthlyTokenBudget = 55_000_000 // ~$10 at $0.18/1M tokens

func currentMonth() int {
	now := time.Now()
	return now.Year()*100 + int(now.Month())
}

func RecordTokenUsage(db *gorm.DB, userID uint, repoID uint, tokens int) error {
	usage := models.TokenUsage{
		UserID: userID,
		RepoID: repoID,
		Tokens: tokens,
		Month:  currentMonth(),
	}
	if err := db.Create(&usage).Error; err != nil {
		return fmt.Errorf("failed to record token usage: %w", err)
	}
	return nil
}

func GetMonthlyTokenUsage(db *gorm.DB) (int, error) {
	var total int
	err := db.Model(&models.TokenUsage{}).
		Where("month = ?", currentMonth()).
		Select("COALESCE(SUM(tokens), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get monthly token usage: %w", err)
	}
	return total, nil
}

func ExceedsMonthlyBudget(db *gorm.DB) (bool, error) {
	total, err := GetMonthlyTokenUsage(db)
	if err != nil {
		return false, err
	}
	return total >= MonthlyTokenBudget, nil
}