package models

import (
    "time"
    "gorm.io/gorm"
)

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

type User struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time  

    Name      string    `gorm:"type:varchar(255)"`
    Password  string    `gorm:"not null"`
    Token     string    `gorm:"type:text"`
    Email     string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	IsAdmin   bool      `gorm:"default:false"`

    Workspaces []Workspace `gorm:"foreignKey:UserID"`
}

type Workspace struct {
    ID        uint           `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`

    UserID uint `gorm:"index;not null"`
    User   User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

    Repositories []Repository `gorm:"foreignKey:WorkspaceID"`
    Conversations []Conversation `gorm:"foreignKey:WorkspaceID"`
}


type Conversation struct {
    ID        uint           `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`

    Title string `gorm:"type:varchar(255)"`

    WorkspaceID uint    `gorm:"index;not null"`
    Workspace   Workspace `gorm:"foreignKey:WorkspaceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

    Messages []Message `gorm:"foreignKey:ConversationID"`
}

type Message struct {
    ID        uint           `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`

    Role    string `gorm:"type:varchar(50);not null"` // "user" | "assistant"
    Content string `gorm:"type:text;not null"`

    ConversationID uint         `gorm:"index;not null"`
    Conversation   Conversation `gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}



type Repository struct {
    ID        uint           `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`

    Name       string `gorm:"type:varchar(255);not null"`
    GithubURL  string `gorm:"type:text"`
    Branch     string `gorm:"default:main"`

    WorkspaceID uint    `gorm:"index;not null"`
    Workspace   Workspace `gorm:"foreignKey:WorkspaceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}