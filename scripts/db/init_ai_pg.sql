-- BotMatrix 全新 AI 核心数据库架构 (PostgreSQL)
-- 彻底重构：语义化命名、snake_case、高性能索引、审计追踪
-- 适配 Go 语言 (GORM/Ent) 与 C# (Dapper)

-- =============================================================================
-- 1. 环境准备
-- =============================================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

-- 通用时间戳更新函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- =============================================================================
-- 2. AI 基础设施模块 (Infrastructure)
-- =============================================================================

-- AI 提供商
CREATE TABLE IF NOT EXISTS ai_providers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL, -- openai, azure, ollama, anthropic, deepseek
    endpoint VARCHAR(255),
    api_key TEXT,
    config JSONB DEFAULT '{}', -- 存储 organization_id, proxy 等
    is_active BOOLEAN DEFAULT TRUE,
    owner_id BIGINT DEFAULT 0, -- 0 表示系统全局，其它表示用户私有 Key
    is_shared BOOLEAN DEFAULT FALSE, -- 是否允许共享给他人使用
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- AI 模型
CREATE TABLE IF NOT EXISTS ai_models (
    id BIGSERIAL PRIMARY KEY,
    provider_id BIGINT REFERENCES ai_providers(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL, -- chat, image, embedding, audio
    context_window INT DEFAULT 4096,
    max_output_tokens INT,
    input_price_per_1k_tokens DECIMAL(12, 8) DEFAULT 0,
    output_price_per_1k_tokens DECIMAL(12, 8) DEFAULT 0,
    base_url VARCHAR(255),
    api_key TEXT,
    api_model_id VARCHAR(100),
    config JSONB DEFAULT '{}', -- 存储默认 temperature, top_p 等
    is_active BOOLEAN DEFAULT TRUE,
    is_paused BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, name)
);

-- =============================================================================
-- 3. 智能体与数字员工模块 (Agent & Evolution)
-- =============================================================================

-- 核心智能体定义
CREATE TABLE IF NOT EXISTS ai_agents (
    id BIGSERIAL PRIMARY KEY,
    guid UUID DEFAULT uuid_generate_v4() UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    system_prompt TEXT,
    user_prompt_template TEXT, -- 预定义的交互模板
    model_id BIGINT REFERENCES ai_models(id),
    tags JSONB DEFAULT '[]',
    config JSONB DEFAULT '{}', -- 存储运行时参数、技能列表等
    owner_id BIGINT, -- 关联创建者
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 岗位定义 (Evolution - Job)
CREATE TABLE IF NOT EXISTS ai_job_definitions (
    id BIGSERIAL PRIMARY KEY,
    job_key VARCHAR(100) NOT NULL UNIQUE, -- 如 software_engineer
    name VARCHAR(100) NOT NULL,
    purpose TEXT NOT NULL,
    inputs_schema JSONB DEFAULT '{}',
    outputs_schema JSONB DEFAULT '{}',
    constraints JSONB DEFAULT '[]', -- 岗位约束
    workflow JSONB DEFAULT '[]', -- 标准执行步骤 (Sequential, DAG)
    tool_schema JSONB DEFAULT '[]', -- 岗位可使用的工具列表 (存储 skill_key 数组)
    model_selection_strategy VARCHAR(50) DEFAULT 'random', -- 模型选择策略
    version INT DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 技能/工具定义 (Evolution - Skill)
CREATE TABLE IF NOT EXISTS ai_skill_definitions (
    id BIGSERIAL PRIMARY KEY,
    skill_key VARCHAR(100) NOT NULL UNIQUE, -- 如 file_read
    name VARCHAR(100) NOT NULL,
    description TEXT,
    action_name VARCHAR(50) NOT NULL, -- 对应 Agent 输出的 action
    parameter_schema JSONB DEFAULT '{}',
    is_builtin BOOLEAN DEFAULT TRUE,
    script_content TEXT, -- 动态脚本内容 (可选)
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 数字员工实例 (Evolution - Employee)
CREATE TABLE IF NOT EXISTS ai_employee_instances (
    id BIGSERIAL PRIMARY KEY,
    employee_id VARCHAR(64) UNIQUE, -- 工号
    bot_id VARCHAR(64) NOT NULL, -- 关联的机器人底层 ID
    agent_id BIGINT REFERENCES ai_agents(id),
    job_id BIGINT REFERENCES ai_job_definitions(id),
    name VARCHAR(100),
    title VARCHAR(100),
    department VARCHAR(100),
    online_status VARCHAR(20) DEFAULT 'offline', -- online, offline, busy
    state VARCHAR(20) DEFAULT 'idle', -- 当前业务状态
    salary_token_used BIGINT DEFAULT 0,
    salary_token_limit BIGINT DEFAULT 1000000,
    kpi_score DECIMAL(5, 2) DEFAULT 100.00,
    experience_data JSONB DEFAULT '{}', -- 存储进化的元数据
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- 4. 执行与日志模块 (Execution & Logs)
-- =============================================================================

-- 任务记录
CREATE TABLE IF NOT EXISTS ai_task_records (
    id BIGSERIAL PRIMARY KEY,
    execution_id UUID DEFAULT uuid_generate_v4() UNIQUE,
    title VARCHAR(255),
    description TEXT,
    initiator_id BIGINT, -- 发起者 ID
    assignee_id BIGINT REFERENCES ai_employee_instances(id),
    status VARCHAR(20) DEFAULT 'pending', -- pending, executing, completed, failed, cancelled
    progress INT DEFAULT 0,
    plan_data JSONB DEFAULT '{}', -- AI 生成的计划
    result_data JSONB DEFAULT '{}', -- 最终执行结果
    parent_task_id BIGINT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 任务执行步骤
CREATE TABLE IF NOT EXISTS ai_task_steps (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT REFERENCES ai_task_records(id) ON DELETE CASCADE,
    step_index INT NOT NULL,
    name VARCHAR(100),
    input_data JSONB,
    output_data JSONB,
    status VARCHAR(20),
    duration_ms INT,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 底层 LLM 调用审计日志
CREATE TABLE IF NOT EXISTS ai_llm_call_logs (
    id BIGSERIAL PRIMARY KEY,
    task_step_id BIGINT REFERENCES ai_task_steps(id),
    agent_id BIGINT,
    model_id BIGINT,
    prompt_tokens INT DEFAULT 0,
    completion_tokens INT DEFAULT 0,
    total_cost DECIMAL(15, 8) DEFAULT 0,
    latency_ms INT,
    is_success BOOLEAN DEFAULT TRUE,
    raw_request JSONB, -- 可选：存储请求元数据
    raw_response JSONB, -- 可选：存储响应元数据
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- 5. 知识库模块 (RAG)
-- =============================================================================

-- 知识库
CREATE TABLE IF NOT EXISTS ai_knowledge_bases (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    owner_id BIGINT,
    is_public BOOLEAN DEFAULT FALSE,
    config JSONB DEFAULT '{}', -- 存储分片策略等
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 知识切片
CREATE TABLE IF NOT EXISTS ai_knowledge_chunks (
    id BIGSERIAL PRIMARY KEY,
    kb_id BIGINT REFERENCES ai_knowledge_bases(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    embedding VECTOR(1536),
    metadata JSONB DEFAULT '{}', -- 存储原始文件引用、页码等
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 知识库文件
CREATE TABLE IF NOT EXISTS ai_knowledge_files (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    storage_path VARCHAR(512) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    upload_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    file_hash VARCHAR(64),
    is_embedded BOOLEAN DEFAULT FALSE,
    embedded_time TIMESTAMPTZ,
    embedding_error TEXT,
    user_id BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 工具调用审计日志
CREATE TABLE IF NOT EXISTS ai_tool_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    guid VARCHAR(64) UNIQUE NOT NULL,
    task_id VARCHAR(64),
    staff_id VARCHAR(64),
    tool_name VARCHAR(128),
    input_args TEXT,
    output_result TEXT,
    risk_level INT DEFAULT 0,
    status VARCHAR(32) DEFAULT 'Success',
    approved_by VARCHAR(64),
    rejection_reason TEXT,
    approved_at TIMESTAMPTZ,
    create_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- 6. 对话与记忆模块 (Conversation & Memory)
-- =============================================================================

-- 对话会话
CREATE TABLE IF NOT EXISTS ai_conversations (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(128) UNIQUE NOT NULL,
    user_id BIGINT,
    bot_id VARCHAR(64),
    title VARCHAR(255),
    summary TEXT, -- 长期记忆：会话摘要
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 对话消息
CREATE TABLE IF NOT EXISTS ai_messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT REFERENCES ai_conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL, -- system, user, assistant, tool
    content TEXT NOT NULL,
    tokens INT DEFAULT 0,
    tool_calls JSONB, -- 如果是 assistant 角色，记录工具调用
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- 7. 资源租赁与计费模块 (Resource Leasing & Billing)
-- =============================================================================

-- AI 账户/钱包
CREATE TABLE IF NOT EXISTS ai_wallets (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT UNIQUE NOT NULL, -- 关联用户或组织
    balance DECIMAL(18, 4) DEFAULT 0.0000, -- 余额
    currency VARCHAR(10) DEFAULT 'CNY',
    frozen_balance DECIMAL(18, 4) DEFAULT 0.0000, -- 冻结余额 (如租赁中)
    total_spent DECIMAL(18, 4) DEFAULT 0.0000,
    config JSONB DEFAULT '{}', -- 存储预警线、自动充值等配置
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 算力租赁资源项
CREATE TABLE IF NOT EXISTS ai_lease_resources (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- gpu_worker, agent_instance, employee_service
    description TEXT,
    provider_id BIGINT, -- 资源提供者 (可以是系统或用户)
    price_per_hour DECIMAL(18, 4) NOT NULL,
    unit_name VARCHAR(20) DEFAULT 'hour',
    max_capacity INT DEFAULT 1,
    current_usage INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'available', -- available, busy, maintenance
    config JSONB DEFAULT '{}', -- 存储规格 (VRAM, CUDA cores 等)
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 租赁合同/订单
CREATE TABLE IF NOT EXISTS ai_lease_contracts (
    id BIGSERIAL PRIMARY KEY,
    resource_id BIGINT REFERENCES ai_lease_resources(id),
    tenant_id BIGINT NOT NULL, -- 承租人
    start_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'active', -- active, completed, terminated
    auto_renew BOOLEAN DEFAULT FALSE,
    total_paid DECIMAL(18, 4) DEFAULT 0.0000,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 计费流水 (包括 Token 消耗和租赁费用)
CREATE TABLE IF NOT EXISTS ai_billing_transactions (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT REFERENCES ai_wallets(id),
    type VARCHAR(20) NOT NULL, -- consume, recharge, refund, lease_fee
    amount DECIMAL(18, 4) NOT NULL,
    related_id BIGINT, -- 关联任务 ID 或 租赁合同 ID
    related_type VARCHAR(50), -- task, lease, recharge
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 智能体订阅
CREATE TABLE IF NOT EXISTS ai_agent_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    agent_id BIGINT REFERENCES ai_agents(id) ON DELETE CASCADE,
    is_sub BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, agent_id)
);

-- 智能体标签
CREATE TABLE IF NOT EXISTS ai_agent_tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    owner_id BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 智能体与标签关联
CREATE TABLE IF NOT EXISTS ai_agent_tag_relations (
    agent_id BIGINT REFERENCES ai_agents(id) ON DELETE CASCADE,
    tag_id BIGINT REFERENCES ai_agent_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (agent_id, tag_id)
);

-- =============================================================================
-- 8. 辅助功能 (Indexes & Triggers)
-- =============================================================================

-- 触发器：自动更新时间戳
DO $$ 
DECLARE 
    t TEXT;
BEGIN
    FOR t IN (
        SELECT table_name FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name LIKE 'ai_%' 
    ) 
    LOOP
        -- 先尝试删除已存在的触发器
        EXECUTE format('DROP TRIGGER IF EXISTS update_at_trigger ON %I', t);
        -- 重新创建触发器
        EXECUTE format('CREATE TRIGGER update_at_trigger BEFORE UPDATE ON %I FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column()', t);
    END LOOP;
END $$;

-- 向量索引 (HNSW)
CREATE INDEX IF NOT EXISTS idx_ai_knowledge_chunks_embedding ON ai_knowledge_chunks USING hnsw (embedding vector_cosine_ops);

-- 业务索引
CREATE INDEX IF NOT EXISTS idx_ai_messages_conversation ON ai_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_ai_task_records_assignee ON ai_task_records(assignee_id);
CREATE INDEX IF NOT EXISTS idx_ai_employee_instances_bot ON ai_employee_instances(bot_id);
