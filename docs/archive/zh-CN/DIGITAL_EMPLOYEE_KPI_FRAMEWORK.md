# 数字员工管理与 KPI 考核体系

## 1. 组织架构管理
数字员工通过 `EnterpriseID` 实现企业级租户隔离，并通过 `Department` 字段进行逻辑分组。

### 1.1 岗位分配 (Role Assignment)
每个数字员工必须关联一个 `RoleTemplate`。模板定义了：
- **核心技能**: 决定了员工能调用的 MCP 工具。
- **基础提示词**: 决定了员工的执行风格和决策逻辑。
- **KPI 权重**: 决定了绩效考核的侧重点。

## 2. 任务待办系统 (Todo System)
每个数字员工拥有独立的任务队列。任务分为：
- **待处理 (Pending)**: 指派给该员工但尚未开始执行的任务。
- **执行中 (Executing)**: 正在由 AI 引擎驱动的任务。
- **需审批 (Pending Approval)**: 正在等待人工介入的高风险任务。

## 3. KPI 考核机制 (KPI Framework)
系统根据任务执行结果自动计算员工绩效。主要维度包括：

| 维度 | 计算逻辑 | 权重 (示例) |
| :--- | :--- | :--- |
| **完成率 (Success Rate)** | `成功任务数 / 总任务数` | 40% |
| **执行效率 (Efficiency)** | `平均步骤耗时 vs 模板基准耗时` | 30% |
| **自主度 (Autonomy)** | `无人工干预执行数 / 总任务数` | 20% |
| **Token 成本 (Cost)** | `消耗 Token 数 vs 任务价值系数` | 10% |

### 3.1 自动调优
KPI 分数会影响员工的 `SalaryToken` 分配。低绩效员工可能会触发“再培训”逻辑（更新 BasePrompt 或技能补丁）。

## 4. API 接口参考 (Admin)

### 4.1 数字员工管理
- **GET** `/api/admin/employees/tasks`: 获取员工待办事项/任务列表。支持 `status` (pending, executing, pending_approval, completed, failed) 过滤。
- **GET** `/api/admin/employees/kpi`: 获取员工 KPI 统计数据。
- **POST** `/api/admin/employees/optimize`: 触发 AI 驱动的员工自动优化逻辑。

### 4.2 绩效回顾 (Performance Review)
1. 调用 `/api/admin/employees/kpi?id={id}` 获取实时的考核数据。
2. 调用 `/api/admin/employees/optimize` 对表现欠佳的员工进行自动化调优。
3. 系统分析最近 5 次失败的任务，并基于 AI 生成更精准的 `Bio` (人设/简介) 以提升后续执行成功率。

## 5. 安全与权限
- **主管视角**: 可查看下属所有员工的任务进度和执行详情。
- **审计跟踪**: 每一项 KPI 的变动都有对应的 `ExecutionID` 支撑，确保透明。
