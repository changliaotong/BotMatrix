package types

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// SensitiveType 定义敏感信息类型
type SensitiveType string

const (
	SensitivePhone  SensitiveType = "PHONE"
	SensitiveEmail  SensitiveType = "EMAIL"
	SensitiveIDCard SensitiveType = "IDCARD"
	SensitiveIP     SensitiveType = "IP"
	SensitiveCustom SensitiveType = "CUSTOM"
)

// PrivacyFilter 处理敏感信息的识别与替换
type PrivacyFilter struct {
	patterns map[SensitiveType]*regexp.Regexp
	mu       sync.RWMutex
}

// NewPrivacyFilter 创建一个新的隐私过滤器
func NewPrivacyFilter() *PrivacyFilter {
	f := &PrivacyFilter{
		patterns: make(map[SensitiveType]*regexp.Regexp),
	}
	f.initDefaultPatterns()
	return f
}

func (f *PrivacyFilter) initDefaultPatterns() {
	// 简单手机号正则
	f.patterns[SensitivePhone] = regexp.MustCompile(`(1[3-9]\d{9})`)
	// 邮箱正则
	f.patterns[SensitiveEmail] = regexp.MustCompile(`([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
	// 身份证正则 (简单 18 位)
	f.patterns[SensitiveIDCard] = regexp.MustCompile(`([1-9]\d{5}[1-9]\d{3}((0\d)|(1[0-2]))(([0|1|2]\d)|3[0-1])\d{3}([0-9]|X))`)
	// IP 地址正则
	f.patterns[SensitiveIP] = regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
}

// MaskContext 包含一次掩码操作的上下文，用于后续还原
type MaskContext struct {
	OriginalMap map[string]string // Placeholder -> Original
	MaskedMap   map[string]string // Original -> Placeholder
	Counter     int
}

// NewMaskContext 创建掩码上下文
func NewMaskContext() *MaskContext {
	return &MaskContext{
		OriginalMap: make(map[string]string),
		MaskedMap:   make(map[string]string),
		Counter:     0,
	}
}

// Mask 替换文本中的敏感信息，并返回替换后的文本和上下文
func (f *PrivacyFilter) Mask(text string, ctx *MaskContext) string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := text
	for sType, pattern := range f.patterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// 如果已经替换过，直接返回对应的占位符
			if placeholder, ok := ctx.MaskedMap[match]; ok {
				return placeholder
			}

			// 生成新的占位符
			ctx.Counter++
			placeholder := fmt.Sprintf("[%s_%d]", string(sType), ctx.Counter)
			ctx.OriginalMap[placeholder] = match
			ctx.MaskedMap[match] = placeholder
			return placeholder
		})
	}
	return result
}

// Unmask 将占位符还原为原始信息
func (f *PrivacyFilter) Unmask(text string, ctx *MaskContext) string {
	result := text
	for placeholder, original := range ctx.OriginalMap {
		result = strings.ReplaceAll(result, placeholder, original)
	}
	return result
}

// AddCustomPattern 添加自定义脱敏正则
func (f *PrivacyFilter) AddCustomPattern(name string, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.patterns[SensitiveType(name)] = re
	return nil
}
