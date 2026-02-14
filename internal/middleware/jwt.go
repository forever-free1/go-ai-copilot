package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go-ai-copilot/pkg/jwt"
)

// AuthMiddleware JWT认证中间件
type AuthMiddleware struct {
	jwt *jwt.JWT
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwtTool *jwt.JWT) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwtTool}
}

// Handler JWT认证处理函数
func (m *AuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证信息",
			})
			c.Abort()
			return
		}

		// Bearer Token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证格式错误",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析Token
		claims, err := m.jwt.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token无效或已过期",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
