# 数字员工进化系统 (DE Evolution System) 规划文档

## 1. 核心理念
将“数字员工”从简单的对话助手，升级为**具有岗位职责、工作流程和自我进化能力的组织成员**。

### 1.1 从 Agent 到 Employee
- **Agent**: 解决单一任务（如“写一段代码”）。
- **Employee**: 负责岗位职责（如“初级开发工程师”，负责维护 A 模块，按 Git Flow 提交代码）。

## 2. 系统架构

### 2.1 岗位定义 (JobDefinition)
- **ID & Name**: 岗位的唯一标识。
- **Purpose**: 岗位的核心目标（用于 LLM 理解职责）。
- **Inputs/Outputs Schema**: 定义标准化的输入输出。
- **Workflow**: 结构化的执行步骤（例如：Plan -> Execute -> Review -> Refine）。
- **Constraints**: 约束条件（如：禁止修改数据库，必须使用 C# 10）。

### 2.2 员工实例 (EmployeeInstance)
- **EmployeeId**: 唯一的工号。
- **JobId**: 关联的岗位。
- **SkillSet**: 绑定的技能插件列表。
- **State**: 当前状态（空闲、忙碌、休假）。
- **Evolution Data**: 累积的经验值、成功率、用户评价。

### 2.3 任务与执行记录 (Task & Execution)
- **TaskRecord**: 记录发起的具体任务需求。
- **TaskExecution**: 记录任务执行的详细过程，包括每一个 Step 的输入输出和 LLM 调用。

## 3. 进化路径

### 3.1 第一阶段：结构化执行 (当前阶段)
- 实现岗位与员工的基础模型。
- 重构 `AgentExecutor`，使其能够识别 `JobDefinition` 中的 `Workflow` 并按步骤执行。
- 引入岗位约束（Constraints）作为 System Prompt 的一部分。

### 3.2 第二阶段：闭环评价 (近期目标)
- 引入 `EvaluationRule`，每个任务执行后，由专门的“评价 Agent”进行打分。
- 记录失败案例，作为后续任务的“负面提示词”。

### 3.3 第三阶段：自主进化 (远期目标)
- 员工能够根据多次任务的成功经验，建议修改 `JobDefinition` 的 `Workflow`（例如发现某个步骤多余）。
- 支持多员工协同：A 员工的输出是 B 员工的输入，形成自动化流水线。

## 4. 立即执行计划 (Action Items)

1.  **[已完成]** 核心数据库表模型设计与初始化。
2.  **[进行中]** 重构 `AgentExecutor`：
    - 增加对 `JobId` 的支持。
    - 在 Prompt 生成时，自动加载 `JobDefinition` 的 `Purpose` 和 `Constraints`。
    - 实现简单的 `SequentialWorkflow`（顺序流）执行引擎。
3.  **[待处理]** API Key 安全加固：
    - 对 `UserAIConfig` 中的 `ApiKey` 进行加密存储，防止数据库泄露。
    - 实现余额/算力额度监控。

---
*此文档随项目进度持续更新*
