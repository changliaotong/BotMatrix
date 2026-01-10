package development_team

import (
	"BotMatrix/common/ai"
	"context"
	"fmt"
	"time"
)

type ProgrammerImpl struct {
	aiSvc      ai.AIService
	skills     []string
	experience int
}

func NewProgrammer(aiSvc ai.AIService) *ProgrammerImpl {
	return &ProgrammerImpl{
		aiSvc:      aiSvc,
		skills:     []string{"code_generation", "refactoring", "bug_fixing", "optimization", "testing"},
		experience: 80,
	}
}

func (p *ProgrammerImpl) GetRole() string {
	return "programmer"
}

func (p *ProgrammerImpl) ExecuteTask(task Task) (Result, error) {
	startTime := time.Now()
	
	var result Result
	var err error
	
	switch task.Type {
	case "generate_code":
		prompt := task.Input["prompt"].(string)
		language := task.Input["language"].(string)
		code := p.GenerateCode(prompt, language)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"code": code,
			},
			Log: "Code generation completed",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	case "refactor_code":
		code := task.Input["code"].(string)
		improvements := task.Input["improvements"].([]string)
		refactored := p.RefactorCode(code, improvements)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"refactored_code": refactored,
			},
			Log: "Code refactoring completed",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	case "fix_bug":
		code := task.Input["code"].(string)
		bug := task.Input["bug_description"].(string)
		fixed := p.FixBug(code, bug)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"fixed_code": fixed,
			},
			Log: "Bug fixing completed",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	default:
		return Result{}, fmt.Errorf("unknown task type: %s", task.Type)
	}
	
	return result, err
}

func (p *ProgrammerImpl) GetSkills() []string {
	return p.skills
}

func (p *ProgrammerImpl) GetExperience() int {
	return p.experience
}

func (p *ProgrammerImpl) Learn(skill string, experience int) {
	for _, s := range p.skills {
		if s == skill {
			p.experience += experience
			return
		}
	}
	p.skills = append(p.skills, skill)
	p.experience += experience
}

func (p *ProgrammerImpl) GenerateCode(prompt string, language string) string {
	fullPrompt := `你是一名资深` + language + `程序员。根据以下需求生成高质量的` + language + `代码：

需求：` + prompt + `

要求：
1. 代码结构清晰，符合最佳实践
2. 包含必要的注释
3. 处理可能的错误情况
4. 提供示例用法（如果适用）

输出格式：纯代码，无需解释`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := p.aiSvc.Chat(ctx, fullPrompt, nil, nil)
	if err != nil {
		return fmt.Sprintf("代码生成失败：%v", err)
	}

	return response
}

func (p *ProgrammerImpl) RefactorCode(code string, improvements []string) string {
	improvementsStr := ""
	for _, imp := range improvements {
		improvementsStr += "- " + imp + "\n"
	}

	prompt := `你是一名资深代码重构专家。请根据以下改进点重构这段代码：

原始代码：
` + code + `

改进点：
` + improvementsStr + `

要求：
1. 保持功能不变
2. 提高代码可读性
3. 优化性能
4. 减少重复代码
5. 遵循最佳实践

输出格式：重构后的代码`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := p.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return fmt.Sprintf("代码重构失败：%v", err)
	}

	return response
}

func (p *ProgrammerImpl) FixBug(code string, bugDescription string) string {
	prompt := `你是一名资深调试专家。请修复这段代码中的bug：

原始代码：
` + code + `

Bug描述：` + bugDescription + `

要求：
1. 找出问题根源
2. 提供修复后的代码
3. 解释修复思路

输出格式：修复后的代码`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := p.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return fmt.Sprintf("bug修复失败：%v", err)
	}

	return response
}