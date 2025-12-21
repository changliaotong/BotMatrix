import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from SQLConn import SQLConn
import common
import json
import datetime

import uuid
from decimal import Decimal

class DateTimeEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, (datetime.date, datetime.datetime)):
            return obj.isoformat()
        if isinstance(obj, uuid.UUID):
            return str(obj)
        if isinstance(obj, Decimal):
            return float(obj)
        return super(DateTimeEncoder, self).default(obj)

def check_data():
    common.is_debug_sql = True
    
    print("\n--- Sample Data from User ---")
    sql = "SELECT TOP 1 * FROM sz84_robot.dbo.[User]"
    res = SQLConn.QueryDict(sql)
    if res:
        print(json.dumps(res[0], indent=2, cls=DateTimeEncoder, ensure_ascii=False))
    else:
        print("User table empty.")

    print("\n--- Sample Data from GroupMember ---")
    sql = "SELECT TOP 1 * FROM sz84_robot.dbo.GroupMember"
    res = SQLConn.QueryDict(sql)
    if res:
        print(json.dumps(res[0], indent=2, cls=DateTimeEncoder, ensure_ascii=False))
    else:
        print("GroupMember table empty.")

    print("\n--- Sample Data from wx_client ---")
    sql = "SELECT TOP 1 * FROM sz84_robot.dbo.wx_client"
    res = SQLConn.QueryDict(sql)
    if res:
        print(json.dumps(res[0], indent=2, cls=DateTimeEncoder, ensure_ascii=False))
    else:
        print("wx_client table empty.")

if __name__ == "__main__":
    check_data()
