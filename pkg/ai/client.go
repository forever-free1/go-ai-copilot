package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
)

// Client AI客户端
type Client struct {
	apiKey   string
	baseURL  string
	model    string
	temp     float64
	maxTokens int
	timeout  int
	client   *openai.Client
}

// NewClient 创建AI客户端
func NewClient(apiKey, baseURL, model string, temperature float64, maxTokens, timeout int) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("AI_API_KEY")
	}
	if apiKey == "" {
		return nil, errors.New("API_KEY未设置")
	}

	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL

	client := openai.NewClientWithConfig(cfg)

	return &Client{
		apiKey:   apiKey,
		baseURL:  baseURL,
		model:    model,
		temp:     temperature,
		maxTokens: maxTokens,
		timeout:  timeout,
		client:   client,
	}, nil
}

// StreamChat 流式对话
// ctx: 用于控制请求生命周期，支持用户断开时自动终止
// messages: 对话历史
// onChunk: 每个token的回调函数
func (c *Client) StreamChat(ctx context.Context, messages []openai.ChatCompletionMessage, onChunk func(string) error) error {
	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: float32(c.temp),
		MaxTokens:   c.maxTokens,
		Stream:      true,
	}

	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("创建流式请求失败: %v", err)
	}
	defer stream.Close()

	// 持续读取直到上下文取消或流结束
	for {
		select {
		case <-ctx.Done():
			// 用户断开连接，主动终止请求
			return ctx.Err()
		default:
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("读取流失败: %v", err)
			}

			if len(resp.Choices) > 0 {
				content := resp.Choices[0].Delta.Content
				if content != "" {
					if err := onChunk(content); err != nil {
						return err
					}
				}
			}
		}
	}
}

// Chat 普通对话（非流式）
func (c *Client) Chat(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: float32(c.temp),
		MaxTokens:   c.maxTokens,
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("AI调用失败: %v", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", errors.New("AI返回为空")
}
