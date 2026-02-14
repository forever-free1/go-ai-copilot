package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go-ai-copilot/internal/cache"
	"go-ai-copilot/internal/database"
	"go-ai-copilot/internal/model"
)

// SessionHandler 会话处理器
type SessionHandler struct {
	historyLimit int // 上下文记忆轮数
}

// NewSessionHandler 创建会话处理器
func NewSessionHandler() *SessionHandler {
	return &SessionHandler{
		historyLimit: 10, // 默认最近10轮对话
	}
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	Title string `json:"title" binding:"required,max=255"`
	Mode  string `json:"mode"`
}

// CreateSession 创建会话
func (h *SessionHandler) CreateSession(c *gin.Context) {
	userID := c.GetUint("userID")

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 如果没有提供标题，默认一个
	title := req.Title
	if title == "" {
		title = "新会话"
	}

	// 默认模式为 chat
	mode := req.Mode
	if mode == "" {
		mode = "chat"
	}

	session := model.Session{
		UserID: userID,
		Title:  title,
		Mode:   mode,
	}

	if err := database.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "会话创建失败",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
		Data:    session,
	})
}

// GetSessions 获取会话列表
func (h *SessionHandler) GetSessions(c *gin.Context) {
	userID := c.GetUint("userID")

	var sessions []model.Session
	if err := database.DB.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "获取会话列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
		Data:    sessions,
	})
}

// GetSession 获取单个会话
func (h *SessionHandler) GetSession(c *gin.Context) {
	userID := c.GetUint("userID")
	sessionID := c.Param("id")

	var session model.Session
	if err := database.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Code:    404,
			Message: "会话不存在",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
		Data:    session,
	})
}

// UpdateSession 更新会话
type UpdateSessionRequest struct {
	Title string `json:"title" binding:"required,max=255"`
}

func (h *SessionHandler) UpdateSession(c *gin.Context) {
	userID := c.GetUint("userID")
	sessionID := c.Param("id")

	var req UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "参数错误",
		})
		return
	}

	result := database.DB.Model(&model.Session{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Update("title", req.Title)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "更新失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, AuthResponse{
			Code:    404,
			Message: "会话不存在",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
	})
}

// DeleteSession 删除会话
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	userID := c.GetUint("userID")
	sessionIDStr := c.Param("id")
	sessionIDUint, _ := strconv.ParseUint(sessionIDStr, 10, 32)

	// 检查会话是否存在
	var session model.Session
	if err := database.DB.Where("id = ? AND user_id = ?", sessionIDUint, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Code:    404,
			Message: "会话不存在",
		})
		return
	}

	// 删除会话（软删除）
	if err := database.DB.Delete(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "删除失败",
		})
		return
	}

	// 删除会话下的所有消息
	database.DB.Where("session_id = ?", sessionIDUint).Delete(&model.Message{})

	// 删除Redis缓存
	cache.DelSessionHistory(uint(sessionIDUint))

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
	})
}

// GetHistory 获取会话历史
func (h *SessionHandler) GetHistory(c *gin.Context) {
	userID := c.GetUint("userID")
	sessionIDStr := c.Param("id")
	sessionIDUint, _ := strconv.ParseUint(sessionIDStr, 10, 32)

	// 检查会话是否存在
	var session model.Session
	if err := database.DB.Where("id = ? AND user_id = ?", sessionIDUint, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Code:    404,
			Message: "会话不存在",
		})
		return
	}

	// 优先从Redis获取
	messages, err := cache.GetSessionHistory(uint(sessionIDUint))
	if err == nil && len(messages) > 0 {
		c.JSON(http.StatusOK, AuthResponse{
			Code:    0,
			Message: "success",
			Data:    messages,
		})
		return
	}

	// 从数据库获取
	if err := database.DB.Where("session_id = ?", sessionIDUint).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "获取历史失败",
		})
		return
	}

	// 更新会话时间
	database.DB.Model(&session).Update("updated_at", time.Now())

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
		Data:    messages,
	})
}

// AddMessage 添加消息到会话
func (h *SessionHandler) AddMessage(sessionID, userID uint, role, content string) error {
	msg := model.Message{
		SessionID: sessionID,
		UserID:    userID,
		Role:      role,
		Content:   content,
	}

	if err := database.DB.Create(&msg).Error; err != nil {
		return err
	}

	// 更新会话时间
	database.DB.Model(&model.Session{}).Where("id = ?", sessionID).Update("updated_at", time.Now())

	// 更新Redis缓存
	h.updateSessionHistoryCache(sessionID)

	return nil
}

// updateSessionHistoryCache 更新会话历史缓存
func (h *SessionHandler) updateSessionHistoryCache(sessionID uint) {
	var messages []model.Message
	database.DB.Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(h.historyLimit).
		Find(&messages)

	// 反转顺序（从旧到新）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	cache.SetSessionHistory(sessionID, messages)
}

// GetHistoryForContext 获取用于上下文的会话历史
// 返回最近N轮对话，用于拼接到Prompt
func (h *SessionHandler) GetHistoryForContext(sessionID uint) []model.Message {
	// 先尝试从Redis获取
	messages, err := cache.GetSessionHistory(sessionID)
	if err == nil && len(messages) > 0 {
		return messages
	}

	// 从数据库获取
	database.DB.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&messages)

	return messages
}
