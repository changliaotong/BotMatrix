import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from SQLConn import SQLConn
import common

def check_pk():
    common.is_debug_sql = True
    print("Checking User table constraints...")
    
    # Check Primary Key
    sql_pk = """
    SELECT COLUMN_NAME 
    FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
    WHERE TABLE_NAME = 'User' AND CONSTRAINT_NAME LIKE 'PK%'
    """
    pk = SQLConn.QueryDict(sql_pk)
    print(f"PK Columns: {pk}")

    # Check Identity
    sql_identity = """
    SELECT name, is_identity
    FROM sys.columns 
    WHERE object_id = object_id('User') AND is_identity = 1
    """
    identity = SQLConn.QueryDict(sql_identity)
    print(f"Identity Columns: {identity}")
    
    # Check if Id is unique
    sql_unique = """
    SELECT COUNT(*) as Cnt, Id FROM [User] GROUP BY Id HAVING COUNT(*) > 1
    """
    dupes = SQLConn.QueryDict(sql_unique)
    print(f"Duplicates on Id: {len(dupes)} found.")
    if len(dupes) > 0:
        print(f"Sample dupe: {dupes[0]}")

if __name__ == "__main__":
    check_pk()
