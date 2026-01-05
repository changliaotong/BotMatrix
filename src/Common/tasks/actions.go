package tasks

import (
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"encoding/json"
	"fmt"
	"time"
)

func (d *Dispatcher) registerDefaultActions() {
	d.actions["send_message"] = d.handleSendMessage
	d.actions["mute_group"] = d.handleMuteGroup
	d.actions["unmute_group"] = d.handleUnmuteGroup
	d.actions["kick_member"] = d.handleKickMember
	d.actions["set_group_admin"] = d.handleSetGroupAdmin
	d.actions["skill_call"] = d.handleSkillCall
}

func (d *Dispatcher) handleKickMember(task models.Task, execution *models.Execution) error {
	var params struct {
		BotID            string `json:"bot_id"`
		GroupID          string `json:"group_id"`
		UserID           string `json:"user_id"`
		RejectAddRequest bool   `json:"reject_add_request"`
	}

	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	// 权限预检查
	members, err := d.manager.GetGroupMembers(params.BotID, params.GroupID)
	if err == nil && len(members) > 0 {
		var botRole, targetRole string
		for _, m := range members {
			if m.UserID == params.BotID {
				botRole = m.Role
			}
			if m.UserID == params.UserID {
				targetRole = m.Role
			}
		}

		canKick := false
		if botRole == "owner" {
			canKick = true
		} else if botRole == "admin" {
			if targetRole == "member" {
				canKick = true
			}
		}

		if !canKick {
			return fmt.Errorf("permission denied: bot (%s) cannot kick target (%s)", botRole, targetRole)
		}
	}

	return d.manager.SendBotAction(params.BotID, "set_group_kick", params)
}

func (d *Dispatcher) handleSetGroupAdmin(task models.Task, execution *models.Execution) error {
	var params struct {
		BotID   string `json:"bot_id"`
		GroupID string `json:"group_id"`
		UserID  string `json:"user_id"`
		Enable  bool   `json:"enable"`
	}

	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	return d.manager.SendBotAction(params.BotID, "set_group_admin", params)
}

func (d *Dispatcher) handleSkillCall(task models.Task, execution *models.Execution) error {
	var params map[string]any
	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	skillName, _ := params["skill"].(string)
	if skillName == "" {
		return fmt.Errorf("missing skill name")
	}

	// 查找目标 Worker ID
	workerID, _ := params["worker_id"].(string)
	if workerID == "" {
		// 自动发现具备该技能的 Worker
		workerID = d.manager.FindWorkerBySkill(skillName)
	}

	if workerID == "" {
		return fmt.Errorf("no worker available for skill: %s", skillName)
	}

	cmd := types.WorkerCommand{
		Type:        "skill_call",
		Skill:       skillName,
		Params:      params,
		TaskID:      task.ID,
		ExecutionID: execution.ExecutionID,
		Timestamp:   time.Now().Unix(),
	}

	return d.manager.SendToWorker(workerID, cmd)
}

func (d *Dispatcher) handleSendMessage(task models.Task, execution *models.Execution) error {
	var params struct {
		BotID    string `json:"bot_id"`
		GroupID  string `json:"group_id"`
		UserID   string `json:"user_id"`
		Message  string `json:"message"`
		IsPublic bool   `json:"is_public"`
	}

	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	action := "send_group_msg"
	actionParams := make(map[string]any)
	actionParams["message"] = params.Message

	if params.GroupID != "" {
		actionParams["group_id"] = params.GroupID
	} else if params.UserID != "" {
		action = "send_private_msg"
		actionParams["user_id"] = params.UserID
	} else {
		return fmt.Errorf("missing group_id or user_id")
	}

	return d.manager.SendBotAction(params.BotID, action, actionParams)
}

func (d *Dispatcher) handleMuteGroup(task models.Task, execution *models.Execution) error {
	var params struct {
		BotID    string `json:"bot_id"`
		GroupID  string `json:"group_id"`
		UserID   string `json:"user_id"`  // 可选，禁言特定用户
		Duration uint32 `json:"duration"` // 禁言时长，秒
	}

	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	if params.UserID != "" {
		return d.manager.SendBotAction(params.BotID, "set_group_ban", struct {
			GroupID  string `json:"group_id"`
			UserID   string `json:"user_id"`
			Duration uint32 `json:"duration"`
		}{
			GroupID:  params.GroupID,
			UserID:   params.UserID,
			Duration: params.Duration,
		})
	}

	// 全员禁言
	return d.manager.SendBotAction(params.BotID, "set_group_whole_ban", struct {
		GroupID string `json:"group_id"`
		Enable  bool   `json:"enable"`
	}{
		GroupID: params.GroupID,
		Enable:  true,
	})
}

func (d *Dispatcher) handleUnmuteGroup(task models.Task, execution *models.Execution) error {
	var params struct {
		BotID   string `json:"bot_id"`
		GroupID string `json:"group_id"`
		UserID  string `json:"user_id"` // 可选，解禁特定用户
	}

	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	if params.UserID != "" {
		return d.manager.SendBotAction(params.BotID, "set_group_ban", struct {
			GroupID  string `json:"group_id"`
			UserID   string `json:"user_id"`
			Duration uint32 `json:"duration"`
		}{
			GroupID:  params.GroupID,
			UserID:   params.UserID,
			Duration: 0,
		})
	}

	// 取消全员禁言
	return d.manager.SendBotAction(params.BotID, "set_group_whole_ban", struct {
		GroupID string `json:"group_id"`
		Enable  bool   `json:"enable"`
	}{
		GroupID: params.GroupID,
		Enable:  false,
	})
}
