import psycopg2

def reset_sequence():
    conn_str = "host=192.168.0.114 dbname=botmatrix user=derlin password=fkueiqiq461686"
    try:
        conn = psycopg2.connect(conn_str)
        conn.autocommit = True
        cursor = conn.cursor()
        
        print("Connected to Postgres botmatrix database")
        
        # Reset ai_agents_id_seq
        cursor.execute("SELECT MAX(id) FROM ai_agents")
        max_id = cursor.fetchone()[0] or 0
        cursor.execute(f"SELECT setval('ai_agents_id_seq', {max_id + 1}, false)")
        print(f"Reset 'ai_agents_id_seq' to {max_id + 1}")
        
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    reset_sequence()
