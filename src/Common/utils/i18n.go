package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Translator handles multi-language translations
type Translator struct {
	Translations map[string]map[string]string
	DefaultLang  string
}

// GlobalTranslator is the global translation instance
var GlobalTranslator *Translator

// InitTranslator initializes the translator
func InitTranslator(localesPath string, defaultLang string) {
	t := &Translator{
		Translations: make(map[string]map[string]string),
		DefaultLang:  defaultLang,
	}

	files, err := os.ReadDir(localesPath)
	if err != nil {
		log.Printf("[WARN] Failed to read locales directory %s: %v", localesPath, err)
		GlobalTranslator = t
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			lang := strings.TrimSuffix(file.Name(), ".json")
			filePath := filepath.Join(localesPath, file.Name())

			data, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("[WARN] Failed to read translation file %s: %v", filePath, err)
				continue
			}

			var translations map[string]string
			if err := json.Unmarshal(data, &translations); err != nil {
				log.Printf("[WARN] Failed to parse translation file %s: %v", filePath, err)
				continue
			}

			t.Translations[lang] = translations
			log.Printf("[INFO] Loaded language pack: %s (%d entries)", lang, len(translations))
		}
	}

	GlobalTranslator = t
}

// T translates a key. Supports "key|default text" format.
func T(lang string, key string, args ...any) string {
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

	// 统一语言代码格式
	lang = strings.ReplaceAll(lang, "_", "-")
	
	// 处理中文变体
	if lang == "zh" {
		lang = "zh-CN"
	} else if strings.HasPrefix(lang, "zh-") {
		// 如果有对应的语言包则不处理，否则默认回退
		if _, ok := GlobalTranslator.Translations[lang]; !ok {
			if strings.Contains(strings.ToLower(lang), "tw") || strings.Contains(strings.ToLower(lang), "hk") {
				lang = "zh-TW"
			} else {
				lang = "zh-CN"
			}
		}
	}

	translations, ok := GlobalTranslator.Translations[lang]
	if !ok {
		baseLang := strings.Split(lang, "-")[0]
		translations, ok = GlobalTranslator.Translations[baseLang]
	}

	if ok {
		if val, ok := translations[id]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(val, args...)
			}
			return val
		}
	}

	// 如果没找到翻译，且是中文，则返回默认文本（通常代码里写的就是中文默认文本）
	if lang == "zh-CN" {
		if len(args) > 0 {
			return fmt.Sprintf(defaultText, args...)
		}
		return defaultText
	}

	if ok {
		if val, exists := translations[id]; exists {
			if len(args) > 0 {
				return fmt.Sprintf(val, args...)
			}
			return val
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(defaultText, args...)
	}
	return defaultText
}

// GetLangFromRequest extracts language from request header or cookie
func GetLangFromRequest(r *http.Request) string {
	lang := r.Header.Get("Accept-Language")
	if lang == "" {
		if cookie, err := r.Cookie("lang"); err == nil {
			lang = cookie.Value
		}
	}
	if lang == "" {
		lang = "zh-CN"
	}
	// Simplified lang extraction (e.g., "en-US,en;q=0.9" -> "en-US")
	if strings.Contains(lang, ",") {
		lang = strings.Split(lang, ",")[0]
	}
	return lang
}
