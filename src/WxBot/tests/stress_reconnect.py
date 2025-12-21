import asyncio
import websockets
import time
import subprocess
import sys

# Configuration
WS_URI = "ws://192.168.0.167:3111"
RESTART_CMD = 'ssh -t derlin@192.168.0.167 "cd /opt/wxbot && docker-compose restart wxbot"'
TEST_CYCLES = 100  # Number of restart cycles

async def run_restart_cmd():
    print("    [Cmd] Sending restart command...")
    # Use create_subprocess_shell for async execution
    proc = await asyncio.create_subprocess_shell(
        RESTART_CMD,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE
    )
    stdout, stderr = await proc.communicate()
    if proc.returncode != 0:
        print(f"    [Cmd] Restart failed: {stderr.decode()}")
    else:
        print("    [Cmd] Restart command finished.")

async def monitor_cycle(cycle_num):
    print(f"\n=== Cycle {cycle_num}/{TEST_CYCLES} ===")
    
    disconnect_time = 0
    reconnect_time = 0
    
    # 1. Connect and hold
    print("[1] Connecting...")
    try:
        # ping_interval=None to avoid internal ping timeouts confusing the logic, 
        # but here we want to detect disconnects. Docker stop usually closes socket cleanly.
        async with websockets.connect(WS_URI, ping_interval=None) as ws:
            print("    Connected. Waiting 1s before restart...")
            await asyncio.sleep(1)
            
            # 2. Trigger restart in background
            restart_task = asyncio.create_task(run_restart_cmd())
            
            # 3. Wait for disconnection
            print("    Waiting for disconnection...")
            try:
                # Wait for connection to close
                await ws.wait_closed()
            except Exception:
                pass
            
            disconnect_time = time.time()
            print(f"    Disconnected at {time.strftime('%H:%M:%S', time.localtime(disconnect_time))}")
            
    except Exception as e:
        print(f"    Connection error (might be already down): {e}")
        disconnect_time = time.time()

    # 4. Loop to reconnect
    print("[2] Attempting to reconnect...")
    while True:
        try:
            async with websockets.connect(WS_URI, open_timeout=2) as ws:
                reconnect_time = time.time()
                print(f"    Reconnected at {time.strftime('%H:%M:%S', time.localtime(reconnect_time))}")
                break
        except Exception:
            await asyncio.sleep(0.5)
            
    downtime = reconnect_time - disconnect_time
    print(f"    [Result] Downtime duration: {downtime:.2f} seconds")

async def main():
    print(f"Starting stress test for {WS_URI}")
    for i in range(1, TEST_CYCLES + 1):
        await monitor_cycle(i)
        await asyncio.sleep(2)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nTest aborted.")
