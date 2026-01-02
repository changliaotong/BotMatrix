package rag

import (
	"regexp"
	"strings"
)

// Chunk 文本片段
type Chunk struct {
	Content string
	Title   string // 片段标题 (如 Markdown 的一级或二级标题)
}

// SimpleMarkdownChunker 简单的 Markdown 切分器
func SimpleMarkdownChunker(content string, minSize int) []Chunk {
	// 按标题切分 (H1, H2, H3)
	re := regexp.MustCompile(`(?m)^(#{1,3}\s+.*)$`)
	indices := re.FindAllStringIndex(content, -1)

	if len(indices) == 0 {
		return []Chunk{{Content: content}}
	}

	var chunks []Chunk
	lastPos := 0
	var currentTitle string

	for _, idx := range indices {
		if lastPos < idx[0] {
			text := strings.TrimSpace(content[lastPos:idx[0]])
			if len(text) >= minSize {
				chunks = append(chunks, Chunk{Content: text, Title: currentTitle})
			}
		}
		currentTitle = strings.TrimSpace(strings.TrimLeft(content[idx[0]:idx[1]], "# "))
		lastPos = idx[0] // 包含标题在内容中
	}

	// 最后一个片段
	if lastPos < len(content) {
		text := strings.TrimSpace(content[lastPos:])
		if len(text) >= minSize {
			chunks = append(chunks, Chunk{Content: text, Title: currentTitle})
		}
	}

	return chunks
}

// SimpleCodeChunker 简单的代码切分器 (按函数或结构体切分)
func SimpleCodeChunker(content string) []Chunk {
	// 匹配 // 或 /* 注释紧跟的 func 或 type 定义
	re := regexp.MustCompile(`(?m)((?://.*\n|/\*[\s\S]*?\*/\s*)*(?:func|type)\s+\w+[\s\S]*?\{)`)
	matches := re.FindAllStringIndex(content, -1)

	if len(matches) == 0 {
		return []Chunk{{Content: content}}
	}

	var chunks []Chunk
	for i, match := range matches {
		start := match[0]
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}

		text := strings.TrimSpace(content[start:end])
		// 提取函数名或结构体名作为标题
		titleRe := regexp.MustCompile(`(?:func|type)\s+(\w+)`)
		titleMatch := titleRe.FindStringSubmatch(text)
		title := "Code Snippet"
		if len(titleMatch) > 1 {
			title = titleMatch[1]
		}

		chunks = append(chunks, Chunk{Content: text, Title: title})
	}

	return chunks
}
