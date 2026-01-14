import psycopg2

POSTGRES_CONFIG = {
    "host": "192.168.0.114",
    "database": "botmatrix",
    "user": "derlin",
    "password": "fkueiqiq461686"
}

def list_tables():
    try:
        conn = psycopg2.connect(**POSTGRES_CONFIG)
        cur = conn.cursor()
        cur.execute("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
        for row in cur.fetchall():
            print(row[0])
        cur.close()
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    list_tables()
