import pymssql

def check_cmd():
    server = '192.168.0.114'
    user = 'derlin'
    password = 'fkueiqiq461686'
    database = 'sz84_robot'

    try:
        conn = pymssql.connect(server, user, password, database)
        cursor = conn.cursor(as_dict=True)
        
        print(f"Connected to {database} on {server}")
        
        # Check all commands
        print("\nAll commands in Cmd table:")
        cursor.execute("SELECT CmdName, CmdText, IsClose FROM Cmd ORDER BY CmdName")
        rows = cursor.fetchall()
        for row in rows:
            print(f"Name: {row['CmdName']}, Text: {row['CmdText']}, Closed: {row['IsClose']}")
            
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    check_cmd()
