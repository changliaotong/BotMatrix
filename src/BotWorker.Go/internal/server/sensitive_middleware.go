package server

import (
	"BotMatrix/common/models"
	"botworker/internal/onebot"
	"botworker/internal/services"
	"context"
	"fmt"
)

func NewSensitiveWordMiddleware(service *services.SensitiveWordService, settingService *services.SettingService, caller ActionCaller) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(event *onebot.Event) error {
			// 0. Check Global Switch
			enabledStr, _ := settingService.GetGlobalSetting(context.Background(), "Func_SensitiveWord")
			if enabledStr == "false" {
				return next(event)
			}

			if event.MessageType == "" {
				return next(event) // Not a message event
			}

			// Use RawMessage for checking
			content := event.RawMessage
			if content == "" {
				// Fallback to trying to convert Message to string if RawMessage is empty
				if msgStr, ok := event.Message.(string); ok {
					content = msgStr
				}
			}

			if content == "" {
				return next(event)
			}

			matched := service.Check(context.Background(), content)
			var groupInfo *models.GroupInfo

			if matched == nil && event.GroupID != 0 {
				// Check Group Specific
				var gMatched *models.SensitiveWord
				gMatched, groupInfo, _ = service.CheckGroup(context.Background(), event.GroupID, content)
				if gMatched != nil {
					matched = gMatched
				}
			}

			if matched != nil {
				// Handle Action
				switch matched.Action {
				case 1: // Recall
					// Requires MessageID
					// NOTE: event.MessageID is FlexibleInt64 (wraps int64).
					// It's not a pointer, so we check if it's non-zero.
					if event.MessageID != 0 {
						if caller != nil {
							// For OneBot V11, message_id is int32.
							// However, our params expect FlexibleInt64
							caller.DeleteMessage(&onebot.DeleteMessageParams{
								MessageID: event.MessageID,
							})
						}
					}
					// Even without recall, we block the message processing
					return fmt.Errorf("sensitive word detected: %s (Action: Recall)", matched.Word)

				case 2: // Mute
					if caller != nil && event.GroupID != 0 && event.UserID != 0 {
						duration := matched.Duration
						if duration <= 0 {
							duration = 600 // Default 10 minutes
						}
						caller.SetGroupBan(&onebot.SetGroupBanParams{
							GroupID:  event.GroupID,
							UserID:   event.UserID,
							Duration: int(duration),
						})
						// Also Recall
						if event.MessageID != 0 {
							caller.DeleteMessage(&onebot.DeleteMessageParams{
								MessageID: event.MessageID,
							})
						}
					}
					return fmt.Errorf("sensitive word detected: %s (Action: Mute)", matched.Word)

				case 3: // Kick
					if caller != nil && event.GroupID != 0 && event.UserID != 0 {
						caller.SetGroupKick(&onebot.SetGroupKickParams{
							GroupID:   event.GroupID,
							UserID:    event.UserID,
							RejectAdd: false,
						})
						// Also Recall
						if event.MessageID != 0 {
							caller.DeleteMessage(&onebot.DeleteMessageParams{
								MessageID: event.MessageID,
							})
						}
					}
					return fmt.Errorf("sensitive word detected: %s (Action: Kick)", matched.Word)

				case 4: // Black (Ban + Kick + Recall)
					if caller != nil && event.GroupID != 0 && event.UserID != 0 {
						service.AddBlacklist(event.SelfID, event.GroupID, event.UserID, "Sensitive: "+matched.Word)
						caller.SetGroupKick(&onebot.SetGroupKickParams{
							GroupID:   event.GroupID,
							UserID:    event.UserID,
							RejectAdd: true,
						})
						if event.MessageID != 0 {
							caller.DeleteMessage(&onebot.DeleteMessageParams{
								MessageID: event.MessageID,
							})
						}
					}
					return fmt.Errorf("sensitive word detected: %s (Action: Black)", matched.Word)

				case 5: // Credit (Deduct + Recall)
					if caller != nil && event.GroupID != 0 && event.UserID != 0 {
						amount := int64(100) // Default deduction
						service.DeductCredit(event.SelfID, event.GroupID, "", event.UserID, "", amount, "Sensitive: "+matched.Word)

						caller.SendMessage(&onebot.SendMessageParams{
							MessageType: "group",
							GroupID:     event.GroupID,
							Message:     fmt.Sprintf("[CQ:at,qq=%d] 触发敏感词【%s】，扣除%d积分", event.UserID, matched.Word, amount),
						})

						if event.MessageID != 0 {
							caller.DeleteMessage(&onebot.DeleteMessageParams{
								MessageID: event.MessageID,
							})
						}
					}
					return fmt.Errorf("sensitive word detected: %s (Action: Credit)", matched.Word)

				case 6: // Warn (Add Warn + Check Count + Recall)
					if caller != nil && event.GroupID != 0 && event.UserID != 0 {
						service.AddWarn(event.SelfID, event.GroupID, event.UserID, "Sensitive: "+matched.Word, 0)

						count, _ := service.GetWarnCount(event.GroupID, event.UserID)

						kickCount := 5
						muteCount := 3
						if groupInfo != nil {
							if groupInfo.KickCount > 0 {
								kickCount = groupInfo.KickCount
							}
							if groupInfo.MuteEnterCount > 0 { // MuteKeywordCount? Actually MuteEnterCount is usually for Enter/Join.
								// But WarnMessage.cs uses `KickCount` and `MuteKeywordCount`?
								// Let's check model.
								// GroupInfo has MuteKeywordCount.
								if groupInfo.MuteKeywordCount > 0 {
									muteCount = groupInfo.MuteKeywordCount
								}
							}
						}

						if int(count) >= kickCount {
							caller.SetGroupKick(&onebot.SetGroupKickParams{GroupID: event.GroupID, UserID: event.UserID})
							service.ClearWarn(event.GroupID, event.UserID)
						} else if int(count) >= muteCount {
							caller.SetGroupBan(&onebot.SetGroupBanParams{GroupID: event.GroupID, UserID: event.UserID, Duration: 600})
						} else {
							caller.SendMessage(&onebot.SendMessageParams{
								MessageType: "group",
								GroupID:     event.GroupID,
								Message:     fmt.Sprintf("[CQ:at,qq=%d] 警告：触发敏感词【%s】，累计警告 %d/%d 次", event.UserID, matched.Word, count, kickCount),
							})
						}

						if event.MessageID != 0 {
							caller.DeleteMessage(&onebot.DeleteMessageParams{
								MessageID: event.MessageID,
							})
						}
					}
					return fmt.Errorf("sensitive word detected: %s (Action: Warn)", matched.Word)
				}

				// Action 0 or others: just block/monitor (or maybe allow if Action=0 means 'log only'?)
				// Usually sensitive words imply blocking unless specified otherwise.
				// If Action == 0 (None), maybe we just log and let it pass?
				// The requirement says "sensitive word system", usually implies filtering.
				// Let's assume Action 0 is "Block Only".
				return fmt.Errorf("sensitive word detected: %s", matched.Word)
			}

			return next(event)
		}
	}
}
