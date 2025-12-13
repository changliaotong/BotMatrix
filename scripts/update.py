import os
import sys
import subprocess
import argparse
import tarfile
import time

# 添加项目根目录到路径
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# 配置信息 (与 deploy.py 保持一致)
SERVER_IP = "192.168.0.167"
USERNAME = "derlin"
REMOTE_DEPLOY_DIR = "/opt/BotMatrix"
REMOTE_TMP_DIR = "/tmp"

# 是否使用 sudo
USE_SUDO = True
SUDO_CMD = "sudo " if USE_SUDO else ""

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
    parser = argparse.ArgumentParser(description="Fast update specific files to the server using tar compression.")
    parser.add_argument('files', metavar='FILE', type=str, nargs='*', help='Files to update (default: all .py files)')
    parser.add_argument('--restart', action='store_true', default=True, help='Restart the container after update (default: True)')
    parser.add_argument('--no-restart', dest='restart', action='store_false', help='Do not restart the container')

    args = parser.parse_args()

    # 1. 确定要上传的文件
    files_to_upload = args.files
    if not files_to_upload:
        print("No files specified. Auto-detecting project files...")
        
        # 主要项目目录和文件
        additional_items = [
            'WxBot',        # Python Worker & Core
            'BotNexus',     # Go Manager
            'scripts',      # 运维脚本
            'docs',         # 文档
            'docker-compose.yml', 
            'README.md'
        ]
        
        for item in additional_items:
            if os.path.exists(item):
                files_to_upload.append(item)
    
    if not files_to_upload:
        print("No files to upload.")
        return

    print(f"Files to update ({len(files_to_upload)}): {files_to_upload}")

    # 2. 本地打包
    tar_filename = "update_package.tar.gz"
    
    # 获取项目根目录
    root_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    # 切换到根目录执行打包，确保路径结构正确
    current_dir = os.getcwd()
    os.chdir(root_dir)
    
    print(f"\n[Step 1/4] Compressing files to {tar_filename}...")
    try:
        with tarfile.open(tar_filename, "w:gz") as tar:
            for file in files_to_upload:
                if os.path.exists(file):
                    tar.add(file)
                else:
                    print(f"Warning: File {file} not found, skipping.")
    except Exception as e:
        print(f"Error compressing files: {e}")
        sys.exit(1)
        
    # 恢复目录
    os.chdir(current_dir)
    
    # 压缩包现在在根目录下，需要构建完整路径
    tar_filepath = os.path.join(root_dir, tar_filename)

    # 3. 上传压缩包
    print("\n[Step 2/4] Uploading package to server...")
    upload_cmd = f"scp {tar_filepath} {USERNAME}@{SERVER_IP}:{REMOTE_TMP_DIR}/{tar_filename}"
    run_command(upload_cmd)

    # 删除本地压缩包
    try:
        os.remove(tar_filepath)
    except:
        pass

    # 4. 服务器端解压并部署
    print("\n[Step 3/4] Extracting on server...")
    
    # 确保远程目录存在
    ensure_dir_cmd = f"sudo mkdir -p {REMOTE_DEPLOY_DIR} && sudo chown {USERNAME}:{USERNAME} {REMOTE_DEPLOY_DIR}"
    # 这里我们直接执行，如果需要输入密码，-t 会处理
    run_command(f'ssh -t {USERNAME}@{SERVER_IP} "{ensure_dir_cmd}"')

    # 命令逻辑：解压到临时目录 -> 移动/覆盖到部署目录 -> 设置权限 -> 清理压缩包
    # 使用 tar -mxzf 覆盖解压
    remote_cmds = [
        f"tar -mxzf {REMOTE_TMP_DIR}/{tar_filename} -C {REMOTE_DEPLOY_DIR}",
        f"rm {REMOTE_TMP_DIR}/{tar_filename}"
    ]
    
    remote_cmd_str = " && ".join(remote_cmds)
    # 如果需要sudo，可能需要调整命令结构，这里假设用户对部署目录有写权限
    if USE_SUDO:
         remote_cmd_str = f"{SUDO_CMD}tar -mxzf {REMOTE_TMP_DIR}/{tar_filename} -C {REMOTE_DEPLOY_DIR} && rm {REMOTE_TMP_DIR}/{tar_filename}"

    ssh_cmd = f'ssh -t {USERNAME}@{SERVER_IP} "{remote_cmd_str}"'
    run_command(ssh_cmd)

    # 5. 重启容器
    if args.restart:
        print("\n[Step 4/4] Rebuilding and Restarting containers...")
        # 因为代码被打包进镜像，必须重新构建才能生效
        # 使用 up -d --build 可以智能构建并重启
        restart_cmd = f'ssh -t {USERNAME}@{SERVER_IP} "cd {REMOTE_DEPLOY_DIR} && {SUDO_CMD}docker-compose up -d --build"'
        run_command(restart_cmd)
        print("\nUpdate and Restart SUCCESS!")
    else:
        print("\nUpdate SUCCESS! (Container not restarted)")

if __name__ == "__main__":
    main()
