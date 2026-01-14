
import psycopg2
from psycopg2.extras import RealDictCursor

conn = psycopg2.connect(
    host="192.168.0.114",
    database="botmatrix",
    user="derlin",
    password="fkueiqiq461686",
    port=5432
)

def check_billing():
    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        # Check wallets
        print("--- Wallets ---")
        cur.execute("SELECT * FROM ai_wallets")
        for row in cur.fetchall():
            print(row)
            
        # Check lease contracts
        print("\n--- Lease Contracts ---")
        cur.execute("SELECT * FROM ai_lease_contracts")
        for row in cur.fetchall():
            print(row)

        # Check for user 123456
        print("\n--- User 123456 Info ---")
        cur.execute("SELECT * FROM ai_wallets WHERE owner_id = 123456")
        wallet = cur.fetchone()
        if not wallet:
            print("User 123456 has no wallet. Creating one...")
            cur.execute("""
                INSERT INTO ai_wallets (owner_id, balance, currency, frozen_balance, total_spent, config)
                VALUES (123456, 1000.0, 'CNY', 0.0, 0.0, '{}')
            """)
            conn.commit()
            print("Wallet created with 1000.0 balance.")
        else:
            print(f"User 123456 already has a wallet: {wallet}")
            if wallet['balance'] < 100:
                print("Updating balance to 1000.0...")
                cur.execute("UPDATE ai_wallets SET balance = 1000.0 WHERE owner_id = 123456")
                conn.commit()
                print("Balance updated.")

        # Check lease resources
        print("\n--- Lease Resources ---")
        cur.execute("SELECT * FROM ai_lease_resources LIMIT 0")
        colnames = [desc[0] for desc in cur.description]
        print(f"Columns: {colnames}")
        
        cur.execute("SELECT * FROM ai_lease_resources")
        resources = cur.fetchall()
        for row in resources:
            print(row)

        type_col = "resource_type" if "resource_type" in colnames else "type"
        cur.execute(f"SELECT * FROM ai_lease_resources WHERE {type_col} = 'ai_service'")
        ai_resource = cur.fetchone()
        if not ai_resource:
            print("No ai_service resource found. Creating one...")
            if type_col == "resource_type":
                cur.execute("""
                    INSERT INTO ai_lease_resources (resource_type, name, provider_id, model_id, total_units, available_units, unit_price, unit_type, config, is_active)
                    VALUES ('ai_service', 'Default AI Service', 0, 'doubao', 100, 100, 0.0, 'hour', '{}', true)
                    RETURNING id
                """)
            else:
                cur.execute("""
                    INSERT INTO ai_lease_resources (type, name, provider_id, price_per_hour, unit_name, max_capacity, current_usage, status, config)
                    VALUES ('ai_service', 'Default AI Service', 0, 0.0, 'hour', 100, 0, 'available', '{}')
                    RETURNING id
                """)
            ai_resource_id = cur.fetchone()['id']
            conn.commit()
            print(f"ai_service resource created with ID {ai_resource_id}")
        else:
            ai_resource_id = ai_resource['id']

        # Ensure user 123456 has a lease
        cur.execute("SELECT * FROM ai_lease_contracts LIMIT 0")
        contract_cols = [desc[0] for desc in cur.description]
        
        status_col = "status" # Assuming status exists as it was in LeaseContract.cs
        cur.execute(f"SELECT * FROM ai_lease_contracts WHERE tenant_id = 123456 AND {status_col} = 'active'")
        lease = cur.fetchone()
        if not lease:
            print("User 123456 has no active lease. Creating one...")
            start_col = "start_at" if "start_at" in contract_cols else "start_time"
            end_col = "end_at" if "end_at" in contract_cols else "end_time"
            units_col = "units" if "units" in contract_cols else "total_paid" # This is a guess, let's use the actual names from the DB if possible
            
            # Re-read column names to be sure
            print(f"Contract columns: {contract_cols}")
            
            if "units" in contract_cols:
                cur.execute(f"""
                    INSERT INTO ai_lease_contracts (tenant_id, resource_id, units, {start_col}, {end_col}, status, config)
                    VALUES (123456, %s, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + interval '1 year', 'active', '{{}}')
                """, (ai_resource_id,))
            else:
                cur.execute(f"""
                    INSERT INTO ai_lease_contracts (tenant_id, resource_id, {start_col}, {end_col}, status, config)
                    VALUES (123456, %s, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + interval '1 year', 'active', '{{}}')
                """, (ai_resource_id,))
            conn.commit()
            print("Lease contract created.")
        else:
            print(f"User 123456 already has an active lease: {lease}")

if __name__ == "__main__":
    check_billing()
    conn.close()
