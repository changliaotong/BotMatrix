package ai

import (
	"BotMatrix/common/models"
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCrossDepartmentCollaboration(t *testing.T) {
	// 1. 初始化内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.DigitalEmployeeGORM{}, &models.DigitalEmployeeTaskGORM{})

	taskSvc := NewDigitalEmployeeTaskService(db, nil)

	ctx := context.Background()

	// 2. 创建两个不同部门的数字员工
	emp1 := models.DigitalEmployeeGORM{
		Name:       "张三",
		Department: "技术部",
		Title:      "架构师",
		BotID:      "bot-zhangsan",
		EmployeeID: "EMP_ZS_001",
	}
	db.Create(&emp1)

	emp2 := models.DigitalEmployeeGORM{
		Name:       "李四",
		Department: "财务部",
		Title:      "财务专员",
		BotID:      "bot-lisi",
		EmployeeID: "EMP_LS_002",
	}
	db.Create(&emp2)

	// 3. 张三创建一个任务并指派给李四 (跨部门协作)
	task := &models.DigitalEmployeeTaskGORM{
		ExecutionID: "collaboration-001",
		Title:       "技术部差旅费核销",
		Description: "请核销 2025 年 12 月技术部北京出差费用。",
		CreatorID:   "employee-zhansan", // 模拟张三创建
		AssigneeID:  emp2.ID,            // 指派给李四
		Status:      "pending",
	}

	err = taskSvc.CreateTask(ctx, task)
	assert.NoError(t, err)

	// 4. 验证任务指派情况
	var savedTask models.DigitalEmployeeTaskGORM
	db.Where("execution_id = ?", "collaboration-001").First(&savedTask)
	assert.Equal(t, emp2.ID, savedTask.AssigneeID)
	assert.Equal(t, "pending", savedTask.Status)

	// 5. 李四处理任务并创建一个子任务 (例如让行政部帮忙查行程单)
	subTask := &models.DigitalEmployeeTaskGORM{
		Title:       "查询北京行程单",
		Description: "查询张三 12 月 10 日北京往返行程单。",
		Status:      "pending",
	}

	err = taskSvc.CreateSubTask(ctx, "collaboration-001", subTask)
	assert.NoError(t, err)

	// 6. 验证子任务关联关系
	var savedSubTask models.DigitalEmployeeTaskGORM
	db.Where("title = ?", "查询北京行程单").First(&savedSubTask)
	assert.Equal(t, savedTask.ID, savedSubTask.ParentTaskID)
	assert.Contains(t, savedSubTask.ExecutionID, "sub-collaboration-001")
}
