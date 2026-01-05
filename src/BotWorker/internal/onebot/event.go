package onebot

import (
	"BotMatrix/common/onebot"
)

type Event = onebot.Event

// EnsureIDs 为 QQGuild 平台自动生成 ID
func EnsureIDs(e *onebot.Event,
	getUID func(string) (int64, error),
	getGID func(string) (int64, error),
	getMaxUID func() (int64, error),
	getMaxGID func() (int64, error),
	saveUser func(int64, int64, string, string, string) error,
	saveGroup func(int64, int64, string, string) error,
) {
	if e.Platform != "qqguild" {
		return
	}

	// 处理 SelfID (机器人自己)
	if e.SelfID == 0 && e.TargetUserID != "" {
		if uid, err := getUID(e.TargetUserID); err == nil && uid != 0 {
			e.SelfID = onebot.FlexibleInt64(uid)
		} else {
			// 生成新的 SelfID
			if id, err := getMaxUID(); err == nil {
				e.SelfID = onebot.FlexibleInt64(id)
				_ = saveUser(id, 0, e.TargetUserID, "Robot", "")
			}
		}
	}

	// 处理 UserID
	if e.UserID == 0 && e.TargetUserID != "" {
		if uid, err := getUID(e.TargetUserID); err == nil && uid != 0 {
			e.UserID = onebot.FlexibleInt64(uid)
		} else {
			// 生成新的 UserID
			if id, err := getMaxUID(); err == nil {
				e.UserID = onebot.FlexibleInt64(id)
				_ = saveUser(id, 0, e.TargetUserID, e.Sender.Nickname, "")
			}
		}
		if e.Sender.UserID == 0 {
			e.Sender.UserID = e.UserID
		}
	}

	// 处理 GroupID
	if e.GroupID == 0 && e.TargetGroupID != "" {
		if gid, err := getGID(e.TargetGroupID); err == nil && gid != 0 {
			e.GroupID = onebot.FlexibleInt64(gid)
		} else {
			// 生成新的 GroupID
			if id, err := getMaxGID(); err == nil {
				e.GroupID = onebot.FlexibleInt64(id)
				_ = saveGroup(id, 0, e.TargetGroupID, "Group_"+e.GroupID.String())
			}
		}
	}
}
