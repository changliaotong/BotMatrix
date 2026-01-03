package tasks

import (
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

// BotManager 定义调度中心需要的机器人管理能力
type BotManager interface {
	SendBotAction(botID string, action string, params any) error
	SendToWorker(workerID string, msg types.WorkerCommand) error
	FindWorkerBySkill(skillName string) string // 返回 WorkerID
	GetTags(targetType string, targetID string) []string
	GetTargetsByTags(targetType string, tags []string, logic string) []string
	GetGroupMembers(botID string, groupID string) ([]types.MemberInfo, error)
}

func (d *Dispatcher) registerDefaultActions() {
	d.actions["send_message"] = d.handleSendMessage
	d.actions["mute_group"] = d.handleMuteGroup
	d.actions["unmute_group"] = d.handleUnmuteGroup
	d.actions["kick_member"] = d.handleKickMember
	d.actions["mute_random"] = d.handleMuteRandom
	d.actions["set_group_admin"] = d.handleSetGroupAdmin
	d.actions["skill_call"] = d.handleSkillCall
}

func (d *Dispatcher) handleKickMember(task Task, execution *Execution) error {
	var params struct {
		BotID            string `json:"bot_id"`
		GroupID          string `json:"group_id"`
		UserID           string `json:"user_id"`
		RejectAddRequest bool   `json:"reject_add_request"`
	}

	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	bm := d.manager.(BotManager)

	// 权限预检查
	members, err := bm.GetGroupMembers(params.BotID, params.GroupID)
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

	return bm.SendBotAction(params.BotID, "set_group_kick", params)
}

func (d *Dispatcher) handleSetGroupAdmin(task Task, execution *Execution) error {
	var params struct {
		BotID   string `json:"bot_id"`
		GroupID string `json:"group_id"`
		UserID  string `json:"user_id"`
		Enable  bool   `json:"enable"`
	}

	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	bm := d.manager.(BotManager)
	return bm.SendBotAction(params.BotID, "set_group_admin", params)
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

	cmd := types.WorkerCommand{
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
	// 使用 Redis Streams (XAdd) 代替 RPush，以匹配 Worker 的实现
	return d.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: queue,
		Values: map[string]interface{}{"payload": payload},
	}).Err()
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

func (d *Dispatcher) handleMuteRandom(task Task, execution *Execution) error {
	var params struct {
		BotID    string `json:"bot_id"`
		GroupID  string `json:"group_id"`
		Duration uint32 `json:"duration"`
		Count    int    `json:"count"`
		Smart    bool   `json:"smart"`
	}

	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	if params.Count <= 0 {
		params.Count = 1
	}

	bm := d.manager.(BotManager)
	members, err := bm.GetGroupMembers(params.BotID, params.GroupID)
	if err != nil {
		return fmt.Errorf("failed to get group members: %v", err)
	}

	if len(members) == 0 {
		return fmt.Errorf("no members found in group %s", params.GroupID)
	}

	// 权限检查：获取机器人自己在群里的角色
	botRole := "member"
	for _, m := range members {
		if m.UserID == params.BotID {
			botRole = m.Role
			break
		}
	}

	// 过滤掉不可禁言的目标：
	// 1. 如果机器人是管理员，只能禁言普通成员
	// 2. 如果机器人是群主，可以禁言除自己外的所有人
	// 3. 如果机器人是普通成员，谁也禁言不了
	var availableMembers []types.MemberInfo
	for _, m := range members {
		if m.UserID == params.BotID {
			continue // 不能禁言自己
		}

		canMute := false
		if botRole == "owner" {
			canMute = true // 群主无敌
		} else if botRole == "admin" {
			if m.Role == "member" {
				canMute = true // 管理员只能禁言普通成员
			}
		}

		if canMute {
			availableMembers = append(availableMembers, m)
		}
	}

	if len(availableMembers) == 0 {
		return fmt.Errorf("no muteable members found (bot role: %s)", botRole)
	}

	// 智能模式：从可禁言的目标中优先选择最近发言的人
	var targets []types.MemberInfo
	if params.Smart {
		activeThreshold := time.Now().Add(-30 * time.Minute)
		for _, m := range availableMembers {
			if m.LastSeen.After(activeThreshold) {
				targets = append(targets, m)
			}
		}
		// 如果活跃的人不够，则用所有可禁言成员兜底
		if len(targets) < params.Count {
			targets = availableMembers
		}
	} else {
		targets = availableMembers
	}

	// 随机打乱并挑选
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(targets), func(i, j int) {
		targets[i], targets[j] = targets[j], targets[i]
	})

	actualCount := params.Count
	if actualCount > len(targets) {
		actualCount = len(targets)
	}

	for i := 0; i < actualCount; i++ {
		member := targets[i]
		_ = bm.SendBotAction(params.BotID, "set_group_ban", struct {
			GroupID  string `json:"group_id"`
			UserID   string `json:"user_id"`
			Duration uint32 `json:"duration"`
		}{
			GroupID:  params.GroupID,
			UserID:   member.UserID,
			Duration: params.Duration,
		})

		// 发送中奖通知
		msg := "🎉 恭喜用户 %s (%s) 抽中随机禁言套餐，禁言 %d 秒！"
		if params.Smart {
			msg = "🔥 智能探测：捕捉到最近发言的活跃用户 %s (%s)，禁言套餐已送达，时长 %d 秒！"
		}
		reply := fmt.Sprintf(msg, member.Nickname, member.UserID, params.Duration)
		_ = bm.SendBotAction(params.BotID, "send_group_msg", map[string]any{
			"group_id": params.GroupID,
			"message":  reply,
		})
	}

	return nil
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
