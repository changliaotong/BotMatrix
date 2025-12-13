import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from wxclient import User
from SQLConn import SQLConn
import common

def verify_exists_logic():
    common.is_debug_sql = True
    print("--- Verifying User.exists logic ---")

    # 1. Find a client_qq that exists in User table
    print("Finding a valid Id from User table...")
    sql = "SELECT TOP 1 Id FROM [User] WHERE Id > 0"
    res = SQLConn.Query(sql)
    
    if not res:
        print("No users found in User table to test with.")
        return

    existing_id = int(res)
    print(f"Found existing User Id: {existing_id}")

    # 2. Test User.exists(id)
    print(f"Testing User.exists({existing_id})...")
    exists = User.exists(existing_id)
    print(f"User.exists({existing_id}) returned: {exists}")

    if exists:
        print("SUCCESS: User.exists correctly identified existing user.")
    else:
        print("FAILURE: User.exists returned False for existing user.")

    # 3. Test with non-existent ID
    non_existent_id = 999999999999
    print(f"Testing User.exists({non_existent_id}) (should be False)...")
    exists_fake = User.exists(non_existent_id)
    print(f"User.exists({non_existent_id}) returned: {exists_fake}")

    if not exists_fake:
         print("SUCCESS: User.exists correctly identified non-existing user.")
    else:
         print("FAILURE: User.exists returned True for non-existing user.")

if __name__ == "__main__":
    verify_exists_logic()
