package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"go-ai-copilot/internal/database"
	"go-ai-copilot/internal/model"
	"go-ai-copilot/pkg/jwt"
)

// UserHandler 用户处理器
type UserHandler struct {
	jwt *jwt.JWT
}

// NewUserHandler 创建用户处理器
func NewUserHandler(jwtTool *jwt.JWT) *UserHandler {
	return &UserHandler{jwt: jwtTool}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser model.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "用户名已存在",
		})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "密码加密失败",
		})
		return
	}

	// 创建用户
	user := model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Email:    req.Email,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "用户创建失败",
		})
		return
	}

	// 生成Token
	token, err := h.jwt.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "Token生成失败",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"nickname": user.Nickname,
				"email":    user.Email,
			},
		},
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "参数错误",
		})
		return
	}

	// 查找用户
	var user model.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Code:    401,
			Message: "用户名或密码错误",
		})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Code:    401,
			Message: "用户名或密码错误",
		})
		return
	}

	// 生成Token
	token, err := h.jwt.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "Token生成失败",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"nickname": user.Nickname,
				"email":    user.Email,
			},
		},
	})
}

// GetUserInfo 获取用户信息
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Code:    401,
			Message: "未登录",
		})
		return
	}

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Code:    404,
			Message: "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"email":    user.Email,
		},
	})
}

// UpdateUserInfo 更新用户信息
type UpdateUserInfoRequest struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Code:    401,
			Message: "未登录",
		})
		return
	}

	var req UpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "参数错误",
		})
		return
	}

	result := database.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"nickname": req.Nickname,
			"email":    req.Email,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "更新失败",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
	})
}

// ChangePassword 修改密码
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=20"`
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Code:    401,
			Message: "未登录",
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "参数错误",
		})
		return
	}

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Code:    404,
			Message: "用户不存在",
		})
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "原密码错误",
		})
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "密码加密失败",
		})
		return
	}

	if err := database.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "密码更新失败",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
	})
}
