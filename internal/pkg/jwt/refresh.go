package jwt

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"cv2/internal/pkg/errx"
)

// RefreshClient Token 刷新客户端
type RefreshClient struct {
	refreshURL string
	httpClient *http.Client
}

// NewRefreshClient 创建 Token 刷新客户端
func NewRefreshClient(refreshURL string) *RefreshClient {
	return &RefreshClient{
		refreshURL: refreshURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RefreshResponse Token 刷新响应
type RefreshResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"` // 新的 token
}

// RefreshToken 刷新 Token
func (c *RefreshClient) RefreshToken(ctx context.Context, oldToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.refreshURL, nil)
	if err != nil {
		return "", errx.Warp(http.StatusInternalServerError, err, "create request")
	}

	req.Header.Set("Authorization", "Bearer "+oldToken)
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errx.Warp(http.StatusInternalServerError, err, "do request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errx.Warp(http.StatusInternalServerError, err, "read response")
	}

	var result RefreshResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", errx.Warp(http.StatusInternalServerError, err, "unmarshal response")
	}

	if result.Code != 200 {
		return "", errx.Newf(http.StatusUnauthorized, "refresh token failed: %s", result.Msg)
	}

	return result.Data, nil
}
