package collaboration

import (
	"fmt"
	"log"
)

// RoleAdapter 角色适配器
// 用于将现有的开发团队角色适配为通用协作角色

type RoleAdapter struct {
	role interface{}
}

// NewRoleAdapter 创建新的角色适配器
func NewRoleAdapter(role interface{}) *RoleAdapter {
	return &RoleAdapter{
		role: role,
	}
}

// GetID 获取角色ID
func (ra *RoleAdapter) GetID() string {
	// 尝试调用GetID方法
	if getID, ok := ra.role.(interface{ GetID() string }); ok {
		return getID.GetID()
	}
	return fmt.Sprintf("%p", ra.role)
}

// GetName 获取角色名称
func (ra *RoleAdapter) GetName() string {
	// 尝试调用GetName方法
	if getName, ok := ra.role.(interface{ GetName() string }); ok {
		return getName.GetName()
	}
	return "Unknown Role"
}

// GetType 获取角色类型
func (ra *RoleAdapter) GetType() string {
	// 尝试调用GetType方法
	if getType, ok := ra.role.(interface{ GetType() string }); ok {
		return getType.GetType()
	}
	return "unknown"
}

// GetSkills 获取角色技能
func (ra *RoleAdapter) GetSkills() []string {
	// 尝试调用GetSkills方法
	if getSkills, ok := ra.role.(interface{ GetSkills() []string }); ok {
		return getSkills.GetSkills()
	}
	return []string{}
}

// ExecuteTask 执行任务
func (ra *RoleAdapter) ExecuteTask(task Task) (Result, error) {
	// 尝试调用ExecuteTask方法
	if executeTask, ok := ra.role.(interface{ ExecuteTask(Task) (Result, error) }); ok {
		return executeTask.ExecuteTask(task)
	}
	return Result{}, fmt.Errorf("role does not support ExecuteTask method")
}

// LearnSkill 学习新技能
func (ra *RoleAdapter) LearnSkill(skill string) error {
	// 尝试调用LearnSkill方法
	if learnSkill, ok := ra.role.(interface{ LearnSkill(string) error }); ok {
		return learnSkill.LearnSkill(skill)
	}
	return fmt.Errorf("role does not support LearnSkill method")
}

// GetStatus 获取角色状态
func (ra *RoleAdapter) GetStatus() RoleStatus {
	// 尝试调用GetStatus方法
	if getStatus, ok := ra.role.(interface{ GetStatus() RoleStatus }); ok {
		return getStatus.GetStatus()
	}
	return RoleStatusIdle
}

// SetStatus 设置角色状态
func (ra *RoleAdapter) SetStatus(status RoleStatus) error {
	// 尝试调用SetStatus方法
	if setStatus, ok := ra.role.(interface{ SetStatus(RoleStatus) error }); ok {
		return setStatus.SetStatus(status)
	}
	return fmt.Errorf("role does not support SetStatus method")
}

// GetCollaborationID 获取协作ID
func (ra *RoleAdapter) GetCollaborationID() string {
	return ra.GetID()
}

// GetCollaborationType 获取协作类型
func (ra *RoleAdapter) GetCollaborationType() string {
	return ra.GetType()
}

// InitiateCollaboration 发起协作请求
func (ra *RoleAdapter) InitiateCollaboration(targetRoleID string, request CollaborationRequest) error {
	// 尝试调用InitiateCollaboration方法
	if initiateCollaboration, ok := ra.role.(interface{ InitiateCollaboration(string, CollaborationRequest) error }); ok {
		return initiateCollaboration.InitiateCollaboration(targetRoleID, request)
	}
	log.Printf("Role %s does not support InitiateCollaboration method", ra.GetName())
	return nil
}

// RespondToCollaboration 响应协作请求
func (ra *RoleAdapter) RespondToCollaboration(requestID string, response CollaborationResponse) error {
	// 尝试调用RespondToCollaboration方法
	if respondToCollaboration, ok := ra.role.(interface{ RespondToCollaboration(string, CollaborationResponse) error }); ok {
		return respondToCollaboration.RespondToCollaboration(requestID, response)
	}
	log.Printf("Role %s does not support RespondToCollaboration method", ra.GetName())
	return nil
}

// CancelCollaboration 取消协作请求
func (ra *RoleAdapter) CancelCollaboration(requestID string) error {
	// 尝试调用CancelCollaboration方法
	if cancelCollaboration, ok := ra.role.(interface{ CancelCollaboration(string) error }); ok {
		return cancelCollaboration.CancelCollaboration(requestID)
	}
	log.Printf("Role %s does not support CancelCollaboration method", ra.GetName())
	return nil
}

// GetCollaborationStatus 获取协作状态
func (ra *RoleAdapter) GetCollaborationStatus(requestID string) (CollaborationStatus, error) {
	// 尝试调用GetCollaborationStatus方法
	if getCollaborationStatus, ok := ra.role.(interface{ GetCollaborationStatus(string) (CollaborationStatus, error) }); ok {
		return getCollaborationStatus.GetCollaborationStatus(requestID)
	}
	return CollaborationStatusUnknown, fmt.Errorf("role does not support GetCollaborationStatus method")
}