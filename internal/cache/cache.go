package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go-ai-copilot/internal/model"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

// Config Redis配置
type Config struct {
	Addr     string
	Password string
	DB       int
}

// Init 初始化Redis连接
func Init(cfg Config) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis连接失败: %v", err)
	}

	fmt.Println("Redis连接成功")
	return nil
}

// GetDB 获取数据库实例（为了兼容旧代码）
func GetDB() *redis.Client {
	return Client
}

// SessionHistoryKey 会话历史缓存Key
func SessionHistoryKey(sessionID uint) string {
	return fmt.Sprintf("session:history:%d", sessionID)
}

// GetSessionHistory 获取会话历史缓存
func GetSessionHistory(sessionID uint) ([]model.Message, error) {
	key := SessionHistoryKey(sessionID)
	data, err := Client.Get(Ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var messages []model.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// SetSessionHistory 设置会话历史缓存
// 默认1小时过期
func SetSessionHistory(sessionID uint, messages []model.Message) error {
	key := SessionHistoryKey(sessionID)
	data, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	return Client.Set(Ctx, key, data, time.Hour).Err()
}

// DelSessionHistory 删除会话历史缓存
func DelSessionHistory(sessionID uint) error {
	key := SessionHistoryKey(sessionID)
	return Client.Del(Ctx, key).Err()
}
