import psycopg2
from psycopg2.extras import RealDictCursor

def check_agents():
    conn_str = "host=192.168.0.114 dbname=botmatrix user=derlin password=fkueiqiq461686"
    try:
        conn = psycopg2.connect(conn_str)
        cursor = conn.cursor(cursor_factory=RealDictCursor)
        
        print("Connected to Postgres botmatrix database")
        
        # Check for dev_orchestrator
        cursor.execute("SELECT * FROM ai_agents WHERE name = %s", ('dev_orchestrator',))
        rows = cursor.fetchall()
        
        if not rows:
            print("No agent found with name 'dev_orchestrator'")
            
            # List all agents
            print("\nAll agents in ai_agents table:")
            cursor.execute("SELECT id, guid, name, is_public FROM ai_agents")
            for row in cursor.fetchall():
                print(row)
        else:
            print(f"Found {len(rows)} agent(s):")
            for row in rows:
                print(row)
                
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    check_agents()
