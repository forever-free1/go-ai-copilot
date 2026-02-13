package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化 Gin 的默认路由引擎 (自带 Logger 和 Recovery 中间件)
	r := gin.Default()

	// 2. 注册 GET 路由 /api/chat
	r.GET("/api/chat", func(c *gin.Context) {
		// 获取 URL 中的 question 参数
		question := c.Query("question")

		// 简单的判空逻辑
		if question == "" {
			question = "你没有问我任何问题哦~"
		}

		// 3. 返回标准的 JSON 响应
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"reply": "我是智能 OnCall Agent 的雏形 (基于 Gin 构建)。你刚才说：" + question,
			},
		})
	})

	// 4. 启动服务，监听 6872 端口
	r.Run(":6872")
}
