import psycopg2
from psycopg2.extras import RealDictCursor

def describe_table():
    conn_str = "host=192.168.0.114 dbname=botmatrix user=derlin password=fkueiqiq461686"
    try:
        conn = psycopg2.connect(conn_str)
        cursor = conn.cursor(cursor_factory=RealDictCursor)
        
        print("Connected to Postgres botmatrix database")
        
        # Get column info for ai_agents
        print("\nStructure of 'ai_agents' table:")
        cursor.execute("""
            SELECT column_name, data_type, is_nullable
            FROM information_schema.columns
            WHERE table_name = 'ai_agents'
            ORDER BY ordinal_position;
        """)
        columns = cursor.fetchall()
        for col in columns:
            print(f"{col['column_name']}: {col['data_type']} (Nullable: {col['is_nullable']})")
            
        if not any(col['column_name'] == 'guid' for col in columns):
            print("\nCRITICAL: 'guid' column is MISSING in 'ai_agents' table!")
        else:
            print("\n'guid' column exists.")

        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    describe_table()
