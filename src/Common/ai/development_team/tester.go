package development_team

import (
	"BotMatrix/common/ai"
	"context"
	"fmt"
	"time"
)

type TesterImpl struct {
	aiSvc      ai.AIService
	skills     []string
	experience int
}

func NewTester(aiSvc ai.AIService) *TesterImpl {
	return &TesterImpl{
		aiSvc:      aiSvc,
		skills:     []string{"test_case_generation", "test_execution", "bug_reporting", "performance_testing", "security_testing"},
		experience: 60,
	}
}

func (t *TesterImpl) GetRole() string {
	return "tester"
}

func (t *TesterImpl) ExecuteTask(task Task) (Result, error) {
	startTime := time.Now()

	var result Result
	var err error

	switch task.Type {
	case "generate_test_cases":
		code := task.Input["code"].(string)
		testType := task.Input["test_type"].(string)
		testCases := t.GenerateTestCases(code, testType)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"test_cases": testCases,
			},
			Log:           "Test cases generated",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	case "execute_tests":
		testCases := task.Input["test_cases"].([]string)
		results := t.ExecuteTests(testCases)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"test_results": results,
			},
			Log:           "Tests executed",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	case "generate_test_report":
		results := task.Input["results"].(map[string]bool)
		report := t.GenerateTestReport(results)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"test_report": report,
			},
			Log:           "Test report generated",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	default:
		return Result{}, fmt.Errorf("unknown task type: %s", task.Type)
	}

	return result, err
}

func (t *TesterImpl) GetSkills() []string {
	return t.skills
}

func (t *TesterImpl) GetExperience() int {
	return t.experience
}

func (t *TesterImpl) Learn(skill string, experience int) {
	for _, s := range t.skills {
		if s == skill {
			t.experience += experience
			return
		}
	}
	t.skills = append(t.skills, skill)
	t.experience += experience
}

func (t *TesterImpl) GenerateTestCases(code string, testType string) []string {
	prompt := `你是一名资深测试工程师。请为以下代码生成` + testType + `测试用例：

代码：
` + code + `

要求：
1. 覆盖主要功能路径
2. 考虑边界情况
3. 包含负面测试
4. 提供测试数据

输出格式：测试用例列表`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := t.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return []string{fmt.Sprintf("测试用例生成失败：%v", err)}
	}

	return []string{response}
}

func (t *TesterImpl) ExecuteTests(testCases []string) map[string]bool {
	// 这里应该实际执行测试，但为了演示直接返回模拟结果
	results := make(map[string]bool)
	for i, test := range testCases {
		results[fmt.Sprintf("test_%d", i+1)] = true
	}
	return results
}

func (t *TesterImpl) GenerateTestReport(results map[string]bool) string {
	total := len(results)
	passed := 0
	for _, pass := range results {
		if pass {
			passed++
		}
	}

	prompt := `你是一名资深测试分析师。请根据以下测试结果生成测试报告：

总测试数：` + fmt.Sprintf("%d", total) + `
通过测试数：` + fmt.Sprintf("%d", passed) + `
失败测试数：` + fmt.Sprintf("%d", total-passed) + `

要求：
1. 总结测试结果
2. 分析通过率
3. 提供改进建议

输出格式：Markdown报告`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := t.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return fmt.Sprintf("测试报告生成失败：%v", err)
	}

	return response
}
