package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// Pet 宠物模型
type Pet struct {
	PetID     string    `json:"pet_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // 宠物类型：猫、狗、兔等
	Level     int       `json:"level"`
	Exp       int       `json:"exp"`
	Hunger    int       `json:"hunger"`    // 饥饿值 0-100
	Happiness int       `json:"happiness"` // 快乐值 0-100
	Health    int       `json:"health"`    // 健康值 0-100
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PetPlugin 宠物系统插件
type PetPlugin struct {
	pets      map[string]*Pet   // key: pet_id
	userPets  map[string][]*Pet // key: user_id
	cmdParser *CommandParser    // 命令解析器
}

// NewPetPlugin 创建宠物系统插件实例
func NewPetPlugin() *PetPlugin {
	return &PetPlugin{
		pets:      make(map[string]*Pet),
		userPets:  make(map[string][]*Pet),
		cmdParser: NewCommandParser(),
	}
}

func (p *PetPlugin) Name() string {
	return "pets"
}

func (p *PetPlugin) Description() string {
	return "宠物系统插件，支持领养宠物、喂食、玩耍、升级等功能"
}

func (p *PetPlugin) Version() string {
	return "1.0.0"
}

func (p *PetPlugin) Init(robot plugin.Robot) {
	log.Println("加载宠物系统插件")

	// 处理领养宠物命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为领养宠物命令
		if match, _ := p.cmdParser.MatchCommand("adopt|领养宠物|领养", event.RawMessage); !match {
			return nil
		}

		userID := event.UserID
		if userID == "" {
			p.sendMessage(robot, event, "无法获取用户ID，领养失败")
			return nil
		}

		// 检查用户是否已经有宠物
		if _, ok := p.userPets[userID]; ok && len(p.userPets[userID]) >= 3 {
			p.sendMessage(robot, event, "你最多只能领养3只宠物")
			return nil
		}

		// 随机生成宠物类型
		petTypes := []string{"🐱 猫咪", "🐶 狗狗", "🐰 兔子", "🐹 仓鼠", "🐻 小熊", "🐼 熊猫", "🐯 老虎", "🦁 狮子"}
		petType := petTypes[rand.Intn(len(petTypes))]

		// 生成宠物ID
		petID := fmt.Sprintf("pet_%d_%s", time.Now().Unix(), userID)

		// 创建新宠物
		pet := &Pet{
			PetID:     petID,
			UserID:    userID,
			Name:      fmt.Sprintf("%s的%s", userID, petType),
			Type:      petType,
			Level:     1,
			Exp:       0,
			Hunger:    80,
			Happiness: 80,
			Health:    100,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 存储宠物
		p.pets[petID] = pet
		p.userPets[userID] = append(p.userPets[userID], pet)

		p.sendMessage(robot, event, fmt.Sprintf("🎉 恭喜你领养了一只%s！\n宠物名字：%s\n等级：%d\n经验：%d\n饥饿值：%d\n快乐值：%d\n健康值：%d",
			petType, pet.Name, pet.Level, pet.Exp, pet.Hunger, pet.Happiness, pet.Health))

		return nil
	})

	// 处理查看宠物命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为查看宠物命令
		if match, _ := p.cmdParser.MatchCommand("pets|我的宠物|宠物", event.RawMessage); !match {
			return nil
		}

		userID := event.UserID
		if userID == "" {
			p.sendMessage(robot, event, "无法获取用户ID")
			return nil
		}

		// 获取用户的宠物
		userPets, ok := p.userPets[userID]
		if !ok || len(userPets) == 0 {
			p.sendMessage(robot, event, "你还没有宠物，使用/领养命令领养一只吧")
			return nil
		}

		// 发送宠物列表
		msg = "🐾 你的宠物 🐾\n"
		msg += "------------------------\n"
		for i, pet := range userPets {
			msg += fmt.Sprintf("%d. %s\n", i+1, pet.Name)
			msg += fmt.Sprintf("   类型：%s\n", pet.Type)
			msg += fmt.Sprintf("   等级：%d\n", pet.Level)
			msg += fmt.Sprintf("   经验：%d/%d\n", pet.Exp, pet.Level*100)
			msg += fmt.Sprintf("   饥饿值：%d/100\n", pet.Hunger)
			msg += fmt.Sprintf("   快乐值：%d/100\n", pet.Happiness)
			msg += fmt.Sprintf("   健康值：%d/100\n", pet.Health)
			msg += "------------------------\n"
		}

		p.sendMessage(robot, event, msg)

		return nil
	})

	// 处理喂食命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为喂食命令
		match, _, params := p.cmdParser.MatchCommandWithParams("feed|喂食", `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		userID := event.UserID
		if userID == "" {
			p.sendMessage(robot, event, "无法获取用户ID")
			return nil
		}

		// 获取用户的宠物
		userPets, ok := p.userPets[userID]
		if !ok || len(userPets) == 0 {
			p.sendMessage(robot, event, "你还没有宠物，使用/领养命令领养一只吧")
			return nil
		}

		// 解析宠物编号
		petIndex := 0
		if len(params) > 0 && params[0] != "" {
			index, err := fmt.Atoi(params[0])
			if err == nil && index > 0 && index <= len(userPets) {
				petIndex = index - 1
			}
		}

		pet := userPets[petIndex]

		// 喂食
		pet.Hunger += 20
		if pet.Hunger > 100 {
			pet.Hunger = 100
		}
		pet.Happiness += 5
		if pet.Happiness > 100 {
			pet.Happiness = 100
		}
		pet.Exp += 10

		// 检查升级
		p.checkLevelUp(pet)

		p.sendMessage(robot, event, fmt.Sprintf("🍖 你给%s喂食了！\n饥饿值：%d → %d\n快乐值：%d → %d\n经验值：%d → %d",
			pet.Name, pet.Hunger-20, pet.Hunger, pet.Happiness-5, pet.Happiness, pet.Exp-10, pet.Exp))

		return nil
	})

	// 处理玩耍命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为玩耍命令
		match, _, params := p.cmdParser.MatchCommandWithParams("play|玩耍", `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		userID := event.UserID
		if userID == "" {
			p.sendMessage(robot, event, "无法获取用户ID")
			return nil
		}

		// 获取用户的宠物
		userPets, ok := p.userPets[userID]
		if !ok || len(userPets) == 0 {
			p.sendMessage(robot, event, "你还没有宠物，使用/领养命令领养一只吧")
			return nil
		}

		// 解析宠物编号
		petIndex := 0
		if len(params) > 0 && params[0] != "" {
			index, err := fmt.Atoi(params[0])
			if err == nil && index > 0 && index <= len(userPets) {
				petIndex = index - 1
			}
		}

		pet := userPets[petIndex]

		// 玩耍
		pet.Happiness += 20
		if pet.Happiness > 100 {
			pet.Happiness = 100
		}
		pet.Hunger -= 10
		if pet.Hunger < 0 {
			pet.Hunger = 0
		}
		pet.Exp += 15

		// 检查升级
		p.checkLevelUp(pet)

		p.sendMessage(robot, event, fmt.Sprintf("🎮 你和%s玩耍了！\n快乐值：%d → %d\n饥饿值：%d → %d\n经验值：%d → %d",
			pet.Name, pet.Happiness-20, pet.Happiness, pet.Hunger+10, pet.Hunger, pet.Exp-15, pet.Exp))

		return nil
	})

	// 处理洗澡命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为洗澡命令
		match, _, params := p.cmdParser.MatchCommandWithParams("wash|洗澡", `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		userID := event.UserID
		if userID == "" {
			p.sendMessage(robot, event, "无法获取用户ID")
			return nil
		}

		// 获取用户的宠物
		userPets, ok := p.userPets[userID]
		if !ok || len(userPets) == 0 {
			p.sendMessage(robot, event, "你还没有宠物，使用/领养命令领养一只吧")
			return nil
		}

		// 解析宠物编号
		petIndex := 0
		if len(params) > 0 && params[0] != "" {
			index, err := fmt.Atoi(params[0])
			if err == nil && index > 0 && index <= len(userPets) {
				petIndex = index - 1
			}
		}

		pet := userPets[petIndex]

		// 洗澡
		pet.Health += 15
		if pet.Health > 100 {
			pet.Health = 100
		}
		pet.Happiness += 10
		if pet.Happiness > 100 {
			pet.Happiness = 100
		}
		pet.Exp += 5

		// 检查升级
		p.checkLevelUp(pet)

		p.sendMessage(robot, event, fmt.Sprintf("🛁 你给%s洗澡了！\n健康值：%d → %d\n快乐值：%d → %d\n经验值：%d → %d",
			pet.Name, pet.Health-15, pet.Health, pet.Happiness-10, pet.Happiness, pet.Exp-5, pet.Exp))

		return nil
	})

	// 定时更新宠物状态（每小时）
	go p.updatePetStatus()
}

// checkLevelUp 检查宠物是否升级
func (p *PetPlugin) checkLevelUp(pet *Pet) {
	requiredExp := pet.Level * 100
	if pet.Exp >= requiredExp {
		pet.Level++
		pet.Exp -= requiredExp
		pet.Health = 100
		pet.Happiness = 100
		pet.Hunger = 80

		log.Printf("宠物%s升级到%d级", pet.Name, pet.Level)
	}
}

// updatePetStatus 定时更新宠物状态
func (p *PetPlugin) updatePetStatus() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		// 每小时减少饥饿值和快乐值
		for _, pet := range p.pets {
			pet.Hunger -= 5
			if pet.Hunger < 0 {
				pet.Hunger = 0
			}

			pet.Happiness -= 5
			if pet.Happiness < 0 {
				pet.Happiness = 0
			}

			// 饥饿值或快乐值过低会影响健康
			if pet.Hunger < 20 || pet.Happiness < 20 {
				pet.Health -= 10
				if pet.Health < 0 {
					pet.Health = 0
				}
			}

			pet.UpdatedAt = time.Now()
		}
	}
}

// sendMessage 发送消息
func (p *PetPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, msg string) {
	if event.MessageType == "group" {
		robot.SendGroupMessage(event.GroupID, msg)
	} else {
		robot.SendPrivateMessage(event.UserID, msg)
	}
}
