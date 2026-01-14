import psycopg2

def update_strategies():
    conn_str = "host=192.168.0.114 dbname=botmatrix user=derlin password=fkueiqiq461686"
    try:
        conn = psycopg2.connect(conn_str)
        cur = conn.cursor()
        cur.execute("UPDATE ai_job_definitions SET model_selection_strategy = 'random' WHERE model_selection_strategy = 'specified'")
        print(f"Updated {cur.rowcount} rows.")
        conn.commit()
        cur.close()
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    update_strategies()
