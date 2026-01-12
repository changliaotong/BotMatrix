# 虚拟软件公司 (Virtual Software Company) 协作规范 v1

## 1. 核心岗位定义 (Core Roles)

### A. 需求分析师 (Requirements Analyst - RA)
- **目标**: 将模糊的用户文档转化为结构化的技术规格说明书。
- **输入**: 原始需求文档 (.md / .txt)。
- **输出**: `TECHNICAL_SPEC.json` (包含功能点、API 定义、数据模型)。

### B. 系统架构师 (System Architect - SA)
- **目标**: 设计项目结构和技术栈，拆解开发任务。
- **输入**: `TECHNICAL_SPEC.json`。
- **输出**: `PROJECT_STRUCTURE.json` (文件树、依赖关系、任务列表)。

### C. 软件开发员 (Software Developer - SD)
- **目标**: 根据任务列表编写具体的代码文件。
- **输入**: 单个文件任务 + 技术规格。
- **输出**: 源代码文件内容。

### D. 质量审计员 (QA/Reviewer)
- **目标**: 检查代码质量，运行静态分析或单元测试。
- **输入**: 源代码。
- **输出**: 评分与修改意见。

## 2. 自动化开发流 (Auto-Dev Workflow)

1. **Step 1: Parsing**: RA 解析文档，生成功能矩阵。
2. **Step 2: Designing**: SA 根据功能矩阵生成目录结构和空文件占位。
3. **Step 3: Coding**: 针对每个文件，启动 SD 实例进行填充。
4. **Step 4: Reviewing**: QA 审计，如果评分低于 80，回退给 SD 重写。
5. **Step 5: Integration**: 合并代码并尝试编译/运行。

## 3. 技术实现方案

- **存储**: 使用 `JobDefinition` 存储上述岗位。
- **驱动**: 引入 `DevWorkflowManager` 负责管理 Agent 之间的消息传递和状态流转。
- **进化**: 利用已实现的 `EvolutionService` 根据开发成功率自动优化 SD 的 Prompt。
