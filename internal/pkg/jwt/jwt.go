package jwt

import (
	"context"
	"errors"
	"net/http"
	"time"

	"cv2/internal/pkg/errx"
	"cv2/internal/types"

	"github.com/golang-jwt/jwt/v4"
)

var (
	// ErrTokenInvalid Token 无效
	ErrTokenInvalid = errx.New(http.StatusUnauthorized, "token is invalid")
	// ErrTokenExpired Token 已过期
	ErrTokenExpired = errx.New(http.StatusUnauthorized, "token has expired")
	// ErrTokenMalformed Token 格式错误
	ErrTokenMalformed = errx.New(http.StatusUnauthorized, "token is malformed")
	// ErrUserContextNotFound 用户上下文未找到
	ErrUserContextNotFound = errx.New(http.StatusUnauthorized, "user context not found")
)

// Parser JWT 解析器
type Parser struct {
	secretKey string
}

// NewParser 创建 JWT 解析器
func NewParser(secretKey string) *Parser {
	return &Parser{
		secretKey: secretKey,
	}
}

// ParseToken 解析 JWT Token
func (p *Parser) ParseToken(tokenString string) (*types.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errx.Newf(http.StatusUnauthorized, "unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(p.secretKey), nil
	})

	if err != nil {
		// 判断具体错误类型
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, errx.Warp(http.StatusUnauthorized, err, "parse token failed")
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*types.JWTClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// ValidateToken 验证 Token 是否有效
func (p *Parser) ValidateToken(tokenString string) error {
	_, err := p.ParseToken(tokenString)
	return err
}

// IsTokenExpiringSoon 检查 Token 是否即将过期（剩余时间小于指定时长）
func (p *Parser) IsTokenExpiringSoon(claims *types.JWTClaims, duration time.Duration) bool {
	if claims.ExpiresAt == nil {
		return false
	}

	expiresAt := claims.ExpiresAt.Time
	now := time.Now()
	remainingTime := expiresAt.Sub(now)

	return remainingTime > 0 && remainingTime <= duration
}

// ExtractUserContext 从 Claims 中提取用户上下文信息
func ExtractUserContext(claims *types.JWTClaims) *types.UserContext {
	return &types.UserContext{
		UserID:   claims.UserID,
		UserName: claims.UserName,
		TenantID: claims.TenantID,
		UserType: claims.UserType,
		DeptID:   claims.DeptID,
		DeptName: claims.DeptName,
		LoginID:  claims.LoginID,
		ClientID: claims.ClientID,
	}
}

// Context keys for user information
type contextKey string

const (
	// UserContextKey 用户上下文在 context 中的 key
	UserContextKey contextKey = "user_context"
)

// SetUserContext 将用户上下文存储到 context 中
func SetUserContext(ctx context.Context, userCtx *types.UserContext) context.Context {
	return context.WithValue(ctx, UserContextKey, userCtx)
}

// GetUserContext 从 context 中获取用户上下文
func GetUserContext(ctx context.Context) (*types.UserContext, error) {
	userCtx, ok := ctx.Value(UserContextKey).(*types.UserContext)
	if !ok {
		return nil, ErrUserContextNotFound
	}
	return userCtx, nil
}
