package collaboration

import (
	"log"
	"math"
	"sort"
	"sync"
)

// TaskAssigner 实现通用任务分配
type TaskAssigner struct {
	mu          sync.RWMutex
	roles       map[string]Role
	messageBus  MessageBus
}

// NewTaskAssigner 创建新的任务分配器实例
func NewTaskAssigner(messageBus MessageBus) *TaskAssigner {
	return &TaskAssigner{
		roles:      make(map[string]Role),
		messageBus: messageBus,
	}
}

// AddRole 添加角色到任务分配器
func (ta *TaskAssigner) AddRole(role Role) error {
	ta.mu.Lock()
	defer ta.mu.Unlock()
	
	ta.roles[role.GetID()] = role
	return nil
}

// RemoveRole 从任务分配器移除角色
func (ta *TaskAssigner) RemoveRole(roleID string) error {
	ta.mu.Lock()
	defer ta.mu.Unlock()
	
	delete(ta.roles, roleID)
	return nil
}

// AssignTask 分配任务给最合适的角色
func (ta *TaskAssigner) AssignTask(task Task) error {
	ta.mu.RLock()
	defer ta.mu.RUnlock()
	
	// 找到最合适的角色
	bestRole, err := ta.findBestRole(task)
	if err != nil {
		return err
	}
	
	// 发送任务分配消息
	message := Message{
		Type:       "task_assignment",
		FromRoleID: "task_assigner",
		ToRoleID:   bestRole.GetID(),
		Content: map[string]interface{}{
			"task_id":      task.GetID(),
			"task_type":    task.GetType(),
			"description":  task.GetDescription(),
			"priority":     task.GetPriority(),
			"input":        task.GetInput(),
		},
	}
	
	return ta.messageBus.SendMessage(message)
}

// findBestRole 找到最合适的角色
func (ta *TaskAssigner) findBestRole(task Task) (Role, error) {
	if len(ta.roles) == 0 {
		return nil, nil // 没有可用角色
	}
	
	// 评估每个角色的适合度
	type RoleScore struct {
		role  Role
	score float64
	}
	
	var scores []RoleScore
	
	for _, role := range ta.roles {
		score := ta.calculateRoleScore(role, task)
		scores = append(scores, RoleScore{role: role, score: score})
	}
	
	// 按分数排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// 返回分数最高的角色
	if len(scores) > 0 {
		return scores[0].role, nil
	}
	
	return nil, nil
}

// calculateRoleScore 计算角色适合度分数
func (ta *TaskAssigner) calculateRoleScore(role Role, task Task) float64 {
	// 技能匹配度（50%权重）
	skillsScore := ta.calculateSkillsScore(role, task)
	
	// 角色类型匹配度（20%权重）
	typeScore := ta.calculateTypeScore(role, task)
	
	// 工作负载（20%权重）
	loadScore := ta.calculateLoadScore(role)
	
	// 历史表现（10%权重）
	performanceScore := ta.calculatePerformanceScore(role, task)
	
	// 总分数
	totalScore := 0.5*skillsScore + 0.2*typeScore + 0.2*loadScore + 0.1*performanceScore
	
	log.Printf("Role %s score: %.2f (skills: %.2f, type: %.2f, load: %.2f, performance: %.2f)",
		role.GetName(), totalScore, skillsScore, typeScore, loadScore, performanceScore)
	
	return totalScore
}

// calculateSkillsScore 计算技能匹配度分数
func (ta *TaskAssigner) calculateSkillsScore(role Role, task Task) float64 {
	taskType := task.GetType()
	skills := role.GetSkills()
	
	// 简单的技能匹配逻辑（可根据需要扩展）
	matchCount := 0
	for _, skill := range skills {
		if skill == taskType || containsSkill(skill, taskType) {
			matchCount++
		}
	}
	
	if len(skills) == 0 {
		return 0
	}
	
	return float64(matchCount) / float64(len(skills)) * 100
}

// calculateTypeScore 计算角色类型匹配度分数
func (ta *TaskAssigner) calculateTypeScore(role Role, task Task) float64 {
	roleType := role.GetType()
	taskType := task.GetType()
	
	// 简单的类型匹配逻辑（可根据需要扩展）
	if roleType == taskType || containsType(roleType, taskType) {
		return 100
	}
	
	return 50 // 默认分数
}

// calculateLoadScore 计算工作负载分数
func (ta *TaskAssigner) calculateLoadScore(role Role) float64 {
	status := role.GetStatus()
	
	// 工作负载越低，分数越高
	load := status.Load
	if load > 100 {
		load = 100
	}
	
	return 100 - load
}

// calculatePerformanceScore 计算历史表现分数
func (ta *TaskAssigner) calculatePerformanceScore(role Role, task Task) float64 {
	// 简单的历史表现逻辑（可根据需要扩展）
	// 这里可以查询数据库获取角色的历史表现数据
	
	// 默认返回80分
	return 80
}

// containsSkill 检查技能是否匹配
func containsSkill(skill, taskType string) bool {
	// 简单的包含检查（可根据需要扩展）
	return len(skill) > 0 && len(taskType) > 0
}

// containsType 检查类型是否匹配
func containsType(roleType, taskType string) bool {
	// 简单的包含检查（可根据需要扩展）
	return len(roleType) > 0 && len(taskType) > 0
}

// BatchAssignTasks 批量分配任务
func (ta *TaskAssigner) BatchAssignTasks(tasks []Task) error {
	for _, task := range tasks {
		err := ta.AssignTask(task)
		if err != nil {
			log.Printf("Failed to assign task %s: %v", task.GetID(), err)
		}
	}
	return nil
}

// GetRoleLoad 获取角色负载
func (ta *TaskAssigner) GetRoleLoad(roleID string) (float64, error) {
	ta.mu.RLock()
	defer ta.mu.RUnlock()
	
	role, ok := ta.roles[roleID]
	if !ok {
		return 0, nil
	}
	
	return role.GetStatus().Load, nil
}

// GetAllRoleLoads 获取所有角色负载
func (ta *TaskAssigner) GetAllRoleLoads() map[string]float64 {
	ta.mu.RLock()
	defer ta.mu.RUnlock()
	
	loads := make(map[string]float64)
	for roleID, role := range ta.roles {
		loads[roleID] = role.GetStatus().Load
	}
	
	return loads
}

// BalanceLoad 平衡角色负载
func (ta *TaskAssigner) BalanceLoad() error {
	ta.mu.RLock()
	defer ta.mu.RUnlock()
	
	if len(ta.roles) <= 1 {
		return nil // 不需要平衡
	}
	
	// 计算平均负载
	totalLoad := 0.0
	for _, role := range ta.roles {
		totalLoad += role.GetStatus().Load
	}
	
	averageLoad := totalLoad / float64(len(ta.roles))
	
	// 调整负载过高的角色
	for _, role := range ta.roles {
		status := role.GetStatus()
		if status.Load > averageLoad*1.2 {
			// 降低负载
			newLoad := math.Max(averageLoad, status.Load-20)
			status.Load = newLoad
			role.SetStatus(status)
		}
	}
	
	return nil
}