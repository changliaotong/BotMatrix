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
func (p *BabyPlugin) initDatabase() {
	if GlobalDB == nil {
		log.Println("警告: 数据库未初始化，宝宝系统将使用模拟数据")
		return
	}
	
	// 创建宝宝表
	createBabyTable := `
	CREATE TABLE IF NOT EXISTS baby (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		name VARCHAR(50) NOT NULL,
		birthday TIMESTAMP NOT NULL,
		growth_value INT NOT NULL DEFAULT 0,
		days_old INT NOT NULL DEFAULT 0,
		level INT NOT NULL DEFAULT 1,
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := GlobalDB.Exec(createBabyTable)
	if err != nil {
		log.Printf("创建宝宝表失败: %v\n", err)
		return
	}
	
	// 创建宝宝事件表
	createBabyEventTable := `
	CREATE TABLE IF NOT EXISTS baby_event (
		id SERIAL PRIMARY KEY,
		baby_id INT NOT NULL REFERENCES baby(id) ON DELETE CASCADE,
		event_type VARCHAR(50) NOT NULL,
		content VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createBabyEventTable)
	if err != nil {
		log.Printf("创建宝宝事件表失败: %v\n", err)
		return
	}
	
	// 创建宝宝系统配置表
	createBabyConfigTable := `
	CREATE TABLE IF NOT EXISTS baby_config (
		id SERIAL PRIMARY KEY,
		is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
		growth_rate INT NOT NULL DEFAULT 1000,
		update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = GlobalDB.Exec(createBabyConfigTable)
	if err != nil {
		log.Printf("创建宝宝系统配置表失败: %v\n", err)
		return
	}
	
	// 初始化配置
	var count int
	err = GlobalDB.QueryRow("SELECT COUNT(*) FROM baby_config").Scan(&count)
	if err != nil {
		log.Printf("查询宝宝系统配置失败: %v\n", err)
		return
	}
	
	if count == 0 {
		_, err = GlobalDB.Exec("INSERT INTO baby_config (is_enabled, growth_rate) VALUES (TRUE, 1000)")
		if err != nil {
			log.Printf("初始化宝宝系统配置失败: %v\n", err)
			return
		}
	}
	
	log.Println("宝宝系统数据库初始化完成")
}

// isSystemEnabled 检查宝宝系统是否开启
func (p *BabyPlugin) isSystemEnabled() bool {
	if GlobalDB == nil {
		// 如果没有数据库连接，默认返回开启状态
		return true
	}
	
	// 查询系统配置
	var isEnabled bool
	err := GlobalDB.QueryRow("SELECT is_enabled FROM baby_config LIMIT 1").Scan(&isEnabled)
	if err != nil {
		// 如果查询失败，默认返回开启状态
		log.Printf("查询宝宝系统配置失败: %v\n", err)
		return true
	}
	
	return isEnabled
}

// babyBirth 宝宝降临功能
func (p *BabyPlugin) babyBirth(robot plugin.Robot, event *onebot.Event) {
	// 检查用户是否已有宝宝
	if GlobalDB != nil {
		var count int
		err := GlobalDB.QueryRow("SELECT COUNT(*) FROM baby WHERE user_id = ? AND status = 'active'", event.UserID).Scan(&count)
		if err != nil {
			log.Printf("查询用户宝宝失败: %v\n", err)
			SendTextReply(robot, event, "❌ 查询失败，请稍后重试")
			return
		}
		
		if count > 0 {
			SendTextReply(robot, event, "❌ 您已经有宝宝了哦~\n💡 发送【我的宝宝】查看宝宝详情")
			return
		}
	}
	
	// 创建新宝宝
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
	if GlobalDB != nil {
		insertQuery := `
		INSERT INTO baby (user_id, name, birthday, growth_value, days_old, level, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err := GlobalDB.Exec(insertQuery, 
			baby.UserID, baby.Name, baby.Birthday, baby.GrowthValue, baby.DaysOld, 
			baby.Level, baby.Status, baby.CreatedAt, baby.UpdatedAt)
		if err != nil {
			log.Printf("创建宝宝失败: %v\n", err)
			SendTextReply(robot, event, "❌ 宝宝降临失败，请稍后重试")
			return
		}
	}

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
	var baby Baby
	if GlobalDB != nil {
		row := GlobalDB.QueryRow("SELECT id, user_id, name, birthday, growth_value, days_old, level FROM baby WHERE user_id = ? AND status = 'active'", event.UserID)
		err := row.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.Birthday, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
		if err != nil {
			SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
			return
		}
	} else {
		// 如果没有数据库连接，使用模拟数据
		baby = Baby{
			Name:        "小宝宝",
			Birthday:    time.Now().AddDate(0, 0, -10),
			GrowthValue: 5000,
			DaysOld:     5,
			Level:       1,
		}
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
	var baby Baby
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}
	
	row := GlobalDB.QueryRow("SELECT id, user_id, name, birthday, growth_value, days_old, level FROM baby WHERE user_id = ? AND status = 'active'", event.UserID)
	err := row.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.Birthday, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
	if err != nil {
		SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
		return
	}

	// 增加成长值
	growthAdd := 100
	newGrowthValue := baby.GrowthValue + growthAdd
	
	// 计算应该增加的天数（每1000成长值=1天）
	newDays := newGrowthValue / 1000
	if newDays > baby.DaysOld {
		// 更新天数和等级
		_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, days_old = ?, level = ? WHERE id = ?", 
			newGrowthValue, newDays, newDays/30+1, baby.ID)
		if err != nil {
			log.Printf("更新宝宝学习数据失败: %v\n", err)
			SendTextReply(robot, event, "❌ 学习过程中出现错误，请稍后重试")
			return
		}
		
		// 更新本地变量用于消息显示
		baby.GrowthValue = newGrowthValue
		baby.DaysOld = newDays
		baby.Level = newDays/30 + 1
	} else {
		// 只更新成长值
		_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ? WHERE id = ?", newGrowthValue, baby.ID)
		if err != nil {
			log.Printf("更新宝宝学习数据失败: %v\n", err)
			SendTextReply(robot, event, "❌ 学习过程中出现错误，请稍后重试")
			return
		}
		
		// 更新本地变量用于消息显示
		baby.GrowthValue = newGrowthValue
	}
	
	// 记录学习事件
	_, err = GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)", 
		baby.ID, "learn", "宝宝学习获得"+IntToString(growthAdd)+"点成长值")
	if err != nil {
		log.Printf("记录宝宝学习事件失败: %v\n", err)
	}

	msg := "📚 宝宝正在学习...\n"
	msg += "✅ 学习完成！获得" + IntToString(growthAdd) + "点成长值\n"
	msg += "📈 当前成长值：" + IntToString(baby.GrowthValue) + "\n"
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

	// 检查全局数据库连接
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}

	// 查询用户的宝宝
	var baby Baby
	row := GlobalDB.QueryRow("SELECT id, user_id, name, growth_value, days_old, level FROM baby WHERE user_id = ? AND status = 'active'", event.UserID)
	err := row.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
	if err != nil {
		SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
		return
	}

	// 检查用户积分是否足够
	// 注意：这里需要积分系统的支持，暂时注释掉
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
	newGrowthValue := baby.GrowthValue + growthAdd
	
	// 计算应该增加的天数（每1000成长值=1天）
	newDays := newGrowthValue / 1000
	newLevel := baby.Level
	if newDays > baby.DaysOld {
		newLevel = newDays/30 + 1 // 每30天升1级
	}
	
	// 更新宝宝信息
	_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, days_old = ?, level = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", 
		newGrowthValue, newDays, newLevel, baby.ID)
	if err != nil {
		log.Printf("更新宝宝购买数据失败: %v\n", err)
		SendTextReply(robot, event, "❌ 购买过程中出现错误，请稍后重试")
		return
	}

	// 记录购买事件
	_, err = GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)", 
		baby.ID, "buy", "购买了"+product.Name+"，获得"+IntToString(growthAdd)+"点成长值")
	if err != nil {
		log.Printf("记录宝宝购买事件失败: %v\n", err)
	}

	// 模拟积分数据用于测试
	userPoints := 500

	msg := "🎉 购买成功！\n"
	msg += "🛍️ 商品：" + product.Name + "\n"
	msg += "💰 花费积分：" + IntToString(product.Price) + "\n"
	msg += "剩余积分：" + IntToString(userPoints-product.Price) + "\n"
	msg += "📈 宝宝获得" + IntToString(growthAdd) + "点成长值\n"
	msg += "👶 宝宝当前成长值：" + IntToString(newGrowthValue) + "\n"
	msg += "⭐ 等级：" + IntToString(newLevel) + "\n"
	msg += "💡 宝宝变得更加强壮了！"

	SendTextReply(robot, event, msg)
}

// babyInteract 宝宝互动功能
func (p *BabyPlugin) babyInteract(robot plugin.Robot, event *onebot.Event) {
	// 查询用户的宝宝
	var baby Baby
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}
	
	row := GlobalDB.QueryRow("SELECT id, user_id, name, birthday, growth_value, days_old, level FROM baby WHERE user_id = ? AND status = 'active'", event.UserID)
	err := row.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.Birthday, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
	if err != nil {
		SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
		return
	}

	// 增加成长值
	growthAdd := 50
	newGrowthValue := baby.GrowthValue + growthAdd
	
	// 计算应该增加的天数（每1000成长值=1天）
	newDays := newGrowthValue / 1000
	if newDays > baby.DaysOld {
		// 更新天数和等级
		_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, days_old = ?, level = ? WHERE id = ?", 
			newGrowthValue, newDays, newDays/30+1, baby.ID)
		if err != nil {
			log.Printf("更新宝宝互动数据失败: %v\n", err)
			SendTextReply(robot, event, "❌ 互动过程中出现错误，请稍后重试")
			return
		}
		
		// 更新本地变量用于消息显示
		baby.GrowthValue = newGrowthValue
		baby.DaysOld = newDays
		baby.Level = newDays/30 + 1
	} else {
		// 只更新成长值
		_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ? WHERE id = ?", newGrowthValue, baby.ID)
		if err != nil {
			log.Printf("更新宝宝互动数据失败: %v\n", err)
			SendTextReply(robot, event, "❌ 互动过程中出现错误，请稍后重试")
			return
		}
		
		// 更新本地变量用于消息显示
		baby.GrowthValue = newGrowthValue
	}
	
	// 记录互动事件
	_, err = GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)", 
		baby.ID, "interact", "与宝宝互动获得"+IntToString(growthAdd)+"点成长值")
	if err != nil {
		log.Printf("记录宝宝互动事件失败: %v\n", err)
	}

	msg := "🎮 您正在和宝宝互动...\n"
	msg += "😊 宝宝很开心！获得" + IntToString(growthAdd) + "点成长值\n"
	msg += "📈 当前成长值：" + IntToString(baby.GrowthValue) + "\n"
	msg += "👶 宝宝名字：" + baby.Name + "\n"
	msg += "📅 年龄：" + p.getBabyAge(baby) + "\n"
	msg += "⭐ 等级：" + IntToString(baby.Level) + "\n"
	msg += "💡 多和宝宝互动可以增进亲子感情哦~"

	SendTextReply(robot, event, msg)
}

// babyWork 宝宝打工功能
func (p *BabyPlugin) babyWork(robot plugin.Robot, event *onebot.Event) {
	// 检查全局数据库连接
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}
	
	// 查询用户的宝宝
	var baby Baby
	row := GlobalDB.QueryRow("SELECT id, user_id, name, birthday, growth_value, days_old, level FROM baby WHERE user_id = ? AND status = 'active'", event.UserID)
	err := row.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.Birthday, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
	if err != nil {
		SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
		return
	}

	// 检查宝宝年龄是否足够打工（至少30天）
	if baby.DaysOld < 30 {
		SendTextReply(robot, event, "❌ 宝宝太小了，至少需要1个月才能打工哦~\n当前宝宝年龄：" + p.getBabyAge(baby))
		return
	}

	// 增加成长值和积分
	growthAdd := 150
	pointsAdd := 50
	newGrowthValue := baby.GrowthValue + growthAdd
	
	// 计算应该增加的天数（每1000成长值=1天）
	newDays := newGrowthValue / 1000
	newLevel := baby.Level
	if newDays > baby.DaysOld {
		newLevel = newDays/30 + 1 // 每30天升1级
	}
	
	// 更新宝宝信息
	_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, days_old = ?, level = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", 
		newGrowthValue, newDays, newLevel, baby.ID)
	if err != nil {
		log.Printf("更新宝宝打工数据失败: %v\n", err)
		SendTextReply(robot, event, "❌ 打工过程中出现错误，请稍后重试")
		return
	}

	// 记录宝宝打工事件
	_, err = GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)", 
		baby.ID, "work", "宝宝打工获得"+IntToString(growthAdd)+"点成长值和"+IntToString(pointsAdd)+"积分")
	if err != nil {
		log.Printf("记录宝宝打工事件失败: %v\n", err)
	}

	msg := "💼 宝宝开始打工了...\n"
	msg += "✅ 打工完成！获得" + IntToString(growthAdd) + "点成长值和" + IntToString(pointsAdd) + "积分\n"
	msg += "📈 当前成长值：" + IntToString(newGrowthValue) + "\n"
	msg += "👶 宝宝名字：" + baby.Name + "\n"
	msg += "📅 年龄：" + p.getBabyAge(baby) + "\n"
	msg += "⭐ 等级：" + IntToString(newLevel) + "\n"
	msg += "💡 打工可以培养宝宝的独立性和责任感哦~"

	SendTextReply(robot, event, msg)
}

// babyRename 宝宝改名功能
func (p *BabyPlugin) babyRename(robot plugin.Robot, event *onebot.Event, newName string) {
	if len(newName) < 2 || len(newName) > 10 {
		SendTextReply(robot, event, "❌ 宝宝名字长度必须在2-10个字符之间")
		return
	}

	// 检查全局数据库连接
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}
	
	// 查询用户的宝宝
	var oldName string
	row := GlobalDB.QueryRow("SELECT name FROM baby WHERE user_id = ? AND status = 'active'", event.UserID)
	err := row.Scan(&oldName)
	if err != nil {
		SendTextReply(robot, event, "❌ 您还没有宝宝哦~ 发送【宝宝降临】迎接新生命吧")
		return
	}

	// 更新宝宝名字
	_, err = GlobalDB.Exec("UPDATE baby SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ? AND status = 'active'", newName, event.UserID)
	if err != nil {
		log.Printf("更新宝宝名字失败: %v\n", err)
		SendTextReply(robot, event, "❌ 改名失败，请稍后重试")
		return
	}
	
	// 记录改名事件
	var babyID int
	row = GlobalDB.QueryRow("SELECT id FROM baby WHERE user_id = ? AND status = 'active'", event.UserID)
	row.Scan(&babyID)
	
	_, err = GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)", 
		babyID, "rename", "宝宝改名：从\""+oldName+"\"改为\""+newName+"\"")
	if err != nil {
		log.Printf("记录宝宝改名事件失败: %v\n", err)
	}

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

	// 检查全局数据库连接
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}

	// 更新系统配置为开启
	_, err := GlobalDB.Exec("UPDATE baby_config SET is_enabled = TRUE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		log.Printf("开启宝宝系统失败: %v\n", err)
		SendTextReply(robot, event, "❌ 操作失败，请稍后重试")
		return
	}

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

	// 检查全局数据库连接
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}

	// 更新系统配置为关闭
	_, err := GlobalDB.Exec("UPDATE baby_config SET is_enabled = FALSE, update_at = CURRENT_TIMESTAMP")
	if err != nil {
		log.Printf("关闭宝宝系统失败: %v\n", err)
		SendTextReply(robot, event, "❌ 操作失败，请稍后重试")
		return
	}

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

	// 检查全局数据库连接
	if GlobalDB == nil {
		SendTextReply(robot, event, "❌ 数据库连接失败，请稍后重试")
		return
	}

	// 查询用户的宝宝
	var count int
	err := GlobalDB.QueryRow("SELECT COUNT(*) FROM baby WHERE user_id = ? AND status = 'active'", userID).Scan(&count)
	if err != nil {
		log.Printf("查询用户宝宝失败: %v\n", err)
		SendTextReply(robot, event, "❌ 查询失败，请稍后重试")
		return
	}
	
	if count == 0 {
		SendTextReply(robot, event, "❌ 该用户没有宝宝")
		return
	}

	// 标记宝宝为已抛弃
	_, err = GlobalDB.Exec("UPDATE baby SET status = 'abandoned', updated_at = CURRENT_TIMESTAMP WHERE user_id = ? AND status = 'active'", userID)
	if err != nil {
		log.Printf("抛弃宝宝失败: %v\n", err)
		SendTextReply(robot, event, "❌ 操作失败，请稍后重试")
		return
	}

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
	
	// 检查全局数据库连接
	if GlobalDB == nil {
		log.Println("警告: 数据库未初始化，无法更新宝宝成长值")
		return
	}
	
	// 查询所有活跃状态的宝宝
	rows, err := GlobalDB.Query("SELECT id, user_id, name, birthday, growth_value, days_old, level FROM baby WHERE status = 'active'")
	if err != nil {
		log.Printf("查询活跃宝宝失败: %v\n", err)
		return
	}
	defer rows.Close()
	
	// 遍历所有宝宝，更新成长值
	for rows.Next() {
		var baby Baby
		err := rows.Scan(&baby.ID, &baby.UserID, &baby.Name, &baby.Birthday, &baby.GrowthValue, &baby.DaysOld, &baby.Level)
		if err != nil {
			log.Printf("扫描宝宝数据失败: %v\n", err)
			continue
		}
		
		growthAdd := 50 // 每日自动增加50成长值
		newGrowthValue := baby.GrowthValue + growthAdd
		
		// 计算应该增加的天数（每1000成长值=1天）
		newDays := newGrowthValue / 1000
		if newDays > baby.DaysOld {
			newLevel := newDays/30 + 1 // 每30天升1级
			
			// 更新宝宝数据到数据库
			_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, days_old = ?, level = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", 
				newGrowthValue, newDays, newLevel, baby.ID)
			if err != nil {
				log.Printf("更新宝宝 %s 数据失败: %v\n", baby.Name, err)
				continue
			}
			
			// 更新本地变量用于后续处理
				baby.GrowthValue = newGrowthValue
				baby.DaysOld = newDays
				baby.Level = newLevel
				
				// 检查是否过生日
				p.checkBirthday(baby)
				log.Printf("宝宝 %s 更新完成：成长值=%d, 天数=%d, 等级=%d\n", baby.Name, baby.GrowthValue, baby.DaysOld, baby.Level)
			} else {
				// 只更新成长值
				_, err = GlobalDB.Exec("UPDATE baby SET growth_value = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", 
					newGrowthValue, baby.ID)
				if err != nil {
					log.Printf("更新宝宝 %s 成长值失败: %v\n", baby.Name, err)
					continue
				}
				log.Printf("宝宝 %s 更新完成：成长值=%d\n", baby.Name, newGrowthValue)
			}
			
			// 检查是否达到宝宝达人徽章条件（成长值达到10000）
			if newGrowthValue >= 10000 && baby.GrowthValue < 10000 {
				// 获取徽章插件实例
				badgePlugin := GetBadgePluginInstance()
				// 发放宝宝达人徽章
				err := badgePlugin.GrantBadgeToUser(baby.UserID, "宝宝达人", "system", "宝宝成长值达到10000")
				if err != nil {
					log.Printf("给宝宝 %s 的用户 %s 发放宝宝达人徽章失败: %v\n", baby.Name, baby.UserID, err)
				} else {
					log.Printf("给宝宝 %s 的用户 %s 成功发放宝宝达人徽章\n", baby.Name, baby.UserID)
				}
			}
	}
	
	if err = rows.Err(); err != nil {
		log.Printf("遍历宝宝数据失败: %v\n", err)
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
		_, err := GlobalDB.Exec("INSERT INTO baby_event (baby_id, event_type, content) VALUES (?, ?, ?)", 
			baby.ID, "birthday", "宝宝今天过生日了！现在"+IntToString(baby.DaysOld)+"天了")
		if err != nil {
			log.Printf("记录宝宝 %s 生日事件失败: %v\n", baby.Name, err)
			return
		}
		
		log.Printf("🎉 宝宝 %s 今天过生日了！现在 %d 天了\n", baby.Name, baby.DaysOld)
	}
}