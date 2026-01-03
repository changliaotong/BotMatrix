package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/models"
	"BotMatrix/common/utils"
	"BotNexus/tasks"
	"context"
	"fmt"
	"strings"
	"time"
)

// AgentCollaborationMCPHost 提供智能体间协作的 MCP 工具
type AgentCollaborationMCPHost struct {
	manager *Manager
}

func NewAgentCollaborationMCPHost(m *Manager) *AgentCollaborationMCPHost {
	return &AgentCollaborationMCPHost{manager: m}
}

func (h *AgentCollaborationMCPHost) ListTools(ctx context.Context, serverID string) ([]ai.MCPTool, error) {
	return []ai.MCPTool{
		{
			Name:        "colleague_list",
			Description: "获取当前企业内所有可用的数字员工（同事）列表及其职责描述",
			InputSchema: map[string]any{
				"type": "object",
			},
		},
		{
			Name:        "colleague_consult",
			Description: "向另一位数字员工咨询问题或请求协助。这会同步等待该员工的答复。",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_employee_id": map[string]any{
						"type":        "string",
						"description": "目标员工的工号 (EmployeeID)",
					},
					"question": map[string]any{
						"type":        "string",
						"description": "你想要咨询的具体问题或请求内容",
					},
				},
				"required": []string{"target_employee_id", "question"},
			},
		},
		{
			Name:        "task_delegate",
			Description: "将一个复杂的任务分解并委派给另一位数字员工处理。这适用于不需要立即得到最终结果，或者需要该员工独立完成的工作。返回 execution_id 用于后续查询进度。",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_employee_id": map[string]any{
						"type":        "string",
						"description": "目标员工的工号 (EmployeeID)",
					},
					"task_description": map[string]any{
						"type":        "string",
						"description": "任务的详细描述、背景信息及交付要求",
					},
					"priority": map[string]any{
						"type": "string",
						"enum": []string{"low", "medium", "high"},
					},
				},
				"required": []string{"target_employee_id", "task_description"},
			},
		},
		{
			Name:        "task_status",
			Description: "查询之前委派任务的执行状态和结果。",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"execution_id": map[string]any{
						"type":        "string",
						"description": "委派任务时返回的 execution_id",
					},
				},
				"required": []string{"execution_id"},
			},
		},
		{
			Name:        "task_report",
			Description: "主动向任务发起者汇报当前任务的进度、阶段性成果或最终结果。",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"execution_id": map[string]any{
						"type":        "string",
						"description": "当前正在处理的任务 execution_id",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "任务状态: running (执行中), success (成功), failed (失败)",
						"enum":        []string{"running", "success", "failed"},
					},
					"result": map[string]any{
						"type":        "string",
						"description": "汇报的具体内容、结果或错误信息",
					},
				},
				"required": []string{"execution_id", "status", "result"},
			},
		},
	}, nil
}

func (h *AgentCollaborationMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	// 从上下文获取当前组织 ID
	orgID, _ := ctx.Value("orgIDNum").(uint)
	if orgID == 0 {
		return nil, fmt.Errorf("unauthorized: organization ID missing in context")
	}

	// 传递父 SessionID 用于链路追踪
	parentSessionID, _ := ctx.Value("sessionID").(string)

	switch toolName {
	case "colleague_list":
		var employees []models.DigitalEmployeeGORM
		if err := h.manager.GORMDB.Where("enterprise_id = ?", orgID).Find(&employees).Error; err != nil {
			return nil, err
		}

		var sb strings.Builder
		sb.WriteString("以下是您可以协作的同事列表：\n")
		for _, e := range employees {
			sb.WriteString(fmt.Sprintf("- [%s] %s (%s): %s\n", e.EmployeeID, e.Name, e.Title, e.Bio))
		}

		return ai.MCPCallToolResponse{
			Content: []ai.MCPContent{
				{
					Type: "text",
					Text: sb.String(),
				},
			},
		}, nil

	case "colleague_consult":
		targetID, _ := arguments["target_employee_id"].(string)
		question, _ := arguments["question"].(string)

		// 1. 查找目标员工
		var targetEmp models.DigitalEmployeeGORM
		if err := h.manager.GORMDB.Where("enterprise_id = ? AND employee_id = ?", orgID, targetID).First(&targetEmp).Error; err != nil {
			return nil, fmt.Errorf("未找到工号为 %s 的同事", targetID)
		}

		// 2. 构造模拟内部消息
		internalMsg := h.manager.NewInternalMessage("mesh", "collaboration", targetEmp.BotID, "system", question)
		if parentSessionID != "" {
			internalMsg.Extras["parentSessionID"] = parentSessionID
		}

		// 3. 调用 AI 服务进行对话
		resp, err := h.manager.AIIntegrationService.ChatWithEmployee(&targetEmp, internalMsg, orgID)
		if err != nil {
			return nil, fmt.Errorf("向同事咨询失败: %v", err)
		}

		return ai.MCPCallToolResponse{
			Content: []ai.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("同事 %s 的答复：\n%s", targetEmp.Name, resp),
				},
			},
		}, nil

	case "task_delegate":
		targetID, _ := arguments["target_employee_id"].(string)
		taskDesc, _ := arguments["task_description"].(string)

		// 1. 查找目标员工
		var targetEmp models.DigitalEmployeeGORM
		if err := h.manager.GORMDB.Where("enterprise_id = ? AND employee_id = ?", orgID, targetID).First(&targetEmp).Error; err != nil {
			return nil, fmt.Errorf("未找到工号为 %s 的同事", targetID)
		}

		// 2. 创建任务执行记录用于追踪
		executionID := utils.GenerateRandomToken(8)
		now := time.Now()
		exec := tasks.Execution{
			ExecutionID: executionID,
			Status:      tasks.ExecRunning,
			TriggerTime: now,
			ActualTime:  &now,
			TraceID:     parentSessionID,
		}

		// 尝试关联到一个虚拟任务或创建一个记录
		if err := h.manager.GORMDB.Create(&exec).Error; err != nil {
			return nil, fmt.Errorf("创建任务追踪失败: %v", err)
		}

		// 3. 异步启动任务 (委派)
		go func() {
			internalMsg := h.manager.NewInternalMessage("mesh", "delegation", targetEmp.BotID, "system", taskDesc)
			internalMsg.Extras["executionID"] = executionID
			if parentSessionID != "" {
				internalMsg.Extras["parentSessionID"] = parentSessionID
			}

			resp, err := h.manager.AIIntegrationService.ChatWithEmployee(&targetEmp, internalMsg, orgID)

			// 更新执行结果
			status := tasks.ExecSuccess
			result := resp
			if err != nil {
				status = tasks.ExecFailed
				result = fmt.Sprintf("Error: %v", err)
			}

			h.manager.GORMDB.Model(&tasks.Execution{}).
				Where("execution_id = ?", executionID).
				Updates(map[string]any{
					"status": status,
					"result": result,
				})
		}()

		return ai.MCPCallToolResponse{
			Content: []ai.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("已成功将任务委派给同事 %s (%s)。任务追踪 ID: %s", targetEmp.Name, targetEmp.EmployeeID, executionID),
				},
			},
		}, nil

	case "task_status":
		execID, _ := arguments["execution_id"].(string)
		var exec tasks.Execution
		if err := h.manager.GORMDB.Where("execution_id = ?", execID).First(&exec).Error; err != nil {
			return nil, fmt.Errorf("未找到 ID 为 %s 的任务记录", execID)
		}

		return ai.MCPCallToolResponse{
			Content: []ai.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("任务 [%s] 当前状态: %s\n执行结果: %s", execID, exec.Status, exec.Result),
				},
			},
		}, nil

	case "task_report":
		execID, _ := arguments["execution_id"].(string)
		statusStr, _ := arguments["status"].(string)
		result, _ := arguments["result"].(string)

		status := tasks.ExecRunning
		switch statusStr {
		case "success":
			status = tasks.ExecSuccess
		case "failed":
			status = tasks.ExecFailed
		}

		err := h.manager.GORMDB.Model(&tasks.Execution{}).
			Where("execution_id = ?", execID).
			Updates(map[string]any{
				"status": status,
				"result": result,
			}).Error

		if err != nil {
			return nil, fmt.Errorf("汇报进度失败: %v", err)
		}

		return ai.MCPCallToolResponse{
			Content: []ai.MCPContent{
				{
					Type: "text",
					Text: "进度已成功汇报并同步至任务追踪系统。",
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
}

func (h *AgentCollaborationMCPHost) ListResources(ctx context.Context, serverID string) ([]ai.MCPResource, error) {
	return nil, nil
}

func (h *AgentCollaborationMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("resource not supported")
}

func (h *AgentCollaborationMCPHost) ListPrompts(ctx context.Context, serverID string) ([]ai.MCPPrompt, error) {
	return nil, nil
}

func (h *AgentCollaborationMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("prompt not supported")
}
