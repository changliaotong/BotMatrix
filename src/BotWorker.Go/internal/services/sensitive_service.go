package services

import (
	"BotMatrix/common/models"
	"context"
	"regexp"
	"strings"
	"sync"
)

type SensitiveWordService struct {
	store      *models.Sz84Store
	cache      []models.SensitiveWord
	cacheMu    sync.RWMutex
	regexCache map[string]*regexp.Regexp
}

func NewSensitiveWordService(store *models.Sz84Store) (*SensitiveWordService, error) {
	s := &SensitiveWordService{
		store:      store,
		regexCache: make(map[string]*regexp.Regexp),
	}
	// Initial load, ignore error if table is empty or missing, but log it in real app
	_ = s.Reload(context.Background())
	return s, nil
}

func (s *SensitiveWordService) Reload(ctx context.Context) error {
	words, err := s.store.GetSensitiveWords()
	if err != nil {
		return err
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.cache = words

	// Rebuild regex cache
	s.regexCache = make(map[string]*regexp.Regexp)
	for _, w := range words {
		if w.Type == 1 { // Regex
			if r, err := regexp.Compile(w.Word); err == nil {
				s.regexCache[w.Word] = r
			}
		}
	}
	return nil
}

func (s *SensitiveWordService) Check(ctx context.Context, content string) *models.SensitiveWord {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	for _, w := range s.cache {
		if w.Type == 0 { // Exact/Contains
			if strings.Contains(content, w.Word) {
				return &w
			}
		} else if w.Type == 1 { // Regex
			if r, ok := s.regexCache[w.Word]; ok {
				if r.MatchString(content) {
					return &w
				}
			}
		}
	}
	return nil
}

// CheckGroup checks group-specific sensitive words
func (s *SensitiveWordService) CheckGroup(ctx context.Context, groupId int64, content string) (*models.SensitiveWord, *models.GroupInfo, string) {
	group := s.store.GetGroup(groupId)
	if group == nil {
		return nil, nil, ""
	}

	// Helper to check regex string (pipe separated)
	checkRegex := func(pattern string) bool {
		if pattern == "" {
			return false
		}
		r, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return r.MatchString(content)
	}

	// Priority: Black > Kick > Mute > Credit > Warn > Recall

	// Black (Action 4)
	if checkRegex(group.BlackKeyword) {
		return &models.SensitiveWord{
			Word:   group.BlackKeyword, // Return the pattern as word
			Action: 4,
		}, group, "Black"
	}

	// Kick (Action 3)
	if checkRegex(group.KickKeyword) {
		return &models.SensitiveWord{
			Word:   group.KickKeyword,
			Action: 3,
		}, group, "Kick"
	}

	// Mute (Action 2)
	if checkRegex(group.MuteKeyword) {
		return &models.SensitiveWord{
			Word:     group.MuteKeyword,
			Action:   2,
			Duration: int64(group.MuteKeywordCount * 60), // Minutes to seconds
		}, group, "Mute"
	}

	// Credit (Action 5)
	if checkRegex(group.CreditKeyword) {
		return &models.SensitiveWord{
			Word:   group.CreditKeyword,
			Action: 5,
		}, group, "Credit"
	}

	// Warn (Action 6)
	if checkRegex(group.WarnKeyword) {
		return &models.SensitiveWord{
			Word:   group.WarnKeyword,
			Action: 6,
		}, group, "Warn"
	}

	// Recall (Action 1)
	if checkRegex(group.RecallKeyword) {
		return &models.SensitiveWord{
			Word:   group.RecallKeyword,
			Action: 1,
		}, group, "Recall"
	}

	return nil, nil, ""
}

func (s *SensitiveWordService) AddWord(ctx context.Context, word string, matchType, action int, duration int64) error {
	if err := s.store.AddSensitiveWord(word, matchType, action, duration); err != nil {
		return err
	}
	return s.Reload(ctx)
}

func (s *SensitiveWordService) RemoveWord(ctx context.Context, id int64) error {
	if err := s.store.RemoveSensitiveWord(id); err != nil {
		return err
	}
	return s.Reload(ctx)
}

// Wrappers for actions
func (s *SensitiveWordService) DeductCredit(botUin, groupId int64, groupName string, userId int64, userName string, amount int64, reason string) error {
	// CoinsType_GroupCredit = 5
	_, err := s.store.AddCoins(botUin, groupId, groupName, userId, userName, 5, -amount, reason)
	return err
}

func (s *SensitiveWordService) AddWarn(botUin, groupId, userId int64, reason string, insertBy int64) error {
	return s.store.AppendWarn(botUin, groupId, userId, reason, insertBy)
}

func (s *SensitiveWordService) GetWarnCount(groupId, userId int64) (int64, error) {
	return s.store.WarnCount(groupId, userId)
}

func (s *SensitiveWordService) AddBlacklist(botUin, groupId, userId int64, reason string) error {
	return s.store.AddBlacklist(botUin, groupId, userId, reason)
}

func (s *SensitiveWordService) ClearWarn(groupId, userId int64) error {
	return s.store.ClearWarn(groupId, userId)
}
