import redis
import json
import time

r = redis.Redis(host='192.168.0.126', port=6379, password='redis_zsYik8', decode_responses=True)

import sys

prompt = sys.argv[1] if len(sys.argv) > 1 else "岗位任务 software_engineer 编写斐波那契数列脚本"

event = {
    "time": int(time.time()),
    "self_id": 123456,
    "post_type": "message",
    "message_type": "group",
    "message_id": int(time.time()),
    "user_id": 123456,
    "group_id": 999999,
    "raw_message": prompt,
    "sender": {
        "user_id": 123456,
        "nickname": "Tester"
    }
}

payload = json.dumps(event)
r.xadd("botmatrix:queue:default", {"payload": payload})
print("Task sent to Redis stream")
