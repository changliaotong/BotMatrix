package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// BotManager 定义调度中心需要的机器人管理能力
type BotManager interface {
	SendBotAction(botID string, action string, params map[string]interface{}) error
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
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(task.ActionParams), &params); err != nil {
		return fmt.Errorf("invalid action params: %v", err)
	}

	skillName, _ := params["skill"].(string)
	if skillName == "" {
		return fmt.Errorf("missing skill name")
	}

	// 查找目标 Worker ID
	workerID, _ := params["worker_id"].(string)
	
	// 如果没有指定 worker_id，可以广播或根据能力选择 (这里简化为广播到默认队列)
	queue := "botmatrix:queue:default"
	if workerID != "" {
		queue = fmt.Sprintf("botmatrix:queue:worker:%s", workerID)
	}

	msg := map[string]interface{}{
		"type":      "skill_call",
		"skill":     skillName,
		"params":    params,
		"task_id":   task.ID,
		"timestamp": time.Now().Unix(),
	}

	payload, _ := json.Marshal(msg)
	
	// 我们需要访问 Redis 客户端。Dispatcher 结构体中目前没有 Redis。
	// 但 Dispatcher 的 manager (即 main.Manager) 中有 Rdb。
	
	type RedisManager interface {
		GetRedis() interface{} // 返回 *redis.Client
	}
	
	// 我们在 manager.go 中手动定义一个接口或者直接反射
	// 为了简单，我们先假设 manager 提供了发送消息的方法
	
	return d.sendToQueue(queue, payload)
}

func (d *Dispatcher) sendToQueue(queue string, payload []byte) error {
	if d.rdb == nil {
		return fmt.Errorf("redis client not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return d.rdb.LPush(ctx, queue, payload).Err()
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
	actionParams := map[string]interface{}{
		"message": params.Message,
	}

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
		return bm.SendBotAction(params.BotID, "set_group_ban", map[string]interface{}{
			"group_id": params.GroupID,
			"user_id":  params.UserID,
			"duration": params.Duration,
		})
	}

	// 全员禁言
	return bm.SendBotAction(params.BotID, "set_group_whole_ban", map[string]interface{}{
		"group_id": params.GroupID,
		"enable":   true,
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
		return bm.SendBotAction(params.BotID, "set_group_ban", map[string]interface{}{
			"group_id": params.GroupID,
			"user_id":  params.UserID,
			"duration": 0,
		})
	}

	return bm.SendBotAction(params.BotID, "set_group_whole_ban", map[string]interface{}{
		"group_id": params.GroupID,
		"enable":   false,
	})
}

// toString 辅助函数
func toString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}
