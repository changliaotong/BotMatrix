import psycopg2

def check_dev_orchestrator():
    pg_conn_str = "host=192.168.0.114 dbname=botmatrix user=derlin password=fkueiqiq461686"
    try:
        conn = psycopg2.connect(pg_conn_str)
        cur = conn.cursor()
        
        cur.execute("SELECT job_key, tool_schema FROM ai_job_definitions WHERE job_key = 'dev_orchestrator'")
        row = cur.fetchone()
        print(f"Job: {row[0]}")
        print(f"ToolSchema: {row[1]}")
        
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    check_dev_orchestrator()
