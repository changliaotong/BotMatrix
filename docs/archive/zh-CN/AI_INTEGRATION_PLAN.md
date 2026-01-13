# BotMatrix AI 融入系统深度方案 (AI Integration Plan)

## 1. 核心目标
构建一个**供应商中立、多模型调度、高可靠**的 AI 基础设施，支撑系统的语义解析、自动化编排、智能对话及技能发现。

## 2. 架构设计

### 2.1 提供商管理 (Provider Management)
支持多种 AI 服务接入，包括但不限于：
- **公有云 API**: OpenAI, DeepSeek, Claude, Google Gemini, 阿里通义千问等。
- **本地部署**: 通过 Ollama 或 LocalAI 接入本地运行的模型。
- **聚合转发**: 支持兼容 OpenAI 格式的第三方代理。

### 2.2 模型管理 (Model Management)
精细化管理模型能力：
- **功能分类**: Chat (对话), Embedding (向量化), Vision (视觉), Rerank (重排序)。
- **动态发现**: 自动从提供商拉取可用模型清单。
- **性能配置**: 设置温度 (Temperature)、最大 Token、上下文窗口限制等。

### 2.3 智能路由与分发
- **任务绑定**: 为不同任务指定最合适的模型（例如：复杂编排用 GPT-4o，简单对话用 DeepSeek-V3）。
- **自动降级/故障切换**: 当主模型请求失败时，自动切换到备用模型。

### 2.4 多租户与私有 Token 支持 (User-specific Tokens)
为了降低系统运营成本并满足高级用户需求，系统支持“自有 Key”模式：
- **个人配置优先**: 当用户在个人设置中配置了特定提供商的 API Key 时，系统将优先使用该 Key 进行请求。
- **配额管理**: 系统管理员可以设置是否允许用户使用系统公共 Key，或仅限使用自有 Key。
- **隐私隔离**: 用户填写的 Token 仅对该用户及其关联的任务生效，管理员无法在后台明文查看。

### 2.5 提示词工程与模板管理 (Prompt Management)
- **模板版本化**: 所有的系统级提示词（意图识别、消息润色、摘要生成）均存储在数据库中，支持在线修改和版本回滚，无需重启服务。
- **变量注入**: 支持在提示词中使用变量（如 `{{user_nickname}}`, `{{current_time}}`），实现个性化回复。
- **A/B 测试**: 支持为同一个任务配置多个提示词模板，对比不同模型的输出效果。

### 2.6 长期记忆与本地知识库 (RAG Integration)
- **PGVector 集成**: 利用 PostgreSQL 的 `vector` 扩展，实现海量消息的向量化存储与检索。
- **知识库挂载**: 用户可以上传 PDF/Markdown 文档，机器人通过 RAG (检索增强生成) 技术基于私有知识库回答问题。
- **会话上下文压缩**: 当对话过长时，AI 自动生成会话摘要存入向量库，作为长期记忆的一部分。

### 2.7 标准化插件系统 (Standardized Function Calling)
- **能力报备**: 任何一个 Worker 都可以向 Nexus 报备其具备的“函数工具”（如：查余额、封禁用户）。
- **模型无关调用**: 无论底层模型是否支持原生的 Function Calling，由 Nexus 统一封装，将 AI 意图转化为具体的 Worker 指令。

### 2.8 内容安全与成本控制 (Safety & Cost Control)
- **敏感词过滤**: 在发送给 AI 之前和 AI 返回之后，进行双向合规性检查，防止敏感信息泄露或生成违规内容。
- **频次限制 (Rate Limiting)**: 针对不同级别的用户设置每分钟请求数 (RPM) 和每日 Token 消耗上限。
- **自动截断**: 监控输出长度，防止模型失控导致的高额 Token 计费。

### 2.9 多模态扩展 (Multi-modal Capabilities)
- **多媒体处理**: 支持图片识别（OCR、场景分析）、语音转文字 (STT) 和文字转语音 (TTS)。
- **跨模态任务**: 机器人可以理解用户发送的图片截图，并根据图片内容执行任务（如：根据报错截图自动给出修复建议）。

### 2.10 多智能体协作 (Multi-Agent Workflows)
- **角色定义**: 定义不同专长的 Agent（如：日志分析专家、群组管理助手、代码助手）。
- **自主协作**: 复杂任务将由“主 Agent”进行拆解，并分发给多个“从 Agent”并行处理，最后汇总结果。

### 2.11 观测性与评估 (Observability & Evaluation)
- **推理全路径追踪**: 在后台可以查看 AI 的完整思考过程（Thought Process），包括检索了哪些知识库、调用了哪些工具。
- **性能基准测试**: 定期自动运行基准测试集，评估不同供应商/模型在特定任务（如意图识别）上的准确率变化。

### 2.12 技能中心 (Skill Center) - 发现与分发
- **技能货架**: 集中展示所有已注册的系统技能和用户自定义技能。
- **一键订阅与挂载**: 用户可以在技能中心选择感兴趣的技能（如：智能财务、代码审查、翻译助手），一键挂载到指定的机器人或群组。
- **权限管理**: 技能开发者可以设置技能的可见范围（私有、组织内共享、全平台公开）。

### 2.13 技能训练中心 (Skill Training Center) - 进化与调优
- **语料标注与反馈**: 用户可以将机器人表现不佳的对话“送入训练中心”，通过手动修正（Labeling）转化为 Few-shot 示例，实时优化 AI 表现。
- **RAG 实验室**: 在线调试知识库检索效果，支持多种分段策略（Chunking）和向量化模型的对比测试。
- **提示词 IDE**: 提供一个可视化的提示词编辑环境，支持多版本预览和“一键跑测”基准测试集。
- **自动微调 (Fine-tuning)**：当标注数据达到一定规模（如 500 条以上）时，支持调用提供商 API（如 OpenAI, DeepSeek）启动自动化微调流程，训练该技能的专属模型。

### 2.14 交互式学习与反馈闭环 (Feedback Loop & RLHF)
- **满意度打分**: 所有的 AI 回复均支持“赞/踩”评价。
- **自动迭代**: 系统收集评价数据，作为后续提示词优化或模型微调（Fine-tuning）的训练素材。
- **纠错学习**: 用户修正 AI 的错误指令后，系统自动记录“纠错对”，防止同类错误再次发生。

### 2.14 自愈式系统运维 (Autonomous Maintenance)
- **智能日志诊断**: 当系统报错时，AI 自动调取上下文日志，分析根本原因并给出修复建议。
- **自动扩缩容建议**: AI 根据流量预测，建议管理员增加或减少 Worker 节点的数量。

### 2.15 智能意图分发系统 (Intelligent Intent Dispatch)
- **双层调度架构**: 区分“系统级调度”与“用户级调度”。
- **数字员工 (Digital Employee)**: 将机器人从“工具”抽象为“员工”，引入工号、职位、部门及 KPI 考评体系。
- **虚拟薪资 (Salary)**: 通过 Token 消耗计算数字员工的“运营成本”，实现 ROI 量化分析。

### 2.16 边缘 AI 策略 (Edge AI & Offline Support)
- **端侧处理**: 对于敏感度极高或网络环境受限的场景，支持在具备算力的 Worker 节点直接运行小参数模型（如 Phi-3, TinyLlama）。
- **离线指令集**: 预置基础意图识别模型，确保在断网情况下仍能执行基础的开关指令。

### 2.17 跨平台自媒体支持 (Multi-platform Expansion)
- **全渠道接入**: 系统将逐步支持 **微信公众号 (MP)**、**抖音 (TikTok)**、**微博** 等主流自媒体平台。
- **统一身份体系**: 无论用户从哪个平台接入，数字员工都能通过统一的 UserID 识别用户，并保持对话上下文的连贯性。
- **平台适配层**: 针对不同平台的交互特性（如：公众号的被动回复、抖音的评论区互动）提供专属的适配器插件。

### 2.18 企业级 B2B 协作与跨企业通信 (Enterprise B2B Collaboration)
- **企业数字边界**: 引入“企业 (Enterprise)”概念，数字员工隶属于特定企业，受企业策略管控。
- **自然语言对接协议**: 不同企业的数字员工之间可以使用自然语言进行业务对接。例如：企业 A 的采购员工直接与企业 B 的销售员工沟通库存。
- **通信效率优化**: 在保持自然语言灵活性的同时，系统会自动协商一种“结构化元数据 (Metadata)”协议，以减少 Token 消耗并提高处理精度。
- **安全认证与双向信任**: 基于企业级公私钥对进行通信加密和身份核验，确保跨企业业务对接的真实性与合规性。

### 2.19 移动端 App 规划 (Mobile App & Management Platform)
- **数字员工管理中枢**: 开发原生的移动 App，定位为“数字员工的移动管理平台”。
- **实时监控与干预**: 管理者可以在 App 上实时查看数字员工的工作状态、KPI 进度，并在必要时人工接管对话（Human-in-the-loop）。
- **随时随地部署**: 通过 App 快速扫描授权第三方平台账号，或为新员工分配技能。

## 3. 技术实现细节

### 3.1 数据库模型 (GORM Models)
在 `Common/models/gorm_models.go` 中已定义以下核心扩展模型：

```go
// EnterpriseGORM 企业/组织模型
type EnterpriseGORM struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string         `gorm:"size:255;uniqueIndex;not null;column:name" json:"name"` // 企业名称
	Code        string         `gorm:"size:100;uniqueIndex;not null;column:code" json:"code"` // 企业唯一代码 (用于 B2B 通信)
	Description string         `gorm:"type:text;column:description" json:"description"`
	OwnerID     uint           `gorm:"index;column:owner_id" json:"owner_id"`                // 企业所有者 (关联 UserGORM)
	Config      string         `gorm:"type:text;column:config" json:"config"`                // 企业级全局配置 (JSON)
	Status      string         `gorm:"size:20;default:'active';column:status" json:"status"` // active, suspended
	PublicKey   string         `gorm:"type:text;column:public_key" json:"public_key"`        // 用于 B2B 安全认证的公钥
	PrivateKey  string         `gorm:"type:text;column:private_key" json:"private_key"`      // 用于 B2B 安全认证的私钥 (加密存储)
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

// PlatformAccountGORM 第三方平台账号配置 (公众号, 抖音等)
type PlatformAccountGORM struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	EnterpriseID uint      `gorm:"index;column:enterprise_id" json:"enterprise_id"`
	Platform     string    `gorm:"size:50;not null;column:platform" json:"platform"` // wechat_mp, tiktok, weibo, etc.
	AccountName  string    `gorm:"size:100;column:account_name" json:"account_name"` // 账号名称
	AccountID    string    `gorm:"size:100;column:account_id" json:"account_id"`     // 平台内部 ID (如 AppID)
	Config       string    `gorm:"type:text;column:config" json:"config"`            // 平台配置 (JSON: AppSecret, Token, AESKey 等)
	Status       string    `gorm:"size:20;default:'active';column:status" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// B2BConnectionGORM 企业间 B2B 连接
type B2BConnectionGORM struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SourceEntID  uint      `gorm:"index:idx_b2b_conn;column:source_ent_id" json:"source_ent_id"`     // 发起方企业
	TargetEntID  uint      `gorm:"index:idx_b2b_conn;column:target_ent_id" json:"target_ent_id"`     // 接收方企业
	Status       string    `gorm:"size:20;default:'pending';column:status" json:"status"`            // pending, active, blocked
	AuthProtocol string    `gorm:"size:50;default:'mtls';column:auth_protocol" json:"auth_protocol"` // mtls, oauth2, custom
	Config       string    `gorm:"type:text;column:config" json:"config"`                            // 连接特定配置
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}
```

### 3.2 统一服务接口 (`Common/ai`)
定义统一的 AI 驱动接口，屏蔽底层 API 差异：

```go
type AIService interface {
    // Chat 对话接口
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    // Embed 向量化接口
    Embed(ctx context.Context, input string) ([]float32, error)
    // GetAvailableModels 获取当前可用的模型列表
    GetAvailableModels() []AIModelGORM
}
```

### 3.3 系统集成点
- **BotNexus/AIParser**: 从硬编码或环境变量转向从 `AIProvider` 数据库获取配置。
- **任务编排**: 允许用户在创建任务时选择“智能增强”选项。
- **实时翻译/分析**: 利用 Worker 节点的流式响应能力。

## 4. WebUI 管理界面规划

### 4.1 提供商控制面板
- 列表展示所有已配置的提供商及其连接状态。
- 提供“测试连接”功能。

### 4.2 模型策略配置
- 全局默认模型设置。
- 场景化模型分配（如：意图识别模型、回复生成模型）。

### 4.3 使用统计 (可选)
- 统计各提供商的 Token 消耗、成功率、响应延迟。

### 2.20 隐私堡垒与敏感信息脱敏 (Privacy Bastion & PII Masking)
为了解决用户在调用云端大模型时的隐私顾虑，系统引入了“隐私堡垒”技术：
- **【自动掩码替换】**：系统在将用户消息发送至第三方 LLM 前，会自动识别并使用占位符（如 `[PHONE_1]`, `[EMAIL_1]`）替换手机号、邮箱、身份证等敏感信息。
- **【本地安全还原】**：AI 生成回复后，Nexus 节点会在本地将占位符还原为原始信息。第三方大模型仅能感知到数据结构，无法获取真实的个人隐私。
- **【私有化部署闭环】**：支持 100% 离线部署模式。用户可以在 NAS、家庭服务器或具备算力的本地设备上运行全套系统（包括向量库、数据库和本地 LLM），实现真正的“数据不出户”。

## 5. 实施路线图

1.  **Phase 1**: 数据库模型建立与迁移，完成 `Common/ai` 基础架构。
2.  **Phase 2**: 在 BotNexus 中实现提供商与模型的 CRUD 接口。
3.  **Phase 3**: 重构 `AIParser` 使其支持动态模型选择。
4.  **Phase 4**: WebUI 增加 AI 管理页面。
5.  **Phase 5**: 引入本地 Ollama 自动发现与集成。

---
*请确认方案后，我将开始 Phase 1 的开发工作。*
