from SQLConn import SQLConn
import sys

def fix_schema():
    print("Checking wx_group schema...")
    try:
        # SQL Server syntax to check column length
        # But easier to just try ALTER.
        # Ensure we have a connection first
        conn = SQLConn.conn()
        if not conn:
            print("Could not connect to database.")
            return

        print("Attempting to increase group_uid column length to 255...")
        sql = "ALTER TABLE wx_group ALTER COLUMN group_uid VARCHAR(255)"
        
        cursor = conn.cursor()
        try:
            cursor.execute(sql)
            conn.commit()
            print("Schema update executed successfully.")
        except Exception as e:
            print(f"Schema update failed (might already be updated or permission denied): {e}")
        finally:
            conn.close()
            
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    fix_schema()