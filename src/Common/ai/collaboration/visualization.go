package collaboration

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// Visualization 协作可视化
// 实现数字员工协作的可视化展示
type Visualization struct {
	mu          sync.RWMutex
	messageBus  MessageBus
	workflowManager *WorkflowManager
	visualizationData map[string]interface{}
}

// VisualizationData 可视化数据
type VisualizationData struct {
	Timestamp    time.Time              `json:"timestamp"`
	Workflows    []WorkflowVisualization `json:"workflows"`
	Roles        []RoleVisualization    `json:"roles"`
	Messages     []MessageVisualization `json:"messages"`
	Tasks        []TaskVisualization    `json:"tasks"`
}

// WorkflowVisualization 工作流可视化
type WorkflowVisualization struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Status      WorkflowStatus    `json:"status"`
	Steps       []StepVisualization `json:"steps"`
}

// StepVisualization 步骤可视化
type StepVisualization struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Status      WorkflowStepStatus    `json:"status"`
	RoleType    string                `json:"role_type"`
}

// RoleVisualization 角色可视化
type RoleVisualization struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Status      RoleStatus        `json:"status"`
	Skills      []string          `json:"skills"`
	TaskCount   int               `json:"task_count"`
}

// MessageVisualization 消息可视化
type MessageVisualization struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	FromRoleID  string            `json:"from_role_id"`
	ToRoleID    string            `json:"to_role_id"`
	Timestamp   time.Time         `json:"timestamp"`
	Content     map[string]interface{} `json:"content"`
}

// TaskVisualization 任务可视化
type TaskVisualization struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Status      TaskStatus        `json:"status"`
	AssignedTo  string            `json:"assigned_to"`
	Priority    Priority          `json:"priority"`
}

// NewVisualization 创建新的可视化实例
func NewVisualization(messageBus MessageBus, workflowManager *WorkflowManager) *Visualization {
	return &Visualization{
		messageBus:     messageBus,
		workflowManager: workflowManager,
		visualizationData: make(map[string]interface{}),
	}
}

// GetVisualizationData 获取可视化数据
func (v *Visualization) GetVisualizationData() VisualizationData {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// 收集工作流数据
	workflows := v.getWorkflowVisualizations()
	
	// 收集角色数据
	roles := v.getRoleVisualizations()
	
	// 收集消息数据
	messages := v.getMessageVisualizations()
	
	// 收集任务数据
	tasks := v.getTaskVisualizations()

	return VisualizationData{
		Timestamp:    time.Now(),
		Workflows:    workflows,
		Roles:        roles,
		Messages:     messages,
		Tasks:        tasks,
	}
}

// getWorkflowVisualizations 获取工作流可视化数据
func (v *Visualization) getWorkflowVisualizations() []WorkflowVisualization {
	// 从工作流管理器获取工作流
	// 这里需要集成工作流管理器
	return []WorkflowVisualization{}
}

// getRoleVisualizations 获取角色可视化数据
func (v *Visualization) getRoleVisualizations() []RoleVisualization {
	// 从动态角色加载器获取角色
	// 这里需要集成动态角色加载器
	return []RoleVisualization{}
}

// getMessageVisualizations 获取消息可视化数据
func (v *Visualization) getMessageVisualizations() []MessageVisualization {
	// 从消息总线获取消息
	// 这里需要集成消息总线
	return []MessageVisualization{}
}

// getTaskVisualizations 获取任务可视化数据
func (v *Visualization) getTaskVisualizations() []TaskVisualization {
	// 从任务分配器获取任务
	// 这里需要集成任务分配器
	return []TaskVisualization{}
}

// ExportToJSON 导出为JSON
func (v *Visualization) ExportToJSON() (string, error) {
	data := v.GetVisualizationData()
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// ExportToGraphviz 导出为Graphviz格式
func (v *Visualization) ExportToGraphviz() (string, error) {
	data := v.GetVisualizationData()

	var graph string
	graph += "digraph Collaboration {
"
	graph += "  rankdir=LR;
"
	graph += "  node [shape=box];
"

	// 添加角色节点
	for _, role := range data.Roles {
		graph += fmt.Sprintf("  role_%s [label=\"%s\\n(%s)\"];
", role.ID, role.Name, role.Type)
	}

	// 添加工作流节点
	for _, workflow := range data.Workflows {
		graph += fmt.Sprintf("  workflow_%s [label=\"%s\\n%s\"];
", workflow.ID, workflow.Name, workflow.Status)
		
		// 添加步骤节点
		for _, step := range workflow.Steps {
			graph += fmt.Sprintf("  step_%s [label=\"%s\\n%s\"];
", step.ID, step.Name, step.Status)
			graph += fmt.Sprintf("  workflow_%s -> step_%s;
", workflow.ID, step.ID)
		}
	}

	// 添加消息边
	for _, msg := range data.Messages {
		if msg.FromRoleID != "" && msg.ToRoleID != "" {
			graph += fmt.Sprintf("  role_%s -> role_%s [label=\"%s\"];
", msg.FromRoleID, msg.ToRoleID, msg.Type)
		}
	}

	graph += "}"

	return graph, nil
}

// ExportToMermaid 导出为Mermaid格式
func (v *Visualization) ExportToMermaid() (string, error) {
	data := v.GetVisualizationData()

	var mermaid string
	mermaid += "flowchart LR\n"

	// 添加角色
	for _, role := range data.Roles {
		mermaid += fmt.Sprintf("  role_%s[\"%s\\n(%s)\"]\n", role.ID, role.Name, role.Type)
	}

	// 添加工作流
	for _, workflow := range data.Workflows {
		mermaid += fmt.Sprintf("  workflow_%s[\"%s\\n%s\"]\n", workflow.ID, workflow.Name, workflow.Status)
		
		// 添加步骤
		for _, step := range workflow.Steps {
			mermaid += fmt.Sprintf("  step_%s[\"%s\\n%s\"]\n", step.ID, step.Name, step.Status)
			mermaid += fmt.Sprintf("  workflow_%s --> step_%s\n", workflow.ID, step.ID)
		}
	}

	// 添加消息
	for _, msg := range data.Messages {
		if msg.FromRoleID != "" && msg.ToRoleID != "" {
			mermaid += fmt.Sprintf("  role_%s -->|%s| role_%s\n", msg.FromRoleID, msg.Type, msg.ToRoleID)
		}
	}

	return mermaid, nil
}