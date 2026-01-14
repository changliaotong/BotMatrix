import psycopg2
from psycopg2.extras import RealDictCursor

def check_and_update_doubao():
    try:
        conn = psycopg2.connect(
            dbname='botmatrix',
            user='derlin',
            password='fkueiqiq461686',
            host='192.168.0.114'
        )
        cur = conn.cursor(cursor_factory=RealDictCursor)

        # 1. 检查 Provider
        cur.execute("SELECT id, name FROM ai_providers WHERE name ILIKE '%doubao%' OR type = 'volcengine'")
        providers = cur.fetchall()
        print(f"Found providers: {providers}")

        # 2. 检查 ai_models 的列名
        cur.execute("SELECT column_name FROM information_schema.columns WHERE table_name = 'ai_models'")
        columns = [row['column_name'] for row in cur.fetchall()]
        print(f"ai_models columns: {columns}")

        # 3. 检查 Models
        cur.execute("SELECT id, model_name, capabilities FROM ai_models WHERE model_name ILIKE '%doubao%'")
        models = cur.fetchall()
        print(f"Found models: {models}")

        # 4. 如果没有 1.8 模型，则尝试添加
        doubao_1_8_models = [
            {"model_name": "Doubao-1.8-pro", "capabilities": "chat,vision"},
            {"model_name": "Doubao-1.8-lite", "capabilities": "chat,vision"}
        ]

        for m in doubao_1_8_models:
            exists = any(m['model_name'].lower() in (existing['model_name'] or '').lower() for existing in models)
            if not exists:
                print(f"Adding model: {m['model_name']}")
                p_id = providers[0]['id'] if providers else 1
                cur.execute(
                    "INSERT INTO ai_models (provider_id, model_name, display_name, capabilities) VALUES (%s, %s, %s, %s)",
                    (p_id, m['model_name'], m['model_name'], m['capabilities'])
                )
        
        conn.commit()
        cur.close()
        conn.close()
        print("Done.")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    check_and_update_doubao()
