import psycopg2
from psycopg2.extras import RealDictCursor
import json

POSTGRES_CONFIG = {
    'host': '192.168.0.114',
    'database': 'botmatrix',
    'user': 'derlin',
    'password': 'fkueiqiq461686'
}

def check_status():
    try:
        conn = psycopg2.connect(**POSTGRES_CONFIG)
        cur = conn.cursor(cursor_factory=RealDictCursor)

        print("--- Doubao Models ---")
        cur.execute("SELECT id, name, capabilities, provider_id FROM ai_models WHERE name ILIKE '%doubao%'")
        for row in cur.fetchall():
            print(row)

        print("\n--- Programmer Agent (1038) ---")
        cur.execute("SELECT id, name, model_id FROM ai_agents WHERE id = 1038")
        for row in cur.fetchall():
            print(row)

        print("\n--- Digital Employee (EMP-001) ---")
        cur.execute("SELECT id, employee_id, name, agent_id, job_id, online_status, state FROM ai_employee_instances WHERE employee_id = 'EMP-001'")
        for row in cur.fetchall():
            print(row)

        print("\n--- Recent Tasks ---")
        cur.execute("SELECT id, title, status, assignee_id, result_data, created_at FROM ai_task_records ORDER BY created_at DESC LIMIT 5")
        for row in cur.fetchall():
            print(row)

        cur.close()
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    check_status()
