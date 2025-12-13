
import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from SQLConn import SQLConn
import pyodbc
import pymssql

def check_driver():
    print("Checking DB Driver...")
    try:
        conn = SQLConn.conn()
        print(f"Connection object: {conn}")
        print(f"Type: {type(conn)}")
        
        if isinstance(conn, pymssql.Connection):
            print("Driver: pymssql (uses %s)")
        elif isinstance(conn, pyodbc.Connection):
            print("Driver: pyodbc (uses ?)")
        else:
            print("Driver: Unknown")
            
        conn.close()
    except Exception as e:
        print(f"Connection failed: {e}")

if __name__ == "__main__":
    check_driver()
