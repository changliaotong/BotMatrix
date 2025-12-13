import asyncio
import json
import websockets

URI = 'ws://192.168.0.167:3111'

async def main():
    async with websockets.connect(URI) as ws:
        print('[client] connected')
        # for i in range(3):
        #    msg = await ws.recv()
        #    print('[event]', msg)
        action = {
            'action': 'send_group_msg',
            'params': {'group_id': 20001, 'message': '服务器动作测试'},
            'echo': 'e1'
        }
        await ws.send(json.dumps(action, ensure_ascii=False))
        res = await ws.recv()
        print('[action_result]', res)

if __name__ == '__main__':
    asyncio.run(main())
