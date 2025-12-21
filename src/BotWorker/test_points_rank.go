package main

import (
	"fmt"
	"log"
)

// PointsPlugin 积分系统插件
type PointsPlugin struct {
	// 存储用户积分，key为用户ID，value为积分数量
	points map[string]int
}

// NewPointsPlugin 创建积分系统插件实例
func NewPointsPlugin() *PointsPlugin {
	return &PointsPlugin{
		points: make(map[string]int),
	}
}

// getPointsRank 获取积分排行榜
func (p *PointsPlugin) getPointsRank() []PointsRankItem {
	// 转换为排行榜项列表
	var rank []PointsRankItem
	for userID, points := range p.points {
		if points > 0 {
			rank = append(rank, PointsRankItem{UserID: userID, Points: points})
		}
	}

	// 按积分降序排序
	for i := 0; i < len(rank); i++ {
		for j := i + 1; j < len(rank); j++ {
			if rank[j].Points > rank[i].Points {
				rank[i], rank[j] = rank[j], rank[i]
			}
		}
	}

	// 返回前10名
	if len(rank) > 10 {
		return rank[:10]
	}
	return rank
}

// PointsRankItem 排行榜项
type PointsRankItem struct {
	UserID string // 用户ID
	Points int    // 积分数量
}

func main() {
	log.Println("测试积分排行榜功能...")

	// 创建积分系统插件实例
	plugin := NewPointsPlugin()

	// 添加测试数据
	plugin.points["user1"] = 200
	plugin.points["user2"] = 150
	plugin.points["user3"] = 100
	plugin.points["user4"] = 80
	plugin.points["user5"] = 70
	plugin.points["user6"] = 60
	plugin.points["user7"] = 50
	plugin.points["user8"] = 40
	plugin.points["user9"] = 30
	plugin.points["user10"] = 20
	plugin.points["user11"] = 10

	// 测试排行榜生成
	rank := plugin.getPointsRank()

	// 输出排行榜
	log.Println("🏆 积分排行榜 🏆")
	log.Println("------------------------")
	for i, item := range rank {
		var medal string
		switch i {
		case 0:
			medal = "🥇"
		case 1:
			medal = "🥈"
		case 2:
			medal = "🥉"
		default:
			medal = fmt.Sprintf("%d.", i+1)
		}
		log.Printf("%s 用户%s：%d积分", medal, item.UserID, item.Points)
	}
	log.Println("------------------------")
	log.Printf("总参与人数：%d人", len(plugin.points))

	log.Println("积分排行榜功能测试通过!")
}