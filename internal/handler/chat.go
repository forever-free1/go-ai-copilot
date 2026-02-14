package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"go-ai-copilot/internal/config"
	"go-ai-copilot/pkg/ai"
)

// ChatHandler 对话处理器
type ChatHandler struct {
	client        *ai.Client
	sessionHandler *SessionHandler
}

// NewChatHandler 创建对话处理器
func NewChatHandler() (*ChatHandler, error) {
	cfg := config.GlobalConfig
	client, err := ai.NewClient(
		"", // API_KEY从环境变量读取
		cfg.AI.BaseURL,
		cfg.AI.Model,
		cfg.AI.Temperature,
		cfg.AI.MaxTokens,
		cfg.AI.Timeout,
	)
	if err != nil {
		return nil, err
	}

	return &ChatHandler{
		client:        client,
		sessionHandler: NewSessionHandler(),
	}, nil
}

// ChatRequest 对话请求
type ChatRequest struct {
	Message   string `json:"message" binding:"required"`
	SessionID uint   `json:"session_id,omitempty"` // 会话ID
	APIKey    string `json:"api_key,omitempty"`    // 用户可选传入自己的API_KEY
	Model     string `json:"model,omitempty"`     // 用户可选指定模型
	Temperature float64 `json:"temperature,omitempty"` // 用户可选温度
}

// ChatResponse 对话响应
type ChatResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Chat 普通对话接口（支持会话上下文）
func (h *ChatHandler) Chat(c *gin.Context) {
	// 检查AI客户端是否可用
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, ChatResponse{
			Code:    503,
			Message: "AI服务暂不可用，请配置API_KEY",
		})
		return
	}

	userID := c.GetUint("userID")

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ChatResponse{
			Code:    400,
			Message: "参数错误",
		})
		return
	}

	// 构建消息（包含上下文）
	messages := h.buildMessages(req.SessionID, userID, req.Message, "")

	// 调用AI
	ctx := c.Request.Context()
	reply, err := h.client.Chat(ctx, messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ChatResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	// 保存消息到数据库
	if req.SessionID > 0 {
		h.sessionHandler.AddMessage(req.SessionID, userID, "user", req.Message)
		h.sessionHandler.AddMessage(req.SessionID, userID, "assistant", reply)
	}

	c.JSON(http.StatusOK, ChatResponse{
		Code:    0,
		Message: "success",
		Data:    gin.H{"reply": reply, "session_id": req.SessionID},
	})
}

// buildMessages 构建消息列表（包含上下文）
func (h *ChatHandler) buildMessages(sessionID, userID uint, newMessage, systemPrompt string) []openai.ChatCompletionMessage {
	var messages []openai.ChatCompletionMessage

	// 添加系统提示
	if systemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}

	// 添加历史上下文
	if sessionID > 0 {
		history := h.sessionHandler.GetHistoryForContext(sessionID)
		for _, msg := range history {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 添加当前消息
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: newMessage,
	})

	return messages
}

// StreamChat SSE流式对话接口
// 核心亮点：使用Goroutine+Channel处理流式响应
func (h *ChatHandler) StreamChat(c *gin.Context) {
	// 检查AI客户端是否可用
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, ChatResponse{
			Code:    503,
			Message: "AI服务暂不可用，请配置API_KEY",
		})
		return
	}

	userID := c.GetUint("userID")

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ChatResponse{
			Code:    400,
			Message: "参数错误",
		})
		return
	}

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 创建带超时的上下文（默认2分钟超时）
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// 构建消息（包含上下文）
	messages := h.buildMessages(req.SessionID, userID, req.Message, "")

	// 用于收集完整回复
	fullReply := ""

	// 创建Channel用于传递token
	tokenChan := make(chan string, 100)
	errChan := make(chan error, 1)

	// 启动Goroutine调用AI流式接口
	// 核心亮点：将AI调用放到独立Goroutine，通过Channel实时推送Token
	go func() {
		err := h.client.StreamChat(ctx, messages, func(chunk string) error {
			fullReply += chunk
			tokenChan <- chunk
			return nil
		})
		if err != nil {
			errChan <- err
		}
		close(tokenChan)
	}()

	// 主Goroutine：监听Channel和Context
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, ChatResponse{
			Code:    500,
			Message: "不支持流式响应",
		})
		return
	}

	for {
		select {
		case <-ctx.Done():
			// 用户断开连接，终止流
			return
		case token, ok := <-tokenChan:
			if !ok {
				// 流结束，保存消息到数据库
				if req.SessionID > 0 && fullReply != "" {
					h.sessionHandler.AddMessage(req.SessionID, userID, "user", req.Message)
					h.sessionHandler.AddMessage(req.SessionID, userID, "assistant", fullReply)
				}
				// 发送结束标记
				c.SSEvent("done", map[string]string{"status": "completed"})
				flusher.Flush()
				return
			}
			// 发送token到前端
			c.SSEvent("message", token)
			flusher.Flush()
		case err := <-errChan:
			// 发生错误
			c.SSEvent("error", err.Error())
			flusher.Flush()
			return
		}
	}
}

// Health 健康检查
func (h *ChatHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, ChatResponse{
		Code:    0,
		Message: "success",
	})
}

// HandleChatWithMode 处理带模式的对话请求
// mode: chat(通用对话) / code_generate(代码生成) / code_explain(代码解释)
//                            / code_optimize(代码优化) / code_vuln(漏洞检测) / code_test(单元测试)
func (h *ChatHandler) HandleChatWithMode(c *gin.Context) {
	// 检查AI客户端是否可用
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, ChatResponse{
			Code:    503,
			Message: "AI服务暂不可用，请配置API_KEY",
		})
		return
	}

	userID := c.GetUint("userID")
	mode := c.DefaultPostForm("mode", "chat")

	// 根据模式选择不同的System Prompt
	systemPrompt := getSystemPromptByMode(mode)

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ChatResponse{
			Code:    400,
			Message: "参数错误",
		})
		return
	}

	// 构建消息（包含上下文）
	messages := h.buildMessages(req.SessionID, userID, req.Message, systemPrompt)

	// 调用AI
	ctx := c.Request.Context()
	reply, err := h.client.Chat(ctx, messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ChatResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	// 保存消息到数据库
	if req.SessionID > 0 {
		h.sessionHandler.AddMessage(req.SessionID, userID, "user", req.Message)
		h.sessionHandler.AddMessage(req.SessionID, userID, "assistant", reply)
	}

	c.JSON(http.StatusOK, ChatResponse{
		Code:    0,
		Message: "success",
		Data:    gin.H{"reply": reply},
	})
}

// getSystemPromptByMode 根据模式获取System Prompt
func getSystemPromptByMode(mode string) string {
	switch mode {
	case "code_generate":
		return "你是一个专业的Go后端开发工程师。请根据用户需求生成符合Go规范的代码，包含错误处理、注释、单元测试。"
	case "code_explain":
		return "你是一个专业的Go后端开发工程师。请逐行解释用户提供的Go代码的逻辑、用途、设计思路。"
	case "code_optimize":
		return "你是一个专业的Go后端开发工程师。请优化用户提供的Go代码的性能、可读性、规范度，指出优化点。"
	case "code_vuln":
		return "你是一个专业的Go安全工程师。请检测用户提供的Go代码中的安全漏洞、内存泄漏、并发问题、错误处理缺陷。"
	case "code_test":
		return "你是一个专业的Go测试工程师。请为用户提供的Go代码生成单元测试用例，提升测试覆盖率。"
	default:
		return "你是一个专业的AI助手，请用简洁清晰的语言回答用户的问题。"
	}
}
