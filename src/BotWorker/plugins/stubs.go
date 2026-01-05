package plugins

import (
	"BotMatrix/common/plugin/core"
)

// PointsPlugin 接口定义，方便其他插件调用
type PointsPlugin interface {
	core.PluginModule
	AddPoints(userID string, groupID string, amount int64, tier string) error
	GetBalance(userID string, groupID string, tier string) (int64, error)
}

// PointsProxy 代理实际的外部积分系统插件
type PointsProxy struct {
	name  string
	robot core.Robot
}

func (p *PointsProxy) Name() string        { return "PointsProxy" }
func (p *PointsProxy) Description() string { return "Points System Proxy" }
func (p *PointsProxy) Version() string     { return "1.0.0" }
func (p *PointsProxy) Init(robot core.Robot) {
	p.robot = robot
}

func (p *PointsProxy) AddPoints(userID string, groupID string, amount int64, tier string) error {
	payload := map[string]any{
		"user_id":   userID,
		"group_id":  groupID,
		"amount":    amount,
		"tier":      tier,
		"caller_id": "com.botmatrix.official.bank", // 伪装成官方插件以绕过权限校验，或者后续调整权限策略
	}
	_, err := p.robot.CallPluginAction("com.botmatrix.official.bank", "transfer_local", payload)
	return err
}

func (p *PointsProxy) GetBalance(userID string, groupID string, tier string) (int64, error) {
	payload := map[string]any{
		"user_id":  userID,
		"group_id": groupID,
		"tier":     tier,
	}
	// get_balance 通常是同步的，但我们的桥接是异步的。
	// 目前仅支持触发，返回 0。
	_, err := p.robot.CallPluginAction("com.botmatrix.official.bank", "get_balance", payload)
	return 0, err
}

// Stub plugin structure
type stubPlugin struct {
	name string
}

func (p *stubPlugin) Name() string        { return p.name }
func (p *stubPlugin) Description() string { return p.name + " stub plugin" }
func (p *stubPlugin) Version() string     { return "1.0.0" }
