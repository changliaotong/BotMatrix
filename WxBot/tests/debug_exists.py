import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from wxclient import User
from SQLConn import SQLConn
from MetaData import MetaData
import common

def debug_exists():
    common.is_debug_sql = True
    target_id = 90000058220
    
    print(f"--- Debugging User.exists({target_id}) ---")
    
    # 1. Test via User.exists
    try:
        exists = User.exists(target_id)
        print(f"User.exists({target_id}) result: {exists}")
    except Exception as e:
        print(f"User.exists failed with error: {e}")

    # 2. Test raw SQL with brackets
    print("\n--- Testing Raw SQL with [User] ---")
    sql = f"SELECT 1 FROM sz84_robot.dbo.[User] WHERE Id = {target_id}"
    res = SQLConn.Query(sql)
    print(f"Raw SQL Result (with brackets): {res}")

    # 3. Test raw SQL without brackets
    print("\n--- Testing Raw SQL with User (no brackets) ---")
    sql = f"SELECT 1 FROM sz84_robot.dbo.User WHERE Id = {target_id}"
    res = SQLConn.Query(sql)
    print(f"Raw SQL Result (no brackets): {res}")

    # 4. Test MetaData generated SQL (manual simulation)
    print("\n--- Testing MetaData-style Parameterized SQL ---")
    sql = "SELECT 1 FROM sz84_robot.dbo.User WHERE Id = %s"
    res = SQLConn.Query(sql, (target_id,))
    print(f"MetaData SQL Result: {res}")

    # 5. Test MetaData-style Parameterized SQL with Brackets
    print("\n--- Testing MetaData-style Parameterized SQL with [User] ---")
    sql = "SELECT 1 FROM sz84_robot.dbo.[User] WHERE Id = %s"
    res = SQLConn.Query(sql, (target_id,))
    print(f"MetaData SQL Result (with brackets): {res}")

if __name__ == "__main__":
    debug_exists()
