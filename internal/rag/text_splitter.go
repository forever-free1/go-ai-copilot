package rag

import (
	"unicode"
)

// TextSplitter 文本分块器
type TextSplitter struct {
	ChunkSize    int // 块大小（字符数）
	ChunkOverlap int // 块重叠字符数
}

// NewTextSplitter 创建文本分块器
func NewTextSplitter(chunkSize, chunkOverlap int) *TextSplitter {
	if chunkSize <= 0 {
		chunkSize = 1024
	}
	if chunkOverlap <= 0 {
		chunkOverlap = 256
	}
	if chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 4
	}
	return &TextSplitter{
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
	}
}

// SplitText 将文本分割成块
func (s *TextSplitter) SplitText(text string) []string {
	if text == "" {
		return nil
	}

	// 按段落分割
	paragraphs := s.splitByParagraph(text)
	if len(paragraphs) == 0 {
		return nil
	}

	// 如果单个段落就超过chunkSize，需要进一步分割
	var chunks []string
	currentChunk := ""

	for _, para := range paragraphs {
		paraLen := len(para)

		// 如果单个段落就超过chunkSize
		if paraLen > s.ChunkSize {
			// 先保存当前的chunk
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
				currentChunk = ""
			}
			// 对这个段落进行分割
			chunks = append(chunks, s.splitLargeChunk(para)...)
			continue
		}

		// 如果加上当前段落超过chunkSize，保存当前chunk，开始新的
		if len(currentChunk)+paraLen+1 > s.ChunkSize {
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
			}
			// 新chunk从overlap部分开始
			if len(currentChunk) > s.ChunkOverlap {
				currentChunk = currentChunk[len(currentChunk)-s.ChunkOverlap:]
			} else {
				currentChunk = ""
			}
		}

		// 添加段落
		if currentChunk != "" {
			currentChunk += "\n" + para
		} else {
			currentChunk = para
		}
	}

	// 保存最后一个chunk
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// splitByParagraph 按段落分割
func (s *TextSplitter) splitByParagraph(text string) []string {
	var paragraphs []string
	var current []rune

	for _, r := range text {
		if r == '\n' {
			// 遇到换行，检查是否连续
			if len(current) > 0 {
				para := string(current)
				para = s.trimWhitespace(para)
				if para != "" {
					paragraphs = append(paragraphs, para)
				}
				current = nil
			}
		} else {
			current = append(current, r)
		}
	}

	// 处理最后一个段落
	if len(current) > 0 {
		para := s.trimWhitespace(string(current))
		if para != "" {
			paragraphs = append(paragraphs, para)
		}
	}

	return paragraphs
}

// splitLargeChunk 分割过大的块
func (s *TextSplitter) splitLargeChunk(text string) []string {
	var chunks []string
	runes := []rune(text)

	for i := 0; i < len(runes); i += s.ChunkSize - s.ChunkOverlap {
		end := i + s.ChunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunk := string(runes[i:end])
		chunk = s.trimWhitespace(chunk)
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// trimWhitespace 去除首尾空白
func (s *TextSplitter) trimWhitespace(text string) string {
	runes := []rune(text)
	start := 0
	end := len(runes)

	for start < end && unicode.IsSpace(runes[start]) {
		start++
	}
	for end > start && unicode.IsSpace(runes[end-1]) {
		end--
	}

	if start >= end {
		return ""
	}
	return string(runes[start:end])
}
