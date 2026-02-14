package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"go-ai-copilot/internal/config"
	"go-ai-copilot/internal/handler"
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

	// 2. 初始化 Gin
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	// 3. 初始化处理器
	chatHandler, err := handler.NewChatHandler()
	if err != nil {
		log.Fatalf("AI客户端初始化失败: %v", err)
	}

	// 4. 注册路由
	r.GET("/health", chatHandler.Health)

	// v1 API 路由组
	v1 := r.Group("/api/v1")
	{
		// 对话接口
		v1.POST("/chat", chatHandler.Chat)
		v1.POST("/chat/stream", chatHandler.StreamChat)
		v1.POST("/chat/mode", chatHandler.HandleChatWithMode)
	}

	// 5. 启动服务
	port := cfg.Server.Port
	log.Printf("服务启动成功，监听端口: %s", port)
	log.Printf("AI模型: %s", cfg.AI.Model)
	log.Printf("AI地址: %s", cfg.AI.BaseURL)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("服务启动失败: ", err)
	}
}

// curl测试命令：
// 普通对话
// curl -X POST http://localhost:8080/api/v1/chat \
//   -H "Content-Type: application/json" \
//   -d '{"message": "你好，请介绍一下自己"}'

// 流式对话
// curl -X POST http://localhost:8080/api/v1/chat/stream \
//   -H "Content-Type: application/json" \
//   -d '{"message": "用Go写一个Hello World程序"}'
