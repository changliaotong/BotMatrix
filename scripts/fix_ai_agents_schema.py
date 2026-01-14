import psycopg2

def fix_table():
    conn_str = "host=192.168.0.114 dbname=botmatrix user=derlin password=fkueiqiq461686"
    try:
        conn = psycopg2.connect(conn_str)
        conn.autocommit = True
        cursor = conn.cursor()
        
        print("Connected to Postgres botmatrix database")
        
        # List of columns to add
        columns_to_add = [
            ("user_prompt_template", "TEXT"),
            ("tags", "JSONB DEFAULT '[]'"),
            ("config", "JSONB DEFAULT '{}'"),
            ("is_public", "BOOLEAN DEFAULT FALSE")
        ]
        
        for col_name, col_type in columns_to_add:
            try:
                cursor.execute(f"ALTER TABLE ai_agents ADD COLUMN {col_name} {col_type};")
                print(f"Added '{col_name}' column.")
            except Exception as e:
                if "already exists" in str(e):
                    print(f"'{col_name}' column already exists.")
                else:
                    print(f"Error adding '{col_name}': {e}")

        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    fix_table()
