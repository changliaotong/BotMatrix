# coding: utf-8
from SQLConn import SQLConn
import traceback

def handle(context):
    """
    Plugin to log messages to Database (ChatLog table).
    Ensure 'ChatLog' table exists in your database.
    """
    try:
        # 1. Extract message details
        content = context.get("content", "")
        # Truncate content if too long
        if len(content) > 3500:
            content = content[:3500] + "..."
            
        msg_type = context.get("msg_type", "private")
        user_id = str(context.get("user_id", ""))
        user_name = context.get("sender", "")
        group_id = str(context.get("group_id", ""))
        group_name = context.get("group", "")
        
        # 2. Prepare SQL
        # Using parameterized query for safety (PyODBC uses ?)
        sql = """
            INSERT INTO ChatLog (msg_type, user_id, user_name, group_id, group_name, content, create_time)
            VALUES (?, ?, ?, ?, ?, ?, GETDATE())
        """
        
        params = (msg_type, user_id, user_name, group_id, group_name, content)
        
        # 3. Execute
        # Note: SQLConn.Exec suppresses errors unless common.is_debug_sql is True.
        # If the table doesn't exist, this will fail silently in production mode.
        SQLConn.Exec(sql, params)
        
    except Exception as e:
        # Log error to console only if needed
        # print(f"[db_logger] Error: {e}")
        pass
        
    # Return None to allow other plugins to continue processing
    return None
