package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go-ai-copilot/internal/model"
)

var (
	ErrTokenExpired     = errors.New("token已过期")
	ErrTokenInvalid    = errors.New("token无效")
	ErrTokenNotYetValid = errors.New("token未生效")
)

// Claims JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Config JWT配置
type Config struct {
	SecretKey     string        // 密钥
	ExpireTime    time.Duration // 过期时间
	Issuer        string        // 签发者
}

// JWT JWT工具
type JWT struct {
	cfg Config
}

// New 创建JWT实例
func New(secretKey string, expireTime time.Duration, issuer string) *JWT {
	return &JWT{
		cfg: Config{
			SecretKey:  secretKey,
			ExpireTime: expireTime,
			Issuer:     issuer,
		},
	}
}

// GenerateToken 生成Token
func (j *JWT) GenerateToken(user *model.User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.cfg.ExpireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    j.cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.cfg.SecretKey))
}

// ParseToken 解析Token
func (j *JWT) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.cfg.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotYetValid
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// RefreshToken 刷新Token
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 可以在这里添加刷新逻辑，比如检查是否需要强制重新登录
	return j.GenerateToken(&model.User{
		ID:       claims.UserID,
		Username: claims.Username,
	})
}
