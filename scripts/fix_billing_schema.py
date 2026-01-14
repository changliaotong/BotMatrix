import psycopg2
from psycopg2.extras import RealDictCursor

POSTGRES_CONFIG = {
    "host": "192.168.0.114",
    "database": "botmatrix",
    "user": "derlin",
    "password": "fkueiqiq461686"
}

def fix_schema():
    try:
        conn = psycopg2.connect(**POSTGRES_CONFIG)
        cur = conn.cursor()

        # 1. ai_lease_resources
        print("Checking ai_lease_resources...")
        cur.execute("SELECT column_name FROM information_schema.columns WHERE table_name = 'ai_lease_resources'")
        cols = [r[0] for r in cur.fetchall()]
        
        if 'is_active' not in cols:
            print("Adding is_active to ai_lease_resources...")
            cur.execute("ALTER TABLE ai_lease_resources ADD COLUMN is_active BOOLEAN DEFAULT TRUE")
        
        # Ensure 'type' exists (it does, but let's be safe)
        if 'type' not in cols and 'resource_type' in cols:
            print("Renaming resource_type to type in ai_lease_resources...")
            cur.execute("ALTER TABLE ai_lease_resources RENAME COLUMN resource_type TO type")

        # 2. ai_lease_contracts
        print("Checking ai_lease_contracts...")
        cur.execute("SELECT column_name FROM information_schema.columns WHERE table_name = 'ai_lease_contracts'")
        cols = [r[0] for r in cur.fetchall()]
        
        if 'start_time' not in cols and 'begin_time' in cols:
            print("Renaming begin_time to start_time in ai_lease_contracts...")
            cur.execute("ALTER TABLE ai_lease_contracts RENAME COLUMN begin_time TO start_time")

        # 3. ai_wallets
        print("Checking ai_wallets...")
        cur.execute("SELECT column_name FROM information_schema.columns WHERE table_name = 'ai_wallets'")
        cols = [r[0] for r in cur.fetchall()]
        
        # Add any missing common columns if needed
        
        conn.commit()
        print("Schema fixed successfully.")
        
        # Update existing resources to be active
        cur.execute("UPDATE ai_lease_resources SET is_active = TRUE")
        conn.commit()
        print("Updated all resources to is_active = TRUE")

        cur.close()
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    fix_schema()
