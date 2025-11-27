package algorithm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
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
			Timeout: 3 * time.Minute, // 文件转换可能较慢
		},
	}
}

// DataURL 获取数据服务 URL
func (c *Client) DataURL() string {
	return c.dataURL
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

// AIWrite 流式调用 AI 帮写接口，返回数据通道和错误通道
func (c *Client) AIWrite(ctx context.Context, req *AIWriteRequest) (<-chan string, <-chan error) {
	dataCh := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(dataCh)
		defer close(errCh)

		body, err := json.Marshal(req)
		if err != nil {
			errCh <- fmt.Errorf("marshal request: %w", err)
			return
		}

		// 使用独立的 context 和超时
		reqCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		// 监听外部 context 取消
		go func() {
			<-ctx.Done()
			cancel()
		}()
		defer cancel()

		httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost,
			c.generateURL+"/writer/resume_item_gen", bytes.NewReader(body))
		if err != nil {
			errCh <- fmt.Errorf("create request: %w", err)
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			errCh <- fmt.Errorf("do request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
			return
		}

		// 解析 SSE 流
		reader := bufio.NewReader(resp.Body)
		var eventType string
		var dataLines []string

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				errCh <- fmt.Errorf("read sse: %w", err)
				return
			}

			line = strings.TrimRight(line, "\r\n")

			// 空行表示一个事件结束
			if line == "" {
				if eventType == "add" && len(dataLines) > 0 {
					content := strings.Join(dataLines, "\n")
					select {
					case dataCh <- content:
					case <-ctx.Done():
						return
					}
				}
				if eventType == "end" {
					return
				}
				dataLines = nil
				continue
			}

			// 解析 SSE 字段
			if parts := strings.SplitN(line, ":", 2); len(parts) == 2 {
				field := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				switch field {
				case "event":
					eventType = value
				case "data":
					dataLines = append(dataLines, value)
				}
			}
		}
	}()

	return dataCh, errCh
}

// File2Markdown 文件转 Markdown (pdf/doc/docx → markdown)
func (c *Client) File2Markdown(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	formFile, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(formFile, file); err != nil {
		return "", fmt.Errorf("copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.dataURL+"/convert/file_to_markdown", body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Content, nil
}

// Markdown2Struct Markdown 转结构化数据
func (c *Client) Markdown2Struct(ctx context.Context, md string) (*ResumeData, error) {
	reqBody, _ := json.Marshal(map[string]string{"content": md})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.dataURL+"/extractor/resume_struct", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ResumeData *ResumeData `json:"resume_data"`
		Success    bool        `json:"success"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("简历解析失败，请检查文件格式")
	}

	return result.ResumeData, nil
}

// Struct2Markdown 结构化数据转 Markdown
func (c *Client) Struct2Markdown(ctx context.Context, data *ResumeData, title string) (string, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"json_input": data,
		"title":      title,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.dataURL+"/convert/json_to_markdown", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Message         string `json:"message"`
		MarkdownContent string `json:"markdown_content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.MarkdownContent, nil
}

// Markdown2PDF Markdown 转 PDF，返回 PDF 文件内容
func (c *Client) Markdown2PDF(ctx context.Context, mdContent, filename string) ([]byte, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"markdown_content": mdContent,
		"filename":         filename,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.dataURL+"/convert/markdown_to_pdf", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}
