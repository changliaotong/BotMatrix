package services

import (
	"BotMatrix/common/models"
	"regexp"
	"strings"
	"sync"
)

// PlaceholderContext contains the context for placeholder resolution
type PlaceholderContext struct {
	BotID   string
	GroupID string
	UserID  string
	Name    string
	// 复刻 sz84 原版，预加载核心对象以减少重复查询
	Group  *models.GroupInfo
	User   *models.UserInfo
	Member *models.GroupMember
}

// PlaceholderHandler is a function that returns a replacement string based on context
type PlaceholderHandler func(ctx *PlaceholderContext) string

// PlaceholderService handles registration and resolution of placeholders in messages
type PlaceholderService struct {
	handlers     map[string]PlaceholderHandler
	descriptions map[string]string
	enabled      map[string]bool
	mu           sync.RWMutex
	pattern      *regexp.Regexp
	ifPattern    *regexp.Regexp
}

func NewPlaceholderService() *PlaceholderService {
	return &PlaceholderService{
		handlers:     make(map[string]PlaceholderHandler),
		descriptions: make(map[string]string),
		enabled:      make(map[string]bool),
		// Pattern: {(?P<key>[^}:|]+)(\|(?P<default>[^}]+))?}
		pattern:   regexp.MustCompile(`\{(?P<key>[^}:|]+)(\|(?P<default>[^}]+))?\}`),
		ifPattern: regexp.MustCompile(`\{if:(?P<cond>[^{}?]+)\?(?P<trueVal>[^:{}]*)\:(?P<falseVal>[^}]+)\}`),
	}
}

// Register adds a placeholder handler
func (s *PlaceholderService) Register(name string, handler PlaceholderHandler, description string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[name] = handler
	s.descriptions[name] = description
	s.enabled[name] = true
}

// Replace resolves placeholders in the given string with context
func (s *PlaceholderService) Replace(input string, ctx *PlaceholderContext) string {
	result := input
	maxDepth := 5

	for i := 0; i < maxDepth; i++ {
		if !strings.Contains(result, "{") {
			break
		}

		// 1. Replace IF conditions
		result = s.replaceIf(result, ctx)

		// 2. Replace standard placeholders
		result = s.replacePlaceholders(result, ctx)
	}

	// Unescape \{ and \}
	result = strings.ReplaceAll(result, `\{`, "{")
	result = strings.ReplaceAll(result, `\}`, "}")

	return result
}

func (s *PlaceholderService) replaceIf(input string, ctx *PlaceholderContext) string {
	return s.ifPattern.ReplaceAllStringFunc(input, func(match string) string {
		submatches := s.ifPattern.FindStringSubmatch(match)
		if len(submatches) < 4 {
			return match
		}

		cond := strings.TrimSpace(submatches[1])
		trueVal := submatches[2]
		falseVal := submatches[3]

		// Simple condition evaluation: check if placeholder exists and is not empty
		if s.evaluateCondition(cond, ctx) {
			return trueVal
		}
		return falseVal
	})
}

func (s *PlaceholderService) evaluateCondition(cond string, ctx *PlaceholderContext) bool {
	// For now, simple implementation: if it's a registered placeholder, check its value
	// In the future, this can be expanded to support ==, !=, >, < etc.
	s.mu.RLock()
	handler, ok := s.handlers[cond]
	enabled := s.enabled[cond]
	s.mu.RUnlock()

	if ok && enabled {
		val := handler(ctx)
		return val != "" && val != "0" && val != "false"
	}

	// Support direct comparison if needed
	if strings.Contains(cond, "==") {
		parts := strings.Split(cond, "==")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			return s.getVal(left, ctx) == s.getVal(right, ctx)
		}
	}

	return false
}

func (s *PlaceholderService) getVal(name string, ctx *PlaceholderContext) string {
	s.mu.RLock()
	handler, ok := s.handlers[name]
	enabled := s.enabled[name]
	s.mu.RUnlock()

	if ok && enabled {
		return handler(ctx)
	}
	return name // Return as literal if not found
}

func (s *PlaceholderService) replacePlaceholders(input string, ctx *PlaceholderContext) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.pattern.ReplaceAllStringFunc(input, func(match string) string {
		submatches := s.pattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		key := submatches[1]
		defaultVal := ""
		if len(submatches) > 3 && submatches[3] != "" {
			// submatches[2] is "|default", submatches[3] is "default"
			defaultVal = submatches[3]
		}

		if handler, ok := s.handlers[key]; ok && s.enabled[key] {
			val := handler(ctx)
			if val == "" {
				return defaultVal
			}
			return val
		}

		if defaultVal != "" {
			return defaultVal
		}
		return match
	})
}
