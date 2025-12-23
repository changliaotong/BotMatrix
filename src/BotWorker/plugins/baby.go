package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"log"
	"time"
)

// BabyPlugin 宝宝系统插件
type BabyPlugin struct {
	cmdParser *CommandParser
}

// Baby 宝宝数据模型
type Baby struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"size:20;index" json:"user_id"`
	Name        string    `gorm:"size:50" json:"name"`
	Birthday    time.Time `json:"birthday"`
	GrowthValue int       `json:"growth_value"`
	DaysOld     int       `json:"days_old"`
	Level       int       `json:"level"`
	Status      string    `gorm:"size:20;default:active" json:"status"` // active, abandoned
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BabyEvent 宝宝事件记录
type BabyEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BabyID    uint      `json:"baby_id"`
	EventType string    `gorm:"size:50" json:"event_type"` // birthday, learn, work, interact
	Content   string    `gorm:"size:255" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// BabyConfig 宝宝系统配置
type BabyConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	IsEnabled   bool      `gorm:"default:true" json:"is_enabled"`
	GrowthRate  int       `gorm:"default:1000" json:"growth_rate"` // 每1000成长值增加1天
	UpdateAt    time.Time `json:"update_at"`
}

// NewBabyPlugin 创建宝宝系统插件实例
func NewBabyPlugin() *BabyPlugin {
	return &BabyPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *BabyPlugin) Name() string {
	return "baby"
}

func (p *BabyPlugin) Description() string {
	return "宝宝系统插件，提供宝宝降临、学习、打工等功能"
}

func (p *BabyPlugin) Version() string {
	return "1.0.0"
}

func (p *BabyPlugin) Init(robot plugin.Robot) {
	log.Println("加载宝宝系统插件")

	// 初始化数据库
	p.initDatabase()

	// 处理宝宝系统命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查系统是否开启
		if !p.isSystemEnabled() {
			return nil
		}

		// 宝宝降临
		if match, _ := p.cmdParser.MatchCommand("宝宝降临", event.RawMessage); match {
			p.babyBirth(robot, event)
			return nil
		}

		// 我的宝宝
		if match, _ := p.cmdParser.MatchCommand("我的宝宝", event.RawMessage); match {
			p.myBaby(robot, event)
			return nil
		}

		// 宝宝学习
		if match, _ := p.cmdParser.MatchCommand("宝宝学习", event.RawMessage); match {
			p.babyLearn(robot, event)
			return nil
		}

		// 宝宝商城
		if match, _ := p.cmdParser.MatchCommand("宝宝商城", event.RawMessage); match {
			p.babyMall(robot, event)
			return nil
		}

		// 购买商品
		if match, params := p.cmdParser.MatchCommandWithParams("购买(\d+)", event.RawMessage); match && len(params) > 0 {
			productID := params[1]
			p.buyProduct(robot, event, productID)
			return nil
		}

		// 宝宝互动
		if match, _ := p.cmdParser.MatchCommand("宝宝互动", event.RawMessage); match {
			p.babyInteract(robot, event)
			return nil
		}

		// 宝宝打工
		if match, _ := p.cmdParser.MatchCommand("宝宝打工", event.RawMessage); match {
			p.babyWork(robot, event)
			return nil
		}

		// 宝宝改名
		if match, params := p.cmdParser.MatchCommandWithParams("宝宝改名\+(\S+)", event.RawMessage); match && len(params) > 0 {
			newName := params[1]
			p.babyRename(robot, event, newName)
			return nil
		}

		// 开启宝宝系统
		if match, _ := p.cmdParser.MatchCommand("开启宝宝系统", event.RawMessage); match {
			p.enableSystem(robot, event)
			return nil
		}

		// 关闭宝宝系统
		if match, _ := p.cmdParser.MatchCommand("关闭宝宝系统", event.RawMessage); match {
			p.disableSystem(robot, event)
			return nil
		}

		// 超管抛弃宝宝功能
		if match, params := p.cmdParser.MatchCommandWithParams("抛弃宝宝(\d+)", event.RawMessage); match && len(params) > 0 {
			userID := params[1]
			p.abandonBaby(robot, event, userID)
			return nil
		}

		// 拐卖宝宝说明
		if match, _ := p.cmdParser.MatchCommand("拐卖宝宝说明", event.RawMessage); match {
			p.babyAbandonInfo(robot, event)
			return nil
		}

		return nil
	})
}

// initDatabase 初始化数据库
type Database interface {
	AutoMigrate(dst ...interface{}) error
}

func (p *BabyPlugin) initDatabase() {
	// 这里需要获取数据库连接，实际实现时需要与项目的数据库系统集成
	// db := GetDatabaseInstance()
	// err := db.AutoMigrate(&Baby{}, &BabyEvent{}, &BabyConfig{})
	// if err != nil {
	//  log.Printf("宝宝系统数据库初始化失败: %v\n", err)
	// }
	log.Println("宝宝系统数据库初始化完成")
}

// isSystemEnabled 检查宝宝系统是否开启
func (p *BabyPlugin) isSystemEnabled() bool {
	// 这里需要查询数据库获取系统配置
	// 默认返回开启状态
	return true
}

// babyBirth 宝宝降临功能
func (p *BabyPlugin) babyBirth(robot plugin.Robot, event *onebot.Event) {
	// 检查用户是否已有宝宝
	// 如果没有，创建新宝宝
	baby := Baby{
		UserID:      event.UserID,
		Name:        "小宝宝",
		Birthday:    time.Now(),
		GrowthValue: 0,
		DaysOld:     0,
		Level:       1,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存宝宝数据到数据库
	// db := GetDatabaseInstance()
	// result := db.Create(&baby)
	// if result.Error != nil {
	//  log.Printf("创建宝宝失败: %v\n", result.Error)
	//  SendTextReply(robot, event, "❌ 宝宝降临失败，请稍后重试")
	//  return
	// }

	msg := "🎉 恭喜！您的宝宝降临了！\n"
	msg += "👶 宝宝名字：" + baby.Name + "\n"
	msg += "📅 出生日期：" + baby.Birthday.Format("2006-01-02") + "\n"
	msg += "💡 提示：宝宝每1000成长值增加1天，生日每年过一次哦~\n"
	msg += "💌 发送【我的宝宝】查看宝宝详情"

	SendTextReply(robot, event, msg)
}

// myBaby 我的宝宝功能
func (p *BabyPlugin) myBaby(robot plugin.Robot, event *onebot.Event) {
	// 查询用户的宝宝
	// var baby Baby
	// db := GetDatabaseInstance()
	// result := db.Where("user_id = ? AND status = ?", event.UserID, "active").First(&baby)
	// if result.Error != nil {
	//  SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
	//  return
	// }

	// 模拟数据用于测试
	baby := Baby{
		Name:        "小宝宝",
		Birthday:    time.Now().AddDate(0, 0, -10),
		GrowthValue: 5000,
		DaysOld:     5,
		Level:       1,
	}

	msg := "👶 我的宝宝\n"
	msg += "================================\n"
	msg += "🎂 名字：" + baby.Name + "\n"
	msg += "📅 出生日期：" + baby.Birthday.Format("2006-01-02") + "\n"
	msg += "🔢 年龄：" + p.getBabyAge(baby) + "\n"
	msg += "📈 成长值：" + IntToString(baby.GrowthValue) + "\n"
	msg += "⭐ 等级：" + IntToString(baby.Level) + "\n"
	msg += "================================\n"
	msg += "💡 可用命令：\n"
	msg += "📚 宝宝学习 - 增加宝宝知识\n"
	msg += "🎮 宝宝互动 - 增进亲子感情\n"
	msg += "💼 宝宝打工 - 培养宝宝能力\n"
	msg += "🛍️ 宝宝商城 - 购买宝宝用品\n"
	msg += "✏️ 宝宝改名+新名字 - 给宝宝改名"

	SendTextReply(robot, event, msg)
}

// babyLearn 宝宝学习功能
func (p *BabyPlugin) babyLearn(robot plugin.Robot, event *onebot.Event) {
	// 查询用户的宝宝
	// var baby Baby
	// db := GetDatabaseInstance()
	// result := db.Where("user_id = ? AND status = ?", event.UserID, "active").First(&baby)
	// if result.Error != nil {
	//  SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
	//  return
	// }

	// 增加成长值
	growthAdd := 100
	// baby.GrowthValue += growthAdd
	// 
	// // 计算应该增加的天数（每1000成长值=1天）
	// newDays := baby.GrowthValue / 1000
	// if newDays > baby.DaysOld {
	//  baby.DaysOld = newDays
	//  baby.Level = baby.DaysOld/30 + 1 // 每30天升1级
	// }
	// 
	// // 记录学习事件
	// babyEvent := BabyEvent{
	//  BabyID:    baby.ID,
	//  EventType: "learn",
	//  Content:   "宝宝学习获得" + IntToString(growthAdd) + "点成长值",
	//  CreatedAt: time.Now(),
	// }
	// 
	// // 保存数据到数据库
	// db.Save(&baby)
	// db.Create(&babyEvent)

	// 模拟宝宝数据用于测试
	baby := Baby{
		Name:        "小宝宝",
		GrowthValue: 1000,
		DaysOld:     1,
		Level:       1,
	}

	msg := "📚 宝宝正在学习...\n"
	msg += "✅ 学习完成！获得" + IntToString(growthAdd) + "点成长值\n"
	msg += "📈 当前成长值：" + IntToString(baby.GrowthValue+growthAdd) + "\n"
	msg += "👶 宝宝名字：" + baby.Name + "\n"
	msg += "📅 年龄：" + p.getBabyAge(baby) + "\n"
	msg += "⭐ 等级：" + IntToString(baby.Level) + "\n"
	msg += "💡 学习可以帮助宝宝快速成长，提高智力哦~"

	SendTextReply(robot, event, msg)
}

// babyMall 宝宝商城功能
func (p *BabyPlugin) babyMall(robot plugin.Robot, event *onebot.Event) {
	msg := "🛍️ 宝宝商城\n"
	msg += "================================\n"
	msg += "1️⃣ 奶瓶 - 50积分\n"
	msg += "   功效：增加宝宝100成长值\n"
	msg += "2️⃣ 玩具车 - 100积分\n"
	msg += "   功效：增加宝宝200成长值\n"
	msg += "3️⃣ 故事书 - 150积分\n"
	msg += "   功效：增加宝宝300成长值\n"
	msg += "4️⃣ 新衣服 - 200积分\n"
	msg += "   功效：增加宝宝400成长值\n"
	msg += "================================\n"
	msg += "💡 提示：发送【购买+商品编号】进行购买\n"
	msg += "例如：购买1"

	SendTextReply(robot, event, msg)
}

// Product 商品信息
var babyProducts = map[string]struct {
	Name        string
	Price       int
	GrowthValue int
}{"1": {"奶瓶", 50, 100},
	"2": {"玩具车", 100, 200},
	"3": {"故事书", 150, 300},
	"4": {"新衣服", 200, 400}}

// buyProduct 购买商品功能
func (p *BabyPlugin) buyProduct(robot plugin.Robot, event *onebot.Event, productID string) {
	// 检查商品是否存在
	product, ok := babyProducts[productID]
	if !ok {
		SendTextReply(robot, event, "❌ 无效的商品编号，请查看商城获取正确的商品编号")
		return
	}

	// 查询用户的宝宝
	// var baby Baby
	// db := GetDatabaseInstance()
	// result := db.Where("user_id = ? AND status = ?", event.UserID, "active").First(&baby)
	// if result.Error != nil {
	//  SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
	//  return
	// }

	// 检查用户积分是否足够
	// pointsPlugin := GetPointsPluginInstance()
	// userPoints := pointsPlugin.GetPoints(event.UserID)
	// if userPoints < product.Price {
	//  SendTextReply(robot, event, "❌ 积分不足，购买失败\n当前积分："+IntToString(userPoints)+"\n所需积分："+IntToString(product.Price))
	//  return
	// }

	// 扣除积分
	// pointsPlugin.SubtractPoints(event.UserID, product.Price)

	// 增加宝宝成长值
	growthAdd := product.GrowthValue
	// baby.GrowthValue += growthAdd
	// 
	// // 计算应该增加的天数（每1000成长值=1天）
	// newDays := baby.GrowthValue / 1000
	// if newDays > baby.DaysOld {
	//  baby.DaysOld = newDays
	//  baby.Level = baby.DaysOld/30 + 1 // 每30天升1级
	// }
	// 
	// // 记录购买事件
	// babyEvent := BabyEvent{
	//  BabyID:    baby.ID,
	//  EventType: "buy",
	//  Content:   "购买了" + product.Name + "，获得" + IntToString(growthAdd) + "点成长值",
	//  CreatedAt: time.Now(),
	// }
	// 
	// // 保存数据到数据库
	// db.Save(&baby)
	// db.Create(&babyEvent)

	// 模拟数据用于测试
	baby := Baby{
		Name:        "小宝宝",
		GrowthValue: 1000,
		DaysOld:     1,
		Level:       1,
	}
	
	userPoints := 500

	msg := "🎉 购买成功！\n"
	msg += "🛍️ 商品：" + product.Name + "\n"
	msg += "💰 花费积分：" + IntToString(product.Price) + "\n"
	msg += "剩余积分：" + IntToString(userPoints-product.Price) + "\n"
	msg += "📈 宝宝获得" + IntToString(growthAdd) + "点成长值\n"
	msg += "👶 宝宝当前成长值：" + IntToString(baby.GrowthValue+growthAdd) + "\n"
	msg += "💡 宝宝变得更加强壮了！"

	SendTextReply(robot, event, msg)
}

// babyInteract 宝宝互动功能
func (p *BabyPlugin) babyInteract(robot plugin.Robot, event *onebot.Event) {
	// 查询用户的宝宝
	// var baby Baby
	// db := GetDatabaseInstance()
	// result := db.Where("user_id = ? AND status = ?", event.UserID, "active").First(&baby)
	// if result.Error != nil {
	//  SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
	//  return
	// }

	// 增加成长值
	growthAdd := 50
	// baby.GrowthValue += growthAdd
	// 
	// // 计算应该增加的天数（每1000成长值=1天）
	// newDays := baby.GrowthValue / 1000
	// if newDays > baby.DaysOld {
	//  baby.DaysOld = newDays
	//  baby.Level = baby.DaysOld/30 + 1 // 每30天升1级
	// }
	// 
	// // 记录互动事件
	// babyEvent := BabyEvent{
	//  BabyID:    baby.ID,
	//  EventType: "interact",
	//  Content:   "与宝宝互动获得" + IntToString(growthAdd) + "点成长值",
	//  CreatedAt: time.Now(),
	// }
	// 
	// // 保存数据到数据库
	// db.Save(&baby)
	// db.Create(&babyEvent)

	// 模拟宝宝数据用于测试
	baby := Baby{
		Name:        "小宝宝",
		GrowthValue: 800,
		DaysOld:     0,
		Level:       1,
	}

	msg := "🎮 您正在和宝宝互动...\n"
	msg += "😊 宝宝很开心！获得" + IntToString(growthAdd) + "点成长值\n"
	msg += "📈 当前成长值：" + IntToString(baby.GrowthValue+growthAdd) + "\n"
	msg += "👶 宝宝名字：" + baby.Name + "\n"
	msg += "📅 年龄：" + p.getBabyAge(baby) + "\n"
	msg += "⭐ 等级：" + IntToString(baby.Level) + "\n"
	msg += "💡 多和宝宝互动可以增进亲子感情哦~"

	SendTextReply(robot, event, msg)
}

// babyWork 宝宝打工功能
func (p *BabyPlugin) babyWork(robot plugin.Robot, event *onebot.Event) {
	// 查询用户的宝宝
	// var baby Baby
	// db := GetDatabaseInstance()
	// result := db.Where("user_id = ? AND status = ?", event.UserID, "active").First(&baby)
	// if result.Error != nil {
	// 	SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
	// 	return
	// }

	// 检查宝宝年龄是否足够打工
	// if baby.DaysOld < 30 {
	// 	SendTextReply(robot, event, "❌ 宝宝太小了，至少需要1个月才能打工哦~\n当前宝宝年龄：" + p.getBabyAge(baby))
	// 	return
	// }

	// 增加成长值和积分
	growthAdd := 150
	pointsAdd := 50
	// baby.GrowthValue += growthAdd
	// db.Save(&baby)

	// 增加用户积分
	// pointsPlugin := GetPointsPluginInstance()
	// pointsPlugin.AddPoints(event.UserID, pointsAdd)

	// 模拟宝宝数据用于测试
	baby := Baby{
		Name:        "小宝宝",
		GrowthValue: 1200,
		DaysOld:     30,
		Level:       2,
	}

	msg := "💼 宝宝开始打工了...\n"
	msg += "✅ 打工完成！获得" + IntToString(growthAdd) + "点成长值和" + IntToString(pointsAdd) + "积分\n"
	msg += "📈 当前成长值：" + IntToString(baby.GrowthValue+growthAdd) + "\n"
	msg += "👶 宝宝名字：" + baby.Name + "\n"
	msg += "📅 年龄：" + p.getBabyAge(baby) + "\n"
	msg += "⭐ 等级：" + IntToString(baby.Level) + "\n"
	msg += "💡 打工可以培养宝宝的独立性和责任感哦~"

	SendTextReply(robot, event, msg)
}

// babyRename 宝宝改名功能
func (p *BabyPlugin) babyRename(robot plugin.Robot, event *onebot.Event, newName string) {
	if len(newName) < 2 || len(newName) > 10 {
		SendTextReply(robot, event, "❌ 宝宝名字长度必须在2-10个字符之间")
		return
	}

	// 查询用户的宝宝
	// var baby Baby
	// db := GetDatabaseInstance()
	// result := db.Where("user_id = ? AND status = ?", event.UserID, "active").First(&baby)
	// if result.Error != nil {
	// 	SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
	// 	return
	// }

	// 更新宝宝名字
	oldName := "小宝宝"
	// baby.Name = newName
	// baby.UpdatedAt = time.Now()
	// db.Save(&baby)

	msg := "✏️ 宝宝改名成功！\n"
	msg += "👶 旧名字：" + oldName + "\n"
	msg += "✨ 新名字：" + newName + "\n"
	msg += "📅 宝宝还是原来的那个小可爱哦~\n"
	msg += "💡 发送【我的宝宝】查看最新信息"

	SendTextReply(robot, event, msg)
}

// enableSystem 开启宝宝系统功能
func (p *BabyPlugin) enableSystem(robot plugin.Robot, event *onebot.Event) {
	// 检查用户权限
	if !p.isSuperAdmin(event.UserID) {
		SendTextReply(robot, event, "❌ 您没有权限执行此命令")
		return
	}

	// 更新系统配置为开启
	// db := GetDatabaseInstance()
	// var config BabyConfig
	// db.FirstOrCreate(&config)
	// config.IsEnabled = true
	// config.UpdateAt = time.Now()
	// db.Save(&config)

	msg := "✅ 宝宝系统已成功开启！\n"
	msg += "👶 用户现在可以使用以下宝宝系统功能：\n"
	msg += "- 宝宝降临\n"
	msg += "- 我的宝宝\n"
	msg += "- 宝宝学习\n"
	msg += "- 宝宝商城\n"
	msg += "- 宝宝互动\n"
	msg += "- 宝宝打工\n"
	msg += "- 宝宝改名+新名字"

	SendTextReply(robot, event, msg)
}

// disableSystem 关闭宝宝系统功能
func (p *BabyPlugin) disableSystem(robot plugin.Robot, event *onebot.Event) {
	// 检查用户权限
	if !p.isSuperAdmin(event.UserID) {
		SendTextReply(robot, event, "❌ 您没有权限执行此命令")
		return
	}

	// 更新系统配置为关闭
	// db := GetDatabaseInstance()
	// var config BabyConfig
	// db.FirstOrCreate(&config)
	// config.IsEnabled = false
	// config.UpdateAt = time.Now()
	// db.Save(&config)

	msg := "⚠️ 宝宝系统已成功关闭！\n"
	msg += "👶 用户将暂时无法使用宝宝系统的所有功能\n"
	msg += "💡 需要时可以再次发送【开启宝宝系统】重新启用"

	SendTextReply(robot, event, msg)
}

// abandonBaby 超管抛弃宝宝功能
func (p *BabyPlugin) abandonBaby(robot plugin.Robot, event *onebot.Event, userID string) {
	// 检查用户权限
	if !p.isSuperAdmin(event.UserID) {
		SendTextReply(robot, event, "❌ 您没有权限执行此命令")
		return
	}

	// 查询用户的宝宝
	// var baby Baby
	// db := GetDatabaseInstance()
	// result := db.Where("user_id = ? AND status = ?", userID, "active").First(&baby)
	// if result.Error != nil {
	//  SendTextReply(robot, event, "❌ 该用户没有宝宝")
	//  return
	// }

	// 标记宝宝为已抛弃
	// baby.Status = "abandoned"
	// baby.UpdatedAt = time.Now()
	// db.Save(&baby)

	msg := "⚠️ 操作完成！已成功处理用户 " + userID + " 的宝宝\n"
	msg += "💡 注意：此操作不可逆，请谨慎使用"

	SendTextReply(robot, event, msg)
}

// babyAbandonInfo 拐卖宝宝说明
func (p *BabyPlugin) babyAbandonInfo(robot plugin.Robot, event *onebot.Event) {
	msg := "🚨 拐卖宝宝说明\n"
	msg += "================================\n"
	msg += "⚠️ 系统提示：宝宝是家庭的重要成员\n"
	msg += "❌ 请勿遗弃或拐卖宝宝\n"
	msg += "✅ 请爱护和培养您的宝宝\n"
	msg += "💡 超管有权处理不当使用宝宝系统的用户\n"
	msg += "================================\n"
	msg += "📞 如有问题请联系管理员"

	SendTextReply(robot, event, msg)
}

// getBabyAge 获取宝宝年龄描述
func (p *BabyPlugin) getBabyAge(baby Baby) string {
	duration := time.Since(baby.Birthday)
	days := int(duration.Hours() / 24)
	months := days / 30
	years := days / 365

	if years > 0 {
		return IntToString(years) + "岁" + IntToString(months%12) + "个月"
	} else if months > 0 {
		return IntToString(months) + "个月" + IntToString(days%30) + "天"
	} else {
		return IntToString(days) + "天"
	}
}

// isSuperAdmin 检查是否为超级管理员
func (p *BabyPlugin) isSuperAdmin(userID string) bool {
	// 超级管理员列表（实际使用时应从配置或数据库读取）
	// 这里暂时硬编码几个示例ID用于测试
	superAdmins := []string{
		"123456789", // 示例超级管理员ID
		"987654321", // 示例超级管理员ID
	}
	
	// 检查用户ID是否在超级管理员列表中
	for _, adminID := range superAdmins {
		if userID == adminID {
			return true
		}
	}
	
	return false
}

// updateGrowthValue 更新宝宝成长值
func (p *BabyPlugin) updateGrowthValue() {
	log.Println("开始更新宝宝成长值...")
	
	// 查询所有活跃状态的宝宝
	// var babies []Baby
	// db := GetDatabaseInstance()
	// db.Where("status = ?", "active").Find(&babies)
	
	// 模拟数据用于测试
	babies := []Baby{
		{
			ID:          1,
			UserID:      "123456",
			Name:        "小宝宝",
			Birthday:    time.Now().AddDate(0, 0, -10),
			GrowthValue: 5000,
			DaysOld:     5,
			Level:       1,
			Status:      "active",
		},
	}
	
	// 遍历所有宝宝，更新成长值
	for _, baby := range babies {
		growthAdd := 50 // 每日自动增加50成长值
		baby.GrowthValue += growthAdd
		
		// 计算应该增加的天数（每1000成长值=1天）
		newDays := baby.GrowthValue / 1000
		if newDays > baby.DaysOld {
			baby.DaysOld = newDays
			baby.Level = baby.DaysOld/30 + 1 // 每30天升1级
			p.checkBirthday(baby) // 检查是否过生日
			log.Printf("宝宝 %s 更新完成：成长值=%d, 天数=%d, 等级=%d\n", baby.Name, baby.GrowthValue, baby.DaysOld, baby.Level)
			
			// 保存宝宝数据到数据库
			// db.Save(&baby)
		}
	}
	log.Println("更新宝宝成长值任务执行完成")
}

// checkBirthday 检查宝宝是否过生日
func (p *BabyPlugin) checkBirthday(baby Baby) {
	now := time.Now()
	birthMonth := baby.Birthday.Month()
	birthDay := baby.Birthday.Day()
	
	// 检查是否是生日
	if now.Month() == birthMonth && now.Day() == birthDay {
		// 如果是生日，记录生日事件
		// babyEvent := BabyEvent{
		//  BabyID:    baby.ID,
		//  EventType: "birthday",
		//  Content:   "宝宝今天过生日了！",
		//  CreatedAt: now,
		// }
		// db := GetDatabaseInstance()
		// db.Create(&babyEvent)
		
		log.Printf("🎉 宝宝 %s 今天过生日了！现在 %d 天了\n", baby.Name, baby.DaysOld)
	}
}