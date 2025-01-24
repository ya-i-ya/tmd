package db

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Model struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Chat struct {
	ChatID   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title    string    `gorm:"size:255"`
	Messages []Message `gorm:"foreignKey:ChatID;references:ChatID"`
}

func (m *Model) BeforeCreate(db *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

type User struct {
	Model
	TelegramUserID int64     `gorm:"uniqueIndex;not null"`
	Username       string    `gorm:"size:255"`
	FirstName      string    `gorm:"size:255"`
	LastName       string    `gorm:"size:255"`
	Messages       []Message `gorm:"foreignKey:UserID"`
}

type ChatUser struct {
	Model
	ChatID     uuid.UUID
	UserID     uuid.UUID
	DialogName string    `gorm:"size:255"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

type Message struct {
	Model
	MessageID int       `gorm:"uniqueIndex:idx_message_unique,priority:1;not null"`
	UserID    uuid.UUID `gorm:"uniqueIndex:idx_message_unique,priority:2;not null"`
	ChatID    uuid.UUID `gorm:"uniqueIndex:idx_message_unique,priority:3;not null"`
	Content   string    `gorm:"type:text"`
	MediaURL  string    `gorm:"type:text"`
}
