package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"go-ai-copilot/internal/cache"
	"go-ai-copilot/internal/config"
	"go-ai-copilot/internal/database"
	"go-ai-copilot/internal/handler"
	"go-ai-copilot/internal/middleware"
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

	// 5. 初始化 Gin
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	// 6. 初始化处理器
	chatHandler, err := handler.NewChatHandler()
	if err != nil {
		log.Fatalf("AI客户端初始化失败: %v", err)
	}
	userHandler := handler.NewUserHandler(jwtTool)
	sessionHandler := handler.NewSessionHandler()

	// 7. 初始化中间件
	authMiddleware := middleware.NewAuthMiddleware(jwtTool)

	// 8. 注册路由

	// 健康检查
	r.GET("/health", chatHandler.Health)

	// v1 API 路由组
	v1 := r.Group("/api/v1")
	{
		// 用户认证接口（无需登录）
		user := v1.Group("/user")
		{
			user.POST("/register", userHandler.Register)
			user.POST("/login", userHandler.Login)
		}

		// 需要登录的接口 - 直接使用middleware
		authorized := v1.Group("")
		authorized.Use(authMiddleware.Handler())

		// 用户信息
		authorized.GET("/user/info", userHandler.GetUserInfo)
		authorized.PUT("/user/info", userHandler.UpdateUserInfo)
		authorized.PUT("/user/password", userHandler.ChangePassword)

		// 会话管理
		authorized.GET("/session/list", sessionHandler.GetSessions)
		authorized.POST("/session", sessionHandler.CreateSession)
		authorized.GET("/session/:id", sessionHandler.GetSession)
		authorized.PUT("/session/:id", sessionHandler.UpdateSession)
		authorized.DELETE("/session/:id", sessionHandler.DeleteSession)
		authorized.GET("/session/:id/history", sessionHandler.GetHistory)

		// 对话接口
		authorized.POST("/chat", chatHandler.Chat)
		authorized.POST("/chat/stream", chatHandler.StreamChat)
		authorized.POST("/chat/mode", chatHandler.HandleChatWithMode)
	}

	// 9. 启动服务
	port := cfg.Server.Port
	log.Printf("服务启动成功，监听端口: %s", port)
	log.Printf("AI模型: %s", cfg.AI.Model)
	log.Printf("AI地址: %s", cfg.AI.BaseURL)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("服务启动失败: ", err)
	}
}
