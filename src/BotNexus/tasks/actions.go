package tasks

import (
	"BotMatrix/common"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// BotManager 定义调度中心需要的机器人管理能力
type BotManager interface {
	SendBotAction(botID string, action string, params any) error
	SendToWorker(workerID string, msg common.WorkerCommand) error
	FindWorkerBySkill(skillName string) string // 返回 WorkerID
	GetTags(targetType string, targetID string) []string
	GetTargetsByTags(targetType string, tags []string, logic string) []string
}

func (d *Dispatcher) registerDefaultActions() {
	d.actions["send_message"] = d.handleSendMessage
	d.actions["mute_group"] = d.handleMuteGroup
	d.actions["unmute_group"] = d.handleUnmuteGroup
	d.actions["skill_call"] = d.handleSkillCall
}

func (d *Dispatcher) handleSkillCall(task Task, execution *Execution) error {
	var params map[string]any
	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	skillName, _ := params["skill"].(string)
	if skillName == "" {
		return fmt.Errorf("missing skill name")
	}

	bm := d.manager.(BotManager)

	// 查找目标 Worker ID
	workerID, _ := params["worker_id"].(string)
	if workerID == "" {
		// 自动发现具备该技能的 Worker
		workerID = bm.FindWorkerBySkill(skillName)
	}

	if workerID == "" {
		return fmt.Errorf("no worker available for skill: %s", skillName)
	}

	cmd := common.WorkerCommand{
		Type:        "skill_call",
		Skill:       skillName,
		Params:      params,
		TaskID:      task.ID,
		ExecutionID: execution.ExecutionID,
		Timestamp:   time.Now().Unix(),
	}

	return bm.SendToWorker(workerID, cmd)
}

func (d *Dispatcher) sendToQueue(queue string, payload []byte) error {
	if d.rdb == nil {
		return fmt.Errorf("redis client not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 使用 RPush 配合 Worker 端的 BLPop 实现 FIFO 队列
	return d.rdb.RPush(ctx, queue, payload).Err()
}

func (d *Dispatcher) handleSendMessage(task Task, execution *Execution) error {
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

	bm := d.manager.(BotManager)
	return bm.SendBotAction(params.BotID, action, actionParams)
}

func (d *Dispatcher) handleMuteGroup(task Task, execution *Execution) error {
	var params struct {
		BotID    string `json:"bot_id"`
		GroupID  string `json:"group_id"`
		UserID   string `json:"user_id"`  // 可选，禁言特定用户
		Duration uint32 `json:"duration"` // 禁言时长，秒
	}

	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	bm := d.manager.(BotManager)
	if params.UserID != "" {
		return bm.SendBotAction(params.BotID, "set_group_ban", struct {
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
	return bm.SendBotAction(params.BotID, "set_group_whole_ban", struct {
		GroupID string `json:"group_id"`
		Enable  bool   `json:"enable"`
	}{
		GroupID: params.GroupID,
		Enable:  true,
	})
}

func (d *Dispatcher) handleUnmuteGroup(task Task, execution *Execution) error {
	var params struct {
		BotID   string `json:"bot_id"`
		GroupID string `json:"group_id"`
		UserID  string `json:"user_id"`
	}

	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	bm := d.manager.(BotManager)
	if params.UserID != "" {
		return bm.SendBotAction(params.BotID, "set_group_ban", struct {
			GroupID  string `json:"group_id"`
			UserID   string `json:"user_id"`
			Duration uint32 `json:"duration"`
		}{
			GroupID:  params.GroupID,
			UserID:   params.UserID,
			Duration: 0,
		})
	}

	return bm.SendBotAction(params.BotID, "set_group_whole_ban", struct {
		GroupID string `json:"group_id"`
		Enable  bool   `json:"enable"`
	}{
		GroupID: params.GroupID,
		Enable:  false,
	})
}

// toString 辅助函数
func toString(v any) string {
	return fmt.Sprintf("%v", v)
}
