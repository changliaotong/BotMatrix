import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from SQLConn import SQLConn
import traceback

def list_tables():
    # Force debug mode to see errors
    import common
    common.common.is_debug_sql = True
    
    print("Attempting to connect...")
    try:
        sql = "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME"
        results = SQLConn.QueryDict(sql)
        print("Tables in DB:")
        for row in results:
            print(row['TABLE_NAME'])
    except Exception as e:
        print("Error:", e)
        traceback.print_exc()

if __name__ == "__main__":
    list_tables()
