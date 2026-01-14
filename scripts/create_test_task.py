import psycopg2
import uuid

def create_code_task():
    try:
        conn = psycopg2.connect(
            dbname='botmatrix',
            user='derlin',
            password='fkueiqiq461686',
            host='192.168.0.114'
        )
        cur = conn.cursor()

        # 1. 任务信息
        execution_id = str(uuid.uuid4())
        title = "编写斐波那契数列脚本"
        description = "编写一个 Python 脚本 fib.py，计算斐波那契数列的前 10 个数字，并将结果保存到 fib.txt 文件中。完成后运行脚本验证结果。"
        assignee_id = 1  # 刚才创建的代码专家 ID

        cur.execute("""
            INSERT INTO ai_task_records (execution_id, title, description, assignee_id, status)
            VALUES (%s, %s, %s, %s, %s)
            RETURNING id
        """, (execution_id, title, description, assignee_id, 'pending'))
        
        task_id = cur.fetchone()[0]
        print(f"Task created: {title} (ID: {task_id}, ExecutionID: {execution_id})")

        conn.commit()
        cur.close()
        conn.close()
        print("Task insertion complete.")
        return execution_id

    except Exception as e:
        print(f"Error: {e}")
        return None

if __name__ == "__main__":
    create_code_task()
