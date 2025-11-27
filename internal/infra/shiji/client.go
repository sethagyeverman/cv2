package shiji

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cv2/internal/pkg/errx"
)

// Client 世纪服务客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建世纪服务客户端
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListArticles 获取文章列表
func (c *Client) ListArticles(ctx context.Context, pageNum, pageSize int, title string) (*ArticleListResponse, error) {
	q := url.Values{}
	q.Set("pageNum", fmt.Sprintf("%d", pageNum))
	q.Set("pageSize", fmt.Sprintf("%d", pageSize))
	q.Set("status", "1") // 只查询已发布的文章
	if title != "" {
		q.Set("title", title)
	}

	target := c.baseURL + "/prod-api/system/article/list?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "构造请求失败")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errx.Warp(http.StatusBadGateway, err, "调用外部接口失败")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "读取响应失败")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errx.Newf(http.StatusBadGateway, "外部接口响应非200: %d", resp.StatusCode)
	}

	var result ArticleListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "解析响应失败")
	}

	if result.Code != 200 {
		return nil, errx.Newf(http.StatusBadGateway, "外部接口返回错误: %s", result.Msg)
	}

	return &result, nil
}

// GetArticle 获取文章详情
func (c *Client) GetArticle(ctx context.Context, articleID int64) (*Article, error) {
	target := c.baseURL + "/prod-api/system/article/" + formatInt64(articleID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "构造请求失败")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errx.Warp(http.StatusBadGateway, err, "调用外部接口失败")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "读取响应失败")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errx.Newf(http.StatusBadGateway, "外部接口响应非200: %d", resp.StatusCode)
	}

	var result ArticleDetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "解析响应失败")
	}

	if result.Code != 200 {
		return nil, errx.Newf(http.StatusBadGateway, "外部接口返回错误: %s", result.Msg)
	}

	return &result.Data, nil
}

func formatInt64(n int64) string {
	return fmt.Sprintf("%d", n)
}

// Login 用户登录
// grantType: 授权类型, credentials: 凭证内容(JSON字符串)
func (c *Client) Login(ctx context.Context, clientID, clientSecret, grantType, credentials string) (*LoginResponse, error) {
	data := map[string]string{
		"grant_type":    grantType,
		"client_id":     clientID,
		"client_secret": clientSecret,
		"credentials":   credentials,
	}

	target := c.baseURL + "/prod-api/oauth2/token"

	jsonStr, err := json.Marshal(data)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "构造请求失败")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader(string(jsonStr)))
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "构造请求失败")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errx.Warp(http.StatusBadGateway, err, "调用登录接口失败")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "读取响应失败")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errx.Newf(http.StatusBadGateway, "登录接口响应非200: %d, body: %s", resp.StatusCode, string(body))
	}

	var result LoginResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "解析登录响应失败")
	}

	return &result, nil
}
