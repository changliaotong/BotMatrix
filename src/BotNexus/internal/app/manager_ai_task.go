package app

import (
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// ExecuteAIDraft 执行 AI 生成的任务草稿 (实现 tasks.TaskExecutor 接口)
func (m *Manager) ExecuteAIDraft(draft *models.AIDraft) error {
	return m.executeAIDraft(draft)
}

// executeAIDraft 执行 AI 生成的任务草稿
func (m *Manager) executeAIDraft(draft *models.AIDraft) error {
	var err error
	switch draft.Intent {
	case string(types.AIActionCreateTask):
		var task models.Task
		if err := json.Unmarshal([]byte(draft.Data), &task); err != nil {
			return fmt.Errorf("解析任务数据失败: %v", err)
		}
		task.CreatorID = draft.UserID
		// AI 生成的任务，默认视为企业版权限 (或根据 UserRole 判断)
		isEnterprise := draft.UserRole == "admin" || draft.UserRole == "owner"
		err = m.TaskManager.CreateTask(&task, isEnterprise)
	case string(types.AIActionAdjustPolicy):
		var policyReq struct {
			StrategyName string `json:"strategy_name"`
			Action       string `json:"action"` // enable, disable
			Config       string `json:"config"`
		}
		if err := json.Unmarshal([]byte(draft.Data), &policyReq); err != nil {
			return fmt.Errorf("解析策略数据失败: %v", err)
		}

		isEnabled := policyReq.Action == "enable"
		err = m.GORMDB.Model(&models.Strategy{}).
			Where("name = ?", policyReq.StrategyName).
			Updates(map[string]any{
				"is_enabled": isEnabled,
				"config":     policyReq.Config,
			}).Error
		if err != nil {
			return fmt.Errorf("更新策略失败: %v", err)
		}

	case string(types.AIActionManageTags):
		var tagReq struct {
			TargetType string   `json:"target_type"`
			TargetID   string   `json:"target_id"`
			Tags       []string `json:"tags"`
			Action     string   `json:"action"` // add, remove
		}
		if err := json.Unmarshal([]byte(draft.Data), &tagReq); err != nil {
			return fmt.Errorf("解析标签数据失败: %v", err)
		}

		for _, tagName := range tagReq.Tags {
			if tagReq.Action == "remove" {
				err = m.TaskManager.Tagging.RemoveTag(tagReq.TargetType, tagReq.TargetID, tagName)
			} else {
				err = m.TaskManager.Tagging.AddTag(tagReq.TargetType, tagReq.TargetID, tagName)
			}
			if err != nil {
				break
			}
		}

	case string(types.AIActionCancelTask):
		var cancelReq struct {
			TaskID uint `json:"task_id"`
		}
		if err := json.Unmarshal([]byte(draft.Data), &cancelReq); err != nil {
			return fmt.Errorf("解析取消任务数据失败: %v", err)
		}
		err = m.GORMDB.Model(&models.Task{}).Where("id = ?", cancelReq.TaskID).Update("status", "disabled").Error

	case string(types.AIActionSkillCall):
		var skillReq struct {
			SkillName string         `json:"skill_name"`
			Params    map[string]any `json:"params"`
		}
		if err := json.Unmarshal([]byte(draft.Data), &skillReq); err != nil {
			return fmt.Errorf("解析技能调用数据失败: %v", err)
		}
		// 异步执行技能调用，不等待结果
		go m.SyncSkillCall(context.Background(), skillReq.SkillName, skillReq.Params)
		err = nil

	case string(types.AIActionBatch):
		var batchData struct {
			Actions []*types.ParseResult `json:"actions"`
		}
		if err := json.Unmarshal([]byte(draft.Data), &batchData); err != nil {
			return fmt.Errorf("解析批量任务数据失败: %v", err)
		}

		for _, action := range batchData.Actions {
			// 转换为 AIDraft 进行递归处理
			subDraft := &models.AIDraft{
				UserID:   draft.UserID,
				UserRole: draft.UserRole,
				Intent:   string(action.Intent),
				Data:     fmt.Sprint(action.Data),
			}
			if subErr := m.executeAIDraft(subDraft); subErr != nil {
				clog.Error("[AI] 批量执行子任务失败", zap.Error(subErr))
			}
		}
		err = nil

	default:
		err = fmt.Errorf("未知的 AI 意图: %s", draft.Intent)
	}

	return err
}
