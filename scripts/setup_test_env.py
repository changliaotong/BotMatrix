import pymssql

def setup_group():
    server = '192.168.0.114'
    user = 'derlin'
    password = 'fkueiqiq461686'
    database = 'sz84_robot'
    
    try:
        conn = pymssql.connect(server=server, user=user, password=password, database=database)
        cur = conn.cursor()
        
        # SQL for checking and inserting/updating
        sql = """
        IF NOT EXISTS (SELECT 1 FROM [Group] WHERE Id = 123456)
            INSERT INTO [Group] (Id, GroupName, UseRight, IsOpen, IsValid, RobotOwner, IsSz84)
            VALUES (123456, 'Test Group', 1, 1, 1, 10001, 0)
        ELSE
            UPDATE [Group] SET UseRight = 1, IsOpen = 1, IsValid = 1, IsSz84 = 0 WHERE Id = 123456
        """
        
        cur.execute(sql)
        conn.commit()
        print("Group 123456 configured successfully.")
        
        # Also ensure the user 10001 exists and has AI rights
        sql_user = """
        IF NOT EXISTS (SELECT 1 FROM [User] WHERE Id = 10001)
            INSERT INTO [User] (Id, Name, IsAI, Credit)
            VALUES (10001, 'Test User', 1, 1000)
        ELSE
            UPDATE [User] SET IsAI = 1, Credit = 1000 WHERE Id = 10001
        """
        cur.execute(sql_user)
        conn.commit()
        print("User 10001 configured successfully.")
        
        conn.close()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    setup_group()
