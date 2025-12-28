package common

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Translator 处理多语言翻译
type Translator struct {
	Translations map[string]map[string]string
	DefaultLang  string
}

// GlobalTranslator 全局翻译实例
var GlobalTranslator *Translator

// InitTranslator 初始化翻译器
func InitTranslator(localesPath string, defaultLang string) {
	t := &Translator{
		Translations: make(map[string]map[string]string),
		DefaultLang:  defaultLang,
	}

	files, err := os.ReadDir(localesPath)
	if err != nil {
		log.Printf("[WARN] 无法读取本地化目录 %s: %v", localesPath, err)
		// 如果读取失败，至少保证一个空的翻译器
		GlobalTranslator = t
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			lang := strings.TrimSuffix(file.Name(), ".json")
			filePath := filepath.Join(localesPath, file.Name())

			data, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("[WARN] 无法读取翻译文件 %s: %v", filePath, err)
				continue
			}

			var translations map[string]string
			if err := json.Unmarshal(data, &translations); err != nil {
				log.Printf("[WARN] 无法解析翻译文件 %s: %v", filePath, err)
				continue
			}

			t.Translations[lang] = translations
			log.Printf("[INFO] 已加载语言包: %s (%d 条词条)", lang, len(translations))
		}
	}

	GlobalTranslator = t
}

// T 翻译指定的键。支持 "键名|默认文本" 格式。
// 如果语言是中文，则优先使用默认文本。
// 如果语言不是中文，则根据键名在语言包中查找翻译。
func T(lang string, key string, args ...any) string {
	// 解析 key，支持 "ID|默认文本" 格式
	id := key
	defaultText := key
	if strings.Contains(key, "|") {
		parts := strings.SplitN(key, "|", 2)
		id = parts[0]
		defaultText = parts[1]
	}

	if GlobalTranslator == nil {
		if len(args) > 0 {
			return fmt.Sprintf(defaultText, args...)
		}
		return defaultText
	}

	if lang == "" {
		lang = GlobalTranslator.DefaultLang
	}

	// 如果目标语言是中文，直接使用默认文本（代码中的中文）
	if lang == "zh-CN" || lang == "zh" || strings.HasPrefix(lang, "zh-") {
		if len(args) > 0 {
			return fmt.Sprintf(defaultText, args...)
		}
		return defaultText
	}

	// 尝试在翻译包中查找
	// 尝试精确匹配 (例如 en-US)
	translations, ok := GlobalTranslator.Translations[lang]
	if !ok {
		// 尝试基础语言匹配 (例如 en)
		baseLang := strings.Split(lang, "-")[0]
		translations, ok = GlobalTranslator.Translations[baseLang]
		if !ok {
			// 退回到默认语言，但如果默认语言是中文，我们已经处理过了
			// 这里处理的是默认语言非中文的情况
			translations, ok = GlobalTranslator.Translations[GlobalTranslator.DefaultLang]
		}
	}

	if !ok {
		if len(args) > 0 {
			return fmt.Sprintf(defaultText, args...)
		}
		return defaultText
	}

	val, ok := translations[id]
	if !ok {
		// 如果当前语言没有这个键，尝试在默认语言中查找
		if lang != GlobalTranslator.DefaultLang {
			if defaultTrans, ok := GlobalTranslator.Translations[GlobalTranslator.DefaultLang]; ok {
				if defaultVal, ok := defaultTrans[id]; ok {
					val = defaultVal
				} else {
					if len(args) > 0 {
						return fmt.Sprintf(defaultText, args...)
					}
					return defaultText
				}
			} else {
				if len(args) > 0 {
					return fmt.Sprintf(defaultText, args...)
				}
				return defaultText
			}
		} else {
			if len(args) > 0 {
				return fmt.Sprintf(defaultText, args...)
			}
			return defaultText
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}
	return val
}

// GetLangFromRequest 从请求中获取语言偏好
func GetLangFromRequest(r any) string {
	// 这里可以扩展，比如从 Cookie, Header 或 Query 参数中获取
	// 目前简单实现，优先从 Header "Accept-Language" 获取
	if req, ok := r.(*http.Request); ok {
		lang := req.Header.Get("Accept-Language")
		if lang != "" {
			// 取第一个首选语言
			return strings.Split(strings.Split(lang, ",")[0], ";")[0]
		}
	}
	return ""
}
