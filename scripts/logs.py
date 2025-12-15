import argparse
import subprocess
import sys
import os

# 配置信息 (与 update.py 保持一致)
SERVER_IP = "192.168.0.167"
USERNAME = "derlin"

def main():
    parser = argparse.ArgumentParser(description="Fetch logs from remote container")
    parser.add_argument('service', type=str, nargs='?', default="system-worker", help='Service name (e.g. system-worker, manager)')
    parser.add_argument('-f', '--follow', action='store_true', help='Follow logs')
    parser.add_argument('-n', '--lines', type=str, default="50", help='Number of lines to show')
    
    args = parser.parse_args()
    
    # 映射简写到容器名 (for docker logs) or service name (for docker-compose logs)
    service_map = {
        "system-worker": "system-worker",
        "worker": "system-worker",
        "sys": "system-worker",
        "manager": "bot-manager",
        "bot-manager": "bot-manager",
        "nexus": "bot-manager",
        "wxbot": "wxbot",
        "wx": "wxbot",
        "bot": "wxbot",
        "tencent": "tencent-bot",
        "dingtalk": "dingtalk-bot"
    }

    target = args.service
    
    # Special mode: debug (watch sys + manager)
    if target == "debug":
        cmd = f"ssh -t {USERNAME}@{SERVER_IP} \"cd /opt/BotMatrix && sudo docker-compose logs {'-f' if args.follow else ''} --tail {args.lines} system-worker bot-manager\""
    else:
        # Default single service check
        service_name = service_map.get(target, target)
        
        # Try to use docker-compose logs for everything as it's cleaner with colors/prefixes
        # But we need to know the service name in docker-compose.yml
        # The mapping above uses service names.
        
        cmd = f"ssh -t {USERNAME}@{SERVER_IP} \"cd /opt/BotMatrix && sudo docker-compose logs {'-f' if args.follow else ''} --tail {args.lines} {service_name}\""
    
    print(f"Executing: {cmd}")
    try:
        subprocess.run(cmd, shell=True)
    except KeyboardInterrupt:
        pass
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()
