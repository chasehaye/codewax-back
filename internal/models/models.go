package models

import (
    "time"
    "gorm.io/gorm"
)


type User struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time  

    Name      string    `gorm:"type:varchar(255)"`
    Password  string    `gorm:"not null"`
    Token     string    `gorm:"type:text"`
    Email     string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	IsAdmin   bool      `gorm:"default:false"`
}

type PasswordReset struct {
    gorm.Model
    UserID    uint      `gorm:"index"`
    Token     string    `gorm:"uniqueIndex"`
    ExpiresAt time.Time `gorm:"index"`
    Used      bool      `gorm:"default:false"`
    User      User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type EmailChangeRequest struct {
    gorm.Model
    UserID    uint      `gorm:"index"`
    NewEmail  string    `gorm:"type:varchar(255);not null"`
    Token     string    `gorm:"uniqueIndex"`
    ExpiresAt time.Time `gorm:"index"`
    Used      bool      `gorm:"default:false"`
    User      User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}