import psycopg2

def fix_table():
    conn_str = "host=192.168.0.114 dbname=botmatrix user=derlin password=fkueiqiq461686"
    try:
        conn = psycopg2.connect(conn_str)
        conn.autocommit = True
        cursor = conn.cursor()
        
        print("Connected to Postgres botmatrix database")
        
        # Check if uuid-ossp extension exists
        cursor.execute("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
        print("Ensured uuid-ossp extension exists.")
        
        # Add guid column if it doesn't exist
        try:
            cursor.execute("ALTER TABLE ai_agents ADD COLUMN guid UUID DEFAULT uuid_generate_v4() UNIQUE;")
            print("Added 'guid' column to 'ai_agents' table.")
            
            # Backfill guid for existing rows
            cursor.execute("UPDATE ai_agents SET guid = uuid_generate_v4() WHERE guid IS NULL;")
            print("Backfilled 'guid' values.")
        except Exception as e:
            if "already exists" in str(e):
                print("'guid' column already exists.")
            else:
                raise e

        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    fix_table()
