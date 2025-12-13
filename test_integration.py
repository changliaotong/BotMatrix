import requests
import json
import time

BOTNEXUS_API = "http://127.0.0.1:5000/api"

def test_get_bots():
    print("[-] Testing GET /api/bots...")
    try:
        # Add auth header (admin for full access)
        headers = {
            "Authorization": "Bearer admin"
        }
        resp = requests.get(f"{BOTNEXUS_API}/bots", headers=headers)
        print(f"Status: {resp.status_code}")
        print(f"Response: {resp.text}")
        data = resp.json()
        if isinstance(data, list) and len(data) > 0:
            print(f"[+] Found {len(data)} bots connected.")
            # Find the one with self_id 1098299491 (our wxbot) or just take the first
            target = next((b['self_id'] for b in data if b.get('self_id') == "1098299491"), None)
            if not target and data:
                 target = data[0].get('self_id')
            return target
        elif isinstance(data, dict) and data.get("bots"):
            print(f"[+] Found {len(data['bots'])} bots connected.")
            return list(data['bots'].keys())[0]
        else:
            print("[!] No bots found.")
            return None
    except Exception as e:
        print(f"[!] Error: {e}")
        return None

def test_send_msg(self_id):
    if not self_id:
        print("[!] Skipping send_msg test (no self_id)")
        return
        
    print(f"[-] Testing Send Message via {self_id}...")
    # Send to filehelper as a safe test
    payload = {
        "self_id": self_id,
        "action": "send_private_msg",
        "params": {
            "user_id": "filehelper", 
            "message": "Hello from BotNexus Test Script"
        }
    }
    
    try:
        headers = {
            "Authorization": "Bearer admin"
        }
        resp = requests.post(f"{BOTNEXUS_API}/action", json=payload, headers=headers)
        print(f"Status: {resp.status_code}")
        print(f"Response: {resp.text}")
    except Exception as e:
        print(f"[!] Error: {e}")

if __name__ == "__main__":
    print("=== Starting BotNexus Integration Test ===")
    target_bot_id = test_get_bots()
    if target_bot_id:
        test_send_msg(target_bot_id)
    print("=== Test Finished ===")
