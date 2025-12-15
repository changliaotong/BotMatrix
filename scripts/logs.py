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
    
    # 映射简写到容器名
    service_map = {
        "system-worker": "botmatrix-system-worker",
        "worker": "botmatrix-system-worker",
        "sys": "botmatrix-system-worker",
        "manager": "botmatrix-manager",
        "bot-manager": "botmatrix-manager",
        "nexus": "botmatrix-manager",
        "wxbot": "wxbot",
        "wx": "wxbot",
        "bot": "wxbot"
    }
    
    container_name = service_map.get(args.service, args.service)
    
    # 构建 SSH 命令
    # 使用 -t 强制分配伪终端，以便 sudo 可能需要密码时能交互 (在本地终端运行时)
    # 注意：在自动化环境中如果需要密码可能会卡住，但之前的 update.py 似乎能跑通
    cmd = f"ssh -t {USERNAME}@{SERVER_IP} \"sudo docker logs {'-f' if args.follow else ''} --tail {args.lines} {container_name}\""
    
    print(f"Executing: {cmd}")
    try:
        subprocess.run(cmd, shell=True)
    except KeyboardInterrupt:
        pass
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()
