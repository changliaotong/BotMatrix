package fission

import (
	"database/sql"
	"fmt"
	"botworker/internal/db"
)

// Service 裂变核心服务
type Service struct {
	db *sql.DB
}

func NewService(sqlDB *sql.DB) *Service {
	return &Service{db: sqlDB}
}

// ProcessBind 处理绑定逻辑 (包含防刷和奖励)
func (s *Service) ProcessBind(inviterID, inviteeID int64, platform, code, ip, deviceID string) (string, error) {
	// 1. 获取配置
	config, err := db.GetFissionConfig(s.db)
	if err != nil || !config.Enabled {
		return "裂变系统暂未开启", fmt.Errorf("fission disabled")
	}

	// 2. 基础校验：不能绑定自己
	if inviterID == inviteeID {
		return "不能绑定自己的邀请码", fmt.Errorf("cannot bind self")
	}

	// 3. 防刷：检查当日上限
	if config.MaxDailyInvites > 0 {
		count, _ := db.GetDailyInviteCount(s.db, inviterID)
		if count >= config.MaxDailyInvites {
			return "该邀请者今日名额已满", fmt.Errorf("daily limit reached")
		}
	}

	// 4. 防刷：IP/设备校验
	if config.AntiFraudEnabled {
		isFraud, reason, _ := db.CheckInvitationFraud(s.db, ip, deviceID)
		if isFraud {
			return fmt.Sprintf("异常操作：%s", reason), fmt.Errorf("anti-fraud: %s", reason)
		}
	}

	// 5. 执行绑定
	err = db.CreateInvitation(s.db, inviterID, inviteeID, code, ip, deviceID)
	if err != nil {
		return "绑定失败，您可能已经绑定过或邀请码无效", err
	}

	// 6. 发放基础奖励 (邀请者)
	if config.InviteRewardPoints > 0 {
		reason := fmt.Sprintf("成功邀请新用户: %d", inviteeID)
		_ = db.AddPoints(s.db, 0, inviterID, 0, int64(config.InviteRewardPoints), reason, "fission_invite")
		_ = db.CreateFissionRewardLog(s.db, inviterID, "points", config.InviteRewardPoints, reason)
	}

	// 7. 发放基础奖励 (被邀请者)
	if config.NewUserRewardPoints > 0 {
		reason := fmt.Sprintf("填写邀请码奖励: %s", code)
		_ = db.AddPoints(s.db, 0, inviteeID, 0, int64(config.NewUserRewardPoints), reason, "fission_bind")
		_ = db.CreateFissionRewardLog(s.db, inviteeID, "points", config.NewUserRewardPoints, reason)
	}

	// 8. 触发注册任务完成
	_ = db.CompleteFissionTask(s.db, inviteeID, "register")

	return "绑定成功！奖励已发放。", nil
}

// TriggerTask 触发任务进度
func (s *Service) TriggerTask(userID int64, taskType string) {
	_ = db.CompleteFissionTask(s.db, userID, taskType)
}

// GetUserStats 获取用户裂变数据
func (s *Service) GetUserStats(userID int64) (map[string]any, error) {
	record, err := db.GetUserFissionRecord(s.db, userID)
	if err != nil {
		return nil, err
	}
	
	stats := map[string]any{
		"invite_count": record.InviteCount,
		"points":       record.Points,
		"invite_code":  record.InviteCode,
		"level":        record.Level,
	}
	return stats, nil
}

// GetLeaderboard 获取排行榜
func (s *Service) GetLeaderboard(limit int) ([]map[string]any, error) {
	return db.GetFissionRank(s.db, limit)
}
