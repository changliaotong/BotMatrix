package services

import (
	"BotMatrix/common/models"
	"BotMatrix/common/onebot"
	"context"
	"strings"
)

type SecurityService struct {
	store *models.Sz84Store
}

func NewSecurityService(store *models.Sz84Store) *SecurityService {
	return &SecurityService{store: store}
}

// AuditResult defines the action to take after auditing a message
type AuditResult struct {
	Blocked bool
	Action  string // "recall", "mute", "warn", "kick", "black"
	Reason  string
}

func (s *SecurityService) AuditMessage(ctx context.Context, e *onebot.Event) (*AuditResult, error) {
	if e.GroupID.String() == "" {
		return &AuditResult{Blocked: false}, nil
	}

	gid := e.GroupID.Int64()
	group := s.store.GetGroup(gid)
	if group == nil {
		return &AuditResult{Blocked: false}, nil
	}

	msg := e.RawMessage

	// 1. Check Recall Keywords
	if group.RecallKeyword != "" {
		keywords := strings.Split(group.RecallKeyword, "|")
		for _, k := range keywords {
			if k != "" && strings.Contains(msg, k) {
				return &AuditResult{Blocked: true, Action: "recall", Reason: "触发撤回关键词"}, nil
			}
		}
	}

	// 2. Check Mute Keywords
	if group.MuteKeyword != "" {
		keywords := strings.Split(group.MuteKeyword, "|")
		for _, k := range keywords {
			if k != "" && strings.Contains(msg, k) {
				return &AuditResult{Blocked: true, Action: "mute", Reason: "触发禁言关键词"}, nil
			}
		}
	}

	// 3. Check Warn Keywords
	if group.WarnKeyword != "" {
		keywords := strings.Split(group.WarnKeyword, "|")
		for _, k := range keywords {
			if k != "" && strings.Contains(msg, k) {
				return &AuditResult{Blocked: false, Action: "warn", Reason: "触发警告关键词"}, nil
			}
		}
	}

	return &AuditResult{Blocked: false}, nil
}

func (s *SecurityService) IsBlacklisted(ctx context.Context, botUin, groupId, targetId int64) bool {
	return s.store.IsBlacklisted(botUin, groupId, targetId)
}

func (s *SecurityService) IsWhitelisted(ctx context.Context, botUin, groupId, targetId int64) bool {
	return s.store.IsWhitelisted(botUin, groupId, targetId)
}

func (s *SecurityService) AddBlacklist(ctx context.Context, botUin, groupId, targetId int64, info string) error {
	return s.store.AddBlacklist(botUin, groupId, targetId, info)
}

func (s *SecurityService) RemoveBlacklist(ctx context.Context, botUin, groupId, targetId int64) error {
	return s.store.RemoveBlacklist(botUin, groupId, targetId)
}

func (s *SecurityService) AddWhitelist(ctx context.Context, botUin, groupId, targetId int64, info string) error {
	return s.store.AddWhitelist(botUin, groupId, targetId, info)
}

func (s *SecurityService) RemoveWhitelist(ctx context.Context, botUin, groupId, targetId int64) error {
	return s.store.RemoveWhitelist(botUin, groupId, targetId)
}
