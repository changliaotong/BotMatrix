import pyodbc
import psycopg2
from psycopg2.extras import RealDictCursor
import json
import logging

# 配置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# 数据库连接配置
SQL_SERVER_CONFIG = {
    'driver': '{SQL Server}',
    'server': '192.168.0.114',
    'database': 'sz84_robot',
    'uid': 'derlin',
    'password': 'fkueiqiq461686'
}

POSTGRES_CONFIG = {
    'host': '192.168.0.114',
    'database': 'botmatrix',
    'user': 'derlin',
    'password': 'fkueiqiq461686'
}

def get_sql_server_conn():
    conn_str = f"DRIVER={SQL_SERVER_CONFIG['driver']};SERVER={SQL_SERVER_CONFIG['server']};DATABASE={SQL_SERVER_CONFIG['database']};UID={SQL_SERVER_CONFIG['uid']};PWD={SQL_SERVER_CONFIG['password']};Encrypt=no;"
    return pyodbc.connect(conn_str)

def get_postgres_conn():
    return psycopg2.connect(**POSTGRES_CONFIG)

def import_data():
    ss_conn = None
    pg_conn = None
    try:
        logger.info("Connecting to databases...")
        ss_conn = get_sql_server_conn()
        pg_conn = get_postgres_conn()
        ss_cursor = ss_conn.cursor()
        pg_cursor = pg_conn.cursor()

        # 1. 导入 AIProvider (LLMProvider in SQL Server)
        logger.info("Importing LLM Providers...")
        ss_cursor.execute("SELECT Id, Name, ProviderType, BaseUrl, APIKey, Status, CreateAt, UpdateAt FROM LLMProvider")
        providers = ss_cursor.fetchall()
        
        for p in providers:
            p_id, p_name, p_type, p_base_url, p_api_key, p_status, p_created, p_updated = p
            
            is_enabled = (p_status == 1 or str(p_status).lower() == "active")
            
            pg_cursor.execute("""
                INSERT INTO ai_providers (id, name, type, base_url, api_key, is_enabled, created_at, updated_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    name = EXCLUDED.name,
                    type = EXCLUDED.type,
                    base_url = EXCLUDED.base_url,
                    api_key = EXCLUDED.api_key,
                    is_enabled = EXCLUDED.is_enabled,
                    updated_at = EXCLUDED.updated_at
            """, (p_id, p_name, p_type.lower() if p_type else 'openai', p_base_url, p_api_key, is_enabled, p_created, p_updated))

        logger.info(f"Successfully imported/updated {len(providers)} providers.")

        # 2. 导入 AIModel (LLMModel in SQL Server)
        logger.info("Importing LLM Models...")
        ss_cursor.execute("SELECT Id, ProviderId, Name, ContextLength, SupportsVision, Status, ModelType FROM LLMModel")
        models = ss_cursor.fetchall()

        for m in models:
            m_id, m_provider_id, m_name, m_context, m_vision, m_status, m_type_val = m
            
            # 确定模型类型/能力
            # SQL Server ModelType: 0 通常是 Chat, 1 通常是 Image
            # 此外通过名称关键词识别 Embedding
            model_type = "chat"
            if m_type_val == 1:
                model_type = "image"
            elif "embed" in m_name.lower():
                model_type = "embedding"
            
            if m_vision and model_type == "chat":
                model_type = "chat,vision"
            
            is_default = (m_status == 1 or str(m_status).lower() == "active")

            pg_cursor.execute("""
                INSERT INTO ai_models (id, provider_id, model_name, display_name, api_model_id, capabilities, context_size, is_default)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    provider_id = EXCLUDED.provider_id,
                    model_name = EXCLUDED.model_name,
                    display_name = EXCLUDED.display_name,
                    api_model_id = EXCLUDED.api_model_id,
                    capabilities = EXCLUDED.capabilities,
                    context_size = EXCLUDED.context_size,
                    is_default = EXCLUDED.is_default,
                    updated_at = CURRENT_TIMESTAMP
            """, (m_id, m_provider_id, m_name, m_name, m_name, model_type, m_context, is_default))

        logger.info(f"Successfully imported/updated {len(models)} models.")

        pg_conn.commit()
        logger.info("Migration completed successfully.")

    except Exception as e:
        if pg_conn:
            pg_conn.rollback()
        logger.error(f"Migration failed: {e}")
    finally:
        if ss_conn:
            ss_conn.close()
        if pg_conn:
            pg_conn.close()

if __name__ == "__main__":
    import_data()
