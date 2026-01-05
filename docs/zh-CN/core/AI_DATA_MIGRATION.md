# AI 数据迁移指南 (SQL Server to PostgreSQL)

本指南详细说明了如何将旧版 `sz84` 系统（SQL Server）中的 AI 相关数据迁移到新的 `BotMatrix` 系统（PostgreSQL）中。

## 1. 核心表映射关系

| 源表 (SQL Server) | 目标表 (PostgreSQL) | 关键字段映射 | 说明 |
| :--- | :--- | :--- | :--- |
| `LLMProvider` | `ai_providers` | `Name` -> `name`<br>`BaseUrl` -> `api_base` | AI 供应商信息 |
| `LLMModel` | `ai_models` | `Name` -> `model_name`<br>`ProviderId` -> `provider_id` | 模型定义 |
| `Agents` | `ai_agents` | `Name` -> `name`<br>`Prompt` -> `system_prompt`<br>`Info` -> `description` | 智能体核心配置 |
| `LLMCallLog` | `ai_usage_logs` | `AgentId` -> `agent_id`<br>`InputTokens` -> `prompt_tokens` | 使用记录与统计 |
| `UserMessage` | `ai_chat_messages` | `Content` -> `content`<br>`Role` -> `role` | 历史对话记录 |

## 2. 迁移准备

在执行迁移前，请确保：
1. `BotNexus` 已经成功启动，且 `PostgreSQL` 中的表结构已通过自动迁移生成。
2. 拥有 SQL Server 的读取权限。
3. 拥有 PostgreSQL 的写入权限。

## 3. 迁移工具

我们提供了一个专门的 Python 脚本用于自动化迁移核心 AI 数据：
- 脚本路径：`scripts/migrate_ai_data.py`

### 使用方法：
1. 安装依赖：`pip install pyodbc psycopg2`
2. 修改脚本顶部的 `MSSQL_CONFIG` 和 `PG_CONFIG` 为你的实际数据库信息。
3. 运行脚本：`python scripts/migrate_ai_data.py`

## 4. SQL 迁移脚本模板 (手动迁移)

### 4.1 供应商数据迁移 (ai_providers)
```sql
-- SQL Server 导出逻辑
SELECT 
    Id, 
    Name, 
    BaseUrl AS ApiBase, 
    'openai' AS ProviderType, -- 根据实际情况调整
    '' AS ApiKey -- 需手动补充或加密迁移
FROM LLMProvider;
```

### 4.2 智能体数据迁移 (ai_agents)
```sql
-- SQL Server 导出逻辑
SELECT 
    Id, 
    Name, 
    Info AS Description, 
    Prompt AS SystemPrompt, 
    ModelId AS ModelID,
    0.7 AS Temperature,
    2048 AS MaxTokens,
    IsVoice,
    VoiceId
FROM Agents;
```

### 4.3 模型数据迁移 (ai_models)
```sql
-- SQL Server 导出逻辑
SELECT 
    Id, 
    Name AS ModelName, 
    ProviderId AS ProviderID,
    128000 AS MaxContext -- 默认值
FROM LLMModel;
```

## 5. 迁移注意事项

1. **ID 冲突**: PostgreSQL 使用 `serial/bigserial` 作为自增主键。迁移时如果保留原始 ID，请记得在迁移后更新序列：
   ```sql
   SELECT setval('ai_agents_id_seq', (SELECT MAX(id) FROM ai_agents));
   ```
2. **API Key 安全**: 由于 `LLMProvider` 中可能未存储 `ApiKey`（旧版可能在配置文件中），请在迁移后通过管理后台手动补充 `ai_providers` 表中的 `api_key` 字段。
3. **数据清洗**: 
   - `Agents.Prompt` 中的占位符（如 `{客服QQ}`）可能需要批量替换为新系统的变量格式。
   - 确保 `digital_employees` 表也根据 `Agents` 表进行相应的初始化。

## 6. 验证迁移

执行以下查询确认数据已正确载入：
```sql
SELECT a.name, m.model_name, p.name as provider 
FROM ai_agents a
JOIN ai_models m ON a.model_id = m.id
JOIN ai_providers p ON m.provider_id = p.id;
```
