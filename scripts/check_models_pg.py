import psycopg2
from psycopg2.extras import RealDictCursor

def check_models():
    conn_str = "host=192.168.0.114 dbname=botmatrix user=derlin password=fkueiqiq461686"
    try:
        conn = psycopg2.connect(conn_str)
        cursor = conn.cursor(cursor_factory=RealDictCursor)
        
        print("Connected to Postgres botmatrix database")
        
        cursor.execute("SELECT id, name, provider_id FROM ai_models")
        models = cursor.fetchall()
        
        print("\nAvailable models in ai_models table:")
        for m in models:
            print(m)
            
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    check_models()
