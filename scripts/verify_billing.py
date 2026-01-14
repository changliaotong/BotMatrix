import psycopg2

POSTGRES_CONFIG = {
    "host": "192.168.0.114",
    "database": "botmatrix",
    "user": "derlin",
    "password": "fkueiqiq461686"
}

def verify_billing():
    conn = psycopg2.connect(**POSTGRES_CONFIG)
    cur = conn.cursor()
    
    user_id = 123456
    
    # Check wallet
    cur.execute("SELECT id, balance, owner_id FROM ai_wallets WHERE owner_id = %s", (user_id,))
    wallet = cur.fetchone()
    if wallet:
        print(f"Wallet found: ID={wallet[0]}, Balance={wallet[1]}, Owner={wallet[2]}")
    else:
        print(f"No wallet found for user {user_id}. Creating one...")
        cur.execute("INSERT INTO ai_wallets (owner_id, balance, currency, config) VALUES (%s, 1000.0, 'CNY', '{}') RETURNING id", (user_id,))
        new_wallet_id = cur.fetchone()[0]
        print(f"Created wallet ID={new_wallet_id}")
    
    # Check resources
    cur.execute("SELECT id, name, type, is_active FROM ai_lease_resources WHERE type = 'ai_service'")
    resources = cur.fetchall()
    print(f"AI Service resources: {resources}")
    
    if not resources:
        print("No ai_service resource found. Creating one...")
        cur.execute("""
            INSERT INTO ai_lease_resources (name, type, description, provider_id, price_per_hour, unit_name, max_capacity, status, is_active, config)
            VALUES ('Default AI Service', 'ai_service', 'System default AI service', 0, 0.0, 'hour', 999, 'available', true, '{}')
            RETURNING id
        """)
        resource_id = cur.fetchone()[0]
        print(f"Created resource ID={resource_id}")
    else:
        resource_id = resources[0][0]
    
    # Check contracts
    cur.execute("SELECT id, status, end_time FROM ai_lease_contracts WHERE tenant_id = %s AND resource_id = %s", (user_id, resource_id))
    contract = cur.fetchone()
    if contract:
        print(f"Contract found: ID={contract[0]}, Status={contract[1]}, EndTime={contract[2]}")
    else:
        print(f"No contract found for user {user_id}. Creating one...")
        cur.execute("""
            INSERT INTO ai_lease_contracts (tenant_id, resource_id, start_time, end_time, status, auto_renew, total_paid, config)
            VALUES (%s, %s, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + interval '1 year', 'active', true, 0.0, '{}')
        """, (user_id, resource_id))
        print("Created active contract for 1 year.")
    
    conn.commit()
    cur.close()
    conn.close()

if __name__ == "__main__":
    verify_billing()
