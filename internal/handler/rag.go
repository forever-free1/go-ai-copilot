package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"go-ai-copilot/internal/config"
	"go-ai-copilot/internal/database"
	"go-ai-copilot/internal/model"
	"go-ai-copilot/internal/rag"
	"go-ai-copilot/pkg/ai"
)

// RAGHandler RAG处理器
type RAGHandler struct {
	embeddingClient *ai.EmbeddingClient
	textSplitter    *rag.TextSplitter
}

// NewRAGHandler 创建RAG处理器
func NewRAGHandler() (*RAGHandler, error) {
	cfg := config.GlobalConfig

	// 创建embedding客户端
	embeddingClient, err := ai.NewEmbeddingClient(
		os.Getenv("EMBEDDING_API_KEY"),
		cfg.AI.BaseURL, // 使用与AI相同的base URL
		"text-embedding-3-small", // 默认模型
	)
	if err != nil {
		return nil, fmt.Errorf("embedding客户端初始化失败: %v", err)
	}

	return &RAGHandler{
		embeddingClient: embeddingClient,
		textSplitter:    rag.NewTextSplitter(1024, 256),
	}, nil
}

// UploadRequest 上传请求
type UploadRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// UploadDocument 上传文档
func (h *RAGHandler) UploadDocument(c *gin.Context) {
	userID := c.GetUint("userID")

	// 获取文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "请上传文件",
		})
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".txt", ".md", ".go", ".json", ".yaml", ".yml"}
	isAllowed := false
	for _, e := range allowedExts {
		if ext == e {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "不支持的文件类型，仅支持: .txt, .md, .go, .json, .yaml, .yml",
		})
		return
	}

	// 验证文件大小 (10MB)
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "文件大小不能超过10MB",
		})
		return
	}

	// 创建上传目录
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	// 保存文件
	filePath := filepath.Join(uploadDir, fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), ext))
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "文件保存失败",
		})
		return
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "文件读取失败",
		})
		return
	}

	// 创建文档记录
	doc := model.RAGDocument{
		UserID:   userID,
		FileName: file.Filename,
		FileType: ext[1:], // 去掉点
		FileSize: file.Size,
		Status:   "processing",
	}

	if err := database.DB.Create(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "文档创建失败",
		})
		return
	}

	// 异步处理文档（分块、向量化）
	go h.processDocument(doc.ID, userID, string(content))

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "文档上传成功，正在处理中",
		Data: gin.H{
			"document_id": doc.ID,
			"file_name":   doc.FileName,
			"status":      "processing",
		},
	})
}

// processDocument 处理文档（分块、向量化）
func (h *RAGHandler) processDocument(docID, userID uint, content string) {
	// 文本分块
	chunks := h.textSplitter.SplitText(content)
	if len(chunks) == 0 {
		database.DB.Model(&model.RAGDocument{}).Where("id = ?", docID).Update("status", "failed")
		return
	}

	// 批量向量化
	ctx := database.DB.Statement.Context
	embeddings, err := h.embeddingClient.GetEmbeddings(ctx, chunks)
	if err != nil {
		database.DB.Model(&model.RAGDocument{}).Where("id = ?", docID).Update("status", "failed")
		return
	}

	// 保存分块和向量
	for i, chunk := range chunks {
		chunkModel := model.RAGChunk{
			DocumentID: docID,
			UserID:     userID,
			Content:    chunk,
			ChunkIndex: i,
		}
		if i < len(embeddings) {
			chunkModel.Embedding = embeddings[i]
		}
		database.DB.Create(&chunkModel)
	}

	// 更新文档状态
	database.DB.Model(&model.RAGDocument{}).Where("id = ?", docID).Update("status", "completed")
}

// GetDocuments 获取文档列表
func (h *RAGHandler) GetDocuments(c *gin.Context) {
	userID := c.GetUint("userID")

	var documents []model.RAGDocument
	if err := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "获取文档列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
		Data:    documents,
	})
}

// GetDocument 获取单个文档
func (h *RAGHandler) GetDocument(c *gin.Context) {
	userID := c.GetUint("userID")
	docIDStr := c.Param("id")
	docID, _ := strconv.ParseUint(docIDStr, 10, 32)

	var doc model.RAGDocument
	if err := database.DB.Where("id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Code:    404,
			Message: "文档不存在",
		})
		return
	}

	// 获取分块
	var chunks []model.RAGChunk
	database.DB.Where("document_id = ?", docID).Order("chunk_index").Find(&chunks)

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"document": doc,
			"chunks":   chunks,
		},
	})
}

// DeleteDocument 删除文档
func (h *RAGHandler) DeleteDocument(c *gin.Context) {
	userID := c.GetUint("userID")
	docIDStr := c.Param("id")
	docID, _ := strconv.ParseUint(docIDStr, 10, 32)

	// 检查文档是否存在
	var doc model.RAGDocument
	if err := database.DB.Where("id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Code:    404,
			Message: "文档不存在",
		})
		return
	}

	// 删除分块
	database.DB.Where("document_id = ?", docID).Delete(&model.RAGChunk{})

	// 删除文档
	database.DB.Delete(&doc)

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
	})
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query     string  `json:"query" binding:"required"`
	TopK      int     `json:"top_k"`
	Threshold float64 `json:"threshold"`
}

// Search 搜索相关文档
func (h *RAGHandler) Search(c *gin.Context) {
	userID := c.GetUint("userID")

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "参数错误",
		})
		return
	}

	if req.TopK <= 0 {
		req.TopK = 3
	}
	if req.Threshold <= 0 {
		req.Threshold = 0.5
	}

	// 将查询向量化
	ctx := c.Request.Context()
	embedding, err := h.embeddingClient.GetEmbedding(ctx, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "向量化失败: " + err.Error(),
		})
		return
	}

	// 使用余弦相似度搜索（简化版本：取前TopK）
	var chunks []model.RAGChunk
	database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(req.TopK * 2).
		Find(&chunks)

	// 简单相似度计算（实际生产应该用SQL的向量运算）
	type scoredChunk struct {
		chunk     model.RAGChunk
		score     float64
	}
	var scoredChunks []scoredChunk

	for _, chunk := range chunks {
		if len(chunk.Embedding) == 0 {
			continue
		}
		score := cosineSimilarity(embedding, chunk.Embedding)
		if score >= req.Threshold {
			scoredChunks = append(scoredChunks, scoredChunk{chunk: chunk, score: score})
		}
	}

	// 按相似度排序
	for i := 0; i < len(scoredChunks)-1; i++ {
		for j := i + 1; j < len(scoredChunks); j++ {
			if scoredChunks[j].score > scoredChunks[i].score {
				scoredChunks[i], scoredChunks[j] = scoredChunks[j], scoredChunks[i]
			}
		}
	}

	// 取TopK
	result := make([]gin.H, 0, req.TopK)
	for i := 0; i < len(scoredChunks) && i < req.TopK; i++ {
		result = append(result, gin.H{
			"content":    scoredChunks[i].chunk.Content,
			"score":      scoredChunks[i].score,
			"chunk_index": scoredChunks[i].chunk.ChunkIndex,
		})
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
		Data:    result,
	})
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}

// RAGChat RAG增强的对话
func (h *RAGHandler) RAGChat(c *gin.Context) {
	userID := c.GetUint("userID")

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Code:    400,
			Message: "参数错误",
		})
		return
	}

	// 1. 将问题向量化
	ctx := c.Request.Context()
	embedding, err := h.embeddingClient.GetEmbedding(ctx, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: "向量化失败",
		})
		return
	}

	// 2. 搜索相关文档
	var chunks []model.RAGChunk
	database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(5).
		Find(&chunks)

	// 计算相似度，取Top3
	type scoredChunk struct {
		chunk model.RAGChunk
		score float64
	}
	var scoredChunks []scoredChunk

	for _, chunk := range chunks {
		if len(chunk.Embedding) == 0 {
			continue
		}
		score := cosineSimilarity(embedding, chunk.Embedding)
		if score >= 0.5 {
			scoredChunks = append(scoredChunks, scoredChunk{chunk: chunk, score: score})
		}
	}

	// 排序
	for i := 0; i < len(scoredChunks)-1; i++ {
		for j := i + 1; j < len(scoredChunks); j++ {
			if scoredChunks[j].score > scoredChunks[i].score {
				scoredChunks[i], scoredChunks[j] = scoredChunks[j], scoredChunks[i]
			}
		}
	}

	// 3. 构建Prompt（包含检索到的知识）
	var context strings.Builder
	for i := 0; i < len(scoredChunks) && i < 3; i++ {
		context.WriteString(fmt.Sprintf("[相关文档 %d]:\n%s\n\n", i+1, scoredChunks[i].chunk.Content))
	}

	prompt := fmt.Sprintf(`你是一个专业的AI助手。请根据以下参考资料回答用户的问题。

参考资料：
%s

用户问题：%s

请根据参考资料回答，如果参考资料中没有相关信息，请如实说明。`, context.String(), req.Message)

	// 4. 调用AI
	chatHandler, _ := NewChatHandler()

	// 简单调用（非流式）
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: prompt},
	}
	reply, err := chatHandler.client.Chat(ctx, messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"reply":   reply,
			"context": context.String(),
		},
	})
}
