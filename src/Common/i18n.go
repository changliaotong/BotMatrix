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

// T 翻译指定的键
func T(lang string, key string, args ...interface{}) string {
	if GlobalTranslator == nil {
		return key
	}

	if lang == "" {
		lang = GlobalTranslator.DefaultLang
	}

	// 尝试精确匹配 (例如 en-US)
	translations, ok := GlobalTranslator.Translations[lang]
	if !ok {
		// 尝试基础语言匹配 (例如 en)
		baseLang := strings.Split(lang, "-")[0]
		translations, ok = GlobalTranslator.Translations[baseLang]
		if !ok {
			// 退回到默认语言
			translations, ok = GlobalTranslator.Translations[GlobalTranslator.DefaultLang]
		}
	}

	if !ok {
		return key
	}

	val, ok := translations[key]
	if !ok {
		// 如果当前语言没有这个键，尝试在默认语言中查找
		if lang != GlobalTranslator.DefaultLang {
			if defaultTrans, ok := GlobalTranslator.Translations[GlobalTranslator.DefaultLang]; ok {
				if defaultVal, ok := defaultTrans[key]; ok {
					val = defaultVal
				} else {
					return key
				}
			} else {
				return key
			}
		} else {
			return key
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}
	return val
}

// GetLangFromRequest 从请求中获取语言偏好
func GetLangFromRequest(r interface{}) string {
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
