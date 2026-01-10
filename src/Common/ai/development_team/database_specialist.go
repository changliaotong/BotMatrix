package development_team

import (
	"BotMatrix/common/ai"
	"context"
	"fmt"
	"time"
)

type DatabaseSpecialistImpl struct {
	aiSvc      ai.AIService
	skills     []string
	experience int
}

func NewDatabaseSpecialist(aiSvc ai.AIService) *DatabaseSpecialistImpl {
	return &DatabaseSpecialistImpl{
		aiSvc:      aiSvc,
		skills:     []string{"database_design", "query_optimization", "data_migration", "performance_tuning", "security"},
		experience: 70,
	}
}

func (d *DatabaseSpecialistImpl) GetRole() string {
	return "database_specialist"
}

func (d *DatabaseSpecialistImpl) ExecuteTask(task Task) (Result, error) {
	startTime := time.Now()
	
	var result Result
	var err error
	
	switch task.Type {
	case "generate_schema":
		requirements := task.Input["requirements"].(string)
		schema := d.GenerateDatabaseSchema(requirements)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"schema": schema,
			},
			Log: "Database schema generated",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	case "optimize_queries":
		queries := task.Input["queries"].([]string)
		optimized := d.OptimizeQueries(queries)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"optimized_queries": optimized,
			},
			Log: "Queries optimized",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	case "migrate_database":
		oldSchema := task.Input["old_schema"].(string)
		newSchema := task.Input["new_schema"].(string)
		migration := d.MigrateDatabase(oldSchema, newSchema)
		result = Result{
			Success: true,
			Output: map[string]interface{}{
				"migration": migration,
			},
			Log: "Database migration plan generated",
			ExecutionTime: time.Since(startTime).Seconds(),
		}
	default:
		return Result{}, fmt.Errorf("unknown task type: %s", task.Type)
	}
	
	return result, err
}

func (d *DatabaseSpecialistImpl) GetSkills() []string {
	return d.skills
}

func (d *DatabaseSpecialistImpl) GetExperience() int {
	return d.experience
}

func (d *DatabaseSpecialistImpl) Learn(skill string, experience int) {
	for _, s := range d.skills {
		if s == skill {
			d.experience += experience
			return
		}
	}
	d.skills = append(d.skills, skill)
	d.experience += experience
}

func (d *DatabaseSpecialistImpl) GenerateDatabaseSchema(requirements string) string {
	prompt := `你是一名资深数据库设计师。根据以下需求生成数据库Schema：

需求：` + requirements + `

要求：
1. 使用PostgreSQL语法
2. 包含表结构、字段类型、约束
3. 考虑索引设计
4. 提供ER图描述
5. 包含数据迁移建议

输出格式：SQL语句`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := d.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return fmt.Sprintf("Schema生成失败：%v", err)
	}

	return response
}

func (d *DatabaseSpecialistImpl) OptimizeQueries(queries []string) []string {
	queriesStr := ""
	for _, q := range queries {
		queriesStr += "- " + q + "\n"
	}

	prompt := `你是一名SQL优化专家。请优化以下查询：

查询列表：
` + queriesStr + `

要求：
1. 分析性能瓶颈
2. 提供优化后的查询
3. 解释优化思路

输出格式：优化后的SQL语句`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := d.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return []string{fmt.Sprintf("查询优化失败：%v", err)}
	}

	return []string{response}
}

func (d *DatabaseSpecialistImpl) MigrateDatabase(oldSchema string, newSchema string) string {
	prompt := `你是一名数据库迁移专家。请生成从旧Schema到新Schema的迁移计划：

旧Schema：
` + oldSchema + `

新Schema：
` + newSchema + `

要求：
1. 提供数据迁移脚本
2. 考虑数据一致性
3. 提供回滚方案
4. 评估迁移风险

输出格式：SQL迁移脚本`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := d.aiSvc.Chat(ctx, prompt, nil, nil)
	if err != nil {
		return fmt.Sprintf("迁移计划生成失败：%v", err)
	}

	return response
}