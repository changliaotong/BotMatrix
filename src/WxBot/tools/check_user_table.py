import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from SQLConn import SQLConn
import common

def check_user_table():
    common.is_debug_sql = True
    print("Checking User table columns:")
    sql = "SELECT COLUMN_NAME FROM sz84_robot.INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'User'"
    cols = SQLConn.QueryDict(sql)
    if cols:
        print(f"Columns: {[c['COLUMN_NAME'] for c in cols]}")
    else:
        print("Table User not found.")

if __name__ == "__main__":
    check_user_table()
