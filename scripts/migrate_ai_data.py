import pyodbc
import psycopg2
from psycopg2.extras import execute_values
import datetime

# 数据库配置 (根据实际情况修改)
MSSQL_CONFIG = {
    "DRIVER": "{ODBC Driver 17 for SQL Server}",
    "SERVER": "192.168.0.114,1433",
    "DATABASE": "sz84_robot",
    "UID": "derlin",
    "PWD": "fkueiqiq461686"
}

PG_CONFIG = {
    "host": "192.168.0.114",
    "port": 5432,
    "user": "derlin",
    "password": "fkueiqiq461686",
    "dbname": "botmatrix"
}

def migrate():
    try:
        # 连接数据库
        mssql_conn = pyodbc.connect(
            f"DRIVER={MSSQL_CONFIG['DRIVER']};SERVER={MSSQL_CONFIG['SERVER']};DATABASE={MSSQL_CONFIG['DATABASE']};UID={MSSQL_CONFIG['UID']};PWD={MSSQL_CONFIG['PWD']}"
        )
        pg_conn = psycopg2.connect(**PG_CONFIG)
        pg_cursor = pg_conn.cursor()

        print("--- 开始迁移 AI 数据 ---")

        # 1. 迁移 ai_providers
        mssql_cursor = mssql_conn.cursor()
        mssql_cursor.execute("SELECT Id, Name, BaseUrl FROM LLMProvider")
        providers = mssql_cursor.fetchall()
        
        for p in providers:
            # 尝试获取该 Provider 的 API Key
            mssql_cursor.execute("SELECT TOP 1 ApiKey FROM LLMCredential WHERE ProviderId = ? AND IsActive = 1", (p.Id,))
            cred = mssql_cursor.fetchone()
            api_key = cred.ApiKey if cred else ''
            
            pg_cursor.execute(
                "INSERT INTO ai_providers (id, name, endpoint, type, api_key, created_at, updated_at) VALUES (%s, %s, %s, %s, %s, %s, %s) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, endpoint = EXCLUDED.endpoint, api_key = EXCLUDED.api_key",
                (p.Id, p.Name, p.BaseUrl, 'openai', api_key, datetime.datetime.now(), datetime.datetime.now())
            )

        # 2. 迁移 ai_models
        print("迁移 ai_models...")
        mssql_cursor.execute("SELECT Id, Name, ProviderId, ContextLength, SupportsVision, SupportsFunctionCalling, SupportsStreaming FROM LLMModel")
        models = mssql_cursor.fetchall()
        for m in models:
            # 构建能力列表
            capabilities = []
            if m.SupportsVision: capabilities.append("vision")
            if m.SupportsFunctionCalling: capabilities.append("tools")
            if m.SupportsStreaming: capabilities.append("stream")
            capabilities_json = ",".join(capabilities)
            
            # 确定模型类型
            model_type = "chat"
            if "embedding" in m.Name.lower():
                model_type = "embedding"
            elif "image" in m.Name.lower() or "dall-e" in m.Name.lower():
                model_type = "image"
            
            # 确定 BaseUrl (如果是豆包)
            base_url = None
            if "doubao" in m.Name.lower():
                base_url = "https://ark.cn-beijing.volces.com/api/v3"

            pg_cursor.execute(
                """INSERT INTO ai_models (
                    id, provider_id, name, api_model_id, context_window, capabilities, type, base_url, is_active, is_paused, created_at, updated_at
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s) 
                ON CONFLICT (id) DO UPDATE SET 
                    provider_id = EXCLUDED.provider_id,
                    name = EXCLUDED.name,
                    api_model_id = EXCLUDED.api_model_id,
                    context_window = EXCLUDED.context_window,
                    capabilities = EXCLUDED.capabilities,
                    type = EXCLUDED.type,
                    base_url = EXCLUDED.base_url,
                    updated_at = EXCLUDED.updated_at""",
                (
                    m.Id, m.ProviderId, m.Name, m.Name, 
                    m.ContextLength or 4096, capabilities_json,
                    model_type, base_url, True, False,
                    datetime.datetime.now(), datetime.datetime.now()
                )
            )
        pg_conn.commit()
        print(f"ai_models 迁移完成，共计 {len(models)} 条记录。")

        # 3. 迁移 ai_agents
        print("迁移 ai_agents (直接使用 UsedTimes 字段)...")
        
        mssql_cursor.execute("SELECT Id, Name, Info, Prompt, ModelId, IsVoice, VoiceId, VoiceLang, VoiceName, VoiceRate, Plugins, UserId, UsedTimes, InsertDate, Private FROM Agents")
        agents = mssql_cursor.fetchall()
        
        # 获取所有已迁移的模型 ID
        pg_cursor.execute("SELECT id FROM ai_models")
        valid_model_ids = {row[0] for row in pg_cursor.fetchall()}
        
        for a in agents:
            # 如果 model_id 不在 valid_model_ids 中，设为 None
            model_id = a.ModelId if a.ModelId in valid_model_ids else None
            
            # 映射可见性
            visibility = 'private' if a.Private == 1 else 'public'
            
            pg_cursor.execute(
                """INSERT INTO ai_agents (
                    id, name, description, system_prompt, model_id, 
                    temperature, max_tokens, is_voice, voice_id, voice_name, voice_lang, voice_rate, tools,
                    owner_id, call_count, visibility, created_at, updated_at
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s) 
                ON CONFLICT (id) DO UPDATE SET 
                    name = EXCLUDED.name, 
                    description = EXCLUDED.description, 
                    system_prompt = EXCLUDED.system_prompt, 
                    model_id = EXCLUDED.model_id,
                    is_voice = EXCLUDED.is_voice,
                    voice_id = EXCLUDED.voice_id,
                    voice_name = EXCLUDED.voice_name,
                    voice_lang = EXCLUDED.voice_lang,
                    voice_rate = EXCLUDED.voice_rate,
                    tools = EXCLUDED.tools,
                    owner_id = EXCLUDED.owner_id,
                    call_count = EXCLUDED.call_count,
                    visibility = EXCLUDED.visibility,
                    updated_at = EXCLUDED.updated_at""",
                (
                    a.Id, a.Name, a.Info, a.Prompt, model_id, 
                    0.7, 2048, a.IsVoice, a.VoiceId, a.VoiceName, a.VoiceLang, a.VoiceRate or 1.0, a.Plugins,
                    a.UserId, a.UsedTimes or 0, visibility, a.InsertDate or datetime.datetime.now(), a.InsertDate or datetime.datetime.now()
                )
            )
        pg_conn.commit()
        print(f"ai_agents 迁移完成，共计 {len(agents)} 条记录。")

        # 4. 迁移 ai_usage_logs (可选，如果数据量太大可以跳过)
        skip_logs = True # 默认跳过海量日志迁移
        if not skip_logs:
            # 分批迁移，每批 5000 条
            batch_size = 5000
            offset = 0
            total_migrated = 0
            
            while True:
                # 查询所有需要的字段
                mssql_cursor.execute(f"""
                    SELECT Id, Guid, GroupId, GroupName, UserId, UserName, InsertDate, MsgId, 
                           Messages, Question, Answer, AgentId, ModelId, 
                           TokensInput, TokensOutput, CostTime, Credit 
                    FROM AgentLog 
                    ORDER BY InsertDate DESC 
                    OFFSET {offset} ROWS FETCH NEXT {batch_size} ROWS ONLY
                """)
                logs = mssql_cursor.fetchall()
                if not logs:
                    break
                    
                for l in logs:
                    status = 'success' if l.Answer else 'failed'
                    # 尝试根据 ModelId 获取模型名称
                    mssql_cursor.execute("SELECT Name FROM LLMModel WHERE Id = ?", (l.ModelId,))
                    m_res = mssql_cursor.fetchone()
                    model_name = m_res.Name if m_res else str(l.ModelId)
                    
                    # 清洗 NUL 字符
                    def clean_nul(val):
                        if isinstance(val, str):
                            return val.replace('\x00', '')
                        return val

                    pg_cursor.execute(
                        """INSERT INTO ai_usage_logs (
                            user_id, agent_id, model_name, input_tokens, output_tokens, 
                            duration_ms, status, error_message, 
                            guid, group_id, group_name, user_name, msg_id, 
                            question, answer, messages, credit, created_at
                        ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)""",
                        (
                            l.UserId, l.AgentId, clean_nul(model_name), l.TokensInput or 0, l.TokensOutput or 0, 
                            int(float(l.CostTime or 0) * 1000), status, '', 
                            clean_nul(l.Guid), clean_nul(l.GroupId), clean_nul(l.GroupName), 
                            clean_nul(l.UserName), clean_nul(l.MsgId), 
                            clean_nul(l.Question), clean_nul(l.Answer), clean_nul(l.Messages), 
                            float(l.Credit or 0), l.InsertDate
                        )
                    )
                
                pg_conn.commit()
                total_migrated += len(logs)
                offset += batch_size
                print(f"已迁移 {total_migrated} 条记录...")

            print(f"全量迁移完成，共计 {total_migrated} 条记录。")
        print("--- 迁移完成 ---")

    except Exception as e:
        print(f"迁移失败: {e}")
        if pg_conn:
            pg_conn.rollback()
    finally:
        if mssql_conn: mssql_conn.close()
        if pg_conn: pg_conn.close()

if __name__ == "__main__":
    migrate()
