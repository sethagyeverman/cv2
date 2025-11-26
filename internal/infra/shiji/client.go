package shiji

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// ArticleListResponse 文章列表响应
type ArticleListResponse struct {
	Code  int       `json:"code"`
	Msg   string    `json:"msg"`
	Rows  []Article `json:"rows"`
	Total int64     `json:"total"`
}

// ArticleDetailResponse 文章详情响应
type ArticleDetailResponse struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data Article `json:"data"`
}

// Article 文章
type Article struct {
	ArticleId    int64  `json:"articleId,string"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Content      string `json:"content"`
	ThumbnailUrl string `json:"thumbnailUrl"`
	Status       int    `json:"status,string"`
	ViewCount    int    `json:"viewCount"`
	PublishTime  string `json:"publishTime"`
	CreateTime   string `json:"createTime"`
	UpdateTime   string `json:"updateTime"`
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
