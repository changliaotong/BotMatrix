package development_team

import (
	"BotMatrix/common/ai"
	"context"
	"fmt"
	"time"
)

type ArchitectImpl struct {
	aiSvc      ai.AIService
	skills     []string
	experience int
}

func NewArchitect(aiSvc ai.AIService) *ArchitectImpl {
	return &ArchitectImpl{
		aiSvc:      aiSvc,
		skills:     []string{"system_design", "tech_stack_selection", "module_architecture", "scalability_design"},
		experience: 100,
	}
}

func (a *ArchitectImpl) GetRole() string {
	return "architect"
}

func (a *ArchitectImpl) ExecuteTask(task Task) (Result, error) {
	startTime := time.Now()
	
	var result Result
	var err error
	
	switch task.Type {
	case "design_architecture":
		requirements := task.Input["requirements"].(string)
		architecture := a.DesignArchitecture(requirements)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"architecture": architecture,
			},
			Log: "Architecture design completed",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	case "generate_tech_stack":
		techStack := a.GenerateTechStack()
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"tech_stack": techStack,
			},
			Log: "Tech stack generated",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	case "create_module_structure":
		modules := a.CreateModuleStructure()
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"modules": modules,
			},
			Log: "Module structure created",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	default:
		return Result{}, fmt.Errorf("unknown task type: %s", task.Type)
	}
	
	return result, err
}

func (a *ArchitectImpl) GetSkills() []string {
	return a.skills
}

func (a *ArchitectImpl) GetExperience() int {
	return a.experience
}

func (a *ArchitectImpl) Learn(skill string, experience int) {
	for _, s := range a.skills {
		if s == skill {
			a.experience += experience
			return
		}
	}
	a.skills = append(a.skills, skill)
	a.experience += experience
}

func (a *ArchitectImpl) DesignArchitecture(requirements string) string {
	prompt := `你是一名资深系统架构师，根据以下需求设计系统架构：

需求：` + requirements + `

请提供详细的架构设计，包括：
1. 系统总体架构图描述
2. 核心组件划分
3. 技术栈选择
4. 扩展性设计
5. 安全考虑

输出格式：Markdown文档`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := a.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return fmt.Sprintf("架构设计失败：%v", err)
	}

	return response
}

func (a *ArchitectImpl) GenerateTechStack() []string {
	prompt := `为一个现代分布式系统推荐技术栈，包括：
1. 后端框架
2. 数据库
3. 缓存
4. 消息队列
5. 前端框架
6. DevOps工具

输出格式：JSON数组`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := a.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return []string{fmt.Sprintf("技术栈生成失败：%v", err)}
	}

	// 这里应该解析JSON，但为了简单直接返回
	return []string{response}
}

func (a *ArchitectImpl) CreateModuleStructure() map[string]interface{} {
	prompt := `为一个微服务系统创建模块结构，包括：
1. 核心业务模块
2. 基础设施模块
3. 公共组件模块
4. API网关
5. 监控和日志模块

输出格式：JSON对象`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := a.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("模块结构创建失败：%v", err),
		}
	}

	return map[string]interface{}{
		"structure": response,
	}
}