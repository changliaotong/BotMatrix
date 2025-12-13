import subprocess
import sys

SERVER_IP = "192.168.0.167"
USERNAME = "derlin"

def run_command(cmd):
    print(f"Executing: {cmd}")
    try:
        subprocess.run(cmd, shell=True)
    except Exception as e:
        print(f"Error: {e}")

def main():
    print("========================================")
    print(f"   Diagnostic Tool for {SERVER_IP}")
    print("========================================")
    
    # 1. Check Container Status
    print("\n[1] Checking Container Status:")
    run_command(f'ssh -t {USERNAME}@{SERVER_IP} "sudo docker ps -a | grep wxbot"')
    
    # 2. Check Logs
    print("\n[2] Fetching Last 50 Lines of Logs:")
    run_command(f'ssh -t {USERNAME}@{SERVER_IP} "sudo docker logs --tail 50 wxbot"')
    
    print("\nDone.")

if __name__ == "__main__":
    main()
