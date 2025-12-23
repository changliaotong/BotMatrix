package plugins

import (
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
	return "坐骑系统插件，管理用户坐骑"
}

func (p *MountPlugin) Version() string {
	return "1.0.0"
}

// NewMountPlugin 创建坐骑系统插件实例
func NewMountPlugin() *MountPlugin {
	return &MountPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *MountPlugin) Init(robot plugin.Robot) {
	log.Println("加载坐骑系统插件")

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
		usage := "🐎 坐骑系统命令使用说明:\n"
		usage += "====================\n"
		usage += "/坐骑 商店 - 查看坐骑商店\n"
		usage += "/坐骑 我的 - 查看我的坐骑\n"
		usage += "/坐骑 装备 <坐骑ID> - 装备坐骑\n"
		usage += "/坐骑 升级 <坐骑ID> - 升级坐骑\n"
		usage += "/坐骑 排行 - 查看坐骑排行榜\n"
		p.sendMessage(robot, event, usage)
		return
	}

	// 处理子命令
	subCmd := args[1]
	switch subCmd {
	case "商店", "shop":
		p.showMountShop(robot, event)
	case "我的", "my":
		p.showMyMounts(robot, event, userIDStr)
	case "装备", "equip":
		if len(args) >= 3 {
			p.equipMount(robot, event, userIDStr, args[2])
		} else {
			p.sendMessage(robot, event, "❌ 请指定坐骑ID")
		}
	case "升级", "upgrade":
		if len(args) >= 3 {
			p.upgradeMount(robot, event, userIDStr, args[2])
		} else {
			p.sendMessage(robot, event, "❌ 请指定坐骑ID")
		}
	case "排行", "rank":
		p.showMountRank(robot, event)
	default:
		p.sendMessage(robot, event, "❌ 未知子命令，请使用/坐骑查看帮助")
	}
}

// showMountShop 显示坐骑商店
func (p *MountPlugin) showMountShop(robot plugin.Robot, event *onebot.Event) {
	var mounts []Mount
	if err := p.db.Find(&mounts).Error; err != nil {
		log.Printf("[Mount] 查询坐骑商店失败: %v", err)
		p.sendMessage(robot, event, "❌ 查询坐骑商店失败")
		return
	}

	var msg string
	msg += "🐎 坐骑商店:\n"
	msg += "====================\n\n"

	for _, mount := range mounts {
		msg += fmt.Sprintf("%s %s\n", mount.Icon, mount.Name)
		msg += fmt.Sprintf("📝 %s\n", mount.Description)
		msg += fmt.Sprintf("⭐ 稀有度: %s\n", mount.Rarity)
		msg += fmt.Sprintf("⚡ 速度: %d\n", mount.Speed)
		msg += fmt.Sprintf("💰 价格: %d 积分\n\n", mount.Price)
	}

	if len(mounts) == 0 {
		msg += "暂无坐骑"
	}

	p.sendMessage(robot, event, msg)
}

// showMyMounts 显示用户坐骑
func (p *MountPlugin) showMyMounts(robot plugin.Robot, event *onebot.Event, userID string) {
	var userMounts []UserMount
	if err := p.db.Where("user_id = ?", userID).Find(&userMounts).Error; err != nil {
		log.Printf("[Mount] 查询用户坐骑失败: %v", err)
		p.sendMessage(robot, event, "❌ 查询用户坐骑失败")
		return
	}

	var msg string
	msg += "🐎 我的坐骑:\n"
	msg += "====================\n\n"

	for _, userMount := range userMounts {
		var mount Mount
		if err := p.db.First(&mount, "id = ?", userMount.MountID).Error; err == nil {
			status := ""
			if userMount.IsActive {
				status = "(已装备)"
			}
			msg += fmt.Sprintf("%s %s %s\n", mount.Icon, mount.Name, status)
			msg += fmt.Sprintf("📊 等级: %d\n", userMount.Level)
			msg += fmt.Sprintf("💪 经验: %d/%d\n", userMount.Experience, userMount.Level*1000)
			msg += fmt.Sprintf("⚡ 速度: %d\n\n", mount.Speed+userMount.Level*10)
		}
	}

	if len(userMounts) == 0 {
		msg += "暂无坐骑，快去商店购买吧！"
	}

	p.sendMessage(robot, event, msg)
}

// equipMount 装备坐骑
func (p *MountPlugin) equipMount(robot plugin.Robot, event *onebot.Event, userID, mountID string) {
	// 检查用户是否拥有该坐骑
	var userMount UserMount
	if err := p.db.Where("user_id = ? AND mount_id = ?", userID, mountID).First(&userMount).Error; err != nil {
		p.sendMessage(robot, event, "❌ 你没有该坐骑")
		return
	}

	// 取消其他坐骑的装备状态
	if err := p.db.Model(&UserMount{}).Where("user_id = ? AND is_active = ?", userID, true).Update("is_active", false).Error; err != nil {
		log.Printf("[Mount] 取消其他坐骑装备失败: %v", err)
		p.sendMessage(robot, event, "❌ 装备坐骑失败")
		return
	}

	// 装备当前坐骑
	if err := p.db.Model(&userMount).Update("is_active", true).Error; err != nil {
		log.Printf("[Mount] 装备坐骑失败: %v", err)
		p.sendMessage(robot, event, "❌ 装备坐骑失败")
		return
	}

	p.sendMessage(robot, event, fmt.Sprintf("✅ 成功装备坐骑: %s", userMount.MountID))
}

// upgradeMount 升级坐骑
func (p *MountPlugin) upgradeMount(robot plugin.Robot, event *onebot.Event, userID, mountID string) {
	// 检查用户是否拥有该坐骑
	var userMount UserMount
	if err := p.db.Where("user_id = ? AND mount_id = ?", userID, mountID).First(&userMount).Error; err != nil {
		p.sendMessage(robot, event, "❌ 你没有该坐骑")
		return
	}

	// 检查是否可以升级
	if userMount.Experience < userMount.Level*1000 {
		p.sendMessage(robot, event, "❌ 经验不足，无法升级")
		return
	}

	// 升级坐骑
	userMount.Level++
	userMount.Experience = 0

	if err := p.db.Save(&userMount).Error; err != nil {
		log.Printf("[Mount] 升级坐骑失败: %v", err)
		p.sendMessage(robot, event, "❌ 升级坐骑失败")
		return
	}

	p.sendMessage(robot, event, fmt.Sprintf("✅ 坐骑升级成功，当前等级: %d", userMount.Level))
}

// showMountRank 显示坐骑排行榜
func (p *MountPlugin) showMountRank(robot plugin.Robot, event *onebot.Event) {
	// 查询用户坐骑总价值排行榜
	var rankData []struct {
		UserID     string
		TotalValue int
	}

	query := `SELECT um.user_id, SUM(m.price) as total_value FROM user_mounts um JOIN mounts m ON um.mount_id = m.id GROUP BY um.user_id ORDER BY total_value DESC LIMIT 10`
	if err := p.db.Raw(query).Scan(&rankData).Error; err != nil {
		log.Printf("[Mount] 查询坐骑排行失败: %v", err)
		p.sendMessage(robot, event, "❌ 查询坐骑排行失败")
		return
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

	p.sendMessage(robot, event, msg)
}

// sendMessage 发送消息
func (p *MountPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

// InitializeMounts 初始化坐骑数据
func (p *MountPlugin) InitializeMounts() error {
	mounts := []Mount{
		{
			ID:          "horse",
			Name:        "普通战马",
			Description: "普通的战马，适合长途旅行",
			Icon:        "🐎",
			Rarity:      "普通",
			Speed:       100,
			Price:       1000,
			Type:        "陆地",
		},
		{
			ID:          "unicorn",
			Name:        "独角兽",
			Description: "神秘的独角兽，拥有魔法力量",
			Icon:        "🦄",
			Rarity:      "稀有",
			Speed:       200,
			Price:       5000,
			Type:        "陆地",
		},
		{
			ID:          "dragon",
			Name:        "火焰巨龙",
			Description: "强大的火焰巨龙，拥有毁灭力量",
			Icon:        "🐉",
			Rarity:      "传说",
			Speed:       500,
			Price:       20000,
			Type:        "飞行",
		},
		{
			ID:          "phoenix",
			Name:        "不死凤凰",
			Description: "浴火重生的凤凰，拥有永恒生命",
			Icon:        "🔥",
			Rarity:      "神话",
			Speed:       800,
			Price:       50000,
			Type:        "飞行",
		},
	}

	for _, mount := range mounts {
		if err := p.db.FirstOrCreate(&mount, "id = ?", mount.ID).Error; err != nil {
			return err
		}
	}

	return nil
}
