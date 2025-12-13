import os
import subprocess
import pack_project
import sys
import argparse

# 默认配置
DEFAULT_SERVER_IP = "192.168.0.167"
DEFAULT_USERNAME = "derlin"
REMOTE_TMP_PATH = "/tmp/botmatrix_deploy.zip"
REMOTE_DEPLOY_DIR = "/opt/BotMatrix"

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
    
    remote_cmds = [
        # 1. 创建目录
        f"echo '--> Creating directory {REMOTE_DEPLOY_DIR} ...'",
        f"sudo mkdir -p {REMOTE_DEPLOY_DIR}",
        
        # 2. 解压
        f"echo '--> Unzipping files...'",
        f"sudo unzip -o {REMOTE_TMP_PATH} -d {REMOTE_DEPLOY_DIR}",
        
        # 3. 清理 zip
        f"echo '--> Cleaning up temporary zip...'",
        f"sudo rm {REMOTE_TMP_PATH}",
        
        # 4. 进入目录
        f"echo '--> Switching directory...'",
        f"cd {REMOTE_DEPLOY_DIR}",

        # 5. Docker Compose 部署
        f"echo '--> Starting services with Docker Compose...'",
    ]

    docker_cmd = ""
    if FAST_MODE:
        # Fast Mode: No build, just restart to pick up mounted code changes
        if TARGET == 'manager':
            docker_cmd = "sudo docker-compose restart bot-manager"
        elif TARGET == 'wxbot':
            docker_cmd = "sudo docker-compose restart wxbot"
        else:
            docker_cmd = "sudo docker-compose up -d --remove-orphans && sudo docker-compose restart wxbot"
    else:
        # Full Mode: Rebuild images
        if TARGET == 'manager':
            docker_cmd = "sudo docker-compose up -d --build --no-deps bot-manager"
        elif TARGET == 'wxbot':
            docker_cmd = "sudo docker-compose up -d --build --no-deps wxbot"
        else:
            docker_cmd = "sudo docker-compose up -d --build --remove-orphans"

    remote_cmds.append(docker_cmd)
    
    remote_cmds.extend([
        # 6. 显示状态
        f"echo '--> Deployment finished. Checking status...'",
        f"sudo docker-compose ps",
        f"echo '--> Logs (tail 20):'",
        f"sudo docker-compose logs --tail=20"
    ])

    full_cmd = " && ".join(remote_cmds)
    # 使用 -t 强制分配伪终端，以便 sudo 可以提示输入密码
    ssh_cmd = f"ssh -t {USERNAME}@{SERVER_IP} \"{full_cmd}\""
    
    run_command(ssh_cmd)
    
    print("\n========================================")
    print("   Deployment Successful!")
    print(f"   Dashboard: http://{SERVER_IP}:5000")
    print(f"   WxBot WebUI: http://{SERVER_IP}:5001")
    print(f"   Gateway:   ws://{SERVER_IP}:3111")
    print("========================================")

if __name__ == "__main__":
    main()