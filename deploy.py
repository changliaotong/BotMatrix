import os
import subprocess
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
    root_dir = os.path.dirname(os.path.abspath(__file__))
    bot_dirs = [
        "src/DingTalkBot", "src/DiscordBot", "src/EmailBot", "src/FeishuBot", 
        "src/KookBot", "src/SlackBot", "src/TelegramBot", "src/TencentBot", 
        "src/WeComBot", "src/WxBot"
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
    parser.add_argument('--docker-fix', action='store_true', help='Use Docker fix mode (minimal build)')
    parser.add_argument('--no-source', action='store_true', help='Deploy without source code (build locally)')
    args = parser.parse_args()

    SERVER_IP = args.ip
    USERNAME = args.user
    TARGET = args.target
    FAST_MODE = args.fast
    DOCKER_FIX_MODE = args.docker_fix
    NO_SOURCE = args.no_source

    # Interactive Menu if no target specified
    if TARGET is None:
        print("\nSelect Deployment Target:")
        print("  1. [All] Deploy Everything (Default)")
        print("  2. [NoWx] Deploy All EXCEPT WxBot (Preserves Login)")
        print("  3. [Mgr] Bot Manager Only")
        print("  4. [Wx] WxBot Only")
        print("  5. [Tencent] TencentBot Only")
        print("  6. [Sys] System Worker Only")
        print("  7. [Overmind] Overmind + Bot Manager (Builds Frontend)")
        
        choice = input("\nEnter choice (1-7) [1]: ").strip()
        
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
        elif choice == "7":
            TARGET = "overmind"
        else:
            print("Invalid choice. Aborting.")
            sys.exit(1)

    print("========================================")
    print(f"   Automated Deployment to {USERNAME}@{SERVER_IP}")
    print(f"   Target: {TARGET}")
    print(f"   Mode: {'Fast (Update & Restart)' if FAST_MODE else 'Full (Rebuild & Recreate)'}")
    print(f"   Docker Fix: {'Yes (Minimal Build)' if DOCKER_FIX_MODE else 'No'}")
    print("========================================")

    # 0. Build Overmind (If requested)
    if TARGET == "overmind":
        print("\n[Step 0/3] Building Overmind (Flutter Web)...")
        root_dir = os.path.dirname(os.path.abspath(__file__))
        overmind_dir = os.path.join(root_dir, "src", "Overmind")
        
        # Build Command
        build_cmd = "flutter build web --release --base-href /overmind/"
        print(f"Executing: {build_cmd} in {overmind_dir}")
        
        ret = subprocess.run(build_cmd, shell=True, cwd=overmind_dir)
        if ret.returncode != 0:
            print("Error: Flutter build failed!")
            sys.exit(1)
            
        # Copy Artifacts
        src_dir = os.path.join(overmind_dir, "build", "web")
        dest_dir = os.path.join(root_dir, "src", "BotNexus", "overmind")
        
        print(f"Copying artifacts from {src_dir} to {dest_dir}...")
        if os.path.exists(dest_dir):
            shutil.rmtree(dest_dir)
        shutil.copytree(src_dir, dest_dir)
        
        # Switch target to manager for the rest of deployment logic
        TARGET = "manager"
        FAST_MODE = False # Force rebuild to include new static files

    # 0. 版本号自增
    print("\n[Step 0/3] Bumping version...")
    try:
        import bump_version
        bump_version.bump_version('patch')
    except ImportError:
        print("Warning: bump_version module not found, skipping version bump")

    # Ensure configs exist before packing
    ensure_configs()

    # 0.5 Local Build (If no-source requested)
    if NO_SOURCE:
        print("\n[Step 0.5/3] Building binaries locally (GOOS=linux GOARCH=amd64)...")
        root_dir = os.path.dirname(os.path.abspath(__file__))
        
        go_projects = {
            "BotNexus": {"path": "src/BotNexus", "bin": "BotNexus"},
            "BotWorker": {"path": "src/BotWorker", "bin": "bot-worker"},
            "TencentBot": {"path": "src/TencentBot", "bin": "TencentBot"},
            "WxBotGo": {"path": "src/WxBotGo", "bin": "wxbot-go"},
            "DingTalkBot": {"path": "src/DingTalkBot", "bin": "DingTalkBot"},
            "FeishuBot": {"path": "src/FeishuBot", "bin": "FeishuBot"},
            "TelegramBot": {"path": "src/TelegramBot", "bin": "TelegramBot"},
            "DiscordBot": {"path": "src/DiscordBot", "bin": "DiscordBot"},
            "SlackBot": {"path": "src/SlackBot", "bin": "slack-bot"},
            "KookBot": {"path": "src/KookBot", "bin": "kook-bot"},
            "EmailBot": {"path": "src/EmailBot", "bin": "email-bot"},
            "WeComBot": {"path": "src/WeComBot", "bin": "wecom-bot"},
        }

        env = os.environ.copy()
        env["GOOS"] = "linux"
        env["GOARCH"] = "amd64"
        env["CGO_ENABLED"] = "0" # Ensure static binary for alpine

        for name, info in go_projects.items():
            proj_path = os.path.join(root_dir, info["path"])
            if not os.path.exists(proj_path):
                print(f"  [-] Skipping {name}: path not found")
                continue
            
            print(f"  [+] Building {name} in {info['path']}...")
            build_target = "."
            if os.path.exists(os.path.join(proj_path, "cmd", "main.go")):
                build_target = "./cmd"
            
            ret = subprocess.run(["go", "build", "-o", info["bin"], build_target], 
                                 cwd=proj_path, env=env)
            if ret.returncode != 0:
                print(f"Error: Local build for {name} failed!")
                sys.exit(1)
        
        print("Local builds successful.")

    # 1. 打包项目
    print("\n[Step 1/3] Packing project files...")
    
    pack_script = os.path.join(os.path.dirname(os.path.abspath(__file__)), "pack_project.py")
    pack_cmd = f"{sys.executable} {pack_script}"
    if NO_SOURCE:
        pack_cmd += " --no-source"
    
    run_command(pack_cmd)
    
    root_dir = os.path.dirname(os.path.abspath(__file__))
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
        f"echo '--> Creating directory {REMOTE_DEPLOY_DIR} ...'",
        f"sudo mkdir -p {REMOTE_DEPLOY_DIR}",
        f"echo '--> Checking for directory/file conflicts...'",
        f"sudo find {REMOTE_DEPLOY_DIR} -name config.json -type d -exec rm -rf {{}} + || true",
    ]

    if NO_SOURCE:
        remote_cmds.append(f"echo '--> Cleaning up ALL stale Go source files...'")
        remote_cmds.append(f"sudo find {REMOTE_DEPLOY_DIR} -name '*.go' -delete || true")
        remote_cmds.append(f"sudo find {REMOTE_DEPLOY_DIR} -name 'go.mod' -delete || true")
        remote_cmds.append(f"sudo find {REMOTE_DEPLOY_DIR} -name 'go.sum' -delete || true")
    else:
        remote_cmds.append(f"echo '--> Cleaning up stale Go files in BotNexus...'")
        remote_cmds.append(f"sudo rm -f {REMOTE_DEPLOY_DIR}/src/BotNexus/*.go || true")

    remote_cmds.extend([
        f"echo '--> Unzipping files...'",
        f"sudo unzip -o {REMOTE_TMP_PATH} -d {REMOTE_DEPLOY_DIR}",
        f"echo '--> Initializing missing config files from samples...'",
        f"sudo find {REMOTE_DEPLOY_DIR}/src -name 'config.sample.json' -exec bash -c 'p=\"{{}}\"; cp -n \"$p\" \"${{p%.sample.json}}.json\"' \\; || true",
        f"sudo find {REMOTE_DEPLOY_DIR}/src -name 'config.example.json' -exec bash -c 'p=\"{{}}\"; cp -n \"$p\" \"${{p%.example.json}}.json\"' \\; || true",
        f"sudo find {REMOTE_DEPLOY_DIR}/src -name 'config.example.yaml' -exec bash -c 'p=\"{{}}\"; cp -n \"$p\" \"${{p%.example.yaml}}.yaml\"' \\; || true",
        f"echo '--> Cleaning up temporary zip...'",
        f"sudo rm {REMOTE_TMP_PATH}",
        f"echo '--> Switching directory...'",
        f"cd {REMOTE_DEPLOY_DIR}",
        f"echo '--> Starting services with Docker Compose...'",
    ])

    docker_cmd = ""
    if FAST_MODE:
        if TARGET == 'manager':
            docker_cmd = "sudo docker-compose restart bot-manager"
        elif TARGET == 'wxbot':
            docker_cmd = "sudo docker-compose restart wxbot"
        elif TARGET == 'tencent-bot':
            docker_cmd = "sudo docker-compose restart tencent-bot"
        elif TARGET == 'system-worker':
            docker_cmd = "sudo docker-compose restart system-worker"
        elif TARGET == 'no-wx':
             services_to_restart = "bot-manager system-worker tencent-bot dingtalk-bot feishu-bot telegram-bot discord-bot slack-bot kook-bot email-bot wecom-bot"
             docker_cmd = f"sudo docker-compose restart {services_to_restart}"
        else:
            docker_cmd = "sudo docker-compose up -d --remove-orphans && sudo docker-compose restart wxbot"
    else:
        cleanup_cmd = ""
        if TARGET == 'manager':
            cleanup_cmd = "sudo docker rm -f botmatrix-manager || true"
            if NO_SOURCE:
                docker_cmd = f"{cleanup_cmd} && sudo docker build -f src/BotNexus/Dockerfile.prod -t botmatrix-manager src/BotNexus/ && sudo docker-compose up -d --force-recreate --no-deps bot-manager"
            elif DOCKER_FIX_MODE:
                docker_cmd = f"{cleanup_cmd} && sudo docker build -f src/BotNexus/Dockerfile.minimal -t botmatrix-manager src/BotNexus/ && sudo docker-compose up -d --force-recreate --no-deps bot-manager"
            else:
                docker_cmd = f"{cleanup_cmd} && sudo docker-compose up -d --build --force-recreate --no-deps bot-manager"
        elif TARGET == 'wxbot':
            cleanup_cmd = "sudo docker rm -f wxbot || true"
            docker_cmd = f"{cleanup_cmd} && sudo docker-compose up -d --build --force-recreate --no-deps wxbot"
        elif TARGET == 'tencent-bot':
            cleanup_cmd = "sudo docker rm -f btqq || true"
            if NO_SOURCE:
                docker_cmd = f"{cleanup_cmd} && sudo docker build -f src/TencentBot/Dockerfile.prod -t botmatrix-tencent-bot src/TencentBot/ && sudo docker-compose up -d --force-recreate --no-deps tencent-bot"
            else:
                docker_cmd = f"{cleanup_cmd} && sudo docker-compose up -d --build --force-recreate --no-deps tencent-bot"
        elif TARGET == 'system-worker':
            cleanup_cmd = "sudo docker rm -f botmatrix-system-worker || true"
            if NO_SOURCE:
                docker_cmd = f"{cleanup_cmd} && sudo docker build -f src/BotWorker/Dockerfile.prod -t botmatrix-system-worker src/BotWorker/ && sudo docker-compose up -d --force-recreate --no-deps system-worker"
            else:
                docker_cmd = f"{cleanup_cmd} && sudo docker-compose up -d --build --force-recreate --no-deps system-worker"
        elif TARGET == 'no-wx':
            services_to_up = "bot-manager system-worker tencent-bot dingtalk-bot feishu-bot telegram-bot discord-bot slack-bot kook-bot email-bot wecom-bot"
            containers_to_clean = "botmatrix-manager botmatrix-system-worker tencent-bot dingtalk-bot feishu-bot telegram-bot discord-bot slack-bot kook-bot email-bot wecom-bot"
            cleanup_cmd = f"sudo docker rm -f {containers_to_clean} || true"
            
            if NO_SOURCE:
                # Build all Go bots using their Dockerfile.prod
                build_cmds = [
                    "sudo docker build -f src/BotNexus/Dockerfile.prod -t botmatrix-manager src/BotNexus/",
                    "sudo docker build -f src/BotWorker/Dockerfile.prod -t botmatrix-system-worker src/BotWorker/",
                    "sudo docker build -f src/TencentBot/Dockerfile.prod -t botmatrix-tencent-bot src/TencentBot/",
                    "sudo docker build -f src/WxBotGo/Dockerfile.prod -t botmatrix-wxbot-go src/WxBotGo/",
                    "sudo docker build -f src/DingTalkBot/Dockerfile.prod -t botmatrix-dingtalk-bot src/DingTalkBot/",
                    "sudo docker build -f src/FeishuBot/Dockerfile.prod -t botmatrix-feishu-bot src/FeishuBot/",
                    "sudo docker build -f src/TelegramBot/Dockerfile.prod -t botmatrix-telegram-bot src/TelegramBot/",
                    "sudo docker build -f src/DiscordBot/Dockerfile.prod -t botmatrix-discord-bot src/DiscordBot/",
                    "sudo docker build -f src/SlackBot/Dockerfile.prod -t botmatrix-slack-bot src/SlackBot/",
                    "sudo docker build -f src/KookBot/Dockerfile.prod -t botmatrix-kook-bot src/KookBot/",
                    "sudo docker build -f src/EmailBot/Dockerfile.prod -t botmatrix-email-bot src/EmailBot/",
                    "sudo docker build -f src/WeComBot/Dockerfile.prod -t botmatrix-wecom-bot src/WeComBot/"
                ]
                docker_cmd = f"{cleanup_cmd} && {' && '.join(build_cmds)} && sudo docker-compose up -d --force-recreate {services_to_up}"
            elif DOCKER_FIX_MODE:
                docker_cmd = f"{cleanup_cmd} && sudo docker build -f src/BotNexus/Dockerfile.minimal -t botmatrix-manager src/BotNexus/ && sudo docker-compose up -d --force-recreate {services_to_up}"
            else:
                docker_cmd = f"{cleanup_cmd} && sudo docker-compose up -d --build --force-recreate {services_to_up}"
        else:
            if NO_SOURCE:
                 # Full deployment without source
                 build_cmds = [
                     "sudo docker build -f src/BotNexus/Dockerfile.prod -t botmatrix-manager src/BotNexus/",
                     "sudo docker build -f src/BotWorker/Dockerfile.prod -t botmatrix-system-worker src/BotWorker/",
                     "sudo docker build -f src/TencentBot/Dockerfile.prod -t botmatrix-tencent-bot src/TencentBot/",
                     "sudo docker build -f src/WxBotGo/Dockerfile.prod -t botmatrix-wxbot-go src/WxBotGo/",
                     "sudo docker build -f src/DingTalkBot/Dockerfile.prod -t botmatrix-dingtalk-bot src/DingTalkBot/",
                     "sudo docker build -f src/FeishuBot/Dockerfile.prod -t botmatrix-feishu-bot src/FeishuBot/",
                     "sudo docker build -f src/TelegramBot/Dockerfile.prod -t botmatrix-telegram-bot src/TelegramBot/",
                     "sudo docker build -f src/DiscordBot/Dockerfile.prod -t botmatrix-discord-bot src/DiscordBot/",
                     "sudo docker build -f src/SlackBot/Dockerfile.prod -t botmatrix-slack-bot src/SlackBot/",
                     "sudo docker build -f src/KookBot/Dockerfile.prod -t botmatrix-kook-bot src/KookBot/",
                     "sudo docker build -f src/EmailBot/Dockerfile.prod -t botmatrix-email-bot src/EmailBot/",
                     "sudo docker build -f src/WeComBot/Dockerfile.prod -t botmatrix-wecom-bot src/WeComBot/"
                 ]
                 docker_cmd = f"sudo docker-compose down --remove-orphans && {' && '.join(build_cmds)} && sudo docker-compose up -d --force-recreate"
            elif DOCKER_FIX_MODE:
                docker_cmd = "sudo docker-compose down --remove-orphans && sudo docker build -f src/BotNexus/Dockerfile.minimal -t botmatrix-manager src/BotNexus/ && sudo docker-compose up -d --force-recreate"
            else:
                docker_cmd = "sudo docker-compose down --remove-orphans && sudo docker-compose up -d --build --force-recreate"

    remote_cmds.append(docker_cmd)
    remote_cmds.extend([
        f"echo '--> Deployment finished. Checking status...'",
        f"sudo docker-compose ps",
        f"echo '--> Logs (tail 20):'",
        f"sudo docker-compose logs --tail=20"
    ])

    full_cmd = " && ".join(remote_cmds)
    ssh_cmd = f"ssh -t {USERNAME}@{SERVER_IP} \"{full_cmd}\""
    
    run_command(ssh_cmd)
    
    print("\n========================================")
    print("   Deployment Successful!")
    print(f"   Docker Fix Mode: {'Enabled' if DOCKER_FIX_MODE else 'Disabled'}")
    print(f"   Dashboard: http://{SERVER_IP}:5000")
    print(f"   Overmind:  http://{SERVER_IP}:5000/overmind/")
    print(f"   WxBot WebUI: http://{SERVER_IP}:5001")
    print(f"   Gateway:   ws://{SERVER_IP}:3111")
    print("========================================")

if __name__ == "__main__":
    main()
