package development_team

import (
	"BotMatrix/common/ai"
	"context"
	"fmt"
	"time"
)

type ReviewerImpl struct {
	aiSvc      ai.AIService
	skills     []string
	experience int
}

func NewReviewer(aiSvc ai.AIService) *ReviewerImpl {
	return &ReviewerImpl{
		aiSvc:      aiSvc,
		skills:     []string{"code_review", "security_check", "best_practices", "performance_analysis", "documentation_review"},
		experience: 90,
	}
}

func (r *ReviewerImpl) GetRole() string {
	return "reviewer"
}

func (r *ReviewerImpl) ExecuteTask(task Task) (Result, error) {
	startTime := time.Now()
	
	var result Result
	var err error
	
	switch task.Type {
	case "review_code":
		code := task.Input["code"].(string)
		standards := task.Input["standards"].([]string)
		feedback := r.ReviewCode(code, standards)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"review_feedback": feedback,
			},
			Log: "Code review completed",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	case "check_security":
		code := task.Input["code"].(string)
		issues := r.CheckSecurity(code)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"security_issues": issues,
			},
			Log: "Security check completed",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	case "enforce_best_practices":
		code := task.Input["code"].(string)
		recommendations := r.EnforceBestPractices(code)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"best_practices": recommendations,
			},
			Log: "Best practices enforcement completed",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	default:
		return Result{}, fmt.Errorf("unknown task type: %s", task.Type)
	}
	
	return result, err
}

func (r *ReviewerImpl) GetSkills() []string {
	return r.skills
}

func (r *ReviewerImpl) GetExperience() int {
	return r.experience
}

func (r *ReviewerImpl) Learn(skill string, experience int) {
	for _, s := range r.skills {
		if s == skill {
			r.experience += experience
			return
		}
	}
	r.skills = append(r.skills, skill)
	r.experience += experience
}

func (r *ReviewerImpl) ReviewCode(code string, standards []string) []string {
	standardsStr := ""
	for _, std := range standards {
		standardsStr += "- " + std + "\n"
	}

	prompt := `你是一名资深代码审查专家。请根据以下标准审查这段代码：

代码：
` + code + `

审查标准：
` + standardsStr + `

要求：
1. 找出不符合标准的地方
2. 提供改进建议
3. 解释为什么需要改进

输出格式：审查反馈列表`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := r.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return []string{fmt.Sprintf("代码审查失败：%v", err)}
	}

	return []string{response}
}

func (r *ReviewerImpl) CheckSecurity(code string) []string {
	prompt := `你是一名资深安全专家。请检查这段代码中的安全漏洞：

代码：
` + code + `

要求：
1. 找出潜在的安全漏洞
2. 评估漏洞严重程度
3. 提供修复建议

输出格式：安全问题列表`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := r.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return []string{fmt.Sprintf("安全检查失败：%v", err)}
	}

	return []string{response}
}

func (r *ReviewerImpl) EnforceBestPractices(code string) []string {
	prompt := `你是一名资深软件工程师。请检查这段代码是否遵循最佳实践：

代码：
` + code + `

要求：
1. 找出不符合最佳实践的地方
2. 提供改进建议
3. 解释为什么需要改进

输出格式：最佳实践建议列表`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := r.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return []string{fmt.Sprintf("最佳实践检查失败：%v", err)}
	}

	return []string{response}
}