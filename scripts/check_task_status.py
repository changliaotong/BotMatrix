import psycopg2
import sys
from psycopg2.extras import RealDictCursor

conn_str = "host=192.168.0.114 dbname=botmatrix user=derlin password=fkueiqiq461686"

def check_tasks(task_id=None):
    try:
        conn = psycopg2.connect(conn_str)
        cur = conn.cursor(cursor_factory=RealDictCursor)
        
        if task_id:
            print(f"--- Details for Task {task_id} ---")
            cur.execute("SELECT * FROM ai_task_records WHERE id = %s;", (task_id,))
            row = cur.fetchone()
            if row:
                for k, v in row.items():
                    print(f"{k}: {v}")
            
            print(f"\n--- Steps for Task {task_id} ---")
            cur.execute("SELECT * FROM ai_task_steps WHERE task_id = %s ORDER BY step_index ASC;", (task_id,))
            steps = cur.fetchall()
            for step in steps:
                print(f"Step {step['step_index']}: {step['name']} ({step['status']})")
                if step['input_data']:
                    print(f"  Input: {step['input_data'][:100]}...")
                if step['output_data']:
                    print(f"  Output: {step['output_data'][:200]}...")
                if step['error_message']:
                    print(f"  Error: {step['error_message']}")
        else:
            print("--- Latest 5 Task Records ---")
            cur.execute("SELECT id, title, status, progress, started_at, created_at FROM ai_task_records ORDER BY created_at DESC LIMIT 5;")
            rows = cur.fetchall()
            for row in rows:
                print(f"ID: {row['id']}, Title: {row['title']}, Status: {row['status']}, Progress: {row['progress']}%, Created: {row['created_at']}")
            
            print("\n--- Latest 10 Task Steps ---")
            cur.execute("SELECT task_id, step_index, name, status, error_message, created_at FROM ai_task_steps ORDER BY created_at DESC LIMIT 10;")
            steps = cur.fetchall()
            for step in steps:
                print(f"Task: {step['task_id']}, Step {step['step_index']}, Name: {step['name']}, Status: {step['status']}, Created: {step['created_at']}")
                if step['error_message']:
                    print(f"  Error: {step['error_message']}")
            
        cur.close()
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    t_id = None
    if len(sys.argv) > 1:
        t_id = sys.argv[1]
    check_tasks(t_id)
