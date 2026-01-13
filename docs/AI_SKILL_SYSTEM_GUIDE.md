# AI 技能系统技术指南 (AI Skill System Technical Guide)

## 1. 概述 (Overview)
为了实现“通用数字员工”的愿景，本系统采用**数据驱动 (Data-Driven)** 的架构设计。所有的技能（工具）定义、岗位要求以及执行逻辑均与硬编码解耦，存储于数据库中。这种设计不仅提高了系统的灵活性，也为未来从 C# 迁移到 Go 或 Python 提供了标准化的数据协议。

## 2. 核心架构 (Core Architecture)

### 2.1 数据库模式 (Database Schema)
核心表结构定义在 [init_ai_pg.sql](file:///d:/projects/BotMatrix/scripts/db/init_ai_pg.sql) 中：

- **`ai_skill_definitions`**: 存储所有原子技能。
    - `skill_key`: 唯一标识（如 `file_read`）。
    - `action_name`: LLM 决策时使用的行动名（如 `READ`）。
    - `parameter_schema`: JSON Schema 格式的参数要求。
    - `is_builtin`: 是否为内置技能（由 C# 实现）。
    - `script_content`: 动态技能的脚本代码（Python/Shell）。
- **`ai_job_definitions`**: 存储岗位定义。
    - `tool_schema`: 存储该岗位允许使用的技能 Key 列表（如 `["file_read", "git_op"]`）。

### 2.2 关键服务 (Key Services)

#### [SkillService.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Services/SkillService.cs)
技能调度的中心枢纽。
- **查找逻辑**：优先从数据库匹配 `skill_key`。
- **执行逻辑**：
    - 若为 `IsBuiltin`，分发至注入的 `ISkill` 实现类（如 `CommonSkills.cs`）。
    - 若为动态技能，则调用脚本执行引擎（预留 Python 槽位）。

#### [UniversalAgentManager.cs](file:///d:/projects/BotMatrix/src/BotWorker/Modules/AI/Services/UniversalAgentManager.cs)
通用 Agent 决策循环。
- **动态 Prompt**：在运行时根据岗位配置的 `tool_schema`，从数据库抓取技能的描述和参数 Schema，自动生成精准的工具说明，注入 LLM Prompt。
- **解耦决策**：LLM 只需输出 `Action` 和 `Target`，具体的执行细节由 `SkillService` 处理。

## 3. 开发指南 (Developer Guide)

### 3.1 如何添加新技能
1. **数据库注册**：在 `ai_skill_definitions` 中插入新技能记录。
2. **实现方式**：
    - **内置技能**：在 `CommonSkills.cs` 中添加对应的处理逻辑。
    - **动态技能**：直接在 `script_content` 中编写代码（适用于逻辑经常变动的业务）。
3. **岗位绑定**：在 `ai_job_definitions` 的 `tool_schema` 中加入该技能的 `skill_key`。

### 3.2 岗位配置示例
```json
// 软件开发架构师岗位 (dev_orchestrator)
{
  "job_key": "dev_orchestrator",
  "tool_schema": "[\"list_dir\", \"file_read\", \"file_write\", \"shell_exec\", \"git_op\", \"dotnet_build\"]"
}
```

## 4. 迁移与扩展 (Migration & Evolution)

### 4.1 迈向 Go/Python
- **数据一致性**：由于技能定义是 JSON 格式的，Go 或 Python 的 Agent 可以直接读取同一张数据库表，保持完全一致的工具契约。
- **脚本能力**：目前的架构已经支持存储脚本内容，未来只需实现一个多语言的 `ScriptRunner` 即可让数字员工具备无限的扩展能力。

## 5. 备忘 (Memos)
- **Repository 注册**：确保 `ISkillDefinitionRepository` 在 `Program.cs` 中注册为 Singleton。
- **JSONB 处理**：数据库中的 Schema 字段使用 JSONB，在 C# 中通过 `Dapper` 进行解析。
- **Prompt 优化**：生成的工具说明应包含清晰的示例，以降低 LLM 的幻觉。

---
*最后更新日期：2026-01-14*
