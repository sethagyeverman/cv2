package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	jwtpkg "cv2/internal/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

// AuthMiddleware JWT 鉴权中间件
type AuthMiddleware struct {
	jwtParser        *jwtpkg.Parser
	refreshClient    *jwtpkg.RefreshClient
	refreshThreshold time.Duration // Token 刷新阈值
}

// NewAuthMiddleware 创建鉴权中间件
func NewAuthMiddleware(secretKey, refreshURL string, refreshThresholdMinutes int) *AuthMiddleware {
	if refreshThresholdMinutes <= 0 {
		refreshThresholdMinutes = 10 // 默认10分钟
	}

	return &AuthMiddleware{
		jwtParser:        jwtpkg.NewParser(secretKey),
		refreshClient:    jwtpkg.NewRefreshClient(refreshURL),
		refreshThreshold: time.Duration(refreshThresholdMinutes) * time.Minute,
	}
}

// Handle 处理请求
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 获取 Authorization header
		authorization := r.Header.Get("Authorization")
		if authorization == "" || !strings.HasPrefix(authorization, "Bearer ") {
			m.respondUnauthorized(w, "Authorization header is missing or malformed")
			return
		}

		// 2. 提取 token
		tokenString := strings.TrimPrefix(authorization, "Bearer ")

		// 3. 解析 token
		claims, err := m.jwtParser.ParseToken(tokenString)
		if err != nil {
			logx.Errorf("JWT parsing error: %v", err)
			m.respondUnauthorized(w, "Invalid token: "+err.Error())
			return
		}

		// 4. 检查 token 是否即将过期，如果是则刷新
		if m.jwtParser.IsTokenExpiringSoon(claims, m.refreshThreshold) {
			newToken, err := m.refreshClient.RefreshToken(r.Context(), tokenString)
			if err != nil {
				logx.Errorf("Failed to refresh token: %v", err)
				// 刷新失败不影响当前请求，继续执行
			} else {
				// 将新 token 返回给客户端
				w.Header().Set("Authorization", "Bearer "+newToken)
				w.Header().Set("Access-Control-Expose-Headers", "Authorization")
				logx.Infof("Token refreshed for user: %d", claims.UserID)
			}
		}

		// 5. 提取用户上下文并存储到 context
		userCtx := jwtpkg.ExtractUserContext(claims)
		ctx := jwtpkg.SetUserContext(r.Context(), userCtx)

		// 6. 记录请求信息
		logx.WithContext(ctx).Infof("Authenticated request: user_id=%d, tenant_id=%s, path=%s",
			userCtx.UserID, userCtx.TenantID, r.URL.Path)

		// 7. 传递给下一个处理器
		next(w, r.WithContext(ctx))
	}
}

// respondUnauthorized 返回未授权响应
func (m *AuthMiddleware) respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": http.StatusUnauthorized,
		"msg":  message,
	})
}
