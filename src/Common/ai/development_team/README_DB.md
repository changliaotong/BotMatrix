# 数据库驱动的数字开发团队自我进化系统

## 概述

本系统实现了一个数据库驱动的数字开发团队自我进化系统，通过数据库存储和管理数字角色、任务、项目和进化历史，实现了真正的动态配置和自我进化能力。

## 核心组件

### 1. 数据库模型 (models/digital_team_models.go)

- **DigitalRole**：数字角色模型，存储角色信息、技能、经验、提示词模板等
- **DigitalTask**：任务模型，存储任务信息、输入输出数据、状态等
- **DigitalProject**：项目模型，存储项目信息、技术栈、进度等
- **DigitalEvolution**：进化历史模型，记录角色的进化过程

### 2. 数据库管理器 (database_manager.go)

提供数据库操作接口：
- 创建、读取、更新、删除角色
- 创建、读取、更新任务和项目
- 批量初始化角色
- 自动迁移数据库表结构

### 3. 动态角色加载器 (dynamic_role_loader.go)

从数据库动态加载角色：
- 根据角色名称创建对应的角色实例
- 支持加载所有激活的角色
- 支持更新角色实例以反映数据库中的最新数据

### 4. 数据库驱动的自我进化框架 (self_evolution_db.go)

实现基于数据库的自我进化：
- 根据性能数据进化角色
- 自动修复代码
- 自动优化代码
- 学习新技能
- 记录进化历史
- 生成性能报告

## 优势

### 1. 动态配置
- 角色信息存储在数据库中，可随时修改
- 无需重新编译代码即可更新角色配置
- 支持动态添加新角色类型

### 2. 自我进化
- 基于数据库记录的性能数据自动进化
- 完整的进化历史追踪
- 可生成详细的性能报告

### 3. 可扩展性
- 模块化设计，易于添加新功能
- 支持多种数据库类型（PostgreSQL、MySQL等）
- 与现有系统无缝集成

### 4. 可维护性
- 集中管理角色信息
- 统一的数据库操作接口
- 易于监控和调试

## 使用示例

### 基本使用

```go
// 初始化数据库管理器
dbManager, err := NewDatabaseManager()
if err != nil {
    log.Fatalf("Failed to create database manager: %v", err)
}

// 初始化角色（如果不存在）
dbManager.SeedInitialRoles()

// 动态加载角色
loader := NewDynamicRoleLoader(database.GetDB())
programmer, err := loader.LoadRole("程序员")

// 执行任务
result, err := programmer.ExecuteTask(Task{
    ID:          "1",
    Type:        "generate_code",
    Description: "Generate API endpoint",
    Input: map[string]interface{}{
        "language": "Go",
        "framework": "Gin",
    },
})
```

### 自我进化

```go
// 创建进化管理器
evolution := NewDatabaseSelfEvolution(database.GetDB())

// 获取角色模型
roleModel, err := dbManager.GetRole("程序员")

// 进化角色
evolution.EvolveRole(roleModel, map[string]interface{}{
    "performance_score": 95.0,
    "top_skill":         "API design",
})

// 更新角色实例
loader.UpdateRole(programmer)
```

## 数据库表结构

### digital_roles 表

| 字段名       | 类型         | 描述                     |
|-------------|-------------|--------------------------|
| id          | uint        | 主键                     |
| role_name   | string      | 角色名称（唯一）         |
| description | text        | 角色描述                 |
| skills      | jsonb       | 技能列表                 |
| experience  | int         | 经验值                   |
| level       | int         | 角色等级                 |
| prompt      | text        | AI提示词模板             |
| config      | jsonb       | 角色配置                 |
| is_active   | bool        | 是否启用                 |
| created_at  | timestamp   | 创建时间                 |
| updated_at  | timestamp   | 更新时间                 |

### digital_tasks 表

| 字段名       | 类型         | 描述                     |
|-------------|-------------|--------------------------|
| id          | uint        | 主键                     |
| task_name   | string      | 任务名称                 |
| task_type   | string      | 任务类型                 |
| description | text        | 任务描述                 |
| input_data  | jsonb       | 输入数据                 |
| output_data | jsonb       | 输出数据                 |
| status      | string      | 任务状态                 |
| assigned_to | string      | 分配的角色               |
| priority    | int         | 任务优先级               |
| execution_time | float64  | 执行时间（秒）           |
| error_msg   | text        | 错误信息                 |
| created_at  | timestamp   | 创建时间                 |
| updated_at  | timestamp   | 更新时间                 |

### digital_projects 表

| 字段名       | 类型         | 描述                     |
|-------------|-------------|--------------------------|
| id          | uint        | 主键                     |
| project_name| string      | 项目名称                 |
| description | text        | 项目描述                 |
| requirements| text        | 项目需求                 |
| tech_stack  | jsonb       | 技术栈                   |
| status      | string      | 项目状态                 |
| progress    | float64     | 项目进度（0-100）        |
| results     | jsonb       | 项目结果                 |
| created_at  | timestamp   | 创建时间                 |
| updated_at  | timestamp   | 更新时间                 |

### digital_evolutions 表

| 字段名       | 类型         | 描述                     |
|-------------|-------------|--------------------------|
| id          | uint        | 主键                     |
| role_id     | uint        | 角色ID                   |
| role_name   | string      | 角色名称                 |
| evolution_type | string  | 进化类型                 |
| description | text        | 进化描述                 |
| before_data | jsonb       | 进化前数据               |
| after_data  | jsonb       | 进化后数据               |
| effectiveness | float64  | 进化效果（0-100）        |
| created_at  | timestamp   | 创建时间                 |

## 集成指南

### 1. 数据库配置

确保数据库连接已正确配置，支持 PostgreSQL、MySQL 等数据库。

### 2. 自动迁移

系统会自动创建数据库表结构，无需手动执行 SQL 脚本。

### 3. 初始化角色

使用 `SeedInitialRoles()` 方法初始化默认角色：

```go
dbManager.SeedInitialRoles()
```

### 4. 动态加载角色

使用 `DynamicRoleLoader` 动态加载角色：

```go
loader := NewDynamicRoleLoader(database.GetDB())
role, err := loader.LoadRole("程序员")
```

## 未来规划

### 1. 高级功能
- 支持角色间协作
- 更复杂的进化算法
- 多维度性能评估

### 2. 可视化
- 角色进化仪表盘
- 项目进度监控
- 性能分析报告

### 3. 扩展性
- 支持更多角色类型
- 与更多AI模型集成
- 支持分布式部署

## 结论

数据库驱动的数字开发团队自我进化系统提供了一种灵活、可扩展的方式来管理数字角色和实现自我进化。通过将角色信息存储在数据库中，系统实现了真正的动态配置和自我进化能力，为数字人功能的持续发展提供了坚实的基础。