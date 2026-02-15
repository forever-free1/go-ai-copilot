package ai

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

// EmbeddingClient 向量化客户端
type EmbeddingClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *openai.Client
}

// NewEmbeddingClient 创建向量化客户端
func NewEmbeddingClient(apiKey, baseURL, model string) (*EmbeddingClient, error) {
	if apiKey == "" {
		apiKey = os.Getenv("EMBEDDING_API_KEY")
	}
	if apiKey == "" {
		// 如果没有单独配置，使用默认的AI API Key
		apiKey = os.Getenv("AI_API_KEY")
	}
	if apiKey == "" {
		return nil, errors.New("Embedding API_KEY未设置")
	}

	// 默认使用DeepSeek的embedding模型
	if model == "" {
		model = "deepseek-embedding"
	}

	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		// DeepSeek的embedding API需要使用v1路径
		cfg.BaseURL = baseURL + "/v1"
	}

	client := openai.NewClientWithConfig(cfg)

	return &EmbeddingClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  client,
	}, nil
}

// GetEmbedding 获取文本的向量表示
func (c *EmbeddingClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(c.model),
		Input: text,
	}

	resp, err := c.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("embedding请求失败: %v", err)
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("向量化结果为空")
	}

	return resp.Data[0].Embedding, nil
}

// GetEmbeddings 批量获取文本的向量表示
func (c *EmbeddingClient) GetEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	req := openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(c.model),
		Input: texts,
	}

	resp, err := c.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("embedding请求失败: %v", err)
	}

	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}
