package ai

import (
	"context"
	"errors"
	"os"

	"github.com/sashabaranov/go-openai"
)

// EmbeddingClient 向量化客户端
type EmbeddingClient struct {
	apiKey  string
	baseURL string
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

	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(cfg)

	return &EmbeddingClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  client,
	}, nil
}

// GetEmbedding 获取文本的向量表示
func (c *EmbeddingClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2,
		Input: text,
	}

	resp, err := c.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("向量化结果为空")
	}

	return resp.Data[0].Embedding, nil
}

// GetEmbeddings 批量获取文本的向量表示
func (c *EmbeddingClient) GetEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	req := openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2,
		Input: texts,
	}

	resp, err := c.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}
