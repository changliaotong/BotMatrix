package collaboration

import (
	"time"
)

// Role 定义通用角色接口
type Role interface {
	// 获取角色ID
	GetID() string
	
	// 获取角色名称
	GetName() string
	
	// 获取角色类型
	GetType() string
	
	// 获取角色技能
	GetSkills() []string
	
	// 执行任务
	ExecuteTask(task Task) (Result, error)
	
	// 学习新技能
	LearnSkill(skill string) error
	
	// 获取角色状态
	GetStatus() RoleStatus
	
	// 设置角色状态
	SetStatus(status RoleStatus) error
}

// RoleStatus 定义角色状态
type RoleStatus struct {
	State     string    // 在线、离线、忙碌、空闲
	Load      float64   // 工作负载（0-100）
	LastActive time.Time
}

// Task 定义通用任务接口
type Task interface {
	// 获取任务ID
	GetID() string
	
	// 获取任务类型
	GetType() string
	
	// 获取任务描述
	GetDescription() string
	
	// 获取任务优先级
	GetPriority() int
	
	// 获取任务状态
	GetStatus() string
	
	// 设置任务状态
	SetStatus(status string) error
	
	// 获取任务输入
	GetInput() map[string]interface{}
	
	// 设置任务输出
	SetOutput(output map[string]interface{}) error
}

// Result 定义任务执行结果
type Result struct {
	Success     bool
	Output      map[string]interface{}
	Log         string
	Error       error
	ExecutionTime float64
}

// Collaborator 定义通用协作接口
type Collaborator interface {
	// 获取协作ID
	GetCollaborationID() string
	
	// 获取协作类型
	GetCollaborationType() string
	
	// 发起协作请求
	InitiateCollaboration(targetRoleID string, request CollaborationRequest) error
	
	// 响应协作请求
	RespondToCollaboration(requestID string, response CollaborationResponse) error
	
	// 取消协作请求
	CancelCollaboration(requestID string) error
	
	// 获取协作状态
	GetCollaborationStatus(requestID string) (CollaborationStatus, error)
}

// CollaborationRequest 定义协作请求
type CollaborationRequest struct {
	ID          string
	FromRoleID  string
	ToRoleID    string
	Type        string
	Content     map[string]interface{}
	Priority    int
	Deadline    time.Time
	CreatedAt   time.Time
}

// CollaborationResponse 定义协作响应
type CollaborationResponse struct {
	ID          string
	RequestID   string
	FromRoleID  string
	ToRoleID    string
	Status      string // 接受、拒绝、延迟
	Content     map[string]interface{}
	CreatedAt   time.Time
}

// CollaborationStatus 定义协作状态
type CollaborationStatus struct {
	ID          string
	Status      string // 待处理、进行中、已完成、已取消
	Progress    float64 // 协作进度（0-100）
	LastUpdate  time.Time
}

// Message 定义通用消息
type Message struct {
	ID          string
	Type        string // 任务分配、协作请求、状态更新、错误报告等
	FromRoleID  string
	ToRoleID    string
	Content     map[string]interface{}
	Priority    int
	Timestamp   time.Time
	Metadata    map[string]interface{}
}

// MessageHandler 定义消息处理函数
type MessageHandler func(message Message) error

// MessageBus 定义通用消息总线接口
type MessageBus interface {
	// 发送消息
	SendMessage(message Message) error
	
	// 订阅消息
	Subscribe(roleID string, handler MessageHandler) error
	
	// 取消订阅
	Unsubscribe(roleID string, handler MessageHandler) error
	
	// 广播消息
	Broadcast(message Message) error
	
	// 发送延迟消息
	SendDelayedMessage(message Message, delay time.Duration) error
}

// Workflow 定义通用工作流接口
type Workflow interface {
	// 获取工作流ID
	GetID() string
	
	// 获取工作流名称
	GetName() string
	
	// 获取工作流类型
	GetType() string
	
	// 启动工作流
	Start(input map[string]interface{}) error
	
	// 暂停工作流
	Pause() error
	
	// 恢复工作流
	Resume() error
	
	// 终止工作流
	Terminate() error
	
	// 获取工作流状态
	GetStatus() WorkflowStatus
	
	// 获取工作流进度
	GetProgress() float64
}

// WorkflowStatus 定义工作流状态
type WorkflowStatus struct {
	State     string    // 待启动、运行中、已暂停、已完成、已终止
	Progress  float64   // 工作流进度（0-100）
	StartedAt time.Time
	UpdatedAt time.Time
}