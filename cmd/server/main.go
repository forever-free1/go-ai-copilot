package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"go-ai-copilot/internal/cache"
	"go-ai-copilot/internal/config"
	"go-ai-copilot/internal/database"
	"go-ai-copilot/internal/handler"
	"go-ai-copilot/internal/router"
	"go-ai-copilot/pkg/jwt"
)

// @title go-ai-copilot API
// @version 1.0
// @description 基于 Gin 开发的 AI 代码助手
// @host localhost:8080

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 2. 初始化数据库
	dbCfg := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}
	if err := database.Init(dbCfg); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 3. 初始化Redis
	redisCfg := cache.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	if err := cache.Init(redisCfg); err != nil {
		log.Printf("警告: Redis初始化失败，将使用内存缓存: %v", err)
	}

	// 4. 初始化JWT
	jwtTool := jwt.New(cfg.JWT.Secret, cfg.JWT.ExpireTime, cfg.JWT.Issuer)

	// 5. 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 6. 初始化处理器
	chatHandler, err := handler.NewChatHandler()
	if err != nil {
		log.Fatalf("AI客户端初始化失败: %v", err)
	}
	userHandler := handler.NewUserHandler(jwtTool)
	sessionHandler := handler.NewSessionHandler()
	ragHandler, err := handler.NewRAGHandler()
	if err != nil {
		log.Printf("警告: RAG处理器初始化失败: %v", err)
	}

	// 7. 设置路由
	r := router.Setup(jwtTool, chatHandler, userHandler, sessionHandler, ragHandler)

	// 8. 启动服务
	port := cfg.Server.Port
	log.Printf("服务启动成功，监听端口: %s", port)
	log.Printf("AI模型: %s", cfg.AI.Model)
	log.Printf("AI地址: %s", cfg.AI.BaseURL)

	// 检查环境变量中是否有AI API Key
	if apiKey := os.Getenv("AI_API_KEY"); apiKey != "" {
		log.Printf("AI API Key: 已配置")
	} else {
		log.Printf("警告: 未配置 AI_API_KEY 环境变量")
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal("服务启动失败: ", err)
	}
}
