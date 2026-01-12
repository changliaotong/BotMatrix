# 数字员工自我进化系统：从执行到进化的落地路线图

## 1. 远期目标（Vision）
构建一套**数字员工工厂（Digital Employee Factory）**，实现从岗位建模、自动化生产、稳定执行到持续进化的全生命周期管理。
- **核心定位**：不卖“智能”，卖“岗位交付”。
- **商业模式**：按事件计费、按件计费、岗位订阅。
- **技术终局**：边际成本趋近于零，岗位经验持续沉淀。

## 2. 当前项目现状分析（Current State）
- **已具备能力**：
  - `AIService` & `AgentExecutor`：基础执行引擎。
  - `ModelProviderManager`：多模型/多提供商动态切换。
  - `UserAIConfig`：用户自有 Key 与算力租赁（对应商业化雏形）。
  - `DigitalEmployeeToolFilter`：审计与风险控制。
  - `ImageGenerationPlugin`：已抽象出的“技能/插件”。
- **存在差距**：
  - 缺乏“岗位（Job）”的工程化定义，目前多为松散的“智能体（Agent）”。
  - 缺乏“工作流（Workflow）”的强制约束，执行逻辑依赖模型自主发挥。
  - 缺乏“评估（Evaluation）”与“进化（Evolution）”的闭环数据流。

## 3. 近期可执行计划（Immediate Actions）

### Phase 0: 基础设施升级（本周目标）
1. **数据模型落库**：
   - 参照 `PLAN.MD` 第十节，创建 `JobDefinition`、`EmployeeInstance`、`Task`、`TaskExecution` 等核心表。
2. **核心服务框架**：
   - 实现 `JobService`：负责岗位的 CRUD 与版本控制。
   - 实现 `EmployeeService`：负责根据岗位模板生成具体的员工实例。
3. **执行引擎重构**：
   - 升级 `AgentExecutor`，使其支持 `JobDefinition` 中定义的 `Workflow` 和 `Constraints`。

### Phase 1: 首个标杆岗位落地
1. **岗位选择**：**「社群运营执行员 (Community Operator)」**。
2. **能力实现**：
   - 自动欢迎、关键词回复、定时任务。
   - 计费埋点：按处理消息次数或成功触发动作计费。

## 4. 进化阶段规划 (Evolution Phases)

### Phase 0: 基础设施与闭环 (当前阶段 - 已完成)
- 建立 Job/Employee/Task 存储体系。
- 实现自动化评价 (EvaluationService)。
- 实现自动化进化 (EvolutionService)。

### Phase 1: 自我改造与系统增强 (进行中)
- **目标**: 让数字员工能够理解并修改 BotMatrix 源代码。
- **关键任务**:
  1. **代码感知能力**: 注入 CodebaseTool，允许 DE 读取系统文件。
  2. **架构反思岗位**: 定义 `system_architect_de`，定期扫描系统坏味道。
  3. **安全沙箱修改**: 实现受控的代码写入与自动测试验证。

### Phase 2: 多 Agent 协作与生产力爆炸 (下一阶段)
- **目标**: 形成完整的虚拟软件公司，支持复杂项目的全生命周期开发。
- **关键任务**:
  1. 完善 `DevWorkflowManager`。
  2. 实现跨 Agent 的状态机同步。
  3. 引入人类在环 (Human-in-the-loop) 的关键节点审批机制。

## 5. 当前任务：系统自我升级指令
数字员工将接受 `system_upgrade` 任务，目标是优化 `MetaData` 的查询性能或重构过时的 `BuiltinCommandMiddleware`。
