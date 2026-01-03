package app

import (
	"BotMatrix/common/types"
	"BotNexus/tasks"
	"encoding/json"
	"fmt"
)

// ExecuteAIDraft 执行 AI 生成的任务草稿 (实现 tasks.TaskExecutor 接口)
func (m *Manager) ExecuteAIDraft(draft *tasks.AIDraft) error {
	return m.executeAIDraft(draft)
}

// executeAIDraft 执行 AI 生成的任务草稿
func (m *Manager) executeAIDraft(draft *tasks.AIDraft) error {
	var err error
	switch draft.Intent {
	case string(tasks.AIActionCreateTask):
		var task tasks.Task
		if err := json.Unmarshal([]byte(draft.Data), &task); err != nil {
			return fmt.Errorf("解析任务数据失败: %v", err)
		}
		task.CreatorID = draft.UserID
		err = m.TaskManager.CreateTask(&task, true)
	case string(tasks.AIActionAdjustPolicy):
		var policyReq struct {
			StrategyName string `json:"strategy_name"`
			Action       string `json:"action"` // enable, disable
			Config       string `json:"config"`
		}
		if err := json.Unmarshal([]byte(draft.Data), &policyReq); err != nil {
			return fmt.Errorf("解析策略数据失败: %v", err)
		}

		isEnabled := policyReq.Action == "enable"
		err = m.GORMDB.Model(&tasks.Strategy{}).
			Where("name = ?", policyReq.StrategyName).
			Updates(map[string]any{
				"is_enabled": isEnabled,
				"config":     policyReq.Config,
			}).Error
		if err != nil {
			return fmt.Errorf("更新策略失败: %v", err)
		}

	case string(tasks.AIActionManageTags):
		var tagReq struct {
			TargetType string `json:"target_type"` // group, friend
			TargetID   string `json:"target_id"`
			TagName    string `json:"tag_name"`
			Action     string `json:"action"` // add, remove
		}
		if err := json.Unmarshal([]byte(draft.Data), &tagReq); err != nil {
			return fmt.Errorf("解析标签数据失败: %v", err)
		}

		if tagReq.Action == "remove" {
			err = m.TaskManager.Tagging.RemoveTag(tagReq.TargetType, tagReq.TargetID, tagReq.TagName)
		} else {
			err = m.TaskManager.Tagging.AddTag(tagReq.TargetType, tagReq.TargetID, tagReq.TagName)
		}

	case string(tasks.AIActionCancelTask):
		var cancelReq struct {
			TaskID uint `json:"task_id"`
		}
		if err := json.Unmarshal([]byte(draft.Data), &cancelReq); err != nil {
			// 如果不是直接的 task_id，尝试寻找“刚才的任务”
			var tasksList []tasks.Task
			m.GORMDB.Where("creator_id = ? AND status = ?", draft.UserID, tasks.TaskPending).
				Order("created_at DESC").Limit(1).Find(&tasksList)
			if len(tasksList) > 0 {
				cancelReq.TaskID = tasksList[0].ID
			} else {
				return fmt.Errorf("未找到可取消的任务")
			}
		}

		if cancelReq.TaskID == 0 {
			// 兜底：尝试寻找“刚才的任务”
			var tasksList []tasks.Task
			m.GORMDB.Where("creator_id = ? AND status = ?", draft.UserID, tasks.TaskPending).
				Order("created_at DESC").Limit(1).Find(&tasksList)
			if len(tasksList) > 0 {
				cancelReq.TaskID = tasksList[0].ID
			} else {
				return fmt.Errorf("请指定要取消的任务 ID")
			}
		}
		err = m.TaskManager.CancelTask(cancelReq.TaskID, fmt.Sprintf("%d", draft.UserID))

	case string(tasks.AIActionSkillCall):
		var skillReq struct {
			Skill  string         `json:"skill"`
			Params map[string]any `json:"params"`
		}
		if err := json.Unmarshal([]byte(draft.Data), &skillReq); err != nil {
			return fmt.Errorf("解析技能数据失败: %v", err)
		}
		workerID := m.FindWorkerBySkill(skillReq.Skill)
		if workerID == "" {
			err = fmt.Errorf("未找到具备该能力的 Worker: %s", skillReq.Skill)
		} else {
			// 构造指令发送给 Worker
			cmd := types.WorkerCommand{
				Type:   "skill_call",
				Skill:  skillReq.Skill,
				Params: skillReq.Params,
				UserID: fmt.Sprintf("%d", draft.UserID),
			}
			err = m.SendToWorker(workerID, cmd)
		}
	case string(tasks.AIActionBatch):
		var batch []struct {
			Intent string          `json:"intent"`
			Data   json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal([]byte(draft.Data), &batch); err != nil {
			return fmt.Errorf("解析批量任务数据失败: %v", err)
		}

		for _, item := range batch {
			subDraft := &tasks.AIDraft{
				UserID:   draft.UserID,
				GroupID:  draft.GroupID,
				UserRole: draft.UserRole,
				Intent:   item.Intent,
				Data:     string(item.Data),
				Status:   "confirmed",
			}
			if err := m.executeAIDraft(subDraft); err != nil {
				return err // 只要有一个失败就返回错误（或者可以改为收集错误继续执行）
			}
		}

	default:
		err = fmt.Errorf("未知的 AI 意图: %s", draft.Intent)
	}

	return err
}
