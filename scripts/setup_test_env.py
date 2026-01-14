import psycopg2
import json

def setup_test_environment():
    try:
        conn = psycopg2.connect(
            dbname='botmatrix',
            user='derlin',
            password='fkueiqiq461686',
            host='192.168.0.114'
        )
        cur = conn.cursor()

        # 1. 创建岗位定义 (Software Engineer)
        job_key = 'software_engineer'
        job_name = '软件工程师'
        purpose = '负责编写、测试和优化代码，能够独立完成模块开发。'
        tool_schema = json.dumps(['file_read', 'file_write', 'list_dir', 'shell_exec', 'python_test'])
        
        cur.execute("""
            INSERT INTO ai_job_definitions (job_key, name, purpose, tool_schema)
            VALUES (%s, %s, %s, %s)
            ON CONFLICT (job_key) DO UPDATE SET
            name = EXCLUDED.name, purpose = EXCLUDED.purpose, tool_schema = EXCLUDED.tool_schema
            RETURNING id
        """, (job_key, job_name, purpose, tool_schema))
        job_id = cur.fetchone()[0]
        print(f"Job created/updated: {job_name} (ID: {job_id})")

        # 2. 更新“程序员”智能体，关联豆包 1.8 模型 (ID: 40)
        agent_id = 1038
        model_id = 40
        cur.execute("UPDATE ai_agents SET model_id = %s WHERE id = %s", (model_id, agent_id))
        print(f"Agent {agent_id} updated with model {model_id}")

        # 3. 创建数字员工实例
        employee_id = 'EMP-001'
        bot_id = 'test_bot'
        employee_name = '代码专家'
        title = '高级软件工程师'
        
        cur.execute("""
            INSERT INTO ai_employee_instances (employee_id, bot_id, agent_id, job_id, name, title, online_status)
            VALUES (%s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (employee_id) DO UPDATE SET
            bot_id = EXCLUDED.bot_id, agent_id = EXCLUDED.agent_id, job_id = EXCLUDED.job_id,
            name = EXCLUDED.name, title = EXCLUDED.title, online_status = EXCLUDED.online_status
            RETURNING id
        """, (employee_id, bot_id, agent_id, job_id, employee_name, title, 'online'))
        emp_instance_id = cur.fetchone()[0]
        print(f"Employee instance created/updated: {employee_name} (ID: {emp_instance_id})")

        conn.commit()
        cur.close()
        conn.close()
        print("Test environment setup complete.")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    setup_test_environment()
