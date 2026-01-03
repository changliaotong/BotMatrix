package rag

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"BotMatrix/common/ai"
	"BotNexus/tasks"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/xuri/excelize/v2"
)

// Chunk 文本片段
type Chunk struct {
	Content string
	Title   string // 片段标题 (如 Markdown 的一级或二级标题)
}

// ContentParser 文档解析器接口
type ContentParser interface {
	Parse(ctx context.Context, content []byte) []Chunk
}

// AIParser 基础 AI 解析器
type AIParser struct {
	svc     tasks.AIService
	modelID uint
}

// ImageParser 图像解析器实现
type ImageParser struct {
	AIParser
}

func (p *ImageParser) Parse(ctx context.Context, content []byte) []Chunk {
	description, err := DescribeImage(ctx, p.svc, p.modelID, content)
	if err != nil {
		return []Chunk{{Content: "[图片识别失败]"}}
	}
	return []Chunk{{Content: description}}
}

// DescribeImage 使用 AI 视觉模型描述图片
func DescribeImage(ctx context.Context, svc tasks.AIService, modelID uint, imageContent []byte) (string, error) {
	if svc == nil {
		return "", fmt.Errorf("AI service not available")
	}

	// 将图片转换为 base64
	base64Data := base64.StdEncoding.EncodeToString(imageContent)
	// 简单的多模态消息构造
	// 注意：由于 ai.Message 目前只支持 string Content，我们暂时用特殊标记或者期待适配
	// 这里我们直接调用 Chat，如果底层适配器支持 vision 格式，它应该能处理

	// 针对 OpenAI 兼容接口的多模态格式
	messages := []ai.Message{
		{
			Role: ai.RoleUser,
			Content: []ai.ContentPart{
				{
					Type: "text",
					Text: "请详细描述这张图片的内容，如果是文档扫描件，请提取其中的文字。",
				},
				{
					Type: "image_url",
					ImageURL: &ai.ImageURLValue{
						URL: fmt.Sprintf("data:image/jpeg;base64,%s", base64Data),
					},
				},
			},
		},
	}

	resp, err := svc.Chat(ctx, modelID, messages, nil)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	choice := resp.Choices[0]
	if s, ok := choice.Message.Content.(string); ok {
		return s, nil
	}
	if parts, ok := choice.Message.Content.([]ai.ContentPart); ok {
		var sb strings.Builder
		for _, p := range parts {
			if p.Type == "text" {
				sb.WriteString(p.Text)
			}
		}
		return sb.String(), nil
	}

	return "", fmt.Errorf("unexpected content type in AI response")
}

// MarkdownParser Markdown 解析器实现
type MarkdownParser struct {
	MinSize int
}

func (p *MarkdownParser) Parse(ctx context.Context, content []byte) []Chunk {
	return SimpleMarkdownChunker(string(content), p.MinSize)
}

// CodeParser 代码解析器实现
type CodeParser struct{}

func (p *CodeParser) Parse(ctx context.Context, content []byte) []Chunk {
	return SimpleCodeChunker(string(content))
}

// DefaultParser 默认解析器 (不切分)
type DefaultParser struct{}

func (p *DefaultParser) Parse(ctx context.Context, content []byte) []Chunk {
	return []Chunk{{Content: string(content)}}
}

// TxtParser 纯文本解析器实现
type TxtParser struct {
	MinSize int
}

func (p *TxtParser) Parse(ctx context.Context, content []byte) []Chunk {
	text := string(content)
	// 按双换行符切分段落
	paragraphs := strings.Split(text, "\n\n")
	var chunks []Chunk
	var currentChunk strings.Builder

	for _, pText := range paragraphs {
		pText = strings.TrimSpace(pText)
		if pText == "" {
			continue
		}

		if currentChunk.Len()+len(pText) > 1000 { // 超过 1000 字符强制切分
			if currentChunk.Len() >= p.MinSize {
				chunks = append(chunks, Chunk{Content: currentChunk.String()})
			}
			currentChunk.Reset()
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(pText)
	}

	if currentChunk.Len() >= p.MinSize {
		chunks = append(chunks, Chunk{Content: currentChunk.String()})
	}

	// 如果没有切出任何内容 (可能文件太小)，则作为一个整体
	if len(chunks) == 0 && len(text) > 0 {
		chunks = append(chunks, Chunk{Content: text})
	}

	return chunks
}

// ExcelParser Excel 解析器实现
type ExcelParser struct{}

func (p *ExcelParser) Parse(ctx context.Context, content []byte) []Chunk {
	text, err := ExcelToText(content)
	if err != nil {
		return []Chunk{{Content: string(content)}} // 降级
	}
	// Excel 内容通常按 Sheet 切分
	return SimpleMarkdownChunker(text, 20)
}

// ExcelToText 从 Excel 二进制数据中提取纯文本
func ExcelToText(content []byte) (string, error) {
	reader := bytes.NewReader(content)
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var sb strings.Builder
	for _, sheetName := range f.GetSheetList() {
		sb.WriteString(fmt.Sprintf("## Sheet: %s\n", sheetName))
		rows, err := f.GetRows(sheetName)
		if err != nil {
			continue
		}
		for _, row := range rows {
			sb.WriteString(strings.Join(row, "\t"))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// PDFParser PDF 解析器实现
type PDFParser struct {
	AIParser
}

func (p *PDFParser) Parse(ctx context.Context, content []byte) []Chunk {
	text, err := PDFToText(content)
	if err != nil {
		text = string(content) // 降级
	}
	chunks := SimpleMarkdownChunker(text, 50)

	// 如果有 AI 服务，尝试提取并识别图片
	if p.svc != nil {
		images, err := ExtractImagesFromPDF(content)
		if err == nil && len(images) > 0 {
			for i, imgData := range images {
				description, err := DescribeImage(ctx, p.svc, p.modelID, imgData)
				if err == nil {
					chunks = append(chunks, Chunk{
						Content: fmt.Sprintf("\n\n### PDF 图片内容 [%d]\n%s", i+1, description),
					})
				}
			}
		}
	}

	return chunks
}

// ExtractImagesFromPDF 从 PDF 二进制数据中提取所有图片
func ExtractImagesFromPDF(content []byte) ([][]byte, error) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "rag_pdf_*.pdf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(content); err != nil {
		return nil, err
	}

	// 创建临时输出目录
	tmpDir, err := os.MkdirTemp("", "rag_pdf_img_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	// 使用 pdfcpu 提取图片
	conf := model.NewDefaultConfiguration()
	if err := api.ExtractImagesFile(tmpFile.Name(), tmpDir, nil, conf); err != nil {
		return nil, err
	}

	// 读取提取出的图片
	var images [][]byte
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		imgPath := filepath.Join(tmpDir, f.Name())
		imgData, err := os.ReadFile(imgPath)
		if err == nil {
			images = append(images, imgData)
		}
	}

	return images, nil
}

// PDFToText 从 PDF 二进制数据中提取纯文本
func PDFToText(content []byte) (string, error) {
	reader := bytes.NewReader(content)
	r, err := pdf.NewReader(reader, int64(len(content)))
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for pageIndex := 1; pageIndex <= r.NumPage(); pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// DocxParser Docx 解析器实现
type DocxParser struct{}

func (p *DocxParser) Parse(ctx context.Context, content []byte) []Chunk {
	text, err := DocxToText(content)
	if err != nil {
		return []Chunk{{Content: string(content)}} // 降级到纯文本
	}
	// Docx 通常按段落切分，或者直接使用 SimpleMarkdownChunker 的逻辑 (如果包含标题)
	return SimpleMarkdownChunker(text, 50)
}

// DocParser 旧版 Doc 解析器 (暂存，目前仅支持提取可见字符)
type DocParser struct{}

func (p *DocParser) Parse(ctx context.Context, content []byte) []Chunk {
	// 简单的可见字符提取 (针对旧版 binary doc 的极简处理)
	var sb strings.Builder
	for _, b := range content {
		if (b >= 32 && b <= 126) || b == '\n' || b == '\r' || b == '\t' {
			sb.WriteByte(b)
		}
	}
	return []Chunk{{Content: sb.String()}}
}

// DocxToText 从 docx 二进制数据中提取纯文本
func DocxToText(content []byte) (string, error) {
	reader := bytes.NewReader(content)
	zipReader, err := zip.NewReader(reader, int64(len(content)))
	if err != nil {
		return "", err
	}

	var xmlFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == "word/document.xml" {
			xmlFile = f
			break
		}
	}

	if xmlFile == nil {
		return "", io.EOF
	}

	rc, err := xmlFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	var sb strings.Builder
	decoder := xml.NewDecoder(rc)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "p" {
				sb.WriteString("\n")
			}
		case xml.CharData:
			sb.Write(se)
		}
	}

	return strings.TrimSpace(sb.String()), nil
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
	var chunks []Chunk

	// 1. 提取包级别注释和声明
	pkgRe := regexp.MustCompile(`(?m)^((?://.*\n|/\*[\s\S]*?\*/\s*)*package\s+\w+)`)
	pkgMatch := pkgRe.FindStringIndex(content)
	if pkgMatch != nil {
		chunks = append(chunks, Chunk{
			Content: strings.TrimSpace(content[pkgMatch[0]:pkgMatch[1]]),
			Title:   "Package Declaration",
		})
	}

	// 2. 匹配 // 或 /* 注释紧跟的 func 或 type 定义
	re := regexp.MustCompile(`(?m)((?://.*\n|/\*[\s\S]*?\*/\s*)*(?:func|type)\s+\w+[\s\S]*?\{)`)
	matches := re.FindAllStringIndex(content, -1)

	if len(matches) == 0 {
		if len(chunks) == 0 {
			return []Chunk{{Content: content}}
		}
		return chunks
	}

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
