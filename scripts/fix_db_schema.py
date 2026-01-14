import psycopg2

def add_column():
    conn_params = {
        "host": "192.168.0.114",
        "database": "botmatrix",
        "user": "derlin",
        "password": "fkueiqiq461686"
    }
    
    try:
        conn = psycopg2.connect(**conn_params)
        cur = conn.cursor()
        
        # Check if is_shared column exists
        cur.execute("""
            SELECT column_name 
            FROM information_schema.columns 
            WHERE table_name='ai_providers' AND column_name='is_shared'
        """)
        
        if not cur.fetchone():
            print("Adding column 'is_shared' to 'ai_providers' table...")
            cur.execute("ALTER TABLE ai_providers ADD COLUMN is_shared BOOLEAN DEFAULT FALSE")
            conn.commit()
            print("Column 'is_shared' added successfully.")
        else:
            print("Column 'is_shared' already exists.")

        # Rename columns in ai_providers to match C# models
        columns_to_rename = {
            'base_url': 'endpoint',
            'is_enabled': 'is_active'
        }
        
        for old_name, new_name in columns_to_rename.items():
            cur.execute(f"""
                SELECT column_name 
                FROM information_schema.columns 
                WHERE table_name='ai_providers' AND column_name='{old_name}'
            """)
            if cur.fetchone():
                print(f"Renaming column '{old_name}' to '{new_name}' in 'ai_providers'...")
                cur.execute(f"ALTER TABLE ai_providers RENAME COLUMN {old_name} TO {new_name}")
                conn.commit()

        # Add missing columns to ai_providers
        cur.execute("ALTER TABLE ai_providers ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'")
        cur.execute("ALTER TABLE ai_providers ADD COLUMN IF NOT EXISTS owner_id BIGINT DEFAULT 0")
        conn.commit()

        # Fix ai_models columns
        # Check current columns in ai_models
        cur.execute("SELECT column_name FROM information_schema.columns WHERE table_name='ai_models'")
        current_cols = [row[0] for row in cur.fetchall()]
        print(f"Current columns in ai_models: {current_cols}")

        # Rename columns in ai_models
        model_renames = {
            'model_name': 'name',
            'is_default': 'is_active'
        }
        for old_name, new_name in model_renames.items():
            if old_name in current_cols and new_name not in current_cols:
                print(f"Renaming column '{old_name}' to '{new_name}' in 'ai_models'...")
                cur.execute(f"ALTER TABLE ai_models RENAME COLUMN {old_name} TO {new_name}")
                conn.commit()

        # Add missing columns to ai_models
        # context_window, max_output_tokens, input_price_per_1k_tokens, output_price_per_1k_tokens, config
        cur.execute("ALTER TABLE ai_models ADD COLUMN IF NOT EXISTS context_window INTEGER DEFAULT 0")
        cur.execute("ALTER TABLE ai_models ADD COLUMN IF NOT EXISTS max_output_tokens INTEGER DEFAULT 0")
        cur.execute("ALTER TABLE ai_models ADD COLUMN IF NOT EXISTS input_price_per_1k_tokens DECIMAL(10,4) DEFAULT 0")
        cur.execute("ALTER TABLE ai_models ADD COLUMN IF NOT EXISTS output_price_per_1k_tokens DECIMAL(10,4) DEFAULT 0")
        cur.execute("ALTER TABLE ai_models ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'")
        conn.commit()
            
        # Fix ai_job_definitions table
        cur.execute("SELECT column_name FROM information_schema.columns WHERE table_name='ai_job_definitions'")
        job_cols = [row[0] for row in cur.fetchall()]
        print(f"Current columns in ai_job_definitions: {job_cols}")
        
        # Add missing columns to ai_job_definitions
        cur.execute("ALTER TABLE ai_job_definitions ADD COLUMN IF NOT EXISTS system_prompt TEXT DEFAULT ''")
        cur.execute("ALTER TABLE ai_job_definitions ADD COLUMN IF NOT EXISTS tool_schema JSONB DEFAULT '[]'")
        cur.execute("ALTER TABLE ai_job_definitions ADD COLUMN IF NOT EXISTS workflow JSONB DEFAULT '[]'")
        cur.execute("ALTER TABLE ai_job_definitions ADD COLUMN IF NOT EXISTS model_selection_strategy VARCHAR(50) DEFAULT 'random'")
        cur.execute("ALTER TABLE ai_job_definitions ADD COLUMN IF NOT EXISTS inputs_schema JSONB DEFAULT '{}'")
        cur.execute("ALTER TABLE ai_job_definitions ADD COLUMN IF NOT EXISTS outputs_schema JSONB DEFAULT '{}'")
        cur.execute("ALTER TABLE ai_job_definitions ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1")
        conn.commit()

        cur.close()
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    add_column()
