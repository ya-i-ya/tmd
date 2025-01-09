package db

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}

type User struct {
	BaseModel
	TelegramUserID int64      `gorm:"uniqueIndex;not null"`
	Username       string     `gorm:"size:255"`
	FirstName      string     `gorm:"size:255"`
	LastName       string     `gorm:"size:255"`
	Messages       []Message  `gorm:"foreignKey:UserID"`
	Chats          []ChatUser `gorm:"foreignKey:UserID"`
}
type Chat struct {
	BaseModel
	ChatID   uuid.UUID `gorm:"uniqueIndex;not null"`
	Title    string    `gorm:"size:255"`
	Messages []Message `gorm:"foreignKey:ChatID"`
	Users    []User    `gorm:"foreignKey:ChatID"`
}
type ChatUser struct {
	BaseModel
	ChatID    uuid.UUID `gorm:"uniqueIndex;not null"`
	UserID    uuid.UUID `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
type Message struct {
	BaseModel
	MessageID int       `gorm:"uniqueIndex:idx_message_unique,priority:1;not null"`
	UserID    uuid.UUID `gorm:"uniqueIndex:idx_message_unique,priority:2;not null"`
	ChatId    uuid.UUID `gorm:"uniqueIndex:idx_message_unique,priority:3;not null"`
	Content   string    `gorm:"type:text"`
	MediaURL  string    `gorm:"type:text"`
}
