import psycopg2
import pymssql
import uuid

def setup():
    # Postgres setup
    pg_conn_str = "host=192.168.0.114 dbname=botmatrix user=derlin password=fkueiqiq461686"
    try:
        pg_conn = psycopg2.connect(pg_conn_str)
        pg_conn.autocommit = True
        pg_cursor = pg_conn.cursor()
        
        print("Connected to Postgres botmatrix database")
        
        # Check if dev_orchestrator exists
        pg_cursor.execute("SELECT id FROM ai_agents WHERE name = 'dev_orchestrator'")
        if pg_cursor.fetchone():
            print("'dev_orchestrator' already exists in Postgres.")
        else:
            agent_guid = str(uuid.uuid4())
            pg_cursor.execute("""
                INSERT INTO ai_agents (
                    guid, name, description, system_prompt, model_id, is_public
                ) VALUES (%s, %s, %s, %s, %s, %s)
            """, (
                agent_guid, 
                'dev_orchestrator', 
                'Autonomous software developer agent', 
                'You are an autonomous senior software developer. Your goal is to plan, implement, and test software based on user requirements. Use the available tools to explore the codebase, write code, and run tests. Follow the Manus protocol: plan first, execute incrementally, and verify results.',
                9, # doubao-seed-1-6-flash-250715
                True
            ))
            print(f"Inserted 'dev_orchestrator' into Postgres with GUID: {agent_guid}")
        
        pg_conn.close()
    except Exception as e:
        print(f"Postgres error: {e}")

    # SQL Server setup
    try:
        mssql_conn = pymssql.connect(server='192.168.0.114', user='derlin', password='fkueiqiq461686', database='sz84_robot')
        mssql_cursor = mssql_conn.cursor()
        
        print("Connected to SQL Server sz84_robot database")
        
        # Check if 岗位任务 exists
        mssql_cursor.execute("SELECT CmdName FROM Cmd WHERE CmdName = '岗位任务'")
        if mssql_cursor.fetchone():
            print("'岗位任务' already exists in SQL Server.")
        else:
            mssql_cursor.execute("INSERT INTO Cmd (CmdName, CmdText, IsClose) VALUES (%s, %s, %s)", ('岗位任务', '岗位任务', 0))
            mssql_conn.commit()
            print("Inserted '岗位任务' into SQL Server.")
            
        mssql_conn.close()
    except Exception as e:
        print(f"SQL Server error: {e}")

if __name__ == "__main__":
    setup()
