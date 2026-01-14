import psycopg2
from psycopg2.extras import RealDictCursor

conn_str = "host=192.168.0.114 dbname=sz84_robot user=derlin password=fkueiqiq461686"

def check_cmd():
    try:
        conn = psycopg2.connect(conn_str)
        cur = conn.cursor(cursor_factory=RealDictCursor)
        cur.execute('SELECT * FROM "Cmd" WHERE "CmdName" = %s', ('岗位任务',))
        row = cur.fetchone()
        if row:
            print("Found Cmd:")
            for k, v in row.items():
                print(f"{k}: {v}")
        else:
            print("Cmd '岗位任务' not found.")
        cur.close()
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    check_cmd()
