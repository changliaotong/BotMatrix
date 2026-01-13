# 🛠️ BotMatrix：系统集成与去插件化架构 (System Integration & De-pluginization)

## 1. 核心设计哲学 (Core Philosophy)
用户不应感知到“插件（Plugins）”的存在。在 BotMatrix 中，所有功能都被抽象为**系统模块（System Modules）**，它们是构成这个数字世界的逻辑基石。

---

## 2. 矩阵市场 (Matrix Market)
作为资源的调度与权限中心，负责管理系统模块的生命周期。

### 2.1 模块定义
| 系统 ID | 用户可见名称 | 描述 | 解锁成本 |
| :--- | :--- | :--- | :--- |
| `game.pet.v2` | **生命模拟系统** | 跨位面的生命形式模拟与培养 | 1000 积分 / Lv.1 |
| `game.marriage.v2` | **协议共鸣系统** | 建立深度逻辑链接与共鸣契约 | 5000 积分 / Lv.5 |
| `game.fishing.v2` | **位面垂钓系统** | 从虚空裂缝中打捞数据残片 | 500 积分 / Lv.1 |
| `game.music` | **音频流转系统** | 解析重构矩阵中的波形数据 | 2000 积分 / Lv.3 |

### 2.2 权限控制
- 存储于 `UserModuleAccess` 表。
- `MenuService` 在渲染时会实时校验此表，未激活系统显示为 `🔒` 锁定状态。

---

## 3. 金融级支撑 (Financial Foundation)
`PointsService` 提供了 `points.transfer` Skill，支持跨系统模块的复式记账转账。
- **系统储备 (SYSTEM_RESERVE)**: 负责积分发行（如签到奖励）。
- **系统收益 (SYSTEM_REVENUE)**: 负责资源回收（如系统激活费用）。

---

## 4. 进化感官 (Sensory Evolution)
### 4.1 系统脉动 (System Pulse)
通过 `EventNexus` 的审计流实现。用户可以实时观测到：
- 资源激活事件
- 大额资金流动
- 位面晋升记录

### 4.2 位面梯度
| 等级区间 | 位面名称 | 视觉标识 |
| :--- | :--- | :--- |
| 1 - 9 | 原质 (Prime) | ⚪ |
| 10 - 29 | 构件 (Component) | 🟢 |
| 30 - 59 | 逻辑 (Logic) | 🔵 |
| 60 - 89 | 协议 (Protocol) | 🟣 |
| 90 - 119 | 矩阵 (Matrix) | 🟡 |
| 120+ | 奇点 (Singularity) | 🔴 |

---

## 5. 矩阵先知：AI 知识库系统 (Matrix Oracle)
为了降低用户的学习成本，系统引入了基于 RAG 的 AI 问答体系。

### 5.1 数据流转
1. **源数据**: `docs/**/*.md` 文档。
2. **切片处理**: 将文档按 Markdown 标题或段落进行物理切片。
3. **向量化 (Embedding)**: 使用向量化模型将文本转化为 768/1536 维向量。
4. **存储**: 存入本地向量数据库。
5. **检索 (Retrieval)**: 当用户提问时，计算问题的向量，检索相似度最高的文档片段。
6. **生成 (Generation)**: 将片段作为上下文喂给 LLM，生成最终解答。

### 5.2 接入点
- 系统模块 ID: `core.oracle`
- 交互指令: `咨询 [问题]`、`问问 [问题]`

---

## 6. 开发者规范
1. **禁止直接提及“插件”**: 所有面向用户的提示文本必须使用“系统”、“模块”或“逻辑”。
2. **强制审计**: 关键业务逻辑必须发布 `SystemAuditEvent` 到 `EventNexus`。
3. **Skill 导向**: 模块间通讯应优先使用 `robot.CallSkillAsync` 而非直接引用。
