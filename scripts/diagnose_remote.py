import subprocess
import sys

SERVER_IP = "192.168.0.167"
USERNAME = "derlin"

def run_ssh_cmd(cmd):
    ssh_cmd = f"ssh {USERNAME}@{SERVER_IP} \"{cmd}\""
    print(f"--- Executing: {cmd} ---")
    try:
        subprocess.run(ssh_cmd, shell=True)
    except Exception as e:
        print(f"Error: {e}")

def main():
    print(f"Diagnosing WxBot on {SERVER_IP}...\n")

    # 1. Check if container is running
    print("1. Checking Container Status:")
    run_ssh_cmd("sudo docker ps -a | grep wxbot")
    print("\n")

    # 2. Check if port is listening
    print("2. Checking Port 3111:")
    run_ssh_cmd("sudo netstat -tulpn | grep 3111 || echo 'Port 3111 not found in netstat'")
    print("\n")

    # 3. Check logs
    print("3. Checking Container Logs (Last 50 lines):")
    run_ssh_cmd("sudo docker logs --tail 50 wxbot")
    print("\n")

if __name__ == "__main__":
    main()
