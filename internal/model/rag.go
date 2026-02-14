package model

import (
	"time"

	"gorm.io/gorm"
)

// RAGDocument RAG文档模型
type RAGDocument struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	FileName  string         `gorm:"size:255;not null" json:"file_name"`
	FileType  string         `gorm:"size:50;not null" json:"file_type"`
	FileSize  int64          `json:"file_size"`
	Status    string         `gorm:"size:20;not null;default:pending" json:"status"` // pending, processing, completed, failed
}

// TableName 表名
func (RAGDocument) TableName() string {
	return "rag_documents"
}

// RAGChunk RAG文档分块模型
type RAGChunk struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	DocumentID uint           `gorm:"index;not null" json:"document_id"`
	UserID     uint           `gorm:"index;not null" json:"user_id"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	// 向量字段使用 pgvector 的 vector 类型
	Embedding []float32       `gorm:"type:vector(1536)" json:"-"`
	ChunkIndex int            `gorm:"not null" json:"chunk_index"`
}

// TableName 表名
func (RAGChunk) TableName() string {
	return "rag_chunks"
}
