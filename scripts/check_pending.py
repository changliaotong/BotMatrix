import redis

r = redis.Redis(host='192.168.0.126', port=6379, password='redis_zsYik8', db=0)
stream_name = 'botmatrix:queue:default'
group_name = 'botmatrix-group'

try:
    pending_info = r.xpending(stream_name, group_name)
    # The return format depends on the redis-py version
    print(f"Pending Info: {pending_info}")
    
    count = pending_info.get('pending') if isinstance(pending_info, dict) else pending_info[0]
    print(f"Total pending: {count}")
    
    if count > 0:
        details = r.xpending_range(stream_name, group_name, '-', '+', 10)
        for d in details:
            print(f"ID: {d['message_id'].decode()}, Consumer: {d['consumer'].decode()}, Idle: {d['time_since_delivered']}, Deliveries: {d['times_delivered']}")
except Exception as ex:
    import traceback
    traceback.print_exc()
