package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// MountPlugin 坐骑系统插件
type MountPlugin struct {
	cmdParser *CommandParser
	db        *gorm.DB
}

// Mount 坐骑结构体
type Mount struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Rarity      string    `json:"rarity"`
	Speed       int       `json:"speed"`
	Price       int       `json:"price"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserMount 用户坐骑结构体
type UserMount struct {
	UserID     string    `json:"user_id" gorm:"primaryKey"`
	MountID    string    `json:"mount_id" gorm:"primaryKey"`
	AcquiredAt time.Time `json:"acquired_at"`
	Level      int       `json:"level"`
	Experience int       `json:"experience"`
	IsActive   bool      `json:"is_active"`
}

func (p *MountPlugin) Name() string {
	return "mount"
}

func (p *MountPlugin) Description() string {
	return common.T("", "mount_plugin_desc|坐骑系统插件，管理用户坐骑")
}

func (p *MountPlugin) Version() string {
	return "1.0.0"
}

// GetSkills 报备插件技能
func (p *MountPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "mount_shop",
			Description: common.T("", "mount_skill_shop_desc|查看坐骑商店"),
			Usage:       "mount_shop",
			Params:      map[string]string{},
		},
		{
			Name:        "my_mounts",
			Description: common.T("", "mount_skill_my_mounts_desc|查看已拥有的坐骑"),
			Usage:       "my_mounts user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "mount_param_user_id_desc|用户QQ号"),
			},
		},
		{
			Name:        "equip_mount",
			Description: common.T("", "mount_skill_equip_desc|装备指定的坐骑"),
			Usage:       "equip_mount user_id=123456 mount_id=horse",
			Params: map[string]string{
				"user_id":  common.T("", "mount_param_user_id_desc|用户QQ号"),
				"mount_id": common.T("", "mount_param_mount_id_desc|坐骑ID"),
			},
		},
		{
			Name:        "upgrade_mount",
			Description: common.T("", "mount_skill_upgrade_desc|升级指定的坐骑"),
			Usage:       "upgrade_mount user_id=123456 mount_id=horse",
			Params: map[string]string{
				"user_id":  common.T("", "mount_param_user_id_desc|用户QQ号"),
				"mount_id": common.T("", "mount_param_mount_id_desc|坐骑ID"),
			},
		},
		{
			Name:        "mount_rank",
			Description: common.T("", "mount_skill_rank_desc|查看坐骑排行榜"),
			Usage:       "mount_rank",
			Params:      map[string]string{},
		},
	}
}

// NewMountPlugin 创建坐骑系统插件实例
func NewMountPlugin() *MountPlugin {
	return &MountPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *MountPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "mount_plugin_loaded_log|加载坐骑系统插件"))

	// 注册技能处理器
	robot.HandleSkill("mount_shop", func(params map[string]string) (string, error) {
		return p.doShowMountShop(), nil
	})

	robot.HandleSkill("my_mounts", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		if userID == "" {
			return "", fmt.Errorf(common.T("", "mount_missing_user_id|missing user_id"))
		}
		return p.doShowMyMounts(userID), nil
	})

	robot.HandleSkill("equip_mount", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		mountID := params["mount_id"]
		if userID == "" || mountID == "" {
			return "", fmt.Errorf(common.T("", "mount_missing_params|missing user_id or mount_id"))
		}
		return p.doEquipMount(userID, mountID), nil
	})

	robot.HandleSkill("upgrade_mount", func(params map[string]string) (string, error) {
		userID := params["user_id"]
		mountID := params["mount_id"]
		if userID == "" || mountID == "" {
			return "", fmt.Errorf(common.T("", "mount_missing_params|missing user_id or mount_id"))
		}
		return p.doUpgradeMount(userID, mountID), nil
	})

	robot.HandleSkill("mount_rank", func(params map[string]string) (string, error) {
		return p.doShowMountRank(), nil
	})

	// 处理坐骑系统命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为坐骑命令
		if match, _ := p.cmdParser.MatchCommand("坐骑|mount|ride", event.RawMessage); match {
			// 处理坐骑命令
			p.handleMountCommand(robot, event)
		}

		return nil
	})
}

// handleMountCommand 处理坐骑命令
func (p *MountPlugin) handleMountCommand(robot plugin.Robot, event *onebot.Event) {
	userIDStr := fmt.Sprintf("%d", event.UserID)

	// 检查命令参数
	args := p.cmdParser.ParseArgs(event.RawMessage)
	if len(args) == 1 {
		// 发送坐骑系统使用说明
		usage := common.T("", "mount_usage_title|🐎 坐骑系统命令使用说明:\n")
		usage += common.T("", "mount_separator|====================\n")
		usage += common.T("", "mount_usage_shop|/坐骑 商店 - 查看坐骑商店\n")
		usage += common.T("", "mount_usage_my|/坐骑 我的 - 查看我的坐骑\n")
		usage += common.T("", "mount_usage_equip|/坐骑 装备 <坐骑ID> - 装备坐骑\n")
		usage += common.T("", "mount_usage_upgrade|/坐骑 升级 <坐骑ID> - 升级坐骑\n")
		usage += common.T("", "mount_usage_rank|/坐骑 排行 - 查看坐骑排行榜\n")
		p.sendMessage(robot, event, usage)
		return
	}

	// 处理子命令
	subCmd := args[1]
	switch subCmd {
	case "商店", "shop":
		p.sendMessage(robot, event, p.doShowMountShop())
	case "我的", "my":
		p.sendMessage(robot, event, p.doShowMyMounts(userIDStr))
	case "装备", "equip":
		if len(args) >= 3 {
			p.sendMessage(robot, event, p.doEquipMount(userIDStr, args[2]))
		} else {
			p.sendMessage(robot, event, common.T("", "mount_specify_id|❌ 请指定坐骑ID"))
		}
	case "升级", "upgrade":
		if len(args) >= 3 {
			p.sendMessage(robot, event, p.doUpgradeMount(userIDStr, args[2]))
		} else {
			p.sendMessage(robot, event, common.T("", "mount_specify_id|❌ 请指定坐骑ID"))
		}
	case "排行", "rank":
		p.sendMessage(robot, event, p.doShowMountRank())
	default:
		p.sendMessage(robot, event, common.T("", "mount_unknown_subcmd|❌ 未知子命令，请使用/坐骑查看帮助"))
	}
}

// doShowMountShop 显示坐骑商店
func (p *MountPlugin) doShowMountShop() string {
	if p.db == nil {
		return common.T("", "mount_db_not_connected|❌ 数据库未连接")
	}
	var mounts []Mount
	if err := p.db.Find(&mounts).Error; err != nil {
		log.Printf(common.T("", "mount_query_shop_failed_log|[Mount] 查询坐骑商店失败: %v"), err)
		return common.T("", "mount_query_shop_failed|❌ 查询坐骑商店失败")
	}

	var msg string
	msg += common.T("", "mount_shop_title|🐎 坐骑商店:\n")
	msg += common.T("", "mount_separator|====================\n")
	msg += "\n"

	for _, mount := range mounts {
		msg += fmt.Sprintf("%s %s\n", mount.Icon, mount.Name)
		msg += fmt.Sprintf("📝 %s\n", mount.Description)
		msg += fmt.Sprintf(common.T("", "mount_rarity_prefix|⭐ 稀有度: %s\n"), mount.Rarity)
		msg += fmt.Sprintf(common.T("", "mount_speed_prefix|⚡ 速度: %d\n"), mount.Speed)
		msg += fmt.Sprintf(common.T("", "mount_price_format|💰 价格: %d 积分\n\n"), mount.Price)
	}

	if len(mounts) == 0 {
		msg += common.T("", "mount_no_mounts|暂无坐骑")
	}

	return msg
}

// doShowMyMounts 显示用户坐骑
func (p *MountPlugin) doShowMyMounts(userID string) string {
	if p.db == nil {
		return common.T("", "mount_db_not_connected|❌ 数据库未连接")
	}
	var userMounts []UserMount
	if err := p.db.Where("user_id = ?", userID).Find(&userMounts).Error; err != nil {
		log.Printf(common.T("", "mount_query_user_mounts_failed_log|[Mount] 查询用户坐骑失败: %v"), err)
		return common.T("", "mount_query_user_mounts_failed|❌ 查询用户坐骑失败")
	}

	var msg string
	msg += common.T("", "mount_my_mounts_title|🐎 我的坐骑:\n")
	msg += common.T("", "mount_separator|====================\n")
	msg += "\n"

	for _, userMount := range userMounts {
		var mount Mount
		if err := p.db.First(&mount, "id = ?", userMount.MountID).Error; err == nil {
			status := ""
			if userMount.IsActive {
				status = common.T("", "mount_status_equipped|(已装备)")
			}
			msg += fmt.Sprintf("%s %s %s\n", mount.Icon, mount.Name, status)
			msg += fmt.Sprintf(common.T("", "mount_level_prefix|📊 等级: %d\n"), userMount.Level)
			msg += fmt.Sprintf(common.T("", "mount_experience_prefix|💪 经验: %d/%d\n"), userMount.Experience, userMount.Level*1000)
			msg += fmt.Sprintf(common.T("", "mount_speed_prefix|⚡ 速度: %d\n\n"), mount.Speed+userMount.Level*10)
		}
	}

	if len(userMounts) == 0 {
		msg += common.T("", "mount_no_mounts_user|暂无坐骑，快去商店购买吧！")
	}

	return msg
}

// doEquipMount 装备坐骑
func (p *MountPlugin) doEquipMount(userID, mountID string) string {
	if p.db == nil {
		return common.T("", "mount_db_not_connected|❌ 数据库未连接")
	}
	// 检查用户是否拥有该坐骑
	var userMount UserMount
	if err := p.db.Where("user_id = ? AND mount_id = ?", userID, mountID).First(&userMount).Error; err != nil {
		return common.T("", "mount_not_owned|❌ 你没有该坐骑")
	}

	// 取消其他坐骑的装备状态
	if err := p.db.Model(&UserMount{}).Where("user_id = ? AND is_active = ?", userID, true).Update("is_active", false).Error; err != nil {
		log.Printf(common.T("", "mount_unequip_others_failed_log|[Mount] 取消其他坐骑装备失败: %v"), err)
		return common.T("", "mount_equip_failed|❌ 装备坐骑失败")
	}

	// 装备当前坐骑
	if err := p.db.Model(&userMount).Update("is_active", true).Error; err != nil {
		log.Printf(common.T("", "mount_equip_failed_log|[Mount] 装备坐骑失败: %v"), err)
		return common.T("", "mount_equip_failed|❌ 装备坐骑失败")
	}

	return fmt.Sprintf(common.T("", "mount_equip_success|✅ 成功装备坐骑: %s"), mountID)
}

// doUpgradeMount 升级坐骑
func (p *MountPlugin) doUpgradeMount(userID, mountID string) string {
	if p.db == nil {
		return "❌ 数据库未连接"
	}
	// 检查用户是否拥有该坐骑
	var userMount UserMount
	if err := p.db.Where("user_id = ? AND mount_id = ?", userID, mountID).First(&userMount).Error; err != nil {
		return "❌ 你没有该坐骑"
	}

	// 检查是否可以升级
	if userMount.Experience < userMount.Level*1000 {
		return "❌ 经验不足，无法升级"
	}

	// 升级坐骑
	userMount.Level++
	userMount.Experience = 0

	if err := p.db.Save(&userMount).Error; err != nil {
		log.Printf("[Mount] 升级坐骑失败: %v", err)
		return "❌ 升级坐骑失败"
	}

	return fmt.Sprintf("✅ 坐骑升级成功，当前等级: %d", userMount.Level)
}

// doShowMountRank 显示坐骑排行榜
func (p *MountPlugin) doShowMountRank() string {
	if p.db == nil {
		return "❌ 数据库未连接"
	}
	// 查询用户坐骑总价值排行榜
	var rankData []struct {
		UserID     string
		TotalValue int
	}

	query := `SELECT um.user_id, SUM(m.price) as total_value FROM user_mounts um JOIN mounts m ON um.mount_id = m.id GROUP BY um.user_id ORDER BY total_value DESC LIMIT 10`
	if err := p.db.Raw(query).Scan(&rankData).Error; err != nil {
		log.Printf("[Mount] 查询坐骑排行失败: %v", err)
		return "❌ 查询坐骑排行失败"
	}

	var msg string
	msg += "🏆 坐骑排行榜:\n"
	msg += "====================\n\n"

	for i, item := range rankData {
		msg += fmt.Sprintf("%d. 用户 %s: %d 积分价值\n", i+1, item.UserID, item.TotalValue)
	}

	if len(rankData) == 0 {
		msg += "暂无坐骑数据"
	}

	return msg
}

// sendMessage 发送消息
func (p *MountPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "mount_send_msg_failed_log|发送消息失败: %v\n"), err)
	}
}

// InitializeMounts 初始化坐骑数据
func (p *MountPlugin) InitializeMounts() error {
	mounts := []Mount{
		{
			ID:          "horse",
			Name:        common.T("", "mount_horse_name|普通战马"),
			Description: common.T("", "mount_horse_desc|普通的战马，适合长途旅行"),
			Icon:        "🐎",
			Rarity:      common.T("", "mount_rarity_common|普通"),
			Speed:       100,
			Price:       1000,
			Type:        common.T("", "mount_type_land|陆地"),
		},
		{
			ID:          "unicorn",
			Name:        common.T("", "mount_unicorn_name|独角兽"),
			Description: common.T("", "mount_unicorn_desc|神秘的独角兽，拥有魔法力量"),
			Icon:        "🦄",
			Rarity:      common.T("", "mount_rarity_rare|稀有"),
			Speed:       200,
			Price:       5000,
			Type:        common.T("", "mount_type_land|陆地"),
		},
		{
			ID:          "dragon",
			Name:        common.T("", "mount_dragon_name|火焰巨龙"),
			Description: common.T("", "mount_dragon_desc|强大的火焰巨龙，拥有毁灭力量"),
			Icon:        "🐉",
			Rarity:      common.T("", "mount_rarity_legendary|传说"),
			Speed:       500,
			Price:       20000,
			Type:        common.T("", "mount_type_flying|飞行"),
		},
		{
			ID:          "phoenix",
			Name:        common.T("", "mount_phoenix_name|不死凤凰"),
			Description: common.T("", "mount_phoenix_desc|浴火重生的凤凰，拥有永恒生命"),
			Icon:        "🔥",
			Rarity:      common.T("", "mount_rarity_mythic|神话"),
			Speed:       800,
			Price:       50000,
			Type:        common.T("", "mount_type_flying|飞行"),
		},
	}

	for _, mount := range mounts {
		if err := p.db.FirstOrCreate(&mount, "id = ?", mount.ID).Error; err != nil {
			return err
		}
	}

	return nil
}
