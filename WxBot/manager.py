# coding: utf-8
import os
import time
import threading
from onebot import BotManager, start_web_ui

def main():
    print("[Manager] Starting centralized management platform...")
    
    # Initialize BotManager (loads config, starts WebSocket gateway)
    # This will NOT start any local bots unless explicitly configured, 
    # but we want this process to be PURE manager if possible.
    # However, BotManager logic currently loads bots from config.
    # We might need to adjust config or BotManager to NOT load local bots in this mode,
    # OR we just assume the config for this container won't have "bots" list,
    # OR we just let it be, but users should move bot configs to worker containers.
    
    # For now, we instantiate BotManager which starts the WS server on port 3001 (default)
    manager = BotManager()
    
    # Start WebUI
    try:
        print("[Manager] Starting WebUI on port 5000...")
        start_web_ui(manager, port=5000)
    except Exception as e:
        print(f"[Manager] WebUI startup failed: {e}")

    # Keep alive
    print("[Manager] Service running.")
    while True:
        time.sleep(1)

if __name__ == "__main__":
    main()
