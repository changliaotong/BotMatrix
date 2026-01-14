import psycopg2

def check_skill_keys():
    pg_conn_str = "host=192.168.0.114 dbname=botmatrix user=derlin password=fkueiqiq461686"
    try:
        conn = psycopg2.connect(pg_conn_str)
        cur = conn.cursor()
        
        cur.execute("SELECT skill_key, action_name FROM ai_skill_definitions")
        rows = cur.fetchall()
        for row in rows:
            print(f"Key: {row[0]}, Action: {row[1]}")
        
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    check_skill_keys()
