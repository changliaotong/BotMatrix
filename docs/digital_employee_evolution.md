# BotMatrix 数字员工演进架构文档

## 1. 概述
本项目旨在将 BotMatrix 打造为一个具备“自进化”能力的数字人开发团队平台。通过数据库驱动的动态组装、A/B 测试进化引擎以及基于 MCP 的能力扩展，实现数字员工从“单兵作战”到“团队协作”再到“对外商业化”的演进。

## 2. 核心架构：数据库驱动的数字员工

传统的硬编码 Agent 模式已被弃用，取而代之的是灵活的“乐高积木”式组装架构。

### 2.1 核心数据模型
- **DigitalJob (职位)**: 定义社会身份、职责范围和 KPI 标准。
  - 例：`Code Repair Expert`, `Marketing Specialist`
- **DigitalCapability (能力)**: 定义底层的工具属性，通常对应 MCP Server 或 API。
  - 例：`fs_read_write`, `browser_access`
- **JobCapabilityRelation (岗位-能力关联)**: 定义某个职位必须具备的底层能力。
- **EmployeeSkillRelation (技能树)**: 记录员工对特定技能的熟练度，支持经验值积累。

### 2.2 动态组装流程 (Recruit)
1.  **职位选择**: 根据 `DigitalJob` ID 确定目标岗位。
2.  **基因注入**: 根据岗位描述自动生成 System Prompt。
3.  **装备分发**: 自动查询并配置该岗位所需的 MCP 工具集。
4.  **变体应用**: 如果属于 A/B 测试，应用特定的 Prompt 变体或参数配置。

## 3. 进化引擎：A/B 测试框架

为了让数字员工不断优化，我们引入了科学的实验机制。

### 3.1 实验模型
- **ABExperiment**: 定义实验目标（如“提升代码修复通过率”）。
- **ABVariant**: 定义实验变体（如“保守型” vs “激进型”）。
  - 支持覆盖 `Prompt`、`Temperature`、`ModelID` 甚至 `MemorySnapshot`。

### 3.2 优胜劣汰
系统通过 KPI 数据（如任务完成率、用户满意度）持续评估不同变体的表现，最终将高分变体的配置固化为标准配置，完成一次“进化”。

## 4. 演进路线图

### 阶段一：单兵进化 (Self-Evolution) [当前阶段]
- [x] 数据库 Schema 设计与迁移
- [x] 数字员工工厂 (Factory) 实现
- [x] A/B 测试基础框架
- [x] 敏感配置外置 (Security)
- [ ] 完善 KPI 反馈循环
- [ ] 实现基于 MemorySnapshot 的记忆移植

### 阶段二：团队协作 (Collaboration)
- [ ] 定义多角色岗位 (PM, Architect, QA)
- [ ] 实现基于 Intent 的任务路由
- [ ] 搭建多 Agent 协作群组

### 阶段三：商业化 (Commercialization)
- [ ] 销售型数字员工开发
- [ ] 外部客户接入接口
- [ ] RAG 知识库深度集成

## 5. 配置与安全
数据库连接等敏感信息已从代码中剥离，通过 `config.json` 进行管理。
- 生产环境：请确保 `config.json` 配置正确且不被提交到版本控制。
- 开发环境：可参考 `config.json.example`。

## 6. 使用指南
### 招聘新员工
```go
factory := employee.NewDigitalEmployeeFactory(db)
emp, err := factory.Recruit(ctx, employee.RecruitParams{
    JobID: jobID,
    EnterpriseID: entID,
    VariantID: &variantID, // 可选，用于 A/B 测试
})
```
