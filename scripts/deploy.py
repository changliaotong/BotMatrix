import os
import subprocess
import pack_project
import sys
import argparse

# 默认配置
DEFAULT_SERVER_IP = "192.168.0.167"
DEFAULT_USERNAME = "derlin"
REMOTE_TMP_PATH = "/tmp/botmatrix_deploy.zip"
REMOTE_DEPLOY_DIR = "/opt/botmatrix"

def run_command(cmd):
    """运行系统命令并检查错误"""
    print(f"Executing: {cmd}")
    try:
        ret = subprocess.run(cmd, shell=True)
        if ret.returncode != 0:
            print(f"Error: Command failed with exit code {ret.returncode}")
            sys.exit(ret.returncode)
    except Exception as e:
        print(f"Error executing command: {e}")
        sys.exit(1)

def main():
    parser = argparse.ArgumentParser(description='Deploy BotMatrix to remote server')
    parser.add_argument('--ip', default=DEFAULT_SERVER_IP, help='Server IP address')
    parser.add_argument('--user', default=DEFAULT_USERNAME, help='SSH Username')
    parser.add_argument('--target', choices=['manager', 'wxbot', 'all'], default='all', help='Target to deploy')
    parser.add_argument('--fast', action='store_true', help='Fast deploy (copy files + restart, no rebuild)')
    args = parser.parse_args()

    SERVER_IP = args.ip
    USERNAME = args.user
    TARGET = args.target
    FAST_MODE = args.fast

    print("========================================")
    print(f"   Automated Deployment to {USERNAME}@{SERVER_IP}")
    print(f"   Target: {TARGET}")
    print(f"   Mode: {'Fast (Update & Restart)' if FAST_MODE else 'Full (Rebuild & Recreate)'}")
    print("========================================")

    # 1. 打包项目
    print("\n[Step 1/3] Packing project files...")
    pack_project.pack_project()
    
    root_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    local_zip = os.path.join(root_dir, "botmatrix_deploy.zip")
    if not os.path.exists(local_zip):
        print(f"Error: {local_zip} not found!")
        sys.exit(1)

    # 2. 上传文件
    print("\n[Step 2/3] Uploading to server...")
    print("Note: You may be asked for your SSH password.")
    upload_cmd = f"scp {local_zip} {USERNAME}@{SERVER_IP}:{REMOTE_TMP_PATH}"
    run_command(upload_cmd)
    
    # 3. 远程部署
    print("\n[Step 3/3] Executing remote deployment commands...")
    
    # 公共初始化命令
    init_cmds = [
        f"echo '--> Creating directory {REMOTE_DEPLOY_DIR}...'",
        f"sudo mkdir -p {REMOTE_DEPLOY_DIR}",
        
        f"echo '--> Unzipping files...'",
        f"sudo unzip -o {REMOTE_TMP_PATH} -d {REMOTE_DEPLOY_DIR}",
        
        f"echo '--> Cleaning up temporary zip...'",
        f"sudo rm {REMOTE_TMP_PATH}",
        
        f"echo '--> Switching directory...'",
        f"cd {REMOTE_DEPLOY_DIR}",

        f"echo '--> Ensuring network botmatrix_net exists...'",
        f"sudo docker network create botmatrix_net || true",
    ]

    deploy_cmds = []

    # Deploy Manager
    if TARGET in ['manager', 'all']:
        if FAST_MODE:
             deploy_cmds.extend([
                f"echo '--> [Manager] Fast Update Mode'",
                f"echo '--> [Manager] Copying updated files to container...'",
                f"sudo docker cp {REMOTE_DEPLOY_DIR}/WxBot/. botmatrix-manager:/app/",
                f"echo '--> [Manager] Restarting container...'",
                f"sudo docker restart botmatrix-manager",
            ])
        else:
            deploy_cmds.extend([
                f"echo '--> [Manager] Stopping existing container...'",
                f"sudo docker stop botmatrix-manager || true",
                f"sudo docker rm botmatrix-manager || true",
                
                f"echo '--> [Manager] Checking and freeing port 3005...'",
                f"sudo fuser -k -9 3005/tcp || true",

                f"echo '--> [Manager] Building and starting...'",
                # Ensure we use the manager compose file
                f"sudo WS_PORT=3005 docker-compose -f docker-compose.manager.yml up -d --build",
            ])

    # Deploy WxBot
    if TARGET in ['wxbot', 'all']:
        if FAST_MODE:
             deploy_cmds.extend([
                f"echo '--> [WxBot] Fast Update Mode'",
                f"echo '--> [WxBot] Copying updated files to container...'",
                f"sudo docker cp {REMOTE_DEPLOY_DIR}/WxBot/. botmatrix-worker-wx:/app/",
                f"echo '--> [WxBot] Restarting container...'",
                f"sudo docker restart botmatrix-worker-wx",
            ])
        else:
            deploy_cmds.extend([
                f"echo '--> [WxBot] Stopping existing container...'",
                f"sudo docker stop botmatrix-worker-wx || true",
                f"sudo docker rm botmatrix-worker-wx || true",
                
                f"echo '--> [WxBot] Checking and freeing port 3111...'",
                f"sudo fuser -k -9 3111/tcp || true",
                
                f"echo '--> [WxBot] Building and starting...'",
                # Ensure we use the wxbot compose file
                f"sudo WXBOT_PORT=3111 docker-compose -f docker-compose.wxbot.yml up -d --build",
            ])

    final_msg = [f"echo '--> Deployment ({TARGET}) SUCCESS!'"]

    # Combine all commands
    remote_cmds = init_cmds + deploy_cmds + final_msg
    
    remote_cmd_str = " && ".join(remote_cmds)
    
    ssh_cmd = f'ssh -t {USERNAME}@{SERVER_IP} "{remote_cmd_str}"'
    
    run_command(ssh_cmd)

if __name__ == "__main__":
    main()