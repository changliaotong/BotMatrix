package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAIAdapter 适配 OpenAI 兼容接口的客户端
type OpenAIAdapter struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func NewOpenAIAdapter(baseURL, apiKey string) *OpenAIAdapter {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIAdapter{
		BaseURL: strings.TrimSuffix(strings.TrimSpace(baseURL), "/"),
		APIKey:  strings.TrimSpace(apiKey),
		HTTP:    &http.Client{},
	}
}

func (a *OpenAIAdapter) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if a.APIKey == "" {
		return nil, fmt.Errorf("API Key is empty, please check your provider configuration")
	}
	req.Stream = false
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

	resp, err := a.HTTP.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (a *OpenAIAdapter) ChatStream(ctx context.Context, req ChatRequest) (<-chan ChatStreamResponse, error) {
	if a.APIKey == "" {
		return nil, fmt.Errorf("API Key is empty, please check your provider configuration")
	}
	req.Stream = true
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	fmt.Printf("[DEBUG] OpenAI Request: POST %s/chat/completions (Key length: %d)\n", a.BaseURL, len(a.APIKey))
	fmt.Printf("[DEBUG] OpenAI Request Body: %s\n", string(body))

	resp, err := a.HTTP.Do(httpReq)
	if err != nil {
		fmt.Printf("[DEBUG] OpenAI Request Error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[DEBUG] OpenAI Response Status: %d, Content-Type: %s\n", resp.StatusCode, resp.Header.Get("Content-Type"))
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		errStr := string(respBody)
		if len(errStr) > 500 {
			errStr = errStr[:500] + "... (truncated)"
		}
		return nil, fmt.Errorf("AI API error (status %d): %s", resp.StatusCode, errStr)
	}

	// 检查 Content-Type，如果不是 stream 却收到了 HTML，说明配置错误
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") && strings.Contains(contentType, "text/html") {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("预期 text/event-stream 但收到 HTML，请检查 BaseURL 是否配置正确。响应前缀: %s", string(respBody[:100]))
	}

	ch := make(chan ChatStreamResponse)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		// 增加缓冲区大小以防单行数据过长 (默认 64K 可能不够)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 10*1024*1024) // 最大支持 10MB
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var streamResp ChatStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				ch <- ChatStreamResponse{Error: err}
				return
			}
			ch <- streamResp
		}
		if err := scanner.Err(); err != nil {
			ch <- ChatStreamResponse{Error: err}
		}
	}()

	return ch, nil
}

func (a *OpenAIAdapter) CreateEmbedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	if a.APIKey == "" {
		return nil, fmt.Errorf("API Key is empty, please check your provider configuration")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	path := "/embeddings"
	// 针对豆包多模态向量模型的特殊处理
	if strings.Contains(strings.ToLower(req.Model), "vision") || strings.Contains(strings.ToLower(req.Model), "multimodal") {
		path = "/embeddings/multimodal"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

	resp, err := a.HTTP.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 针对多模态响应的特殊处理
	if strings.Contains(path, "multimodal") {
		var multiResult struct {
			Data  EmbeddingData `json:"data"`
			Model string        `json:"model"`
			Usage UsageInfo     `json:"usage"`
		}
		if err := json.Unmarshal(respBody, &multiResult); err != nil {
			return nil, err
		}
		return &EmbeddingResponse{
			Data:  []EmbeddingData{multiResult.Data},
			Model: multiResult.Model,
			Usage: multiResult.Usage,
		}, nil
	}

	var result EmbeddingResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
