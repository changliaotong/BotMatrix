import asyncio
import json
import websockets

URI = 'ws://127.0.0.1:3001'

async def main():
    async with websockets.connect(URI) as ws:
        print('[client] connected')
        
        # 接收并打印初始事件 (lifecycle 等)
        # 注意：实际业务中应该在一个死循环中不断 recv
        
        # 示例 1: 发送给个人微信 (假设 self_id=1098299491)
        # 个人微信通常使用映射后的数字 group_id
        action_wechat = {
            'action': 'send_group_msg',
            'params': {
                'self_id': 1098299491, 
                'group_id': 20001, 
                'message': 'Hello via Personal WeChat'
            },
            'echo': 'e1'
        }
        print(f"[send] WeChat: {json.dumps(action_wechat)}")
        await ws.send(json.dumps(action_wechat))

        # 示例 2: 发送给企业微信 (假设 agentid=1000001)
        # 企业微信直接使用字符串 chatid
        action_work = {
            'action': 'send_group_msg',
            'params': {
                'self_id': 1000001,
                'group_id': 'wrHp28CQAA...', # 替换为真实的 chatid
                'message': 'Hello via WXWork'
            },
            'echo': 'e2'
        }
        print(f"[send] WXWork: {json.dumps(action_work)}")
        await ws.send(json.dumps(action_work))

        # 示例 3: 发送给钉钉 (Webhook)
        # 钉钉 Webhook 模式下 group_id 被忽略 (因为 Webhook 已绑定群)
        # 但我们需要正确的 self_id 来选中该 Bot
        # (self_id 是根据 token 生成的 hash，实际开发中建议在 config 中指定固定 id 或先获取列表)
        
        # 持续接收响应
        try:
            while True:
                res = await ws.recv()
                data = json.loads(res)
                if 'echo' in data:
                    print(f"[result] echo={data['echo']} status={data.get('status')}")
                else:
                    print(f"[event] {data.get('post_type')} self_id={data.get('self_id')}")
        except KeyboardInterrupt:
            pass

if __name__ == '__main__':
    asyncio.run(main())
