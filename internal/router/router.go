package router

import (
	"github.com/gin-gonic/gin"
	"go-ai-copilot/internal/handler"
	"go-ai-copilot/internal/middleware"
	"go-ai-copilot/pkg/jwt"
)

// Setup 设置路由
func Setup(jwtTool *jwt.JWT, chatHandler *handler.ChatHandler, userHandler *handler.UserHandler, sessionHandler *handler.SessionHandler, ragHandler *handler.RAGHandler) *gin.Engine {
	// 初始化Gin
	r := gin.Default()

	// 健康检查
	r.GET("/health", chatHandler.Health)

	// 初始化中间件
	authMiddleware := middleware.NewAuthMiddleware(jwtTool)

	// v1 API 路由组
	v1 := r.Group("/api/v1")
	{
		// 用户认证接口（无需登录）
		user := v1.Group("/user")
		{
			user.POST("/register", userHandler.Register)
			user.POST("/login", userHandler.Login)
		}

		// 需要登录的接口
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

		// RAG知识库接口
		ragGroup := authorized.Group("/rag")
		{
			ragGroup.POST("/upload", ragHandler.UploadDocument)
			ragGroup.GET("/list", ragHandler.GetDocuments)
			ragGroup.GET("/:id", ragHandler.GetDocument)
			ragGroup.DELETE("/:id", ragHandler.DeleteDocument)
			ragGroup.POST("/search", ragHandler.Search)
			ragGroup.POST("/chat", ragHandler.RAGChat)
		}
	}

	return r
}
