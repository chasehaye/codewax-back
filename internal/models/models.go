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

    Conversations []Conversation `gorm:"foreignKey:UserID"`
    Repositories  []Repository   `gorm:"foreignKey:UserID"`
}

type Conversation struct {
    ID        uint           `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`

    Title string `gorm:"type:varchar(255)"`
    UserID uint   `gorm:"index;not null"`
    User   User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

    Messages     []Message    `gorm:"foreignKey:ConversationID"`
    Repositories []Repository `gorm:"many2many:conversation_repositories;"`
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
    Status    string `gorm:"type:varchar(50);default:'pending'"` // "pending" | "processing" | "ready" | "failed"

    UserID uint `gorm:"index;not null"`
    User   User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

    Conversations []Conversation `gorm:"many2many:conversation_repositories;"`
    Chunks        []Chunk `gorm:"foreignKey:RepositoryID"`
}

type Chunk struct {
    ID        uint           `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`

    RepositoryID uint       `gorm:"index;not null"`
    Repository   Repository `gorm:"foreignKey:RepositoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

    FilePath  string `gorm:"type:varchar(500);not null"`
    StartLine int    `gorm:"not null"`
    EndLine   int    `gorm:"not null"`
    Content   string `gorm:"type:text;not null"`
    Embedding []byte `gorm:"type:vector(1024)"`
}