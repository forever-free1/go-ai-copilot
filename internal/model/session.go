package model

import (
	"time"

	"gorm.io/gorm"
)

// Session 会话模型
type Session struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	Mode      string         `gorm:"size:20;default:chat" json:"mode"` // chat, code_generate, code_explain, code_optimize, code_vuln, code_test, rag
}

// TableName 表名
func (Session) TableName() string {
	return "sessions"
}

// Message 消息模型
type Message struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	SessionID uint           `gorm:"index;not null" json:"session_id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Role      string         `gorm:"size:20;not null" json:"role"` // user / assistant
	Content   string         `gorm:"type:text;not null" json:"content"`
}

// TableName 表名
func (Message) TableName() string {
	return "chat_messages"
}
