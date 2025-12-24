# 问答与 AI 系统开发文档

> 本文档面向 BotMatrix 开发者，用于快速理解“教机器人说话 + 群知识库 + AI 问答”的整体设计与实现细节。

---

## 1. 总体架构

### 1.1 组件概览

- **知识库插件**：`plugins/knowledge_base.go`
  - 负责“问 / 答”调教与群内问答的主流程
  - 插件类型：`KnowledgeBasePlugin`（`plugins/knowledge_base.go:14-17`）

- **管理员插件扩展**：`plugins/admin.go`
  - 提供“闭嘴 / 本群 / 官方 / 话唠 / 终极”等模式的管理命令
  - 将模式配置落地到 `group_ai_settings` 表（`internal/db/db.go:209-217, 966-972`）

- **工具函数与变量系统**：`plugins/utils.go`
  - 功能开关：`IsFeatureEnabledForGroup`（`plugins/utils.go:330-346`）
  - 变量处理：`NormalizeQuestion`、`SubstituteSystemVariables`、`SubstituteCustomVariables`、`SubstituteAllVariables`（`plugins/utils.go:950-1086`）
  - @ 检测：`IsAtMe`（`plugins/utils.go:1088-1104`）

- **数据库访问层（DAO）**：`internal/db/db.go`
  - 问题与答案模型：`Question`、`Answer`（`internal/db/db.go:974-994`）
  - 群问答模式模型：`GroupAISettings`（`internal/db/db.go:966-972`）
  - DAO：`GetQuestionByGroupAndNormalized`、`CreateQuestion`、`AddAnswer`、`GetRandomApprovedAnswer`、`GetGroupQAMode`、`SetGroupQAMode`（`internal/db/db.go:996-1204`）

- **AI 配置**：`internal/config/config.go`
  - AIConfig：`config.AI`（`internal/config/config.go:217-224, 274-280, 430-447`）
  - 暴露 `OfficialGroupID` 用于模式路由中的“官方群”逻辑

- **主程序加载**：
  - `cmd/main.go` 中加载知识库插件（`cmd/main.go:152-160`）
  - `test_cli.go` 中加载知识库插件用于命令行模拟（`test_cli.go:320-323`）

### 1.2 流程概览

1. **消息进入机器人**（`internal/server/combined.go`）
2. **插件系统广播事件**（`internal/plugin/plugin.go`）
3. **知识库插件接收群消息**（`KnowledgeBasePlugin.Init`，`plugins/knowledge_base.go:38-83`）
   - 优先尝试识别“问 / 答”调教格式 → `handleTeach`
   - 非调教消息再尝试进行问答 → `handleAsk`
4. **Admin 插件根据命令调整问答模式**（`plugins/admin.go:360-441`）
5. **工具层提供变量替换、功能开关与 @ 检测**（`plugins/utils.go`）
6. **DAO 层负责问题/答案持久化与查询**（`internal/db/db.go`）

---

## 2. 数据模型设计

### 2.1 问题与答案表（questions / answers）

**核心目标**：支持“多答案”的问答模型，一个问题可以拥有多条候选答案，答复时随机选取已通过审核的答案。

#### 2.1.1 表结构

建表定义位于 `internal/db/db.go:183-217`：

- `questions` 表：
  - `id SERIAL PRIMARY KEY`
  - `group_id VARCHAR(255) NOT NULL`
  - `question_raw TEXT NOT NULL`
  - `question_normalized TEXT NOT NULL`
  - `status VARCHAR(50) NOT NULL DEFAULT 'approved'`
  - `created_by VARCHAR(255)`
  - `source_group_id VARCHAR(255)`
  - `created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`
  - `updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`
  - `UNIQUE (question_normalized)`：保证不分群组对同一个规范化问题只有一条主问题记录

- `answers` 表：
  - `id SERIAL PRIMARY KEY`
  - `question_id INTEGER NOT NULL REFERENCES questions(id) ON DELETE CASCADE`
  - `answer TEXT NOT NULL`
  - `status VARCHAR(50) NOT NULL DEFAULT 'approved'`
  - `created_by VARCHAR(255)`
  - `created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`
  - `updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`

对应模型结构体定义：`internal/db/db.go:966-994`：

- `GroupAISettings`：记录每个群的问答模式
- `Question`：问题模型（包含 `QuestionRaw` / `QuestionNormalized` / `Status` 等）
- `Answer`：答案模型（包含 `QuestionID` / `Answer` / `Status` 等）

#### 2.1.2 DAO 接口

参考 `internal/db/db.go:996-1204`：

- 模式配置：
  - `GetGroupQAMode(db *sql.DB, groupID string) (string, error)`
  - `SetGroupQAMode(db *sql.DB, groupID string, mode string) error`

- 问题读写：
  - `GetQuestionByGroupAndNormalized(db *sql.DB, groupID, normalized string) (*Question, error)`
  - `CreateQuestion(db *sql.DB, q *Question) (*Question, error)`
    - 使用 `ON CONFLICT (question_normalized)` 做 upsert，实现全局唯一

- 答案读写：
  - `AddAnswer(db *sql.DB, a *Answer) (*Answer, error)`
  - `GetApprovedAnswersByQuestionID(db *sql.DB, questionID int) ([]*Answer, error)`
  - `GetRandomApprovedAnswer(db *sql.DB, questionID int) (*Answer, error)`
    - 逻辑上在已通过 `status='approved'` 的答案中随机挑选一条

### 2.2 群问答模式配置表（group_ai_settings）

建表定义：`internal/db/db.go:209-217`

- `group_id VARCHAR(255) NOT NULL UNIQUE`
- `qa_mode VARCHAR(50) NOT NULL`
- `created_at` / `updated_at`

与 DAO：

- `GetGroupQAMode`：读取指定群的模式，若无记录返回空字符串（交由插件使用默认值）
- `SetGroupQAMode`：持久化模式（`INSERT ... ON CONFLICT (group_id) DO UPDATE`）

问答模式枚举（在插件中使用的字符串）：

- `"silent"`：闭嘴模式
- `"group"`：本群模式
- `"official"`：官方模式
- `"chatty"`：话唠模式
- `"ultimate"`：终极模式

---

## 3. 知识库插件实现（KnowledgeBasePlugin）

文件：`plugins/knowledge_base.go`

### 3.1 插件结构

定义：`plugins/knowledge_base.go:14-23`

- `db *sql.DB`：数据库连接
- `officialGroupID string`：配置中的“官方问答群”ID

构造函数：

- `NewKnowledgeBasePlugin(database *sql.DB, officialGroupID string) *KnowledgeBasePlugin`  
  使用位置：
  - 主程序：`cmd/main.go:152-160`
  - CLI 测试：`test_cli.go:320-323`

插件元信息：

- `Name()` → `"knowledge_base"`（用于功能开关与插件管理，`plugins/knowledge_base.go:26-28`）
- `Description()` → `"教机器人说话和群知识库问答插件"`（`plugins/knowledge_base.go:30-32`）
- `Version()` → `"1.0.0"`（`plugins/knowledge_base.go:34-36`）

### 3.2 Init：消息注册与功能开关

入口：`KnowledgeBasePlugin.Init`（`plugins/knowledge_base.go:38-83`）

- 若 `p.db == nil`：打印日志并直接返回（在测试环境或数据库未配置时不启用）
- 使用两次 `robot.OnMessage` 注册：
  1. 第一次只处理“调教消息”，调用 `handleTeach`
  2. 第二次只处理“问答消息”，调用 `handleAsk`

两处都通过以下逻辑进行前置过滤：

- 只处理群消息：`event.MessageType == "group"`
- 功能开关：`IsFeatureEnabledForGroup(GlobalDB, groupID, "knowledge_base")`（`plugins/utils.go:330-346`）

### 3.3 调教识别：isTeachPattern

函数：`isTeachPattern`（`plugins/knowledge_base.go:85-92`）

- 输入：原始文本 `text`
- 处理：
  - `strings.TrimSpace`
  - 正则：`(?s)^[问Qq][:： ]+(.+?)[\s]+[答Aa][:： ]+(.+)$`
    - 支持中英文冒号、空格
    - 同时兼容 `问` / `Q/q` 和 `答` / `A/a`
    - `(?s)` 允许跨行匹配
- 返回值：是否匹配“问 / 答”模式

### 3.4 调教逻辑：handleTeach

函数：`handleTeach`（`plugins/knowledge_base.go:94-213`）

总体流程：

1. 前置检查：
   - 仅群消息
   - 文本去空格
   - 正则拆分出 `questionRaw` 与 `answerRaw`

2. 基础校验：
   - 若任一为空 → 提示“问题或答案不能为空”，结束

3. 积分检查：
   - 调用 `db.GetPoints(p.db, userIDStr)` 获取当前积分（`plugins/knowledge_base.go:120-125`）
   - 若积分 `< 0` → 提示“积分为负数时将不能再教机器人说话”，结束

4. 内容分析：
   - `containsSensitive`：检查问题或答案中是否包含脏话（`plugins/knowledge_base.go:276-288`）
   - `containsAdvertisement`：检查是否包含广告/推广等（`plugins/knowledge_base.go:290-304`）

5. 扣分策略：
   - 默认扣分：`-10`
   - 含脏话时：`-50`
   - 含广告时：`-100`
   - 调用：`db.AddPoints(p.db, userIDStr, deduct, "教机器人说话", "teach")`

6. 审核与特权判断：
   - 需要审核条件：`needReview := containsSensitive || containsAd`
   - 特权用户（直接生效，无需审核）：
     - `db.IsSuperAdmin(GlobalDB, groupIDStr, userIDStr)`
     - `db.IsUserInGroupWhitelist(GlobalDB, groupIDStr, userIDStr)`
   - 状态字段：
     - 若 `needReview && !isPrivileged` → `status = "pending"`
     - 否则 → `status = "approved"`

7. 问题规范化与持久化：
   - 规范化：`normalized := NormalizeQuestion(questionRaw)`（`plugins/utils.go:950-961`）
   - 问题写入或更新：
     - 构建 `db.Question` 实例
     - 调用 `db.CreateQuestion` 完成 upsert，返回最新 `Question`

8. 答案写入：
   - 构建 `db.Answer` 实例
   - 调用 `db.AddAnswer` 写入数据

9. 响应用户：
   - 默认：`"学习成功，已扣除积分"`
   - 若需要审核且非特权：`"已提交学习，包含敏感内容，需审核通过后生效"`

返回值含义：

- `true`：本次调教消息已被处理（无论成功与否），上层不需要继续处理
- `false`：不是调教消息，留给后续处理链（如其他插件）

### 3.5 问答逻辑：handleAsk

函数：`handleAsk`（`plugins/knowledge_base.go:215-260`）

1. 文本获取：
   - 优先使用 `event.RawMessage`
   - 若为空，尝试将 `event.Message` 断言为 `string`
   - 去首尾空白

2. 前置过滤：
   - 空字符串 → 直接返回
   - 命令前缀过滤：以 `/` 或 `／` 开头的消息视为命令，不进入问答逻辑

3. @ 机器人处理：
   - 使用 `IsAtMe(event)` 检查是否包含 `[CQ:at,qq=SelfID]`（`plugins/utils.go:1088-1104`）
   - 如果是 @ 消息，则先去掉 CQ 码，再做规范化与匹配

4. 规范化：
   - 使用 `NormalizeQuestion(clean)` 将问题标准化（去掉空格、换行等）

5. 读取问答模式：
   - `mode, err := db.GetGroupQAMode(p.db, groupIDStr)`
   - 若为空，则使用默认模式 `"group"`（本群模式）

6. 寻找答案：
   - 调用 `p.findAnswer(normalized, groupIDStr, mode, isAt)` 完成模式路由
   - 若未命中 → 返回 `false`，交给其他插件或 AI

7. 变量替换：
   - 调用 `SubstituteAllVariables(answerText, event)`（系统变量 + 自定义变量）

8. 发送回复：
   - 使用 `p.sendMessage(robot, event, final)`（内部封装了 `SendTextReply`）

### 3.6 模式路由：findAnswer / findFromGroups / findUltimate

核心函数：`findAnswer`（`plugins/knowledge_base.go:248-283`）

输入：

- `normalized`：规范化后问题字符串
- `groupID`：当前群 ID（字符串）
- `mode`：配置中的问答模式
- `isAt`：是否为 @ 机器人消息

逻辑：

1. 若 `isAt && (mode == "official" || mode == "chatty")` → 强制升级为 `"ultimate"` 模式
2. 根据 `effectiveMode` 分派：
   - `"silent"`：直接返回未命中
   - `"group"`：只在当前群查找
   - `"official"`：当前实现等同 `"group"`，但 @ 时会升级为 `"ultimate"`
   - `"chatty"`：优先当前群，其次官方群（`p.officialGroupID`）
   - `"ultimate"`：走 `findUltimate` 跨群搜索

辅助函数：

- `findFromGroups`（`plugins/knowledge_base.go:285-307`）
  - 按列表顺序遍历 `groupID` 列表，逐个：
    - `GetQuestionByGroupAndNormalized`
    - 检查 `Question.Status == "approved"`
    - 调用 `GetRandomApprovedAnswer`，并要求答案状态为 `approved`
  - 一旦命中返回答案文本

- `findUltimate`（`plugins/knowledge_base.go:309-322`）
  - 优先：
    1. 当前群
    2. 官方群（若配置）
  - 仍未命中：
    - 扫描 `questions` 表中所有 `question_normalized == normalized AND status='approved'` 的记录（排除当前群和官方群）
    - 对每个候选 `Question` 调用 `GetRandomApprovedAnswer`，返回首个合法答案

---

## 4. 管理命令与模式切换（AdminPlugin 扩展）

文件：`plugins/admin.go`

### 4.1 功能开关映射

功能开关配置在 `plugins/utils.go`：

- 默认启用：`FeatureDefaults["knowledge_base"] = true`（`plugins/utils.go:59-82`）
- 显示名：`FeatureDisplayNames["knowledge_base"] = "知识库"`（`plugins/utils.go:84-106`）

Admin 插件通过 `normalizeFeatureName` 将中文描述映射到 `featureID`（`plugins/admin.go:448-551`）。

### 4.2 模式切换命令

相关逻辑分布在 `plugins/admin.go:360-409` 及之后新增逻辑：

- 话唠模式：
  - 命令：`话唠` / `chatty`
  - 行为：`db.SetGroupQAMode(p.db, groupID, "chatty")`

- 终极模式：
  - 命令：`终极` / `ultimate`
  - 行为：`db.SetGroupQAMode(p.db, groupID, "ultimate")`

- 闭嘴模式：
  - 命令：`闭嘴` / `silent`
  - 行为：`db.SetGroupQAMode(p.db, groupID, "silent")`

- 本群模式：
  - 命令：`本群模式` / `本群问答` / `本群`
  - 行为：`db.SetGroupQAMode(p.db, groupID, "group")`

- 官方模式：
  - 命令：`官方模式` / `官方问答` / `官方`
  - 行为：`db.SetGroupQAMode(p.db, groupID, "official")`

所有模式切换命令都只对群聊有效，并要求 `p.db != nil`。

---

## 5. 变量系统实现细节

文件：`plugins/utils.go`

### 5.1 问题规范化：NormalizeQuestion

函数位置：`plugins/utils.go:950-961`

行为：

- 去掉首尾空白
- 使用 `strings.NewReplacer` 删除：
  - 普通空格 `" "`
  - 制表符 `"\t"`
  - 换行 `"\n"`
  - 回车 `"\r"`

该函数用于：

-存储 `Question.QuestionNormalized`
- 自定义变量名（如“客服QQ”）的规范化匹配

### 5.2 系统变量替换：SubstituteSystemVariables

函数位置：`plugins/utils.go:963-1037`

关键点：

- 输入：原始答案文本 + `*onebot.Event`
- 自动推导：
  - 提问者昵称 / 名片 / QQ 号
  - 当前时间（年、月、日、时、分、秒）
  - 当前群号
  - 当前星期几
  - 提问者积分（通过 `db.GetPoints` 获取）
- 使用映射表将变量替换：
  - `#你#` / `{你}`、`#我#` / `{我}`、`#积分#` / `{积分}` 等

该函数只负责“简单字符串替换”，不涉及复杂逻辑。为未实现的变量预留空间（值为空时跳过替换）。

### 5.3 自定义变量替换：SubstituteCustomVariables

函数位置：`plugins/utils.go:1039-1077`

语法：

- 使用 `{{变量名}}` 的形式引用自定义变量
  - 例如：`{{客服QQ}}`

实现流程：

1. 正则匹配：`\{\{([^{}]+)\}\}`
2. 对每个变量名：
   - 去掉首尾空白
   - 使用 `NormalizeQuestion` 标准化
   - 在当前群调用 `db.GetQuestionByGroupAndNormalized` 查找对应问题
   - 对命中的 `Question` 调用 `db.GetRandomApprovedAnswer` 获取答案
3. 用答案文本替换 `{{变量名}}`

注意：

- 只在当前群搜索变量对应的问题，不跨群
- 仅使用 `status='approved'` 的答案

### 5.4 综合替换：SubstituteAllVariables

函数位置：`plugins/utils.go:1079-1086`

- 先调用 `SubstituteSystemVariables`
- 再调用 `SubstituteCustomVariables`

这是问答逻辑中的主入口，用于保证所有变量按统一顺序被替换。

### 5.5 @ 机器人检测：IsAtMe

函数位置：`plugins/utils.go:1088-1104`

- 根据 `event.SelfID` 拼出 CQ 码 `[CQ:at,qq=SelfID]`
- 检查 `event.RawMessage` 或 `event.Message.(string)` 中是否包含该子串

用途：

- 在问答逻辑中区分普通消息与“@机器人”的消息
- 在某些模式下（官方 / 话唠），@ 会触发“终极模式”行为

---

## 6. AI 配置与集成点

文件：`internal/config/config.go`

### 6.1 AIConfig

定义：`internal/config/config.go:217-224`

- `APIKey string`：访问 AI 服务的密钥
- `Endpoint string`：API 地址
- `Model string`：模型名称
- `Timeout time.Duration`：调用超时
- `OfficialGroupID string`：官方问答群 ID

默认值：`DefaultConfig` 中设置（`internal/config/config.go:274-280`）。

加载逻辑：

- 从 JSON/YAML 配置文件解析 `jsonCfg.AI` 覆盖默认值（`internal/config/config.go:430-447`）

### 6.2 知识库插件中的 AI 集成扩展点

当前版本中，知识库插件专注于“本地问答”的处理流程：

- 优先命中本群/官方/其他群的问答
- 支持多模式路由、审核与变量替换

集成 AI 的推荐位置：

- 在 `handleAsk` 中，当 `findAnswer` 返回未命中时，可以：
  - 检查当前群是否开启 AI 功能（`FeatureDefaults["ai"]` + 群覆盖配置）
  - 构造 AI 请求（包含原始问题、上下文、群信息等）
  - 调用统一的 AI 客户端（未来可在 `plugins/ai.go` 或 `internal/ai` 中实现）
  - 对 AI 返回内容进行敏感词检测 / 截断 / 审核后再回复

这样可以满足“知识库优先，AI 兜底”的要求，并尽可能保持回答的一致性与安全性。

---

## 7. 扩展与演进建议

### 7.1 官方审核与重罚规则

需求摘要：

- 当用户调教内容被官方审核删除后，该用户后续每条调教内容应扣除 100 分，并且必须经过官方审核才能生效。

可选设计思路：

- 新增“调教黑名单表”或在用户表增加相关标记
- 在 `handleTeach` 中查询该标记：
  - 命中则强制 `deduct=-100` 且 `status="pending"`
  - 可结合 `audit_logs` 表记录操作来源与理由

### 7.2 更丰富的系统变量

现有变量集中在时间、积分、群号等基础信息上。后续可以考虑：

- 天气相关：`#天气预报#` / `{天气预报}`
- 农历相关：`#农历年#` / `{农历年}` 等
- 笑话 / 新闻 / 星座运势等可插拔服务

实现建议：

- 在 `SubstituteSystemVariables` 中对特定变量进行二次解析：
  - 调用对应插件或服务获取动态内容
  - 缓存结果避免频繁调用外部 API

### 7.3 问题模糊匹配与相似度检索

当前匹配采用“去空格后的完全匹配”（`NormalizeQuestion`）。为提升鲁棒性，可以：

- 引入简单的“包含匹配 / 前缀匹配”
- 或接入向量检索引擎实现语义相似度搜索

注意：

- 需谨慎处理歧义和性能问题
- 对结果附带置信度，用于分级：直接回复 / 需要再次确认 / 交给 AI

---

## 8. 快速排查指南

常见问题与排查思路：

1. **知识库完全不触发**
   - 检查是否在群聊中发送消息（插件只处理 `MessageType == "group"`）
   - 检查功能开关：`IsFeatureEnabledForGroup(GlobalDB, groupID, "knowledge_base")`
   - 确认数据库连接有效，且 `InitDatabase` 正常执行

2. **调教成功但不生效**
   - 检查 `questions` / `answers` 中该条记录的 `status` 是否为 `pending`
   - 检查当前群的问答模式是否为 `silent`（闭嘴）或 `group`（只查本群）
   - 若含脏话/广告，确认审核逻辑是否按预期执行

3. **变量未被替换**
   - 确认写法是否正确（如 `#你#` / `{你}` / `{{客服QQ}}`）
   - 对自定义变量，确认当前群下存在问题“客服QQ”，且有 `approved` 状态的答案

4. **跨群问答异常**
   - 检查 `official_group_id` 配置是否正确
   - 确认 `questions` 表中对应群的 `Question.Status` 为 `approved`

本开发文档建议与插件代码保持同步维护，新增功能时可以在本文件中追加章节，确保团队成员可以快速理解问答与 AI 子系统的整体设计。
