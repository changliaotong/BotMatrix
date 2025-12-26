#!/usr/bin/env python3
import json
import sys
import time

class EchoPlugin:
    def __init__(self):
        self.running = True
    
    def handle_event(self, event):
        if event["type"] == "event":
            if event["name"] == "on_message":
                payload = event["payload"]
                text = payload.get("text", "")
                target = payload.get("from", "")
                target_id = payload.get("group_id", "")
                
                response = {
                    "id": event["id"],
                    "ok": True,
                    "actions": [
                        {
                            "type": "send_message",
                            "target": target,
                            "target_id": target_id,
                            "text": f"Python Echo: {text}"
                        }
                    ]
                }
                
                json.dump(response, sys.stdout)
                sys.stdout.write("\n")
                sys.stdout.flush()
            elif event["name"] == "on_health_check":
                response = {
                    "id": event["id"],
                    "ok": True,
                    "actions": []
                }
                json.dump(response, sys.stdout)
                sys.stdout.write("\n")
                sys.stdout.flush()
    
    def run(self):
        while self.running:
            try:
                line = sys.stdin.readline()
                if not line:
                    break
                
                event = json.loads(line)
                self.handle_event(event)
            except json.JSONDecodeError:
                print("Invalid JSON input", file=sys.stderr)
            except Exception as e:
                print(f"Error: {e}", file=sys.stderr)

if __name__ == "__main__":
    plugin = EchoPlugin()
    plugin.run()