package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// EventMessage represents an event received from the core
type EventMessage struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Payload any    `json:"payload"`
}

// Action represents an action to be performed
type Action struct {
	Type     string `json:"type"`
	Target   string `json:"target"`
	TargetID string `json:"target_id"`
	Text     string `json:"text"`
}

// ResponseMessage represents a response to the core
type ResponseMessage struct {
	ID      string   `json:"id"`
	OK      bool     `json:"ok"`
	Actions []Action `json:"actions"`
}

func main() {
	// 初始化数据库
	manager := common.NewManager()
	if err := manager.InitDB(); err != nil {
		fmt.Fprintf(os.Stderr, "数据库初始化失败: %v\n", err)
		return
	}

	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)

	for {
		var msg EventMessage
		err := decoder.Decode(&msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error decoding message: %v\n", err)
			continue
		}

		if msg.Type == "event" && msg.Name == "on_message" {
			payload, ok := msg.Payload.(map[string]any)
			if !ok {
				fmt.Fprintf(os.Stderr, "Invalid payload type\n")
				continue
			}

			text, textOk := payload["text"].(string)
			target, targetOk := payload["from"].(string)
			targetID, targetIDOk := payload["group_id"].(string)
			userID, userIDOk := payload["user_id"].(string)

			if !textOk || !targetOk || !targetIDOk || !userIDOk {
				fmt.Fprintf(os.Stderr, "Missing required fields in payload\n")
				continue
			}

			// 处理签到逻辑
			var response ResponseMessage
			response.ID = msg.ID
			response.OK = true

			// 检查是否是签到命令
			if text == "/签到" || text == "/ sign" || text == "/sign" {
				// 实际签到逻辑
				now := time.Now()
				signTime := now.Format("2006-01-02 15:04:05")

				// 检查今日是否已签到
				var lastSignTime time.Time
				var streak int
				var totalSignDays int
				var totalPoints int

				// 从数据库获取用户签到信息
				query := `SELECT last_sign_time, streak, total_sign_days, total_points FROM member_cache WHERE group_id = $1 AND user_id = $2`
				err := manager.DB.QueryRow(manager.prepareQuery(query), targetID, userID).Scan(&lastSignTime, &streak, &totalSignDays, &totalPoints)

				if err == sql.ErrNoRows {
					// 首次签到
					streak = 1
					totalSignDays = 1
					totalPoints = 10
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "查询签到信息失败: %v\n", err)
					response.Actions = []Action{
						{
							Type:     "send_message",
							Target:   target,
							TargetID: targetID,
							Text:     "签到失败，请稍后重试",
						},
					}
					encoder.Encode(response)
					continue
				} else {
					// 检查是否是今日首次签到
					lastSignDate := lastSignTime.Format("2006-01-02")
					today := now.Format("2006-01-02")

					if lastSignDate == today {
						// 今日已签到
						response.Actions = []Action{
							{
								Type:     "send_message",
								Target:   target,
								TargetID: targetID,
								Text:     "今日已签到，无需重复签到",
							},
						}
						encoder.Encode(response)
						continue
					} else {
						// 检查是否连续签到
						yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
						if lastSignDate == yesterday {
							streak++
						} else {
							streak = 1
						}
						totalSignDays++
						totalPoints += 10 + (streak - 1)
					}
				}

				// 更新数据库中的签到信息
				updateQuery := `UPDATE member_cache SET last_sign_time = $1, streak = $2, total_sign_days = $3, total_points = $4 WHERE group_id = $5 AND user_id = $6`
				_, err = manager.DB.Exec(manager.prepareQuery(updateQuery), now, streak, totalSignDays, totalPoints, targetID, userID)

				if err != nil {
					fmt.Fprintf(os.Stderr, "更新签到信息失败: %v\n", err)
					response.Actions = []Action{
						{
							Type:     "send_message",
							Target:   target,
							TargetID: targetID,
							Text:     "签到失败，请稍后重试",
						},
					}
					encoder.Encode(response)
					continue
				}

				// 构造回复消息
				basePoints := 10
				extraPoints := streak - 1
				totalPointsToday := basePoints + extraPoints

				response.Actions = []Action{
					{
						Type:     "send_message",
						Target:   target,
						TargetID: targetID,
						Text:     fmt.Sprintf("签到成功！\n签到时间：%s\n今日获得积分：%d（基础%d + 连续签到额外%d）\n连续签到天数：%d天\n总签到天数：%d天\n总积分：%d分", signTime, totalPointsToday, basePoints, extraPoints, streak, totalSignDays, totalPoints),
					},
				}
			} else if text == "/signstats" || text == "/ signstats" {
				// 实际签到统计
				var streak int
				var totalSignDays int
				var totalPoints int

				// 从数据库获取用户签到统计
				query := `SELECT streak, total_sign_days, total_points FROM member_cache WHERE group_id = $1 AND user_id = $2`
				err := manager.DB.QueryRow(manager.prepareQuery(query), targetID, userID).Scan(&streak, &totalSignDays, &totalPoints)

				if err == sql.ErrNoRows {
					// 从未签到
					response.Actions = []Action{
						{
							Type:     "send_message",
							Target:   target,
							TargetID: targetID,
							Text:     "您尚未签到过",
						},
					}
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "查询签到统计失败: %v\n", err)
					response.Actions = []Action{
						{
							Type:     "send_message",
							Target:   target,
							TargetID: targetID,
							Text:     "查询统计失败，请稍后重试",
						},
					}
				} else {
					// 构造统计消息
					response.Actions = []Action{
						{
							Type:     "send_message",
							Target:   target,
							TargetID: targetID,
							Text:     fmt.Sprintf("签到统计：\n总签到天数：%d天\n连续签到天数：%d天\n总积分：%d分", totalSignDays, streak, totalPoints),
						},
					}
				}
			} else {
				// 自动签到逻辑（如果今日未签到）
				var lastSignTime time.Time
				var streak int
				var totalSignDays int
				var totalPoints int

				// 从数据库获取用户签到信息
				query := `SELECT last_sign_time, streak, total_sign_days, total_points FROM member_cache WHERE group_id = $1 AND user_id = $2`
				err := manager.DB.QueryRow(manager.prepareQuery(query), targetID, userID).Scan(&lastSignTime, &streak, &totalSignDays, &totalPoints)

				if err == sql.ErrNoRows {
					// 首次签到
					streak = 1
					totalSignDays = 1
					totalPoints = 10
				} else if err != nil {
					// 查询失败，不进行自动签到
					continue
				} else {
					// 检查是否是今日首次签到
					lastSignDate := lastSignTime.Format("2006-01-02")
					today := now.Format("2006-01-02")

					if lastSignDate == today {
						// 今日已签到，不进行自动签到
						continue
					} else {
						// 检查是否连续签到
						yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
						if lastSignDate == yesterday {
							streak++
						} else {
							streak = 1
						}
						totalSignDays++
						totalPoints += 10 + (streak - 1)
					}
				}

				// 更新数据库中的签到信息
				updateQuery := `UPDATE member_cache SET last_sign_time = $1, streak = $2, total_sign_days = $3, total_points = $4 WHERE group_id = $5 AND user_id = $6`
				_, err = manager.DB.Exec(manager.prepareQuery(updateQuery), now, streak, totalSignDays, totalPoints, targetID, userID)

				if err != nil {
					fmt.Fprintf(os.Stderr, "自动签到失败: %v\n", err)
					continue
				}

				// 构造自动签到回复
				response.Actions = []Action{
					{
						Type:     "send_message",
						Target:   target,
						TargetID: targetID,
						Text:     "自动签到成功！今日获得10积分",
					},
				}
			}

			if err := encoder.Encode(response); err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
				continue
			}
		}
	}
}
