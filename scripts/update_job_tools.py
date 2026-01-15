import psycopg2
import json

def update_job():
    try:
        conn = psycopg2.connect("host=192.168.0.114 port=5432 dbname=botmatrix user=derlin password=fkueiqiq461686")
        cur = conn.cursor()
        
        job_key = 'dev_orchestrator'
        cur.execute("SELECT tool_schema FROM ai_job_definitions WHERE job_key = %s", (job_key,))
        row = cur.fetchone()
        
        if row:
            raw_tools = row[0]
            if isinstance(raw_tools, str):
                tools = json.loads(raw_tools)
            else:
                tools = raw_tools
            
            print(f"Current tools for {job_key}: {tools}")
            
            # 确保包含新技能
            new_tools = ["LIST", "READ", "WRITE", "BUILD", "GIT", "COMMAND", "PLAN", "REVIEW"]
            updated_tools = list(set(tools + new_tools))
            
            if len(updated_tools) > len(tools):
                cur.execute("UPDATE ai_job_definitions SET tool_schema = %s WHERE job_key = %s", 
                           (json.dumps(updated_tools), job_key))
                conn.commit()
                print(f"Updated tools for {job_key}: {updated_tools}")
            else:
                print("No updates needed.")
        else:
            print(f"Job {job_key} not found.")
            
        cur.close()
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    update_job()
