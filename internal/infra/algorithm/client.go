package algorithm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 算法服务客户端
type Client struct {
	generateURL string // 生成服务 URL
	dataURL     string // 数据服务 URL
	scoreURL    string // 评分服务 URL
	httpClient  *http.Client
}

// NewClient 创建算法客户端
func NewClient(generateURL, dataURL, scoreURL string) *Client {
	return &Client{
		generateURL: generateURL,
		dataURL:     dataURL,
		scoreURL:    scoreURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SubmitGenerateTask 提交简历生成任务
func (c *Client) SubmitGenerateTask(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.generateURL+"/writer/resume_gen_task", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// GetTaskStatus 查询任务状态
func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (*TaskStatus, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/writer/resume_gen_task/%s", c.generateURL, taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var status TaskStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &status, nil
}

// ScoreResume 简历评分
func (c *Client) ScoreResume(ctx context.Context, req *ScoreRequest) ([]*DimScore, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.scoreURL+"/resume_eval/section_eval", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// 尝试直接解析为数组
	var scores []*DimScore
	if err := json.Unmarshal(respBody, &scores); err == nil {
		return scores, nil
	}

	// 尝试解析包装格式 {code, msg, data}
	var wrapped ScoreResponse
	if err := json.Unmarshal(respBody, &wrapped); err == nil {
		return wrapped.Data, nil
	}

	return nil, fmt.Errorf("failed to parse response: %s", string(respBody))
}
