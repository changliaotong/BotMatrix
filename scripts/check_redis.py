import redis
import json

r = redis.Redis(host='192.168.0.126', port=6379, password='redis_zsYik8', db=0)
stream_name = 'botmatrix:queue:default'

try:
    length = r.xlen(stream_name)
    print(f"Stream length: {length}")
    
    messages = r.xrevrange(stream_name, count=5)
    for msg_id, data in messages:
        print(f"ID: {msg_id}")
        # data is a dict of bytes
        for k, v in data.items():
            print(f"  {k.decode()}: {v.decode()[:100]}...")
except Exception as ex:
    print(f"Error: {ex}")
