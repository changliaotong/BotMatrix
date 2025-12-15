import os
import subprocess
import pack_project
import bump_version
import sys
import argparse

import time
import shutil

# 默认配置
DEFAULT_SERVER_IP = "192.168.0.167"
DEFAULT_USERNAME = "derlin"
# Use unique filename to avoid permission/overwrite issues
TIMESTAMP = int(time.time())
REMOTE_FILENAME = f"botmatrix_deploy_{TIMESTAMP}.zip"
REMOTE_TMP_PATH = f"/home/derlin/{REMOTE_FILENAME}"
REMOTE_DEPLOY_DIR = "/opt/BotMatrix"

def ensure_configs():
    """Ensure config.json exists for all bots by copying config.sample.json if needed"""
    root_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    bot_dirs = [
        "DingTalkBot", "DiscordBot", "EmailBot", "FeishuBot", 
        "KookBot", "SlackBot", "TelegramBot", "TencentBot", 
        "WeComBot", "WxBot"
    ]
    
    print("Checking config files...")
    for bot_dir in bot_dirs:
        dir_path = os.path.join(root_dir, bot_dir)
        config_path = os.path.join(dir_path, "config.json")
        sample_path = os.path.join(dir_path, "config.sample.json")

        if not os.path.exists(config_path) and os.path.exists(sample_path):
            print(f"  [+] Creating {bot_dir}/config.json from sample")
            try:
                shutil.copy(sample_path, config_path)
            except Exception as e:
                print(f"  [-] Failed to create config for {bot_dir}: {e}")

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
    parser.add_argument('--target', default=None, help='Target to deploy (manager, wxbot, all, etc.)')
    parser.add_argument('--fast', action='store_true', help='Fast deploy (copy files + restart, no rebuild)')
    args = parser.parse_args()

    SERVER_IP = args.ip
    USERNAME = args.user
    TARGET = args.target
    FAST_MODE = args.fast

    # Interactive Menu if no target specified
    if TARGET is None:
        print("\nSelect Deployment Target:")
        print("  1. [All] Deploy Everything (Default)")
        print("  2. [NoWx] Deploy All EXCEPT WxBot (Preserves Login)")
        print("  3. [Mgr] Bot Manager Only")
        print("  4. [Wx] WxBot Only")
        print("  5. [Tencent] TencentBot Only")
        print("  6. [Sys] System Worker Only")
        
        choice = input("\nEnter choice (1-6) [1]: ").strip()
        
        if choice == "" or choice == "1":
            TARGET = "all"
        elif choice == "2":
            TARGET = "no-wx"
        elif choice == "3":
            TARGET = "manager"
        elif choice == "4":
            TARGET = "wxbot"
        elif choice == "5":
            TARGET = "tencent-bot"
        elif choice == "6":
            TARGET = "system-worker"
        else:
            print("Invalid choice. Aborting.")
            sys.exit(1)

    print("========================================")
    print(f"   Automated Deployment to {USERNAME}@{SERVER_IP}")
    print(f"   Target: {TARGET}")
    print(f"   Mode: {'Fast (Update & Restart)' if FAST_MODE else 'Full (Rebuild & Recreate)'}")
    print("========================================")

    # 0. 版本号自增
    print("\n[Step 0/3] Bumping version...")
    bump_version.bump_version('patch')

    # Ensure configs exist before packing
    ensure_configs()

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

        # 1.5 Fix: Remove config.json directories if they exist (caused by docker volume mount on missing file)
        f"echo '--> Checking for directory/file conflicts...'",
        f"sudo find {REMOTE_DEPLOY_DIR} -name config.json -type d -exec rm -rf {{}} + || true",
        
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
        elif TARGET == 'tencent-bot':
            docker_cmd = "sudo docker-compose restart tencent-bot"
        elif TARGET == 'system-worker':
            docker_cmd = "sudo docker-compose restart system-worker"
        elif TARGET == 'no-wx':
             # Restart everything EXCEPT wxbot
             # We list explicit services to restart, excluding wxbot
             services_to_restart = "bot-manager system-worker tencent-bot dingtalk-bot feishu-bot telegram-bot discord-bot slack-bot kook-bot email-bot wecom-bot"
             docker_cmd = f"sudo docker-compose restart {services_to_restart}"
        else:
            docker_cmd = "sudo docker-compose up -d --remove-orphans && sudo docker-compose restart wxbot"
    else:
        # Full Mode: Rebuild images
        # Force --no-cache to ensure go build gets latest code
        
        # Pre-cleanup command to handle "name already in use" errors for single service deployments
        cleanup_cmd = ""
        
        if TARGET == 'manager':
            cleanup_cmd = "sudo docker rm -f botmatrix-manager || true"
            docker_cmd = f"{cleanup_cmd} && sudo docker-compose up -d --build --force-recreate --no-deps bot-manager"
        elif TARGET == 'wxbot':
            cleanup_cmd = "sudo docker rm -f wxbot || true"
            docker_cmd = f"{cleanup_cmd} && sudo docker-compose up -d --build --force-recreate --no-deps wxbot"
        elif TARGET == 'tencent-bot':
            cleanup_cmd = "sudo docker rm -f tencent-bot || true"
            docker_cmd = f"{cleanup_cmd} && sudo docker-compose up -d --build --force-recreate --no-deps tencent-bot"
        elif TARGET == 'system-worker':
            cleanup_cmd = "sudo docker rm -f botmatrix-system-worker || true"
            docker_cmd = f"{cleanup_cmd} && sudo docker-compose up -d --build --force-recreate --no-deps system-worker"
        elif TARGET == 'no-wx':
            # Rebuild and recreate everything EXCEPT wxbot
            services_to_up = "bot-manager system-worker tencent-bot dingtalk-bot feishu-bot telegram-bot discord-bot slack-bot kook-bot email-bot wecom-bot"
            # Container names for cleanup (explicit mapping to avoid confusion)
            # bot-manager -> botmatrix-manager
            # system-worker -> botmatrix-system-worker
            # Others match service name
            containers_to_clean = "botmatrix-manager botmatrix-system-worker tencent-bot dingtalk-bot feishu-bot telegram-bot discord-bot slack-bot kook-bot email-bot wecom-bot"
            
            # For multi-service, rely on compose's force-recreate usually, but if conflicts persist, we might need aggressive cleanup.
            # However, for now, let's assume single-target deployments are the most prone to manual interference conflicts.
            # If conflicts occur here, we can add "sudo docker-compose rm -f -s {services_to_up}"
            cleanup_cmd = f"sudo docker rm -f {containers_to_clean} || true"
            docker_cmd = f"{cleanup_cmd} && sudo docker-compose up -d --build --force-recreate {services_to_up}"
        else:
            # Full Deploy: "All"
            # To avoid conflicts, we should really do a full down.
            # This stops WxBot too, but "All" implies fresh start.
            docker_cmd = "sudo docker-compose down --remove-orphans && sudo docker-compose up -d --build --force-recreate"

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