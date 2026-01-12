import asyncio
import json
import time
import websockets
import logging
import os
import sys
import argparse
from datetime import datetime

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class BotSimulator:
    def __init__(self, config_path="test_config.json", state_path="test_state.json", auto=False):
        with open(config_path, 'r', encoding='utf-8') as f:
            self.config = json.load(f)
        
        self.state_path = state_path
        self.passed_tests = self.load_state()
        self.auto = auto
        
        self.uri = self.config['server']['uri']
        self.bot_id = self.config['server']['bot_id']
        self.platform = self.config['server']['platform']
        self.user_id = self.config['tester']['user_id']
        self.nickname = self.config['tester']['nickname']
        
        self.websocket = None
        self.test_results = []
        self.pending_responses = asyncio.Queue()

    def load_state(self):
        if os.path.exists(self.state_path):
            with open(self.state_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        return []

    def save_state(self):
        passed = [r['name'] for r in self.test_results if r['success']]
        # 合并已有的通过记录
        total_passed = list(set(self.passed_tests + passed))
        with open(self.state_path, 'w', encoding='utf-8') as f:
            json.dump(total_passed, f, ensure_ascii=False, indent=4)

    async def connect(self):
        headers = {
            "X-Self-ID": self.bot_id,
            "X-Platform": self.platform
        }
        logger.info(f"Connecting to {self.uri} as Bot {self.bot_id}...")
        try:
            self.websocket = await websockets.connect(self.uri, additional_headers=headers)
            logger.info("Connected to BotNexus")
        except Exception as e:
            logger.error(f"Connection failed: {e}")
            raise

    async def handle_action(self, action):
        echo = action.get("echo")
        action_name = action.get("action")
        params = action.get("params", {})
        
        response = {
            "status": "ok",
            "retcode": 0,
            "data": {},
            "echo": echo
        }
        
        if action_name == "get_login_info":
            response["data"] = {"user_id": self.bot_id, "nickname": "Virtual Bot"}
        elif action_name == "get_group_list":
            response["data"] = [{"group_id": 123456789, "group_name": "Test Group"}]
        elif action_name in ["send_msg", "send_private_msg", "send_group_msg"]:
            message = params.get("message")
            logger.info(f">>> Bot Output: {message}")
            await self.pending_responses.put(message)
            response["data"] = {"message_id": int(time.time())}
        
        await self.websocket.send(json.dumps(response))

    async def send_message(self, content, msg_type="private", group_id=None):
        message_event = {
            "time": int(time.time()),
            "self_id": self.bot_id,
            "post_type": "message",
            "message_type": msg_type,
            "sub_type": "friend" if msg_type == "private" else "normal",
            "message_id": int(time.time()),
            "user_id": self.user_id,
            "message": content,
            "raw_message": content,
            "font": 0,
            "sender": {
                "user_id": self.user_id,
                "nickname": self.nickname,
                "card": self.nickname if msg_type == "group" else ""
            }
        }
        if msg_type == "group":
            message_event["group_id"] = group_id or 123456789
            
        await self.websocket.send(json.dumps(message_event))
        target = f"Group {message_event.get('group_id')}" if msg_type == "group" else "Private"
        logger.info(f"<<< User Input [{target}]: {content}")

    async def run_test_case(self, case):
        if case['name'] in self.passed_tests:
            logger.info(f"Skipping passed test: {case['name']}")
            return

        if not self.auto:
            print(f"\nReady to test: {case['name']} (Input: {case['input']})")
            user_choice = input("Run this test? [Y/n/s(kip)/q(uit)]: ").lower()
            if user_choice == 'n' or user_choice == 's':
                logger.info(f"Skipped by user: {case['name']}")
                return
            if user_choice == 'q':
                logger.info("Quitting tests...")
                return "quit"

        logger.info(f"Running Test Case: {case['name']}")
        
        # 清空之前的回复
        while not self.pending_responses.empty():
            try:
                self.pending_responses.get_nowait()
            except asyncio.QueueEmpty:
                break
                
        start_time = time.time()
        msg_type = case.get("type", "private")
        group_id = case.get("group_id")
        await self.send_message(case['input'], msg_type, group_id)
        
        try:
            # 等待回复，带超时
            timeout = case.get('timeout', 10)
            success = True
            reason = ""
            
            try:
                response = await asyncio.wait_for(self.pending_responses.get(), timeout=timeout)
                
                if case.get('expected_no_response'):
                    success = False
                    reason = f"Expected NO response, but got '{response}'"
                else:
                    # 验证结果
                    if 'expected' in case and response != case['expected']:
                        success = False
                        reason = f"Expected '{case['expected']}', got '{response}'"
                    elif 'expected_contains' in case:
                        for item in case['expected_contains']:
                            if item not in response:
                                success = False
                                reason = f"Expected to contain '{item}', but not found in '{response}'"
                                break
                    elif 'expected_not_empty' in case and not response:
                        success = False
                        reason = "Expected non-empty response"
            except asyncio.TimeoutError:
                if case.get('expected_no_response'):
                    success = True
                    logger.info(f"PASS (No response as expected): {case['name']}")
                else:
                    success = False
                    reason = "Timeout"

            if success:
                if not case.get('expected_no_response'):
                    logger.info(f"PASS: {case['name']}")
            else:
                logger.error(f"FAIL: {case['name']} - {reason}")
            
            self.test_results.append({
                "name": case['name'],
                "success": success,
                "reason": reason,
                "time": time.time() - start_time
            })

        except Exception as e:
            logger.error(f"Error during test {case['name']}: {e}")
            self.test_results.append({
                "name": case['name'],
                "success": False,
                "reason": str(e),
                "time": 0
            })

    async def start(self):
        await self.connect()
        
        # 启动消息监听任务
        async def listen():
            try:
                async for message in self.websocket:
                    data = json.loads(message)
                    if "action" in data:
                        await self.handle_action(data)
            except Exception as e:
                if not isinstance(e, asyncio.CancelledError):
                    logger.error(f"Listen error: {e}")

        listen_task = asyncio.create_task(listen())
        
        # 等待初始化完成
        await asyncio.sleep(2)
        
        # 运行所有测试用例
        try:
            for case in self.config['test_cases']:
                res = await self.run_test_case(case)
                if res == "quit":
                    break
                await asyncio.sleep(1) # 测试间隙
        finally:
            # 清理：确保机器人处于开启状态
            logger.info("Cleaning up: Ensuring bot is Power On...")
            await self.send_message("开机", "group")
            await asyncio.sleep(1)
            await self.send_message("开启 积分系统", "group")
            await asyncio.sleep(1)
            
        # 输出报告
        self.print_report()
        self.save_state()
        
        listen_task.cancel()
        try:
            await listen_task
        except asyncio.CancelledError:
            pass
        await self.websocket.close()

    def print_report(self):
        print("\n" + "="*50)
        print(f"TEST REPORT - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print("="*50)
        new_passed = sum(1 for r in self.test_results if r['success'])
        total_run = len(self.test_results)
        
        if total_run == 0:
            print("No tests were run.")
        else:
            for r in self.test_results:
                status = "✓ PASS" if r['success'] else "✗ FAIL"
                print(f"[{status}] {r['name']} ({r['time']:.2f}s)")
                if not r['success']:
                    print(f"      Reason: {r['reason']}")
        
        print("-" * 50)
        print(f"Newly Passed: {new_passed}/{total_run}")
        print(f"Total Passed (including previous): {len(set(self.passed_tests + [r['name'] for r in self.test_results if r['success']]))}")
        print("="*50 + "\n")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Bot Simulator")
    parser.add_argument("--auto", action="store_true", help="Run tests without user confirmation")
    parser.add_argument("--reset", action="store_true", help="Reset test state (rerun all tests)")
    parser.add_argument("--config", type=str, default="test_config.json", help="Path to config file")
    args = parser.parse_args()

    state_file = "test_state.json"
    if args.reset and os.path.exists(state_file):
        os.remove(state_file)

    sim = BotSimulator(config_path=args.config, auto=args.auto)
    try:
        asyncio.run(sim.start())
    except KeyboardInterrupt:
        pass
    except Exception as e:
        logger.error(f"Fatal error: {e}")
