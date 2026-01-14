import psycopg2
import json

def check_columns():
    POSTGRES_CONFIG = {
        "host": "192.168.0.114",
        "database": "botmatrix",
        "user": "derlin",
        "password": "fkueiqiq461686"
    }
    conn = psycopg2.connect(**POSTGRES_CONFIG)
    cur = conn.cursor()
    
    tables = ['ai_lease_resources', 'ai_lease_contracts', 'knowledge_chunks']
    for table in tables:
        cur.execute(f"SELECT column_name FROM information_schema.columns WHERE table_name = '{table}'")
        columns = [row[0] for row in cur.fetchall()]
        print(f"{table}: {columns}")
    
    cur.close()
    conn.close()

if __name__ == "__main__":
    check_columns()
