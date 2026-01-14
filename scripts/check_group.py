import redis

r = redis.Redis(host='192.168.0.126', port=6379, password='redis_zsYik8', db=0)
stream_name = 'botmatrix:queue:default'
group_name = 'botmatrix-group'

try:
    groups = r.xinfo_groups(stream_name)
    for group in groups:
        if group['name'].decode() == group_name:
            print(f"Group: {group['name'].decode()}")
            print(f"  Consumers: {group['consumers']}")
            print(f"  Pending: {group['pending']}")
            print(f"  Last delivered ID: {group['last-delivered-id'].decode()}")
            
    consumers = r.xinfo_consumers(stream_name, group_name)
    print(f"Total consumers: {len(consumers)}")
    
    # Sort by idle time to find the most recent ones
    consumers.sort(key=lambda x: x['idle'])
    print("\nTop 5 active consumers:")
    for consumer in consumers[:5]:
        print(f"Consumer: {consumer['name'].decode()}")
        print(f"  Pending: {consumer['pending']}")
        print(f"  Idle: {consumer['idle']}")
except Exception as ex:
    print(f"Error: {ex}")
