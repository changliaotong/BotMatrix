# coding: utf-8
import os
import time
import json
import asyncio
import websockets
import redis.asyncio as redis
from SQLConn import SQLConn

class RedisWorker:
    def __init__(self, manager_url, redis_url="redis://:redis_zsYik8@192.168.0.126:6379/0"):
        self.manager_url = manager_url
        if "?" not in self.manager_url:
            self.manager_url += "?role=worker"
        else:
            self.manager_url += "&role=worker"
            
        self.rdb = redis.from_url(redis_url, decode_responses=True)
        self.running = True

    async def sync_permissions(self):
        """Syncs Bot Admin permissions from SQL Server to Redis"""
        print("[Worker] Syncing permissions from SQL Server...")
        try:
            # Query Member table for BotUin <-> AdminId -> User.WxId
            sql = """
                SELECT DISTINCT m.BotUin, u.WxId 
                FROM [Member] m 
                INNER JOIN [User] u ON m.AdminId = u.Id 
                WHERE m.BotUin IS NOT NULL AND m.BotUin <> '' 
                  AND u.WxId IS NOT NULL AND u.WxId <> ''
            """
            # SQLConn is synchronous, so we run it directly (blocking briefly is ok for this worker)
            # In production, use run_in_executor
            rows = SQLConn.QueryDict(sql)
            
            if not rows:
                print("[Worker] No permission records found in DB.")
                return

            # Pipeline Redis updates
            pipe = self.rdb.pipeline()
            
            # Clear old user bot mappings? 
            # To be safe, we might want to clear specific keys or just overwrite.
            # For now, we accumulate/add. Ideally we should have a way to remove stale ones.
            # But let's just add for now.
            
            count = 0
            for row in rows:
                bot_id = str(row['BotUin'])
                user_id = str(row['WxId'])
                
                # 1. Add user to auth:users
                pipe.sadd("auth:users", user_id)
                # 2. Add bot to auth:user:{user_id}:bots
                pipe.sadd(f"auth:user:{user_id}:bots", bot_id)
                # 3. Set owner info for UI
                pipe.hset("auth:bot_owners", bot_id, user_id)
                # 4. Set default password for user if not exists (for login)
                pipe.hsetnx(f"auth:user:{user_id}:pwd", "password", "123456")
                
                count += 1
                
            await pipe.execute()
            print(f"[Worker] Synced {count} permission records to Redis.")
            
        except Exception as e:
            print(f"[Worker] Sync Error: {e}")

    async def sync_loop(self):
        while self.running:
            await self.sync_permissions()
            await asyncio.sleep(300)

    async def start(self):
        print(f"[Worker] connecting to Redis at {self.rdb.connection_pool.connection_kwargs.get('host')}...")
        try:
            await self.rdb.ping()
            print("[Worker] Redis connected!")
            # Initial Sync
            asyncio.create_task(self.sync_loop())
        except Exception as e:
            print(f"[Worker] Redis connection failed: {e}. Running in stateless mode (or failing).")

        while self.running:
            try:
                print(f"[Worker] Connecting to Manager at {self.manager_url}...")
                async with websockets.connect(self.manager_url) as ws:
                    print("[Worker] Connected to Manager.")
                    self.ws = ws
                    
                    async for msg in ws:
                        try:
                            data = json.loads(msg)
                            if "post_type" in data:
                                await self.handle_event(data)
                        except Exception as e:
                            print(f"[Worker] Error: {e}")
                            
            except Exception as e:
                print(f"[Worker] Connection lost: {e}. Retrying...")
                await asyncio.sleep(5)

    async def handle_event(self, event):
        if event.get("post_type") != "message":
            return

        user_id = event.get("user_id")
        msg = event.get("message", "")
        
        print(f"[Worker] Msg from {user_id}: {msg}")

        # --- Redis Business Logic ---
        
        # 1. Conversation State Management
        state_key = f"user:state:{user_id}"
        current_state = await self.rdb.get(state_key)
        
        reply = ""
        
        if msg == "reset":
            await self.rdb.delete(state_key)
            reply = "State reset."
            
        elif current_state is None:
            if msg == "start":
                await self.rdb.set(state_key, "step1", ex=300) # Expire in 5 mins
                reply = "Welcome! You are at Step 1. Send 'next' to proceed."
            else:
                reply = "Send 'start' to begin."
                
        elif current_state == "step1":
            if msg == "next":
                await self.rdb.set(state_key, "step2", ex=300)
                reply = "Great! You are at Step 2. Send 'finish' to complete."
            else:
                reply = "You are at Step 1. Please send 'next'."
                
        elif current_state == "step2":
            if msg == "finish":
                await self.rdb.delete(state_key)
                reply = "Congratulations! Workflow completed."
            else:
                reply = "You are at Step 2. Please send 'finish'."

        if reply:
            await self.send_msg(user_id, reply)

    async def send_msg(self, user_id, message):
        if self.ws:
            await self.ws.send(json.dumps({
                "action": "send_private_msg",
                "params": {
                    "user_id": user_id,
                    "message": message
                }
            }))

if __name__ == "__main__":
    worker = RedisWorker("ws://localhost:3001")
    asyncio.run(worker.start())
