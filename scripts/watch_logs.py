import argparse
import sys
from datetime import datetime
import subprocess

# Configuration
DEFAULT_WS_URL = "ws://192.168.0.167:3001"
DEFAULT_TOKEN = "" # Add token if needed, or leave blank if no auth required for subscriber (BotNexus seems to require token for subscriber)

# Check BotNexus code:
# func serveSubscriber(m *Manager, w http.ResponseWriter, r *http.Request) {
# 	// Auth
# 	token := r.URL.Query().Get("token")
#   ...
#   if user == nil { http.Error(w, "Unauthorized", http.StatusUnauthorized) ... }

# Since we don't have a token generator handy in this script, we might face auth issues if we try to connect as subscriber.
# However, we can connect as a "worker" or "bot" to see some messages, but subscriber is best for logs.
# Let's assume the user has a token or we can bypass auth for localhost/debug if configured.
# Actually, looking at BotNexus/main.go, subscriber auth IS required.

# Alternative: SSH and tail logs.
# This script will wrap the SSH command to tail logs, but with highlighting.

import subprocess

class Colors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

def print_log_line(line):
    line = line.strip()
    if not line:
        return

    # Timestamp
    # Docker logs: 2023-10-27T10:00:00.000000000Z ...
    
    # Highlight known services
    service = ""
    if "bot-manager" in line:
        service = "NEXUS"
        color = Colors.OKCYAN
    elif "tencent-bot" in line:
        service = "TENCENT"
        color = Colors.OKGREEN
    elif "wxbot" in line:
        service = "WXBOT"
        color = Colors.OKBLUE
    elif "system-worker" in line:
        service = "WORKER"
        color = Colors.HEADER
    else:
        service = "OTHER"
        color = Colors.ENDC

    # Highlight Keywords
    content_color = Colors.ENDC
    if "ERROR" in line or "fail" in line.lower():
        content_color = Colors.FAIL
    elif "WARN" in line:
        content_color = Colors.WARNING
    elif "DEBUG" in line:
        content_color = Colors.ENDC # Normal for debug
    elif "[NEXUS-MSG]" in line:
        content_color = Colors.BOLD + Colors.OKGREEN # Special highlight for our new logs

    print(f"{color}[{service}]{Colors.ENDC} {content_color}{line}{Colors.ENDC}")

def main():
    parser = argparse.ArgumentParser(description='Live Log Watcher with Highlighting')
    parser.add_argument('--ip', default='192.168.0.167', help='Server IP')
    parser.add_argument('--user', default='derlin', help='SSH User')
    parser.add_argument('--filter', help='Grep filter', default='')
    args = parser.parse_args()

    print(f"Connecting to {args.user}@{args.ip} to watch logs...")
    
    # Use double quotes for the remote command to avoid issues on Windows
    cmd = f'ssh -t {args.user}@{args.ip} "cd /opt/BotMatrix && sudo docker-compose logs -f --tail=50"'
    
    if args.filter:
        print(f"Filtering for: {args.filter}")
    
    process = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, encoding='utf-8', errors='replace')

    try:
        while True:
            line = process.stdout.readline()
            if not line and process.poll() is not None:
                break
            if line:
                if args.filter and args.filter not in line:
                    continue
                print_log_line(line)
    except KeyboardInterrupt:
        print("\nStopping log watcher...")
        process.terminate()

if __name__ == "__main__":
    main()
