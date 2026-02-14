package database

import (
	"fmt"
	"log"

	"go-ai-copilot/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 数据库连接实例
var DB *gorm.DB

// Config 数据库配置
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Init 初始化数据库连接
func Init(cfg Config) error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(
		&model.User{},
		&model.Session{},
		&model.Message{},
		&model.RAGDocument{},
		&model.RAGChunk{},
	); err != nil {
		return fmt.Errorf("表迁移失败: %v", err)
	}

	// 创建向量索引（如果不存在）
	db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_rag_chunks_embedding ON rag_chunks USING ivfflat (embedding vector_cosine_ops)")

	// 注册向量类型
	// 注意：pgvector 需要数据库支持 vector 扩展

	DB = db
	log.Println("数据库连接成功")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
