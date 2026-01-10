package server

import (
	"context"
	"fmt"
	"botworker/internal/onebot"
	"botworker/internal/services"
)

func NewSecurityMiddleware(service *services.SecurityService, settingService *services.SettingService) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(event *onebot.Event) error {
			// 0. Check Global Switch
			enabledStr, _ := settingService.GetGlobalSetting(context.Background(), "Func_BlackList")
			if enabledStr == "false" {
				return next(event)
			}

			// Extract BotUin, GroupId, UserId from event
			botUin := event.SelfID.Int64()
			groupId := event.GroupID.Int64()
			userId := event.UserID.Int64()

			// Skip check for self (bot itself)
			if userId == botUin {
				return next(event)
			}

			// Blacklist Check
			if service.IsBlacklisted(context.Background(), botUin, groupId, userId) {
				// Blocked
				return fmt.Errorf("blocked by blacklist: user %d in group %d", userId, groupId)
			}
			
			// Whitelist logic could be added here if needed

			return next(event)
		}
	}
}
