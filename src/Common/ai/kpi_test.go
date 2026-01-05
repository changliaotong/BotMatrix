package ai

import (
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestEmployeeKPIAndOptimization(t *testing.T) {
	// 0. 初始化日志
	clog.InitDefaultLogger()

	// 1. 初始化内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.DigitalEmployeeGORM{}, &models.DigitalEmployeeTaskGORM{}, &models.AIModelGORM{}, &models.AIProviderGORM{})

	// 2. 模拟 AI 服务
	mockAI := &MockAIService{}
	kpiSvc := NewDigitalEmployeeKPIService(db, mockAI)

	ctx := context.Background()

	// 3. 创建测试员工与默认模型
	emp := models.DigitalEmployeeGORM{
		Name:       "王五",
		Bio:        "原始简介：我是一个普通的助理。",
		BotID:      "bot-wangwu",
		EmployeeID: "EMP001",
	}
	db.Create(&emp)

	db.Create(&models.AIModelGORM{
		ModelID:   "gpt-4",
		IsDefault: true,
	})

	// 4. 设置 Mock AI 响应
	mockAI.On("Chat", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&ChatResponse{
		Choices: []Choice{
			{Message: Message{Content: "优化后的简介：我是一名专业的资深助理，擅长处理各种复杂任务。"}},
		},
	}, nil)

	// 5. 创建一些失败的任务记录
	for i := 0; i < 3; i++ {
		db.Create(&models.DigitalEmployeeTaskGORM{
			ExecutionID: fmt.Sprintf("exec-fail-%d", i),
			AssigneeID:  emp.ID,
			Status:      "failed",
			ErrorMsg:    "Task execution timed out",
			CreatedAt:   time.Now().AddDate(0, 0, -i),
		})
	}

	// 6. 运行 KPI 计算
	score, err := kpiSvc.CalculateKPI(ctx, emp.ID)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), score) // 3 个任务，0 个完成，KPI 应为 0

	// 7. 运行优化
	err = kpiSvc.OptimizeEmployee(ctx, emp.ID)
	assert.NoError(t, err)

	// 8. 验证优化后的 Bio
	var updatedEmp models.DigitalEmployeeGORM
	db.First(&updatedEmp, emp.ID)
	assert.Contains(t, updatedEmp.Bio, "优化后的简介")

	// 9. 测试生成绩效报告
	report, err := kpiSvc.GetPerformanceReport(ctx, emp.ID, 30)
	assert.NoError(t, err)
	assert.NotEmpty(t, report)
	assert.Contains(t, report, "数字员工绩效报告")
	assert.Contains(t, report, "王五")
	t.Logf("绩效报告内容: %s", report)

	// 10. 测试高绩效场景
	emp2 := models.DigitalEmployeeGORM{
		Name:       "赵六",
		Title:      "高级架构师",
		BotID:      "bot-zhaoliu",
		EmployeeID: "EMP002",
	}
	db.Create(&emp2)
	for i := 0; i < 5; i++ {
		db.Create(&models.DigitalEmployeeTaskGORM{
			ExecutionID: fmt.Sprintf("exec-success-%d", i),
			AssigneeID:  emp2.ID,
			Status:      "completed",
			TokenUsage:  1000, // 低于 5000，成本得分高
			Duration:    120,  // 低于 300s，效率得分高
			CreatedAt:   time.Now().AddDate(0, 0, -i),
		})
	}

	report2, err := kpiSvc.GetPerformanceReport(ctx, emp2.ID, 30)
	assert.NoError(t, err)
	assert.Contains(t, report2, "表现优异")
	assert.Contains(t, report2, "成功 5")
	t.Logf("高绩效报告内容: %s", report2)

	// 11. 测试待办事项 (Todo List) 逻辑
	// 模拟 supervisor 查询赵六的待办事项
	var todoTasks []models.DigitalEmployeeTaskGORM
	db.Where("assignee_id = ? AND status IN ?", emp2.ID, []string{"pending", "executing", "pending_approval"}).Find(&todoTasks)
	// 目前赵六的任务全是 completed，所以 todoTasks 应该是空的
	assert.Equal(t, 0, len(todoTasks))

	// 给赵六指派一个新的待办任务
	db.Create(&models.DigitalEmployeeTaskGORM{
		ExecutionID: "todo-001",
		AssigneeID:  emp2.ID,
		Status:      "pending",
		Title:       "准备季度财报",
	})

	db.Where("assignee_id = ? AND status IN ?", emp2.ID, []string{"pending", "executing", "pending_approval"}).Find(&todoTasks)
	assert.Equal(t, 1, len(todoTasks))
	assert.Equal(t, "准备季度财报", todoTasks[0].Title)
}
