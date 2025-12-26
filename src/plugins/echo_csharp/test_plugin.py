#!/usr/bin/env python3
import json
import subprocess
import sys

def test_csharp_plugin():
    # 启动C#插件
    proc = subprocess.Popen([
        "src/plugins/echo_csharp/bin/Release/net6.0/win-x64/publish/echo_csharp.exe"
    ], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    
    # 测试用例
    test_cases = [
        {"input": "hello", "expected": "C# Echo: hello"},
        {"input": "你好", "expected": "C# Echo: 你好"},
        {"input": "!@#$%^&*()", "expected": "C# Echo: !@#$%^&*()"}
    ]
    
    print("Testing C# Echo Plugin...")
    print("=" * 50)
    
    for i, test_case in enumerate(test_cases):
        # 构造事件消息
        event = {
            "id": f"test_{i}",
            "type": "event",
            "name": "on_message",
            "payload": {
                "text": test_case["input"],
                "from": "group",
                "group_id": "test_group_123",
                "user_id": "test_user_456"
            }
        }
        
        # 发送事件
        proc.stdin.write(json.dumps(event) + "\n")
        proc.stdin.flush()
        
        # 获取响应
        response_line = proc.stdout.readline()
        response = json.loads(response_line)
        
        # 检查响应
        if response["ok"] and len(response["actions"]) > 0:
            action = response["actions"][0]
            if action["text"] == test_case["expected"]:
                print(f"Test {i+1}: PASS - Input: '{test_case['input']}' -> Output: '{action['text']}'")
            else:
                print(f"Test {i+1}: FAIL - Expected: '{test_case['expected']}', Got: '{action['text']}'")
        else:
            print(f"Test {i+1}: FAIL - Invalid response")
    
    # 关闭进程
    proc.stdin.close()
    proc.terminate()
    proc.wait()

if __name__ == "__main__":
    test_csharp_plugin()