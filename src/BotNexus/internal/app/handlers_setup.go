package app

import (
	"BotMatrix/common/bot"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// HandleGetMemberSetup 获取机器人配置列表
func HandleGetMemberSetup(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)
		claims, ok := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			utils.SendJSONResponse(w, false, utils.T(lang, "not_logged_in"), nil)
			return
		}

		// 默认查询所有，如果是普通用户，仅能看到自己名下的机器人
		adminIDStr := r.URL.Query().Get("admin_id")

		var bots []models.BotInfo
		query := m.GORMDB.Model(&models.BotInfo{})

		if !claims.IsAdmin {
			// 普通用户仅能看到自己名下的
			query = query.Where("AdminId = ?", claims.UserID)
		} else if adminIDStr != "" {
			// 管理员可以指定查询某个用户的
			adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
			query = query.Where("AdminId = ?", adminID)
		}

		if err := query.Find(&bots).Error; err != nil {
			utils.SendJSONResponse(w, false, "查询失败: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "", map[string]interface{}{
			"bots": bots,
		})
	}
}

// HandleUpdateMemberSetup 更新机器人配置
func HandleUpdateMemberSetup(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)
		claims, ok := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			utils.SendJSONResponse(w, false, utils.T(lang, "not_logged_in"), nil)
			return
		}

		var updateData models.BotInfo
		if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request"), nil)
			return
		}

		if updateData.BotUin == 0 {
			utils.SendJSONResponse(w, false, "BotUin is required", nil)
			return
		}

		// 如果不是管理员，需要校验是否是自己的机器人
		if !claims.IsAdmin {
			var currentBot models.BotInfo
			if err := m.GORMDB.Where("BotUin = ?", updateData.BotUin).First(&currentBot).Error; err != nil {
				utils.SendJSONResponse(w, false, "未找到该机器人或无权操作", nil)
				return
			}
			if currentBot.AdminId != claims.UserID {
				utils.SendJSONResponse(w, false, "无权操作该机器人", nil)
				return
			}
		}

		// 仅允许更新特定字段，防止误操作
		err := m.GORMDB.Model(&models.BotInfo{}).Where("BotUin = ?", updateData.BotUin).Updates(map[string]interface{}{
			"BotName":        updateData.BotName,
			"BotMemo":        updateData.BotMemo,
			"WemcomeMessage": updateData.WemcomeMessage,
			"IsCredit":       updateData.IsCredit,
			"IsGroup":        updateData.IsGroup,
			"IsPrivate":      updateData.IsPrivate,
			"Valid":          updateData.Valid,
			"IsFreeze":       updateData.IsFreeze,
			"IsBlock":        updateData.IsBlock,
			"IsVip":          updateData.IsVip,
			"AdminId":        updateData.AdminId,
			"Password":       updateData.Password,
			"BotType":        updateData.BotType,
			"ApiIP":          updateData.ApiIP,
			"ApiPort":        updateData.ApiPort,
			"ApiKey":         updateData.ApiKey,
			"IsSignalR":      updateData.IsSignalR,
			"WebUIToken":     updateData.WebUIToken,
			"WebUIPort":      updateData.WebUIPort,
		}).Error

		if err != nil {
			utils.SendJSONResponse(w, false, "更新失败: "+err.Error(), nil)
			return
		}

		// 获取更新后的完整数据并更新缓存
		var updatedBot models.BotInfo
		if m.GORMDB.Where("BotUin = ?", updateData.BotUin).First(&updatedBot).Error == nil {
			if m.Rdb != nil {
				ctx := r.Context()
				pascalData := utils.ToPascalMap(updatedBot)
				jsonData, _ := json.Marshal(pascalData)

				// 1. 简单实体缓存 (DefaultCacheRepository)
				// C# Key: entity:BotInfo:123456
				m.Rdb.Set(ctx, fmt.Sprintf("entity:BotInfo:%d", updatedBot.BotUin), jsonData, 0)

				// 2. ORM 级缓存 (MetaData)
				// C# Key: MetaData:[sz84_robot].[dbo].[Member]:Id:123456
				m.Rdb.Set(ctx, fmt.Sprintf("MetaData:[sz84_robot].[dbo].[Member]:Id:%d", updatedBot.BotUin), jsonData, 0)
			}
		}

		utils.SendJSONResponse(w, true, "更新成功", nil)
	}
}

// HandleDeleteMemberSetup 删除机器人配置
func HandleDeleteMemberSetup(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)
		claims, ok := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			utils.SendJSONResponse(w, false, utils.T(lang, "not_logged_in"), nil)
			return
		}

		botUinStr := r.URL.Query().Get("bot_uin")
		botUin, _ := strconv.ParseInt(botUinStr, 10, 64)

		if botUin == 0 {
			utils.SendJSONResponse(w, false, "BotUin is required", nil)
			return
		}

		// 如果不是管理员，需要校验是否是自己的机器人
		if !claims.IsAdmin {
			var currentBot models.BotInfo
			if err := m.GORMDB.Where("BotUin = ?", botUin).First(&currentBot).Error; err != nil {
				utils.SendJSONResponse(w, false, "未找到该机器人或无权操作", nil)
				return
			}
			if currentBot.AdminId != claims.UserID {
				utils.SendJSONResponse(w, false, "无权操作该机器人", nil)
				return
			}
		}

		if err := m.GORMDB.Where("BotUin = ?", botUin).Delete(&models.BotInfo{}).Error; err != nil {
			utils.SendJSONResponse(w, false, "删除失败: "+err.Error(), nil)
			return
		}

		// 清除缓存
		if m.Rdb != nil {
			// 1. 简单实体缓存 (DefaultCacheRepository)
			m.Rdb.Del(r.Context(), fmt.Sprintf("entity:BotInfo:%d", botUin))
			// 2. ORM 级缓存 (MetaData)
			m.Rdb.Del(r.Context(), fmt.Sprintf("MetaData:[sz84_robot].[dbo].[Member]:Id:%d", botUin))
		}

		utils.SendJSONResponse(w, true, "删除成功", nil)
	}
}

// HandleGetGroupSetup 获取群组配置列表
func HandleGetGroupSetup(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)
		claims, ok := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			utils.SendJSONResponse(w, false, utils.T(lang, "not_logged_in"), nil)
			return
		}

		robotOwnerStr := r.URL.Query().Get("robot_owner")
		groupOwnerStr := r.URL.Query().Get("group_owner")

		var groups []models.GroupInfo
		query := m.GORMDB.Model(&models.GroupInfo{})

		if !claims.IsAdmin {
			// 普通用户仅能看到自己名下的群组（作为机器人主人或群主人）
			// 这里优先匹配 RobotOwnerName 为自己的情况
			query = query.Where("RobotOwnerName = ? OR GroupOwnerName = ?", claims.Username, claims.Username)
		} else {
			if robotOwnerStr != "" {
				query = query.Where("RobotOwnerName = ?", robotOwnerStr)
			}
			if groupOwnerStr != "" {
				query = query.Where("GroupOwnerName = ?", groupOwnerStr)
			}
		}

		if err := query.Find(&groups).Error; err != nil {
			utils.SendJSONResponse(w, false, "查询失败: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "", map[string]interface{}{
			"groups": groups,
		})
	}
}

// HandleUpdateGroupSetup 更新群组配置
func HandleUpdateGroupSetup(m *bot.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := utils.GetLangFromRequest(r)
		claims, ok := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			utils.SendJSONResponse(w, false, utils.T(lang, "not_logged_in"), nil)
			return
		}

		var updateData models.GroupInfo
		if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
			utils.SendJSONResponse(w, false, utils.T(lang, "invalid_request"), nil)
			return
		}

		if updateData.Id == 0 {
			utils.SendJSONResponse(w, false, "Group ID is required", nil)
			return
		}

		// 如果不是管理员，需要校验是否是自己的群组
		if !claims.IsAdmin {
			var currentGroup models.GroupInfo
			if err := m.GORMDB.Where("Id = ?", updateData.Id).First(&currentGroup).Error; err != nil {
				utils.SendJSONResponse(w, false, "未找到该群组或无权操作", nil)
				return
			}
			if currentGroup.RobotOwnerName != claims.Username && currentGroup.GroupOwnerName != claims.Username {
				utils.SendJSONResponse(w, false, "无权操作该群组", nil)
				return
			}
		}

		// 仅允许更新配置相关的字段
		err := m.GORMDB.Model(&models.GroupInfo{}).Where("Id = ?", updateData.Id).Updates(map[string]interface{}{
			"GroupName":             updateData.GroupName,
			"GroupMemo":             updateData.GroupMemo,
			"IsOpen":                updateData.IsOpen,
			"UseRight":              updateData.UseRight,
			"TeachRight":            updateData.TeachRight,
			"AdminRight":            updateData.AdminRight,
			"WelcomeMessage":        updateData.WelcomeMessage,
			"SystemPrompt":          updateData.SystemPrompt,
			"IsPowerOn":             updateData.IsPowerOn,
			"IsWelcomeHint":         updateData.IsWelcomeHint,
			"IsExitHint":            updateData.IsExitHint,
			"IsKickHint":            updateData.IsKickHint,
			"IsChangeHint":          updateData.IsChangeHint,
			"IsRightHint":           updateData.IsRightHint,
			"IsCloudBlack":          updateData.IsCloudBlack,
			"IsCloudAnswer":         updateData.IsCloudAnswer,
			"IsRequirePrefix":       updateData.IsRequirePrefix,
			"IsSz84":                updateData.IsSz84,
			"IsWarn":                updateData.IsWarn,
			"IsBlock":               updateData.IsBlock,
			"IsWhite":               updateData.IsWhite,
			"RobotOwnerName":        updateData.RobotOwnerName,
			"GroupOwnerName":        updateData.GroupOwnerName,
			"IsAcceptNewMember":     updateData.IsAcceptNewMember,
			"RejectMessage":         updateData.RejectMessage,
			"RegexRequestJoin":      updateData.RegexRequestJoin,
			"RecallKeyword":         updateData.RecallKeyword,
			"WarnKeyword":           updateData.WarnKeyword,
			"MuteKeyword":           updateData.MuteKeyword,
			"KickKeyword":           updateData.KickKeyword,
			"BlackKeyword":          updateData.BlackKeyword,
			"WhiteKeyword":          updateData.WhiteKeyword,
			"CreditKeyword":         updateData.CreditKeyword,
			"MuteEnterCount":        updateData.MuteEnterCount,
			"MuteKeywordCount":      updateData.MuteKeywordCount,
			"KickCount":             updateData.KickCount,
			"BlackCount":            updateData.BlackCount,
			"CardNamePrefixBoy":     updateData.CardNamePrefixBoy,
			"CardNamePrefixGirl":    updateData.CardNamePrefixGirl,
			"CardNamePrefixManager": updateData.CardNamePrefixManager,
			"IsMuteRefresh":         updateData.IsMuteRefresh,
			"MuteRefreshCount":      updateData.MuteRefreshCount,
			"IsProp":                updateData.IsProp,
			"IsPet":                 updateData.IsPet,
			"IsBlackRefresh":        updateData.IsBlackRefresh,
			"IsConfirmNew":          updateData.IsConfirmNew,
			"IsCredit":              updateData.IsCredit,
			"IsHintClose":           updateData.IsHintClose,
			"RecallTime":            updateData.RecallTime,
			"IsInvite":              updateData.IsInvite,
			"InviteCredit":          updateData.InviteCredit,
			"IsReplyImage":          updateData.IsReplyImage,
			"IsReplyRecall":         updateData.IsReplyRecall,
			"IsVoiceReply":          updateData.IsVoiceReply,
			"VoiceId":               updateData.VoiceId,
			"IsAI":                  updateData.IsAI,
			"IsOwnerPay":            updateData.IsOwnerPay,
			"ContextCount":          updateData.ContextCount,
			"IsMultAI":              updateData.IsMultAI,
			"IsAutoSignin":          updateData.IsAutoSignin,
			"IsUseKnowledgebase":    updateData.IsUseKnowledgebase,
			"IsSendHelpInfo":        updateData.IsSendHelpInfo,
			"IsRecall":              updateData.IsRecall,
			"IsCreditSystem":        updateData.IsCreditSystem,
			"IsCloseManager":        updateData.IsCloseManager,
			"IsBlackExit":           updateData.IsBlackExit,
			"IsBlackKick":           updateData.IsBlackKick,
			"IsBlackShare":          updateData.IsBlackShare,
			"IsChangeEnter":         updateData.IsChangeEnter,
			"IsMuteEnter":           updateData.IsMuteEnter,
			"IsChangeMessage":       updateData.IsChangeMessage,
			"ParentGroup":           updateData.ParentGroup,
			"BlockMin":              updateData.BlockMin,
			"CityName":              updateData.CityName,
			"FansName":              updateData.FansName,
		}).Error

		if err != nil {
			utils.SendJSONResponse(w, false, "更新失败: "+err.Error(), nil)
			return
		}

		// 获取更新后的完整数据并更新缓存
		var updatedGroup models.GroupInfo
		if m.GORMDB.Where("Id = ?", updateData.Id).First(&updatedGroup).Error == nil {
			if m.Rdb != nil {
				ctx := r.Context()
				pascalData := utils.ToPascalMap(updatedGroup)
				jsonData, _ := json.Marshal(pascalData)

				// 1. 简单实体缓存 (DefaultCacheRepository)
				// C# Key: entity:GroupInfo:123456
				m.Rdb.Set(ctx, fmt.Sprintf("entity:GroupInfo:%d", updatedGroup.Id), jsonData, 0)

				// 2. ORM 级缓存 (MetaData)
				// C# Key: MetaData:[sz84_robot].[dbo].[Group]:Id:123456
				m.Rdb.Set(ctx, fmt.Sprintf("MetaData:[sz84_robot].[dbo].[Group]:Id:%d", updatedGroup.Id), jsonData, 0)
			}
		}

		utils.SendJSONResponse(w, true, "更新成功", nil)
	}
}
