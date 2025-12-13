# coding: utf-8
import asyncio
import websockets
import json
import time
import random
import threading
import re
import collections
import html

import os
from bots.wxbot import WXBot
from bots.wxwork_bot import WXWorkBot
from bots.dingtalk_bot import DingTalkBot
from bots.feishu_bot import FeishuBot
from bots.telegram_bot import TelegramBot
from SQLConn import *
from color import *
from wxgroup import wx_group
from wxclient import wx_client
from common import common
from msg import msg
from web_ui import start_web_ui
from plugin_manager import PluginManager


CONFIG_FILE = "config.json"
DEFAULT_CONFIG = {
    "network": {
        "ws_server": {
            "name": "test",
            "host": "0.0.0.0",
            "port": 3001,
            "heartbeat_interval": 30000,
            "message_format": "string",
            "report_self_message": True,
            "force_push_event": True
        }
    },
    "bots": [
        # 示例：个人微信机器人
        # {"type": "wechat", "self_id": 1098299491},
        
        # 示例：企业微信机器人
        # {
        #     "type": "wxwork",
        #     "corpid": "ww...",
        #     "corpsecret": "...",
        #     "agentid": 1000002
        # }
    ]
}

# ========== OneBot WS 网关 ==========

class OneBotGateway:
    """
    WebSocket 服务器（Python 作为 OneBot 网关，等待 C# 连接）
    - 保存已连接 C# 客户端
    - 向 C# 推送 OneBot 事件
    - 接收 C# 动作（send_group_msg / send_private_msg），回调 bot 发送实际微信消息
    """
    def __init__(self, host="0.0.0.0", port=3001, config=None):
        self.host = host
        self.port = port
        self.config = config or {}
        self.clients = set()
        self.ws_bots = set()
        self.remote_bots = {} # self_id -> websocket
        self.local_bots = {}
        self._pending = {}
        self._loop = None
        self._event_queue = None
        # 消息缓冲区：用于存储无客户端连接时的消息
        self._msg_buffer = [] 
        self._max_buffer_size = 500
        # 消息缓存：用于引用回复 (message_id -> msg_info)
        self._msg_cache = collections.OrderedDict()

    def add_bot(self, bot, self_id):
        self.local_bots[str(self_id)] = bot

    def remove_bot(self, self_id):
        if str(self_id) in self.local_bots:
            del self.local_bots[str(self_id)]

    async def start(self):
        print(f"[OneBot Gateway] 启动 WS 服务器 ws://{self.host}:{self.port}")
        self._loop = asyncio.get_running_loop()
        self._event_queue = asyncio.Queue()
        asyncio.create_task(self._drain_events())
        # Disable ping_interval to avoid disconnecting clients that don't respond to PINGs (code 1006)
        async with websockets.serve(self._client_handler, self.host, self.port, ping_interval=20, ping_timeout=120):
            await asyncio.Future()  # 永不退出

    async def _drain_events(self):
        while True:
            try:
                event = await self._event_queue.get()
                await self.send_event(event)
            except Exception as e:
                if self.config.get("debug"):
                    print("[gateway] drain error:", repr(e))
            finally:
                try:
                    self._event_queue.task_done()
                except Exception:
                    pass

    async def _heartbeat_loop(self, websocket):
        interval = self.config.get("heartbeat_interval", 30000)
        while True:
            try:
                await asyncio.sleep(interval / 1000.0)
                
                self_id = 0
                if self.local_bots:
                    try:
                        self_id = int(next(iter(self.local_bots.keys())))
                    except:
                        pass

                heartbeat_event = {
                    "time": int(time.time()),
                    "self_id": self_id,
                    "post_type": "meta_event",
                    "meta_event_type": "heartbeat",
                    "status": {
                        "online": True,
                        "good": True
                    },
                    "interval": interval
                }
                await websocket.send(json.dumps(heartbeat_event, ensure_ascii=False))
            except asyncio.CancelledError:
                break
            except Exception as e:
                break

    async def _client_handler(self, websocket):
        remote_addr = getattr(websocket, 'remote_address', 'unknown')
        self.clients.add(websocket)
        if self.config.get("debug"):
            print(f"[gateway] client connected from {remote_addr}")
        
        # 启动心跳任务
        heartbeat_task = asyncio.create_task(self._heartbeat_loop(websocket))

        # OneBot 协议：连接成功后发送 lifecycle.connect 事件
        # 帮助客户端确认连接状态，避免部分客户端(如 WebUI/Dashboard)因未收到数据而超时断开
        try:
            self_id = 0
            if self.local_bots:
                # 使用第一个可用的机器人ID
                try:
                    self_id = int(next(iter(self.local_bots.keys())))
                except:
                    pass
            
            lifecycle_event = {
                "time": int(time.time()),
                "self_id": self_id,
                "post_type": "meta_event",
                "meta_event_type": "lifecycle",
                "sub_type": "connect"
            }
            await websocket.send(json.dumps(lifecycle_event, ensure_ascii=False))
        except Exception as e:
            print(f"[gateway] failed to send lifecycle event: {e}")

        # 推送缓冲的消息
        if self._msg_buffer:
            if self.config.get("debug"):
                print(f"[gateway] Flushing {len(self._msg_buffer)} buffered events to new client...")
            while self._msg_buffer:
                event = self._msg_buffer.pop(0)
                try:
                    await websocket.send(json.dumps(event, ensure_ascii=False))
                except Exception as e:
                    print(f"[gateway] failed to flush buffered event: {e}")
                    # 如果发送失败，可能连接断了，把剩下的放回去？或者直接放弃
                    break

        try:
            async for raw in websocket:
                try:
                    data = json.loads(raw)
                except Exception:
                    continue
                if isinstance(data, dict) and data.get("post_type"):
                    self.ws_bots.add(websocket)
                    self.clients.discard(websocket)
                    
                    # Track remote bot self_id
                    if "self_id" in data:
                        sid = str(data["self_id"])
                        if sid not in self.remote_bots:
                            print(f"[gateway] Registered remote bot: {sid}")
                        self.remote_bots[sid] = websocket

                    payload = json.dumps(data, ensure_ascii=False)
                    targets = list(self.clients)
                    for i, ws in enumerate(targets):
                        try:
                            await ws.send(payload)
                        except Exception as e:
                            print(f"[gateway] -> client#{i} deliver error: {repr(e)}")
                            self.clients.discard(ws)
                elif isinstance(data, dict) and (data.get("type") == "action_result" or ("status" in data and "retcode" in data and "echo" in data)):
                    echo = data.get("echo")
                    client = self._pending.pop(echo, None)
                    if client:
                        try:
                            await client.send(json.dumps(data, ensure_ascii=False))
                        except Exception as e:
                            print(f"[gateway] action_result deliver error: {repr(e)}")
                elif isinstance(data, dict) and data.get("action"):
                    self.clients.add(websocket)
                    
                    try:
                        await self._on_action_from_core(websocket, data)
                    except Exception as e:
                        print("[gateway] action handler error:", repr(e))
                        echo = data.get("echo")
                        result = {"status": "failed", "retcode": -1, "msg": str(e or "bot error"), "echo": echo}
                        await websocket.send(json.dumps(result, ensure_ascii=False))
                else:
                    pass
        except websockets.exceptions.ConnectionClosed as e:
            if self.config.get("debug"):
                print(f"[gateway] connection closed by peer from {remote_addr}: code={e.code} reason={e.reason}")
        except Exception as e:
            print(f'[OneBot] handler error from {remote_addr}:', repr(e))
        finally:
            heartbeat_task.cancel()
            self.clients.discard(websocket)
            self.ws_bots.discard(websocket)
            # Remove from remote_bots
            to_remove = [k for k, v in self.remote_bots.items() if v == websocket]
            for k in to_remove:
                del self.remote_bots[k]
                print(f"[gateway] Remote bot disconnected: {k}")

    async def send_event(self, event: dict):
        """
        向所有已连接 C# 客户端推送 OneBot 事件
        """
        total = len(list(self.clients))
        
        # 缓冲逻辑：如果没有客户端连接，且不是心跳事件，则存入缓冲区
        if total == 0 and len(self.ws_bots) == 0:
            if event.get("meta_event_type") != "heartbeat":
                if self.config.get("debug"):
                    print(f"[gateway] No clients. Buffering event: {event.get('post_type')}")
                self._msg_buffer.append(event)
                # 限制缓冲区大小
                if len(self._msg_buffer) > self._max_buffer_size:
                    self._msg_buffer.pop(0)
            return

        payload = json.dumps(event, ensure_ascii=False)
        targets = list(self.clients | self.ws_bots)
        for i, ws in enumerate(targets):
            try:
                await ws.send(payload)
                addr = getattr(ws, 'remote_address', None)
            except Exception as e:
                print(f"[gateway] -> target#{i} send error: {repr(e)}")
                self.clients.discard(ws)
                self.ws_bots.discard(ws)

    async def call_api(self, action: str, params: dict):
        """
        Call OneBot API on the connected implementation.
        """
        # 优先使用本地 bot
        if self.local_bots:
             # TODO: 暂时只支持第一个
            bot = next(iter(self.local_bots.values()))
            # 这里需要模拟一个 websocket 或者直接调用 internal method?
            # 暂时保持原样，只处理 ws_bots
            pass

        if not self.ws_bots and not self.remote_bots:
            print("[gateway] call_api failed: No bot connected")
            return None
        
        # Pick a bot (first one from remote_bots or ws_bots)
        bot_ws = None
        if self.remote_bots:
            bot_ws = next(iter(self.remote_bots.values()))
        elif self.ws_bots:
            bot_ws = next(iter(self.ws_bots))
            
        if not bot_ws:
             return None
        
        echo = str(int(time.time() * 1000)) + str(random.randint(0, 1000))
        data = {
            "action": action,
            "params": params,
            "echo": echo
        }
        
        try:
            await bot_ws.send(json.dumps(data, ensure_ascii=False))
            return echo
        except Exception as e:
            print(f"[gateway] call_api error: {e}")
            return None

    def call_api_nowait(self, action: str, params: dict):
        """
        Sync wrapper for call_api (fire and forget)
        """
        asyncio.create_task(self.call_api(action, params))

    async def _on_action_from_core(self, websocket, action: dict):
        """
        收到 C# 的动作（OneBot Action），转给 bot 执行实际微信发送
        """
        echo = action.get("echo")
        result = {"status": "failed", "retcode": -1, "data": None, "echo": echo}

        params = action.get("params", {}) or {}
        target_id = action.get("self_id") or params.get("robot_qq") or params.get("self_id")
        
        # 1. Try local bots first
        bot = None
        if target_id and str(target_id) in self.local_bots:
            bot = self.local_bots[str(target_id)]
        elif not target_id and len(self.local_bots) > 0:
            bot = next(iter(self.local_bots.values()))
        
        # 2. If no local bot found, check remote bots
        if not bot:
            remote_ws = None
            if target_id and str(target_id) in self.remote_bots:
                remote_ws = self.remote_bots[str(target_id)]
            elif not target_id and self.remote_bots:
                # Pick any remote bot if no ID specified
                remote_ws = next(iter(self.remote_bots.values()))
            
            if remote_ws:
                # Forward to remote bot
                try:
                    # Ensure echo is present
                    if not action.get("echo"):
                        action["echo"] = str(int(time.time() * 1000))
                    
                    # Register pending request to route response back
                    self._pending[action["echo"]] = websocket
                    
                    await remote_ws.send(json.dumps(action, ensure_ascii=False))
                    # We don't send response here, as remote bot will send action_result later
                    return 
                except Exception as e:
                    print(f"[gateway] Forward to remote bot error: {e}")
                    result["msg"] = f"Forward error: {str(e)}"
                    await websocket.send(json.dumps(result, ensure_ascii=False))
                    return

        if not bot:
            result["msg"] = "No bot found"
            await websocket.send(json.dumps(result, ensure_ascii=False))
            return

        name = action.get("action")

        try:
            if name == "send_group_msg":
                group_id = params.get("group_id")
                message = params.get("message", "")
                mid = bot._send_group_message_via_wechat(group_id, message)
                info = getattr(bot, '_last_send_info', {})
                result.update({"status": "ok", "retcode": 0, "data": {"message_id": mid, "debug_ret": info.get('ret'), "http_status": info.get('status')}})

            elif name == "send_private_msg":
                user_id = params.get("user_id")
                message = params.get("message", "")
                mid = bot._send_private_message_via_wechat(user_id, message)
                info = getattr(bot, '_last_send_info', {})
                result.update({"status": "ok", "retcode": 0, "data": {"message_id": mid, "debug_ret": info.get('ret'), "http_status": info.get('status')}})

            elif name == "get_group_list":
                items = []
                for g in getattr(bot, 'group_list', []) or []:
                    gid_uid = g.get('UserName')
                    gname = msg.remove_Emoji(g.get('NickName') or '')
                    try:
                        gid_num = wx_group.get_wx_group(bot.self_id, gid_uid, gname, 0, '')
                    except Exception:
                        gid_num = 0
                    if gid_num:
                        try:
                            if not hasattr(bot, '_group_map_uid_by_id'): bot._group_map_uid_by_id = {}
                            if not hasattr(bot, '_group_map_id_by_uid'): bot._group_map_id_by_uid = {}
                            bot._group_map_uid_by_id[gid_num] = gid_uid
                            bot._group_map_id_by_uid[gid_uid] = gid_num
                        except Exception:
                            pass
                    items.append({"group_id": gid_num, "group_name": gname})
                result.update({"status": "ok", "retcode": 0, "data": items})

            elif name == "get_group_member_info":
                group_id = params.get("group_id")
                user_id = params.get("user_id")
                gid_uid = (bot._group_map_uid_by_id.get(group_id) if hasattr(bot, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                uid = wx_client.get_client_uid(user_id)
                names = bot.get_group_member_name(gid_uid, uid) or {}
                data = {
                    "group_id": group_id,
                    "user_id": user_id,
                    "nickname": names.get("nickname", ""),
                    "card": names.get("display_name", "")
                }
                result.update({"status": "ok", "retcode": 0, "data": data})

            elif name == "get_group_member_list":
                group_id = params.get("group_id")
                gid_uid = (bot._group_map_uid_by_id.get(group_id) if hasattr(bot, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                members = []
                for m in (bot.group_members.get(gid_uid, []) if hasattr(bot, 'group_members') else []):
                    uid = m.get('UserName')
                    user_id = wx_client.get_client_qq_by_uid(uid)
                    nickname = msg.remove_Emoji(m.get('NickName') or '')
                    card = msg.remove_Emoji(m.get('DisplayName') or '')
                    members.append({"user_id": user_id, "nickname": nickname, "card": card})
                result.update({"status": "ok", "retcode": 0, "data": members})

            elif name == "set_group_kick":
                group_id = params.get("group_id")
                user_id = params.get("user_id")
                if not group_id or not user_id:
                    result.update({"status": "failed", "retcode": 10003, "msg": "missing group_id or user_id"})
                else:
                    try:
                        gid_uid = (bot._group_map_uid_by_id.get(group_id) if hasattr(bot, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                        
                        # Server-side permission check
                        owner_uid = bot._fetch_group_owner_uid(gid_uid) if gid_uid else ""
                        is_admin = bool(owner_uid and owner_uid == getattr(bot, 'my_account', {}).get('UserName'))
                        
                        if not is_admin:
                            result.update({"status": "failed", "retcode": 10004, "msg": "not admin"})
                        else:
                            uid = wx_client.get_client_uid(user_id)
                            
                            # 获取被踢成员昵称用于发送通知
                            kicked_nickname = ""
                            try:
                                if hasattr(bot, 'group_members') and gid_uid in bot.group_members:
                                    for m in bot.group_members[gid_uid]:
                                        if m.get('UserName') == uid:
                                            kicked_nickname = m.get('NickName') or m.get('DisplayName') or ""
                                            break
                                if not kicked_nickname:
                                    kicked_nickname = str(user_id)
                                else:
                                    kicked_nickname = msg.remove_Emoji(kicked_nickname)
                            except Exception:
                                kicked_nickname = str(user_id)

                            ok = bot.delete_user_from_group(gid_uid, uid)
                            if ok:
                                result.update({"status": "ok", "retcode": 0, "data": True})
                                try:
                                    # 发送踢出通知
                                    kick_msg = f"{kicked_nickname} 被移出群聊"
                                    bot.send_msg_by_uid(kick_msg, gid_uid)
                                except Exception as e:
                                    print(f"[gateway] send kick msg error: {e}")

                                try:
                                    gname = ""
                                    for g in getattr(bot, 'group_list', []) or []:
                                        if g.get('UserName') == gid_uid:
                                            gname = msg.remove_Emoji(g.get('NickName') or '')
                                            break
                                    event = {
                                        "self_id": bot.self_id,
                                        "time": int(time.time()),
                                        "post_type": "notice",
                                        "notice_type": "group_decrease",
                                        "sub_type": "kick",
                                        "group_id": group_id,
                                        "group_name": gname,
                                        "user_id": user_id,
                                        "operator_id": bot.self_id
                                    }
                                    bot._push_onebot_event(event)
                                except Exception:
                                    pass
                            else:
                                result.update({"status": "failed", "retcode": 10005, "msg": "operation failed"})
                    except Exception as e:
                        print(f"[gateway] set_group_kick error: {e}")
                        result.update({"status": "failed", "retcode": 10002, "msg": str(e)})

            elif name == "get_friend_list":
                items = []
                for c in getattr(bot, 'contact_list', []) or []:
                    uid = c.get('UserName')
                    user_id = 0
                    try:
                        user_id = wx_client.get_client_qq_by_uid(uid)
                    except Exception:
                        user_id = 0
                    nickname = msg.remove_Emoji(c.get('NickName') or '')
                    remark = msg.remove_Emoji(c.get('RemarkName') or '')
                    items.append({"user_id": user_id, "nickname": nickname, "remark": remark})
                result.update({"status": "ok", "retcode": 0, "data": items})

            elif name == "get_group_info":
                group_id = params.get("group_id")
                
                gid_uid = (bot._group_map_uid_by_id.get(group_id) if hasattr(bot, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                gname = ""
                member_count = 0
                for g in getattr(bot, 'group_list', []) or []:
                    if g.get('UserName') == gid_uid:
                        gname = msg.remove_Emoji(g.get('NickName') or '')
                        member_count = g.get('MemberCount', 0)
                        break
                
                # Server-side permission check
                owner_uid = bot._fetch_group_owner_uid(gid_uid) if gid_uid else ""
                is_admin = bool(owner_uid and owner_uid == getattr(bot, 'my_account', {}).get('UserName'))
                
                data = {"group_id": group_id, "group_name": gname, "member_count": member_count, "is_admin": is_admin}
                result.update({"status": "ok", "retcode": 0, "data": data})

            elif name == "set_group_add_member":
                group_id = params.get("group_id")
                user_id = params.get("user_id")
                if not group_id or not user_id:
                    result.update({"status": "failed", "retcode": 10003, "msg": "missing group_id or user_id"})
                else:
                    try:
                        gid_uid = (bot._group_map_uid_by_id.get(group_id) if hasattr(bot, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                        uid = wx_client.get_client_uid(user_id)
                        ok = bot.add_friend_to_group(gid_uid, uid)
                        if ok:
                            result.update({"status": "ok", "retcode": 0, "data": True})
                            try:
                                gname = ""
                                for g in getattr(bot, 'group_list', []) or []:
                                    if g.get('UserName') == gid_uid:
                                        gname = msg.remove_Emoji(g.get('NickName') or '')
                                        break
                                event = {
                                    "self_id": bot.self_id,
                                    "time": int(time.time()),
                                    "post_type": "notice",
                                    "notice_type": "group_increase",
                                    "sub_type": "approve",
                                    "group_id": group_id,
                                    "group_name": gname,
                                    "user_id": user_id,
                                    "operator_id": getattr(common, 'default_robot_qq', 0)
                                }
                                bot._push_onebot_event(event)
                            except Exception:
                                pass
                        else:
                            result.update({"status": "failed", "retcode": 10005, "msg": "operation failed"})
                    except Exception as e:
                        print(f"[gateway] set_group_add_member error: {e}")
                        result.update({"status": "failed", "retcode": 10002, "msg": str(e)})

            elif name == "get_login_info":
                data = {
                    "self_id": bot.self_id,
                    "platform": "wechat-web",
                    "robot_qq": getattr(common, 'default_robot_qq', 0),
                    "online": True
                }
                result.update({"status": "ok", "retcode": 0, "data": data})

            elif name == "get_status":
                qsize = 0
                try:
                    qsize = getattr(self._event_queue, 'qsize', lambda: 0)() if self._event_queue else 0
                except Exception:
                    qsize = 0
                data = {
                    "clients": len(self.clients),
                    "event_queue_size": qsize,
                    "uptime_sec": max(0, int(time.time()) - int(getattr(bot, 'boot_ts', time.time())))
                }
                result.update({"status": "ok", "retcode": 0, "data": data})

            elif name == "get_group_permissions":
                group_id = params.get("group_id")
                gid_uid = (bot._group_map_uid_by_id.get(group_id) if hasattr(bot, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                gname = ""
                for g in getattr(bot, 'group_list', []) or []:
                    if g.get('UserName') == gid_uid:
                        gname = msg.remove_Emoji(g.get('NickName') or '')
                        break
                # Server-side check: ChatRoomOwner equals bot uid => admin/owner
                owner_uid = bot._fetch_group_owner_uid(gid_uid) if gid_uid else ""
                is_admin = bool(owner_uid and owner_uid == getattr(bot, 'my_account', {}).get('UserName'))
                caps = {
                    "can_kick": is_admin,
                    "can_add_member": is_admin,
                    "can_ban": False,
                    "can_notice": False,
                    "can_recall": False,
                    "platform": "wechat-web"
                }
                data = {"group_id": group_id, "group_name": gname, "is_admin": is_admin, "owner_uid": owner_uid, "capabilities": caps}
                result.update({"status": "ok", "retcode": 0, "data": data})

            elif name == "get_all_group_permissions":
                items = []
                for g in getattr(bot, 'group_list', []) or []:
                    gid_uid = g.get('UserName')
                    gname = msg.remove_Emoji(g.get('NickName') or '')
                    # 保持已有映射或创建映射（避免重复数据库调用）
                    gid_num = bot._group_map_id_by_uid.get(gid_uid)
                    if not gid_num:
                        try:
                            # Proactively recover/create group_id mapping using Name if UID changed
                            robot_qq = getattr(common, 'default_robot_qq', 0)
                            gid_num = wx_group.get_wx_group(robot_qq, gid_uid, gname, 0, "")
                        except Exception as e:
                            print(f"[gateway] get_wx_group error for {gname}: {e}")
                            gid_num = 0
                    if gid_num:
                        try:
                            bot._group_map_uid_by_id[gid_num] = gid_uid
                            bot._group_map_id_by_uid[gid_uid] = gid_num
                        except Exception:
                            pass
                    owner_uid = bot._fetch_group_owner_uid(gid_uid)
                    is_admin = bool(owner_uid and owner_uid == getattr(bot, 'my_account', {}).get('UserName'))
                    caps = {
                        "can_kick": is_admin,
                        "can_add_member": is_admin,
                        "can_ban": False,
                        "can_notice": False,
                        "can_recall": False,
                        "platform": "wechat-web"
                    }
                    items.append({"group_id": gid_num, "group_name": gname, "is_admin": is_admin, "owner_uid": owner_uid, "capabilities": caps})
                result.update({"status": "ok", "retcode": 0, "data": items})

            elif name == "send_group_image":
                group_id = params.get("group_id")
                file = (params.get("file") or params.get("image") or "")
                mid = bot._send_group_image_via_wechat(group_id, file)
                result.update({"status": "ok", "retcode": 0, "data": {"message_id": mid}})

            elif name == "send_private_image":
                user_id = params.get("user_id")
                file = (params.get("file") or params.get("image") or "")
                mid = bot._send_private_image_via_wechat(user_id, file)
                result.update({"status": "ok", "retcode": 0, "data": {"message_id": mid}})

            elif name == "send_group_file":
                group_id = params.get("group_id")
                file = params.get("file") or ""
                mid = bot._send_group_file_via_wechat(group_id, file)
                result.update({"status": "ok", "retcode": 0, "data": {"message_id": mid}})

            elif name == "send_private_file":
                user_id = params.get("user_id")
                file = params.get("file") or ""
                mid = bot._send_private_file_via_wechat(user_id, file)
                result.update({"status": "ok", "retcode": 0, "data": {"message_id": mid}})

            elif name == "send_group_music":
                group_id = params.get("group_id")
                title = params.get("title") or ""
                desc = params.get("desc") or ""
                url = params.get("url") or ""
                music_url = params.get("music_url") or ""
                mid = bot._send_group_music_via_wechat(group_id, title, desc, url, music_url)
                result.update({"status": "ok", "retcode": 0, "data": {"message_id": mid}})

            elif name == "send_private_music":
                user_id = params.get("user_id")
                title = params.get("title") or ""
                desc = params.get("desc") or ""
                url = params.get("url") or ""
                music_url = params.get("music_url") or ""
                mid = bot._send_private_music_via_wechat(user_id, title, desc, url, music_url)
                result.update({"status": "ok", "retcode": 0, "data": {"message_id": mid}})

            elif name == "send_group_image_with_text":
                group_id = params.get("group_id")
                file = (params.get("file") or params.get("image") or "")
                message = params.get("message") or ""
                mid_img = bot._send_group_image_via_wechat(group_id, file)
                mid_txt = bot._send_group_message_via_wechat(group_id, message) if message else 0
                result.update({"status": "ok", "retcode": 0, "data": {"image_message_id": mid_img, "text_message_id": mid_txt}})

            elif name == "send_private_image_with_text":
                user_id = params.get("user_id")
                file = (params.get("file") or params.get("image") or "")
                message = params.get("message") or ""
                mid_img = bot._send_private_image_via_wechat(user_id, file)
                mid_txt = bot._send_private_message_via_wechat(user_id, message) if message else 0
                result.update({"status": "ok", "retcode": 0, "data": {"image_message_id": mid_img, "text_message_id": mid_txt}})

            elif name == "get_user_avatar":
                user_id = params.get("user_id")
                fpath = bot._get_user_avatar_path(user_id)
                ok = bool(fpath)
                result.update({"status": "ok" if ok else "failed", "retcode": 0 if ok else 10005, "data": {"file": fpath}})

            elif name == "get_group_avatar":
                group_id = params.get("group_id")
                fpath = bot._get_group_avatar_path(group_id)
                ok = bool(fpath)
                result.update({"status": "ok" if ok else "failed", "retcode": 0 if ok else 10005, "data": {"file": fpath}})

            elif name == "send_group_user_avatar":
                group_id = params.get("group_id")
                user_id = params.get("user_id")
                fpath = bot._get_user_avatar_path(user_id)
                mid = bot._send_group_image_via_wechat(group_id, fpath)
                result.update({"status": "ok", "retcode": 0, "data": {"message_id": mid}})

            elif name == "send_private_user_avatar":
                user_id = params.get("user_id")
                fpath = bot._get_user_avatar_path(user_id)
                mid = bot._send_private_image_via_wechat(user_id, fpath)
                result.update({"status": "ok", "retcode": 0, "data": {"message_id": mid}})

            # ========== 新增/补全功能 ==========
            elif name == "set_group_name":
                group_id = params.get("group_id")
                group_name = params.get("group_name", "")
                gid_uid = (bot._group_map_uid_by_id.get(group_id) if hasattr(bot, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                ok = bot.set_group_name(gid_uid, group_name)
                result.update({"status": "ok" if ok else "failed", "retcode": 0 if ok else 10005})
            
            elif name == "set_group_leave":
                group_id = params.get("group_id")
                is_dismiss = params.get("is_dismiss", False) # 微信Web端无法真正解散，只能退出
                gid_uid = (bot._group_map_uid_by_id.get(group_id) if hasattr(bot, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                ok = bot.quit_group(gid_uid)
                result.update({"status": "ok" if ok else "failed", "retcode": 0 if ok else 10005})
            
            elif name == "delete_msg":
                message_id = params.get("message_id")
                try:
                    message_id = int(message_id)
                except:
                    pass
                
                cached = bot._sent_msg_cache.get(message_id)
                if cached:
                    # 调用 wxbot revoke_msg(client_msg_id, svr_msg_id, to_user_name)
                    ok = bot.revoke_msg(cached['local_id'], cached['svr_id'], cached['to'])
                    result.update({"status": "ok" if ok else "failed", "retcode": 0 if ok else 10002})
                else:
                    print(f"[gateway] delete_msg failed: msg_id={message_id} not in cache")
                    result.update({"status": "failed", "retcode": 10001, "msg": "msg not found in cache"})

            elif name in ["set_group_ban", "set_group_whole_ban", "set_group_admin", "set_group_card", 
                          "set_friend_add_request", "set_group_add_request"]:
                # 协议层支持，但底层能力不支持
                print(f"[gateway] {name} called but not supported by Web WeChat")
                result.update({"status": "failed", "retcode": 10004, "msg": "API not supported by Web WeChat"})

            elif name == "get_stranger_info":
                user_id = params.get("user_id")
                # 尝试用 get_group_member_info 碰运气 (如果在某个群里)
                # 或者直接返回基本信息
                nickname = str(user_id)
                sex = "unknown"
                age = 0
                result.update({"status": "ok", "retcode": 0, "data": {
                    "user_id": user_id,
                    "nickname": nickname,
                    "sex": sex,
                    "age": age
                }})
            
            # 可扩展其他 action
            else:
                result.update({"status": "failed", "retcode": 10001, "msg": f"unsupported action: {name}"})

        except Exception as e:
            print(f"[gateway] action error name={name} params={params} msg={e}")
            result.update({"status": "failed", "retcode": 10002, "msg": str(e)})

        await websocket.send(json.dumps(result, ensure_ascii=False))


# ========== 你的机器人（融合） ==========

class onebot(WXBot):
    """
    - 继承你原来的 WXBot
    - 在 __init__ 中启动 OneBot WS 服务器（另起线程）
    - 接收到微信消息时，组装 OneBot 事件，通过 gateway 推送给 C#
    - 保留 super().handle_msg_all(_msg)
    """

    def __init__(self, gateway, self_id: int = None, client_mode=False):
        super().__init__()
        
        # Override QR config: default to 'tty' for Docker environments
        # Users can check docker logs to scan QR code
        self.conf['qr'] = os.getenv("WX_QR_MODE", "tty")
        print(f"[onebot] Configured QR Code Mode: {self.conf['qr']}")

        # self_id 可传入，否则自动生成
        self.self_id = int(self_id) if self_id else 0
        
        # 多账号支持：设置独立的 session 和 qr 路径
        if self.self_id:
            self.cache_file = os.path.join(self.temp_pwd, f'session_{self.self_id}.json')
            self.qr_file_path = os.path.join(self.temp_pwd, f'wxqr_{self.self_id}.png')
            
            # Migration: If specific session missing but default exists, try to copy default
            default_session = os.path.join(self.temp_pwd, 'session.json')
            if not os.path.exists(self.cache_file) and os.path.exists(default_session):
                print(f"[onebot] Migrating default session to {self.cache_file}")
                try:
                    import shutil
                    shutil.copy(default_session, self.cache_file)
                except Exception as e:
                    print(f"[onebot] Migration failed: {e}")
        else:
            # 如果没有 self_id，可能使用随机生成的 ID 或者默认文件
            # 建议创建时务必传入 self_id
            pass

        self.boot_ts = int(time.time())
        self._seen_msg_ids = set()
        self._last_active_group_uid = None
        self._last_active_ts = 0
        self._group_map_uid_by_id = {}
        self._group_map_id_by_uid = {}
        self._name_index_by_group = {}
        
        # 消息缓存：用于引用回复 (message_id -> msg_info)
        self._msg_cache = collections.OrderedDict()

        # WS 网关 (如果不是 client_mode)
        self.client_mode = client_mode
        self.gateway = gateway
        if not self.client_mode and self.gateway:
            self.gateway.add_bot(self, self.self_id)
        
        # 启动 WS 服务器（后台线程 + 异步事件循环）
        # (已移至 BotManager 统一管理)

    # ... (rest of methods)

    def execute_onebot_action(self, action: str, params: dict = None) -> dict:
        """
        Execute a OneBot action locally on this bot instance.
        Returns a dict with keys: status, retcode, data, msg (optional)
        """
        if params is None:
            params = {}
        
        name = action
        result = {"status": "ok", "retcode": 0, "data": {}}
        
        try:
            if name == "send_group_msg":
                group_id = params.get("group_id")
                message = params.get("message", "")
                mid = self._send_group_message_via_wechat(group_id, message)
                info = getattr(self, '_last_send_info', {})
                result["data"] = {"message_id": mid, "debug_ret": info.get('ret'), "http_status": info.get('status')}

            elif name == "send_private_msg":
                user_id = params.get("user_id")
                message = params.get("message", "")
                mid = self._send_private_message_via_wechat(user_id, message)
                info = getattr(self, '_last_send_info', {})
                result["data"] = {"message_id": mid, "debug_ret": info.get('ret'), "http_status": info.get('status')}

            elif name == "get_group_list":
                items = []
                for g in getattr(self, 'group_list', []) or []:
                    gid_uid = g.get('UserName')
                    gname = msg.remove_Emoji(g.get('NickName') or '')
                    try:
                        gid_num = wx_group.get_wx_group(self.self_id, gid_uid, gname, 0, '')
                    except Exception:
                        gid_num = 0
                    if gid_num:
                        try:
                            if not hasattr(self, '_group_map_uid_by_id'): self._group_map_uid_by_id = {}
                            if not hasattr(self, '_group_map_id_by_uid'): self._group_map_id_by_uid = {}
                            self._group_map_uid_by_id[gid_num] = gid_uid
                            self._group_map_id_by_uid[gid_uid] = gid_num
                        except Exception:
                            pass
                    items.append({"group_id": gid_num, "group_name": gname})
                result["data"] = items

            elif name == "get_group_member_info":
                group_id = params.get("group_id")
                user_id = params.get("user_id")
                gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                uid = wx_client.get_client_uid(user_id)
                names = self.get_group_member_name(gid_uid, uid) or {}
                data = {
                    "group_id": group_id,
                    "user_id": user_id,
                    "nickname": names.get("nickname", ""),
                    "card": names.get("display_name", "")
                }
                result["data"] = data

            elif name == "get_group_member_list":
                group_id = params.get("group_id")
                gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                members = []
                for m in (self.group_members.get(gid_uid, []) if hasattr(self, 'group_members') else []):
                    uid = m.get('UserName')
                    user_id = wx_client.get_client_qq_by_uid(uid)
                    nickname = msg.remove_Emoji(m.get('NickName') or '')
                    card = msg.remove_Emoji(m.get('DisplayName') or '')
                    members.append({"user_id": user_id, "nickname": nickname, "card": card})
                result["data"] = members

            elif name == "set_group_kick":
                group_id = params.get("group_id")
                user_id = params.get("user_id")
                if not group_id or not user_id:
                    result.update({"status": "failed", "retcode": 10003, "msg": "missing group_id or user_id"})
                else:
                    gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                    
                    # Server-side permission check
                    owner_uid = self._fetch_group_owner_uid(gid_uid) if gid_uid else ""
                    is_admin = bool(owner_uid and owner_uid == getattr(self, 'my_account', {}).get('UserName'))
                    
                    if not is_admin:
                        result.update({"status": "failed", "retcode": 10004, "msg": "not admin"})
                    else:
                        uid = wx_client.get_client_uid(user_id)
                        
                        # 获取被踢成员昵称用于发送通知
                        kicked_nickname = ""
                        try:
                            if hasattr(self, 'group_members') and gid_uid in self.group_members:
                                for m in self.group_members[gid_uid]:
                                    if m.get('UserName') == uid:
                                        kicked_nickname = m.get('NickName') or m.get('DisplayName') or ""
                                        break
                            if not kicked_nickname:
                                kicked_nickname = str(user_id)
                            else:
                                kicked_nickname = msg.remove_Emoji(kicked_nickname)
                        except Exception:
                            kicked_nickname = str(user_id)

                        ok = self.delete_user_from_group(gid_uid, uid)
                        if ok:
                            result["data"] = True
                            try:
                                # 发送踢出通知
                                kick_msg = f"{kicked_nickname} 被移出群聊"
                                self.send_msg_by_uid(kick_msg, gid_uid)
                            except Exception as e:
                                print(f"[onebot] send kick msg error: {e}")

                            try:
                                gname = ""
                                for g in getattr(self, 'group_list', []) or []:
                                    if g.get('UserName') == gid_uid:
                                        gname = msg.remove_Emoji(g.get('NickName') or '')
                                        break
                                event = {
                                    "self_id": self.self_id,
                                    "time": int(time.time()),
                                    "post_type": "notice",
                                    "notice_type": "group_decrease",
                                    "sub_type": "kick",
                                    "group_id": group_id,
                                    "group_name": gname,
                                    "user_id": user_id,
                                    "operator_id": self.self_id
                                }
                                self._push_onebot_event(event)
                            except Exception:
                                pass
                        else:
                            result.update({"status": "failed", "retcode": 10005, "msg": "operation failed"})

            elif name == "get_friend_list":
                items = []
                for c in getattr(self, 'contact_list', []) or []:
                    uid = c.get('UserName')
                    user_id = 0
                    try:
                        user_id = wx_client.get_client_qq_by_uid(uid)
                    except Exception:
                        user_id = 0
                    nickname = msg.remove_Emoji(c.get('NickName') or '')
                    remark = msg.remove_Emoji(c.get('RemarkName') or '')
                    items.append({"user_id": user_id, "nickname": nickname, "remark": remark})
                result["data"] = items

            elif name == "get_group_info":
                group_id = params.get("group_id")
                
                gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                gname = ""
                member_count = 0
                for g in getattr(self, 'group_list', []) or []:
                    if g.get('UserName') == gid_uid:
                        gname = msg.remove_Emoji(g.get('NickName') or '')
                        member_count = g.get('MemberCount', 0)
                        break
                
                owner_uid = self._fetch_group_owner_uid(gid_uid) if gid_uid else ""
                is_admin = bool(owner_uid and owner_uid == getattr(self, 'my_account', {}).get('UserName'))
                
                data = {"group_id": group_id, "group_name": gname, "member_count": member_count, "is_admin": is_admin}
                result["data"] = data

            elif name == "set_group_add_member":
                group_id = params.get("group_id")
                user_id = params.get("user_id")
                if not group_id or not user_id:
                    result.update({"status": "failed", "retcode": 10003, "msg": "missing group_id or user_id"})
                else:
                    gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                    uid = wx_client.get_client_uid(user_id)
                    ok = self.add_friend_to_group(gid_uid, uid)
                    if ok:
                        result["data"] = True
                        try:
                            gname = ""
                            for g in getattr(self, 'group_list', []) or []:
                                if g.get('UserName') == gid_uid:
                                    gname = msg.remove_Emoji(g.get('NickName') or '')
                                    break
                            event = {
                                "self_id": self.self_id,
                                "time": int(time.time()),
                                "post_type": "notice",
                                "notice_type": "group_increase",
                                "sub_type": "approve",
                                "group_id": group_id,
                                "group_name": gname,
                                "user_id": user_id,
                                "operator_id": getattr(common, 'default_robot_qq', 0)
                            }
                            self._push_onebot_event(event)
                        except Exception:
                            pass
                    else:
                        result.update({"status": "failed", "retcode": 10005, "msg": "operation failed"})

            elif name == "get_login_info":
                nickname = ""
                try:
                    nickname = msg.remove_Emoji(getattr(self, 'my_account', {}).get('NickName') or "")
                except Exception:
                    nickname = "Unknown"
                    
                data = {
                    "self_id": self.self_id,
                    "nickname": nickname,
                    "platform": "wechat-web",
                    "robot_qq": getattr(common, 'default_robot_qq', 0),
                    "online": True
                }
                result["data"] = data

            elif name == "get_status":
                qsize = 0
                clients_count = 0
                if self.gateway:
                    try:
                        qsize = getattr(self.gateway._event_queue, 'qsize', lambda: 0)() if self.gateway._event_queue else 0
                        clients_count = len(self.gateway.clients)
                    except Exception:
                        pass
                
                data = {
                    "clients": clients_count,
                    "event_queue_size": qsize,
                    "uptime_sec": max(0, int(time.time()) - int(getattr(self, 'boot_ts', time.time())))
                }
                result["data"] = data

            elif name == "get_group_permissions":
                group_id = params.get("group_id")
                gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                gname = ""
                for g in getattr(self, 'group_list', []) or []:
                    if g.get('UserName') == gid_uid:
                        gname = msg.remove_Emoji(g.get('NickName') or '')
                        break
                owner_uid = self._fetch_group_owner_uid(gid_uid) if gid_uid else ""
                is_admin = bool(owner_uid and owner_uid == getattr(self, 'my_account', {}).get('UserName'))
                caps = {
                    "can_kick": is_admin,
                    "can_add_member": is_admin,
                    "can_ban": False,
                    "can_notice": False,
                    "can_recall": False,
                    "platform": "wechat-web"
                }
                data = {"group_id": group_id, "group_name": gname, "is_admin": is_admin, "owner_uid": owner_uid, "capabilities": caps}
                result["data"] = data

            elif name == "get_all_group_permissions":
                items = []
                for g in getattr(self, 'group_list', []) or []:
                    gid_uid = g.get('UserName')
                    gname = msg.remove_Emoji(g.get('NickName') or '')
                    gid_num = self._group_map_id_by_uid.get(gid_uid)
                    if not gid_num:
                        try:
                            robot_qq = getattr(common, 'default_robot_qq', 0)
                            gid_num = wx_group.get_wx_group(robot_qq, gid_uid, gname, 0, "")
                        except Exception as e:
                            print(f"[onebot] get_wx_group error for {gname}: {e}")
                            gid_num = 0
                    if gid_num:
                        try:
                            self._group_map_uid_by_id[gid_num] = gid_uid
                            self._group_map_id_by_uid[gid_uid] = gid_num
                        except Exception:
                            pass
                    owner_uid = self._fetch_group_owner_uid(gid_uid)
                    is_admin = bool(owner_uid and owner_uid == getattr(self, 'my_account', {}).get('UserName'))
                    caps = {
                        "can_kick": is_admin,
                        "can_add_member": is_admin,
                        "can_ban": False,
                        "can_notice": False,
                        "can_recall": False,
                        "platform": "wechat-web"
                    }
                    items.append({"group_id": gid_num, "group_name": gname, "is_admin": is_admin, "owner_uid": owner_uid, "capabilities": caps})
                result["data"] = items

            elif name == "send_group_image":
                group_id = params.get("group_id")
                file = (params.get("file") or params.get("image") or "")
                mid = self._send_group_image_via_wechat(group_id, file)
                result["data"] = {"message_id": mid}

            elif name == "send_private_image":
                user_id = params.get("user_id")
                file = (params.get("file") or params.get("image") or "")
                mid = self._send_private_image_via_wechat(user_id, file)
                result["data"] = {"message_id": mid}

            elif name == "get_login_info":
                user_id = self.self_id
                nickname = ""
                try:
                    nickname = getattr(self, 'my_account', {}).get('NickName') or ""
                    nickname = msg.remove_Emoji(nickname)
                except Exception:
                    pass
                if not nickname:
                    nickname = f"Bot {user_id}"
                
                result["data"] = {
                    "user_id": user_id,
                    "nickname": nickname
                }

            elif name == "send_group_file":
                group_id = params.get("group_id")
                file = params.get("file") or ""
                mid = self._send_group_file_via_wechat(group_id, file)
                result["data"] = {"message_id": mid}

            elif name == "send_private_file":
                user_id = params.get("user_id")
                file = params.get("file") or ""
                mid = self._send_private_file_via_wechat(user_id, file)
                result["data"] = {"message_id": mid}

            elif name == "send_group_music":
                group_id = params.get("group_id")
                title = params.get("title") or ""
                desc = params.get("desc") or ""
                url = params.get("url") or ""
                music_url = params.get("music_url") or ""
                mid = self._send_group_music_via_wechat(group_id, title, desc, url, music_url)
                result["data"] = {"message_id": mid}

            elif name == "send_private_music":
                user_id = params.get("user_id")
                title = params.get("title") or ""
                desc = params.get("desc") or ""
                url = params.get("url") or ""
                music_url = params.get("music_url") or ""
                mid = self._send_private_music_via_wechat(user_id, title, desc, url, music_url)
                result["data"] = {"message_id": mid}

            elif name == "send_group_image_with_text":
                group_id = params.get("group_id")
                file = (params.get("file") or params.get("image") or "")
                message = params.get("message") or ""
                mid_img = self._send_group_image_via_wechat(group_id, file)
                mid_txt = self._send_group_message_via_wechat(group_id, message) if message else 0
                result["data"] = {"image_message_id": mid_img, "text_message_id": mid_txt}

            elif name == "send_private_image_with_text":
                user_id = params.get("user_id")
                file = (params.get("file") or params.get("image") or "")
                message = params.get("message") or ""
                mid_img = self._send_private_image_via_wechat(user_id, file)
                mid_txt = self._send_private_message_via_wechat(user_id, message) if message else 0
                result["data"] = {"image_message_id": mid_img, "text_message_id": mid_txt}

            elif name == "get_user_avatar":
                user_id = params.get("user_id")
                fpath = self._get_user_avatar_path(user_id)
                ok = bool(fpath)
                result.update({"status": "ok" if ok else "failed", "retcode": 0 if ok else 10005})
                result["data"] = {"file": fpath}

            elif name == "get_group_avatar":
                group_id = params.get("group_id")
                fpath = self._get_group_avatar_path(group_id)
                ok = bool(fpath)
                result.update({"status": "ok" if ok else "failed", "retcode": 0 if ok else 10005})
                result["data"] = {"file": fpath}

            elif name == "send_group_user_avatar":
                group_id = params.get("group_id")
                user_id = params.get("user_id")
                fpath = self._get_user_avatar_path(user_id)
                mid = self._send_group_image_via_wechat(group_id, fpath)
                result["data"] = {"message_id": mid}

            elif name == "send_private_user_avatar":
                user_id = params.get("user_id")
                fpath = self._get_user_avatar_path(user_id)
                mid = self._send_private_image_via_wechat(user_id, fpath)
                result["data"] = {"message_id": mid}
            
            elif name == "set_group_name":
                group_id = params.get("group_id")
                group_name = params.get("group_name", "")
                gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                ok = self.set_group_name(gid_uid, group_name)
                result.update({"status": "ok" if ok else "failed", "retcode": 0 if ok else 10005})

            elif name == "set_group_leave":
                group_id = params.get("group_id")
                is_dismiss = params.get("is_dismiss", False)
                gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                ok = self.quit_group(gid_uid)
                result.update({"status": "ok" if ok else "failed", "retcode": 0 if ok else 10005})

            elif name == "delete_msg":
                message_id = params.get("message_id")
                try:
                    message_id = int(message_id)
                except:
                    pass
                cached = self._sent_msg_cache.get(message_id)
                if cached:
                    ok = self.revoke_msg(cached['local_id'], cached['svr_id'], cached['to'])
                    result.update({"status": "ok" if ok else "failed", "retcode": 0 if ok else 10002})
                else:
                    result.update({"status": "failed", "retcode": 10001, "msg": "msg not found in cache"})

            elif name in ["set_group_ban", "set_group_whole_ban", "set_group_admin", "set_group_card", 
                          "set_friend_add_request", "set_group_add_request"]:
                result.update({"status": "failed", "retcode": 10004, "msg": "API not supported by Web WeChat"})

            elif name == "get_stranger_info":
                user_id = params.get("user_id")
                nickname = str(user_id)
                sex = "unknown"
                age = 0
                result["data"] = {
                    "user_id": user_id,
                    "nickname": nickname,
                    "sex": sex,
                    "age": age
                }
            
            else:
                result.update({"status": "failed", "retcode": 10001, "msg": f"unsupported action: {name}"})

        except Exception as e:
            print(f"[onebot] action error name={name} params={params} msg={e}")
            result.update({"status": "failed", "retcode": 10002, "msg": str(e)})
            
        return result

    def _push_onebot_event(self, event):
        if self.client_mode:
            # Client Mode: Push to queue for WorkerClient to pick up
            if hasattr(self, 'event_callback') and callable(self.event_callback):
                self.event_callback(event)
            return

        if self.gateway:
            # Check for new OneBotDriver interface (broadcast_event)
            if hasattr(self.gateway, "broadcast_event"):
                self.gateway.broadcast_event(event)
                return

            # Legacy interface (ReverseGateway)
            loop = getattr(self.gateway, "_loop", None)
            if loop and hasattr(self.gateway, "send_event"):
                asyncio.run_coroutine_threadsafe(self.gateway.send_event(event), loop)

    def _download_to_temp(self, file_or_url: str) -> str:
        try:
            if not file_or_url:
                return ""
            if file_or_url.startswith("http://") or file_or_url.startswith("https://"):
                r = self.session.get(file_or_url)
                if r.status_code != 200:
                    print(f"[download] http error status={r.status_code} url={file_or_url}")
                    return ""
                ext = ".bin"
                try:
                    import mimetypes
                    ct = r.headers.get('Content-Type') or ''
                    ext = mimetypes.guess_extension(ct) or ext
                except Exception:
                    pass
                fn = f"dl_{int(time.time()*1000)}{ext}"
                fpath = os.path.join(self.temp_pwd, fn)
                with open(fpath, 'wb') as f:
                    f.write(r.content)
                return fpath
            return file_or_url
        except Exception as e:
            print(f"[download] error: {e}")
            return ""

    # ========= 发送：图片 / 文件 / 音乐卡片 =========
    def _send_group_image_via_wechat(self, group_id: int, file: str) -> int:
        uid = self.get_send_uid(group_id, "", 0, "")
        if not uid:
            raise RuntimeError(f"group_id={group_id} 未找到 WeChat group_uid")
        fpath = self._download_to_temp(file)
        if not fpath:
            raise RuntimeError("image file unavailable")
        ok = self.send_img_msg_by_uid(fpath, uid)
        # print(f"[robot] send_group_image uid={uid} ok={ok} path={fpath}")
        if not ok:
            raise RuntimeError("send_img_msg_by_uid failed")
        return self._mk_msg_id()

    def _send_private_image_via_wechat(self, user_id: int, file: str) -> int:
        uid = self.get_send_uid(0, "", user_id, "")
        if not uid:
            raise RuntimeError(f"user_id={user_id} 未找到 WeChat client_uid")
        fpath = self._download_to_temp(file)
        if not fpath:
            raise RuntimeError("image file unavailable")
        ok = self.send_file_msg_by_uid(fpath, uid)
        # print(f"[robot] send_private_image uid={uid} ok={ok} path={fpath}")
        if not ok:
            raise RuntimeError("send_img_msg_by_uid failed")
        return self._mk_msg_id()

    def _send_group_file_via_wechat(self, group_id: int, file: str) -> int:
        uid = self.get_send_uid(group_id, "", 0, "")
        if not uid:
            raise RuntimeError(f"group_id={group_id} 未找到 WeChat group_uid")
        fpath = self._download_to_temp(file)
        if not fpath:
            raise RuntimeError("file unavailable")
        ok = self.send_file_msg_by_uid(fpath, uid)
        # print(f"[robot] send_group_file uid={uid} ok={ok} path={fpath}")
        if not ok:
            raise RuntimeError("send_file_msg_by_uid failed")
        return self._mk_msg_id()

    def _send_private_file_via_wechat(self, user_id: int, file: str) -> int:
        uid = self.get_send_uid(0, "", user_id, "")
        if not uid:
            raise RuntimeError(f"user_id={user_id} 未找到 WeChat client_uid")
        fpath = self._download_to_temp(file)
        if not fpath:
            raise RuntimeError("file unavailable")
        ok = self.send_file_msg_by_uid(fpath, uid)
        # print(f"[robot] send_private_file uid={uid} ok={ok} path={fpath}")
        if not ok:
            raise RuntimeError("send_file_msg_by_uid failed")
        return self._mk_msg_id()

    def _send_music_card_by_uid(self, uid: str, title: str, desc: str, url: str, music_url: str) -> bool:
        try:
            import html
            api = self.base_uri + '/webwxsendappmsg?fun=async&f=json&pass_ticket=' + self.pass_ticket
            msg_id = str(int(time.time() * 1000)) + str(random.random())[:5].replace('.', '')
            
            # Escape XML special characters
            title_xml = html.escape(title or "")
            desc_xml = html.escape(desc or "")
            url_xml = html.escape(url or "")
            music_url_xml = html.escape(music_url or "")
            
            # Hardcoded test values if empty (as per user request)
            # 征服 - 那英
            if not title: title_xml = "征服"
            if not desc: desc_xml = "那英"
            if not url: url_xml = "https://i.y.qq.com/v8/playsong.html?hosteuin=7K6PoiSFoKn*&songid=179923&songmid=&type=0&platform=1&appsongtype=1&_wv=1&source=qq&appshare=iphone&media_mid=004LBt3k1d1J9m&ADTAG=qfshare"
            if not music_url: music_url_xml = "http://c6.y.qq.com/rsc/fcgi-bin/fcg_pyq_play.fcg?songid=0&songmid=003Rksq51qnUks&songtype=1&fromtag=50&uin=51437810&code=f606f"
            
            # AppMsg: type=3 音乐
            # 必须包含 dataurl 才能正常播放
            content_xml = (
                "<appmsg appid='wxeb7ec651dd0aefa9' sdkver=''>"
                f"<title>{title_xml}</title>"
                f"<des>{desc_xml}</des>"
                "<action></action>"
                "<type>3</type>"
                "<showtype>0</showtype>"
                "<mediatagname></mediatagname>"
                "<messageext></messageext>"
                "<messageaction></messageaction>"
                "<content></content>"
                "<contentattr>0</contentattr>"
                f"<url>{url_xml}</url>"
                f"<lowurl>{music_url_xml}</lowurl>"
                f"<dataurl>{music_url_xml}</dataurl>"
                f"<lowdataurl>{music_url_xml}</lowdataurl>"
                "<appattach><totallen>0</totallen><attachid></attachid><fileext></fileext></appattach>"
                "<extinfo></extinfo>"
                "</appmsg>"
            )
            data = {
                'BaseRequest': self.base_request,
                'Msg': {
                    'Type': 6,
                    'Content': content_xml,
                    'FromUserName': self.my_account['UserName'],
                    'ToUserName': uid,
                    'LocalID': msg_id,
                    'ClientMsgId': msg_id,
                    'Scene': 0,
                },
            }
            headers = {'content-type': 'application/json; charset=UTF-8'}
            data = json.dumps(data, ensure_ascii=False).encode('utf8')
            r = self.session.post(api, data=data, headers=headers)
            res = json.loads(r.text)
            ret_code = (res.get('BaseResponse') or {}).get('Ret')
            if ret_code != 0:
                print(f"[send] music appmsg failed: ret={ret_code} res={res}")
                return False
            return True
        except Exception as e:
            print(f"[send] music appmsg error: {e}")
            return False

    def _send_link_card_by_uid(self, uid: str, title: str, desc: str, url: str, thumb_url: str) -> bool:
        try:
            import html
            api = self.base_uri + '/webwxsendappmsg?fun=async&f=json&pass_ticket=' + self.pass_ticket
            msg_id = str(int(time.time() * 1000)) + str(random.random())[:5].replace('.', '')
            
            title_xml = html.escape(title or "链接")
            desc_xml = html.escape(desc or "")
            url_xml = html.escape(url or "")
            thumb_xml = html.escape(thumb_url or "")
            
            # AppMsg: type=5 链接
            content_xml = (
                "<appmsg appid='wxeb7ec651dd0aefa9' sdkver=''>"
                f"<title>{title_xml}</title>"
                f"<des>{desc_xml}</des>"
                "<action></action>"
                "<type>5</type>"
                "<showtype>0</showtype>"
                f"<url>{url_xml}</url>"
                f"<thumburl>{thumb_xml}</thumburl>"
                "<appattach><totallen>0</totallen><attachid></attachid><fileext></fileext></appattach>"
                "<extinfo></extinfo>"
                "</appmsg>"
            )
            data = {
                'BaseRequest': self.base_request,
                'Msg': {
                    'Type': 6,
                    'Content': content_xml,
                    'FromUserName': self.my_account['UserName'],
                    'ToUserName': uid,
                    'LocalID': msg_id,
                    'ClientMsgId': msg_id,
                    'Scene': 0,
                },
            }
            headers = {'content-type': 'application/json; charset=UTF-8'}
            data = json.dumps(data, ensure_ascii=False).encode('utf8')
            r = self.session.post(api, data=data, headers=headers)
            res = json.loads(r.text)
            ret_code = (res.get('BaseResponse') or {}).get('Ret')
            if ret_code != 0:
                print(f"[send] link appmsg failed: ret={ret_code} res={res}")
                return False
            return True
        except Exception as e:
            print(f"[send] link appmsg error: {e}")
            return False

    def _send_quote_msg_by_uid(self, uid: str, reply_text: str, refer_info: dict) -> bool:
        try:
            import html
            api = self.base_uri + '/webwxsendappmsg?fun=async&f=json&pass_ticket=' + self.pass_ticket
            msg_id = str(int(time.time() * 1000)) + str(random.random())[:5].replace('.', '')
            
            title_xml = html.escape(reply_text)
            
            ref_svr_id = refer_info.get("svr_id") or ""
            ref_from_uid = refer_info.get("from_uid") or ""
            ref_content = html.escape(refer_info.get("content") or "")
            ref_displayname = html.escape(refer_info.get("sender_name") or "")
            
            # refer_info["group_uid"] is present if it was a group message
            ref_chatusr = refer_info.get("group_uid") or refer_info.get("from_uid") or ""

            # AppMsg: type=57 Quote
            content_xml = (
                "<appmsg appid='' sdkver='0'>"
                f"<title>{title_xml}</title>"
                "<des></des>"
                "<action></action>"
                "<type>57</type>"
                "<showtype>0</showtype>"
                "<content></content>"
                "<url></url>"
                "<lowurl></lowurl>"
                "<appattach><totallen>0</totallen><attachid></attachid><fileext></fileext></appattach>"
                "<refermsg>"
                "<type>1</type>"
                f"<svrid>{ref_svr_id}</svrid>"
                f"<fromusr>{ref_from_uid}</fromusr>"
                f"<chatusr>{ref_chatusr}</chatusr>"
                f"<displayname>{ref_displayname}</displayname>"
                f"<content>{ref_content}</content>"
                "</refermsg>"
                "</appmsg>"
            )
            
            data = {
                'BaseRequest': self.base_request,
                'Msg': {
                    'Type': 6, 
                    'Content': content_xml,
                    'FromUserName': self.my_account['UserName'],
                    'ToUserName': uid,
                    'LocalID': msg_id,
                    'ClientMsgId': msg_id,
                    'Scene': 0,
                },
            }
            headers = {'content-type': 'application/json; charset=UTF-8'}
            data = json.dumps(data, ensure_ascii=False).encode('utf8')
            r = self.session.post(api, data=data, headers=headers)
            res = json.loads(r.text)
            ret_code = (res.get('BaseResponse') or {}).get('Ret')
            if ret_code != 0:
                print(f"[send] quote appmsg failed: ret={ret_code} res={res}")
                return False
            return True
        except Exception as e:
            print(f"[send] quote appmsg error: {e}")
            return False

    def _parse_message_segments(self, message: str) -> list:
        """
        Parse message into segments (text, music, share, image, poke, record, video)
        """
        import re
        segments = []
        # Split by supported CQ codes
        # Capture the delimiter to keep it
        pattern = r"(\[CQ:(?:music|share|image|poke|record|video|reply|at),[^\]]+\])"
        parts = re.split(pattern, message)
        
        for p in parts:
            if not p:
                continue
            
            if p.startswith('[CQ:reply'):
                m = re.match(r"\[CQ:reply,([^\]]+)\]", p)
                if m:
                    kvs = m.group(1)
                    mp = {}
                    for item in kvs.split(','):
                        if '=' in item:
                            k, v = item.split('=', 1)
                            mp[k.strip()] = v.strip()
                    segments.append({'type': 'reply', 'data': mp})
                    continue

            if p.startswith('[CQ:music'):
                m = re.match(r"\[CQ:music,([^\]]+)\]", p)
                if m:
                    kvs = m.group(1)
                    mp = {}
                    for item in kvs.split(','):
                        if '=' in item:
                            k, v = item.split('=', 1)
                            mp[k.strip()] = v.strip()
                    segments.append({'type': 'music', 'data': mp})
                    continue
            
            elif p.startswith('[CQ:share'):
                m = re.match(r"\[CQ:share,([^\]]+)\]", p)
                if m:
                    kvs = m.group(1)
                    mp = {}
                    for item in kvs.split(','):
                        if '=' in item:
                            k, v = item.split('=', 1)
                            mp[k.strip()] = v.strip()
                    segments.append({'type': 'share', 'data': mp})
                    continue

            elif p.startswith('[CQ:image'):
                m = re.match(r"\[CQ:image,([^\]]+)\]", p)
                if m:
                    kvs = m.group(1)
                    mp = {}
                    for item in kvs.split(','):
                        if '=' in item:
                            k, v = item.split('=', 1)
                            mp[k.strip()] = v.strip()
                    segments.append({'type': 'image', 'data': mp})
                    continue

            elif p.startswith('[CQ:poke'):
                m = re.match(r"\[CQ:poke,([^\]]+)\]", p)
                if m:
                    kvs = m.group(1)
                    mp = {}
                    for item in kvs.split(','):
                        if '=' in item:
                            k, v = item.split('=', 1)
                            mp[k.strip()] = v.strip()
                    segments.append({'type': 'poke', 'data': mp})
                    continue

            elif p.startswith('[CQ:record'):
                m = re.match(r"\[CQ:record,([^\]]+)\]", p)
                if m:
                    kvs = m.group(1)
                    mp = {}
                    for item in kvs.split(','):
                        if '=' in item:
                            k, v = item.split('=', 1)
                            mp[k.strip()] = v.strip()
                    segments.append({'type': 'record', 'data': mp})
                    continue

            elif p.startswith('[CQ:video'):
                m = re.match(r"\[CQ:video,([^\]]+)\]", p)
                if m:
                    kvs = m.group(1)
                    mp = {}
                    for item in kvs.split(','):
                        if '=' in item:
                            k, v = item.split('=', 1)
                            mp[k.strip()] = v.strip()
                    segments.append({'type': 'video', 'data': mp})
                    continue
            
            elif p.startswith('[CQ:at'):
                m = re.match(r"\[CQ:at,([^\]]+)\]", p)
                if m:
                    kvs = m.group(1)
                    mp = {}
                    for item in kvs.split(','):
                        if '=' in item:
                            k, v = item.split('=', 1)
                            mp[k.strip()] = v.strip()
                    segments.append({'type': 'at', 'data': mp})
                    continue

            # Default to text
            segments.append({'type': 'text', 'data': p})
        
        return segments

    def _cache_sent_msg(self, mid, ret, uid):
        """
        Cache sent message info for potential recall (if supported later)
        """
        # print(f"[robot] cache sent msg mid={mid} uid={uid} ret={ret}")
        pass

    def _process_send_segments(self, segments: list, uid: str) -> int:
        """
        Send segments sequentially. Returns the message ID of the last sent message.
        """
        last_mid = 0
        sent_any = False
        
        # Pre-scan for reply
        reply_id = None
        for seg in segments:
            if seg['type'] == 'reply':
                reply_id = seg['data'].get('id')
                break 
        
        # Merge text and at segments
        merged_segments = []
        current_text = ""
        
        for seg in segments:
            stype = seg['type']
            data = seg['data']
            
            if stype == 'reply':
                continue
            
            if stype == 'at':
                qq = data.get('qq')
                # Default: keep ID (as per user request for unknown)
                at_str = str(qq)
                
                if qq:
                    if str(qq) == str(self.self_id):
                        at_str = f"@{self.my_account.get('NickName', '我')} "
                    else:
                        try:
                            name = None
                            # 1. Try DB (persistent mapping)
                            name = wx_client.get_client_name_by_qq(int(qq))
                            
                            # 2. Try Memory (group/contact) if DB failed
                            if not name:
                                target_wx_uid = wx_client.get_client_uid(int(qq))
                                if target_wx_uid:
                                    if uid.startswith('@@'):
                                         name_info = self.get_group_member_name(uid, target_wx_uid)
                                         name = name_info.get('nickname2') or name_info.get('remark_name2') or name_info.get('nickname')
                                    if not name:
                                         contact_info = self.get_contact_name(target_wx_uid)
                                         if contact_info:
                                             name = self.get_contact_prefer_name(contact_info)
                            
                            if name:
                                at_str = f"@{name} "
                        except Exception as e:
                            print(f"[robot] resolve at error: {e}")
                
                current_text += at_str

            elif stype == 'text':
                current_text += str(data)
            
            else:
                # Flush text if exists
                if current_text.strip():
                     merged_segments.append({'type': 'text', 'data': current_text})
                     current_text = ""
                merged_segments.append(seg)
        
        # Flush remaining text
        if current_text.strip():
            merged_segments.append({'type': 'text', 'data': current_text})

        for seg in merged_segments:
            stype = seg['type']
            data = seg['data']
            
            if stype == 'text':
                content = data.strip()
                if content:
                    ret = False
                    # Check if we should send as quote reply
                    # Ensure reply_id is string for lookup
                    if reply_id and str(reply_id) in self._msg_cache:
                        refer_info = self._msg_cache[str(reply_id)]
                        ret = self._send_quote_msg_by_uid(uid, content, refer_info)
                        # reply_id = None # Consume it? Usually yes.
                    else:
                        ret = self.send_msg_by_uid(content, uid)
                        
                    if ret:
                        mid = self._mk_msg_id()
                        if isinstance(ret, dict):
                            self._cache_sent_msg(mid, ret, uid)
                        last_mid = mid
                        sent_any = True
                    else:
                        print(f"[robot] send_msg_by_uid FAILED content={content}")
            
            elif stype == 'music':
                # Only support custom for now
                if data.get('type') == 'custom' or data.get('audio') or data.get('url'):
                    title = data.get('title') or ''
                    desc = data.get('desc') or ''
                    url = data.get('url') or ''
                    audio = data.get('audio') or ''
                    ok = self._send_music_card_by_uid(uid, title, desc, url, audio)
                    if ok: 
                        last_mid = self._mk_msg_id()
                        sent_any = True

            elif stype == 'share':
                title = data.get('title') or '链接'
                desc = data.get('content') or data.get('desc') or ''
                url = data.get('url') or ''
                image = data.get('image') or ''
                ok = self._send_link_card_by_uid(uid, title, desc, url, image)
                if ok: 
                    last_mid = self._mk_msg_id()
                    sent_any = True

            elif stype == 'image':
                file_val = data.get('file')
                url_val = data.get('url')
                target_file = url_val or file_val
                if target_file:
                    fpath = self._download_to_temp(target_file)
                    if fpath:
                        res = self.send_img_msg_by_uid(fpath, uid)
                        if res: 
                            last_mid = self._mk_msg_id()
                            sent_any = True

            elif stype == 'poke':
                qq = data.get('qq')
                target_uid = None
                if qq:
                    try:
                        target_uid = wx_client.get_client_uid(int(qq))
                    except:
                        pass
                # Call send_poke. If target_uid is None, it defaults to 'you' or similar in send_poke
                ret = self.send_poke(uid, target_uid)
                if ret: 
                    last_mid = self._mk_msg_id()
                    sent_any = True

            elif stype == 'record' or stype == 'video':
                # Treat record (voice) and video as file transfer
                file_val = data.get('file')
                url_val = data.get('url')
                target_file = url_val or file_val
                if target_file:
                    fpath = self._download_to_temp(target_file)
                    if fpath:
                        # Use send_file_msg_by_uid
                        res = self.send_file_msg_by_uid(fpath, uid)
                        if res: 
                            last_mid = self._mk_msg_id()
                            sent_any = True

        if last_mid == 0 and sent_any:
            # If we sent something but didn't get a proper mid (e.g. legacy methods), generate one
            last_mid = self._mk_msg_id()
            
        return last_mid

    def _send_group_music_via_wechat(self, group_id: int, title: str, desc: str, url: str, music_url: str) -> int:
        uid = self.get_send_uid(group_id, "", 0, "")
        if not uid:
            raise RuntimeError(f"group_id={group_id} 未找到 WeChat group_uid")
        ok = self._send_music_card_by_uid(uid, title or '', desc or '', url or '', music_url or '')
        # print(f"[robot] send_group_music uid={uid} ok={ok} title={title}")
        if not ok:
            raise RuntimeError("send_music_card failed")
        return self._mk_msg_id()

    def _send_private_music_via_wechat(self, user_id: int, title: str, desc: str, url: str, music_url: str) -> int:
        uid = self.get_send_uid(0, "", user_id, "")
        if not uid:
            raise RuntimeError(f"user_id={user_id} 未找到 WeChat client_uid")
        ok = self._send_music_card_by_uid(uid, title or '', desc or '', url or '', music_url or '')
        # print(f"[robot] send_private_music uid={uid} ok={ok} title={title}")
        if not ok:
            raise RuntimeError("send_music_card failed")
        return self._mk_msg_id()

    def _get_user_avatar_path(self, user_id: int) -> str:
        try:
            uid = wx_client.get_client_uid(user_id)
            if not uid:
                return ""
            fn = self.get_head_img(uid)
            return os.path.join(self.temp_pwd, fn)
        except Exception as e:
            print(f"[robot] get_user_avatar error: {e}")
            return ""

    def _get_group_avatar_path(self, group_id: int) -> str:
        try:
            gid_uid = self._group_map_uid_by_id.get(group_id) or wx_group.get_group_uid(group_id)
            if not gid_uid:
                return ""
            fn = self.get_head_img(gid_uid)
            return os.path.join(self.temp_pwd, fn)
        except Exception as e:
            print(f"[robot] get_group_avatar error: {e}")
            return ""

    def _fetch_group_owner_uid(self, gid_uid: str) -> str:
        try:
            url = self.base_uri + '/webwxbatchgetcontact?type=ex&r=%s&pass_ticket=%s' % (int(time.time()), self.pass_ticket)
            params = {
                'BaseRequest': self.base_request,
                'Count': 1,
                'List': [{"UserName": gid_uid, "EncryChatRoomId": ""}]
            }
            r = self.session.post(url, data=json.dumps(params))
            r.encoding = 'utf-8'
            dic = json.loads(r.text)
            cl = dic.get('ContactList') or []
            if not cl:
                return ""
            g0 = cl[0]
            for k in ("ChatRoomOwner", "OwnerUin", "Owner", "OwnerUid"):
                v = g0.get(k)
                if v:
                    return v
            return g0.get('ChatRoomOwner') or ""
        except Exception as e:
            print(f"[robot] fetch_group_owner error: {e}")
            return ""

    # ========= 工具：@检测 =========
    def is_at_me(self, _msg):
        my_names = self.get_group_member_name(self.my_account["UserName"], _msg["user"]["id"]) or {}
        if self.my_account.get("NickName"):
            my_names["nickname2"] = self.my_account["NickName"]
        if self.my_account.get("RemarkName"):
            my_names["remark_name2"] = self.my_account["RemarkName"]

        return any(
            my_names.get(k) == detail["value"]
            for detail in _msg["content"].get("detail", [])
            if detail.get("type") == "at"
            for k in my_names
        )

    def _build_group_name_index(self, gid):
        idx = {
            "display_name": {},
            "remark_name": {},
            "nickname": {},
            "display_name_lower": {},
            "remark_name_lower": {},
            "nickname_lower": {},
        }
        members = getattr(self, 'group_members', {}).get(gid, []) or []
        for m in members:
            uid = m.get('UserName')
            dn = msg.remove_Emoji(m.get('DisplayName') or '')
            rn = msg.remove_Emoji(m.get('RemarkName') or '')
            nn = msg.remove_Emoji(m.get('NickName') or '')
            if dn:
                idx.setdefault("display_name", {}).setdefault(dn, []).append(uid)
                idx.setdefault("display_name_lower", {}).setdefault(dn.lower(), []).append(uid)
            if rn:
                idx.setdefault("remark_name", {}).setdefault(rn, []).append(uid)
                idx.setdefault("remark_name_lower", {}).setdefault(rn.lower(), []).append(uid)
            if nn:
                idx.setdefault("nickname", {}).setdefault(nn, []).append(uid)
                idx.setdefault("nickname_lower", {}).setdefault(nn.lower(), []).append(uid)
        self._name_index_by_group[gid] = idx
        return idx

    def _resolve_member_uid_by_name(self, gid, name):
        if not gid or not name:
            return None
        idx = self._name_index_by_group.get(gid) or self._build_group_name_index(gid)
        n = msg.remove_Emoji(name) or ''
        # 优先精确匹配：群内显示名 -> 备注名 -> 昵称
        for key in ("display_name", "remark_name", "nickname"):
            uids = idx.get(key, {}).get(n, [])
            if len(uids) == 1:
                return uids[0]
        # 次选不区分大小写
        nl = n.lower()
        for key in ("display_name_lower", "remark_name_lower", "nickname_lower"):
            uids = idx.get(key, {}).get(nl, [])
            if len(uids) == 1:
                return uids[0]
        return None

    # ========= 将 OneBot 事件异步推送给所有 C# 客户端 =========
    def _push_onebot_event(self, event: dict):
        # Record message event
        try:
            from web_ui import record_msg
            if event.get("post_type") in ["message", "notice"]:
                record_msg()
        except ImportError:
            pass

        # Filter self message if configured
        if self.gateway and hasattr(self.gateway, 'config'):
            report_self = self.gateway.config.get('report_self_message', True)
            if not report_self and event.get('post_type') == 'message':
                 if str(event.get('user_id')) == str(self.self_id):
                     # print(f"[robot] drop self message event per config")
                     return

        loop = getattr(self.gateway, "_loop", None)
        # print(f"[robot] schedule event message_type={event.get('message_type')} group_id={event.get('group_id')} user_id={event.get('user_id')} has_loop={bool(loop)}")
        if loop:
            try:
                asyncio.run_coroutine_threadsafe(self.gateway._event_queue.put(event), loop)
                # print("[robot] enqueued via run_coroutine_threadsafe")
                return
            except Exception as e:
                print(f"[robot] schedule by run_coroutine_threadsafe failed: {repr(e)}")
                try:
                    loop.call_soon_threadsafe(asyncio.create_task, self.gateway._event_queue.put(event))
                    # print("[robot] enqueued via call_soon_threadsafe(create_task)")
                    return
                except Exception as e2:
                    print(f"[robot] schedule by call_soon_threadsafe failed: {repr(e2)}")
        try:
            loop = asyncio.get_running_loop()
            loop.create_task(self.gateway._event_queue.put(event))
        except RuntimeError:
            pass

    def _cq_to_wx_at(self, message, group_uid):
        """
        Convert [CQ:at,qq=123] to @Name\u2005
        """
        try:
            def repl(m):
                qq = m.group(1)
                if qq == 'all':
                    return '@所有人\u2005'
                
                try:
                    members = self.group_members.get(group_uid, [])
                    target_uid = wx_client.get_client_uid(int(qq))
                    
                    found_member = None
                    
                    # 1. Try by UID (if mapped)
                    if target_uid:
                        for member in members:
                            if member['UserName'] == target_uid:
                                found_member = member
                                break
                    
                    # 2. If not found, try by Name (DB -> Current Member)
                    # This handles the case where WeChat UserName (UID) has changed after restart
                    if not found_member:
                        saved_name = wx_client.get_client_name_by_qq(int(qq))
                        if saved_name:
                            saved_name_clean = msg.remove_Emoji(saved_name)
                            for member in members:
                                m_nick = msg.remove_Emoji(member.get('NickName') or "")
                                m_disp = msg.remove_Emoji(member.get('DisplayName') or "")
                                if m_nick == saved_name_clean or m_disp == saved_name_clean:
                                    found_member = member
                                    # print(f"[robot] _cq_to_wx_at recovered member by name: {saved_name} -> {member.get('NickName')}")
                                    break
                    
                    if found_member:
                        name = found_member.get('DisplayName') or found_member.get('NickName')
                        if name:
                            name = msg.remove_Emoji(name)
                            return f"@{name}\u2005"
                    
                    return f"@{qq}"
                except Exception as e:
                    if self.config.get("debug"):
                        print(f"[robot] _cq_to_wx_at repl error: {e}")
                    return f"@{qq}"

            return re.sub(r'\[CQ:at,qq=(all|\d+)\]', repl, message)
        except Exception as e:
            if self.config.get("debug"):
                print(f"[robot] _cq_to_wx_at error: {e}")
            return message

    # ========= 供 C# 调用的实际微信发送实现 =========
    def _send_group_message_via_wechat(self, group_id: int, message: str) -> int:
        """
        C# -> OneBot Action -> Python -> 实际发群消息
        """
        # print(f"[robot] send_group_message_via_wechat group_id={group_id} message_len={len(message)}")
        # print(f"[robot] resolve uid start group_id={group_id}")
        send_uid = self.get_send_uid(group_id, "", 0, "")
        if not send_uid:
            print(f"[robot] resolve uid failed group_id={group_id}")
            try:
                # print("[robot] try refresh contact & mapping")
                self.get_contact()
                total = 0
                robot_qq = getattr(common, 'default_robot_qq', 0)
                for g in getattr(self, 'group_list', []) or []:
                    gid_uid = g.get('UserName')
                    gname = msg.remove_Emoji(g.get('NickName') or '')
                    try:
                        gid_num = wx_group.get_wx_group(robot_qq, gid_uid, gname, 0, '')
                        if gid_num:
                            self._group_map_uid_by_id[gid_num] = gid_uid
                            self._group_map_id_by_uid[gid_uid] = gid_num
                    except Exception:
                        pass
                    total += 1
                send_uid = self.get_send_uid(group_id, "", 0, "")
            except Exception as e:
                print(f"[robot] refresh mapping error: {e}")
            if not send_uid and self._last_active_group_uid and (int(time.time()) - self._last_active_ts) <= 60:
                if self.config.get("debug"):
                    print(f"[robot] fallback to last_active_group_uid={self._last_active_group_uid}")
                send_uid = self._last_active_group_uid
            if not send_uid:
                raise RuntimeError(f"group_id={group_id} 未找到 WeChat group_uid")
        
        # 修正：确保 send_uid 是群 ID（@@开头）
        # 如果是普通用户 ID（@开头），说明可能错误地获取到了私聊 UID
        if not send_uid.startswith('@@'):
            # 尝试再次强制查找
            real_group_uid = self._group_map_uid_by_id.get(int(group_id))
            if real_group_uid and real_group_uid.startswith('@@'):
                if self.config.get("debug"):
                    print(f"[robot] corrected send_uid from {send_uid} to {real_group_uid}")
                send_uid = real_group_uid
        
        try:
            message = msg.cq_to_wx_face(message)
            message = self._cq_to_wx_at(message, send_uid)
        except Exception:
            pass
            
        # Process message segments (Rich Media + Text)
        segments = self._parse_message_segments(message)
        last_mid = self._process_send_segments(segments, send_uid)
        
        return last_mid

    def _send_private_message_via_wechat(self, user_id: int, message: str) -> int:
        """
        C# -> OneBot Action -> Python -> 实际发私聊
        """
        send_uid = self.get_send_uid(0, "", user_id, "")
        if not send_uid:
            raise RuntimeError(f"user_id={user_id} 未找到 WeChat client_uid")
        try:
            message = msg.cq_to_wx_face(message)
        except Exception:
            pass
            
        # Process message segments (Rich Media + Text)
        segments = self._parse_message_segments(message)
        last_mid = self._process_send_segments(segments, send_uid)
        
        return last_mid

    def _find_group_name_direct(self, group_uid):
        """
        直接从内存 group_list 查找群名称
        """
        if not group_uid:
            return ""
        if hasattr(self, 'group_list'):
            for g in self.group_list:
                if g.get('UserName') == group_uid:
                    return msg.remove_Emoji(g.get('NickName') or "")
        return ""

    def _find_member_uid_by_name(self, group_uid, name):
        """
        根据昵称/群名片查找群成员 UID
        :param group_uid: 群 UID
        :param name: 成员昵称或 "我"
        :return: 成员 UID 或 None
        """
        if not name:
            return None
        if name == "我":
            return self.my_account.get('UserName', '') # 机器人自己的 UID
            
        if not hasattr(self, 'group_members') or group_uid not in self.group_members:
            return None
        
        candidates = self.group_members[group_uid]
        # 1. 尝试精确匹配 DisplayName
        for m in candidates:
            if m.get('DisplayName') == name:
                return m['UserName']
        
        # 2. 尝试精确匹配 NickName
        for m in candidates:
            if m.get('NickName') == name:
                return m['UserName']
                
        # 3. 尝试模糊匹配 (去 Emoji 后)
        clean_name = msg.remove_Emoji(name)
        for m in candidates:
            if msg.remove_Emoji(m.get('DisplayName') or "") == clean_name:
                return m['UserName']
            if msg.remove_Emoji(m.get('NickName') or "") == clean_name:
                return m['UserName']
                
        return None

    def _mk_msg_id(self):
        return int(time.time() * 1000)

    # ========= 统一处理所有消息，并转成 OneBot 事件 =========
    def handle_msg_all(self, _msg):
        self_id = common.default_robot_qq
        ct = _msg.get("create_time")
        if ct and ct < getattr(self, "boot_ts", 0):
            #print(f"[robot] drop history create_time={ct} boot_ts={self.boot_ts}")
            return
        mid_seen = _msg.get("msg_id")
        if mid_seen:
            if mid_seen in self._seen_msg_ids:
                if self.DEBUG:
                    print(f"[robot] drop duplicate msg_id={mid_seen}")
                return
            self._seen_msg_ids.add(mid_seen)
        msg_type_id = _msg.get("msg_type_id")
        content_type = _msg.get("content", {}).get("type")
        # print(f"[robot] {_msg}")
        # print(f"[robot] incoming msg msg_type_id={msg_type_id} content_type={content_type}")
        
        # 基础字段
        user_id_raw = _msg.get("user", {}).get("id", "")
        group_uid = user_id_raw if (msg_type_id == 3 or user_id_raw.startswith("@@")) else None
        
        # 补丁：如果是自己发出的群消息（如撤回），ToUserName 是群 ID
        if not group_uid and _msg.get("to_user_id", "").startswith("@@"):
            group_uid = _msg.get("to_user_id")
            # 尝试补全群名称
            try:
                # 尝试从缓存获取群信息
                # wxbot 中有 get_contact_name(uid) 方法，或者直接查 group_list
                # 这里假设 self 是 WxBot 的子类或 mixin，可以直接调用
                if hasattr(self, 'get_contact_name'):
                    g_info = self.get_contact_name(group_uid)
                    if g_info:
                        raw_gname = self.get_contact_prefer_name(g_info)
                        if raw_gname:
                            # 之前的 group_name 是在 msg_type_id==3 时从 user.name 取的
                            # 这里我们需要手动设置
                            _msg_user_name = _msg.get("user", {}).get("name")
                            # 如果之前的 name 是 self/unknown，我们不覆盖它作为 user name，但我们需要 group_name
                            pass 
                            # 注意：下面 line 1165 定义了 group_name 变量，我们需要覆盖那个变量
            except Exception:
                pass

        group_name = msg.remove_Emoji(_msg.get("user", {}).get("name") or "") if msg_type_id == 3 else ""
        if group_uid and not group_name:
             # 尝试获取群名
             if hasattr(self, 'get_contact_name'):
                 g_info = self.get_contact_name(group_uid)
                 if g_info:
                     group_name = msg.remove_Emoji(self.get_contact_prefer_name(g_info) or "")
             
             # 增强：如果还为空，尝试遍历 group_list
             if not group_name:
                 group_name = self._find_group_name_direct(group_uid)

        client_uid = _msg.get("content", {}).get("user", {}).get("id") or _msg.get("user", {}).get("id")
        name = msg.remove_Emoji(_msg.get("content", {}).get("user", {}).get("name") or _msg.get("user", {}).get("name") or "")
        display_name = remark_name = nick_name = ""
        attr_status = ""

        # 补全群与用户信息
        group_id = 0
        if group_uid:
            names = self.get_group_member_name(group_uid, client_uid) or {}
            display_name = msg.remove_Emoji(names.get("display_name", "") or "")
            remark_name = msg.remove_Emoji(names.get("remark_name", "") or "")
            nick_name = msg.remove_Emoji(names.get("nickname", "") or "")
            attr_status = self.get_attr_status(group_uid, client_uid)
            group_id = wx_group.get_wx_group(self_id, group_uid, group_name, 0, name)
            self._last_active_group_uid = group_uid
            self._last_active_ts = int(time.time())
            if group_id:
                self._group_map_uid_by_id[group_id] = group_uid
                self._group_map_id_by_uid[group_uid] = group_id
        else:
            pass

        user_id = wx_group.get_member(self_id, group_id, client_uid, name, display_name, remark_name, nick_name, attr_status)

        message_id = self._mk_msg_id()
        # OneBot 事件基底
        base_event = {
            "self_id": self.self_id,                  # OneBot self_id（Python 网关的实例 ID）
            "time": int(time.time()),
            "message_id": message_id,          # 对 message 事件有意义
            "user_id": user_id,
            "sender": {
                "user_id": user_id,
                "nickname": name,
                "card": display_name
            }
        }

        # Cache Message for Quote Reply
        try:
            svr_id = _msg.get("msg_id") or _msg.get("new_msg_id") or str(message_id)
            raw_content = _msg.get("content", {}).get("data") or ""
            # 简易内容描述
            if content_type != 0 and not raw_content:
                if content_type == 3: raw_content = "[图片]"
                elif content_type == 4: raw_content = "[语音]"
                elif content_type == 47: raw_content = "[表情]"
                elif content_type == 49: raw_content = "[链接]"
                else: raw_content = "[消息]"

            self._msg_cache[message_id] = {
                "message_id": message_id,
                "svr_id": svr_id,
                "from_uid": client_uid,
                "group_uid": group_uid,
                "sender_name": display_name or nick_name or name,
                "content": raw_content,
                "timestamp": int(time.time())
            }
            if len(self._msg_cache) > 1000:
                self._msg_cache.popitem(last=False)
        except Exception as e:
            print(f"[robot] cache msg error: {e}")
        if group_id:
            base_event["message_type"] = "group"
            base_event["group_id"] = group_id
            base_event["group_name"] = group_name
        else:
            base_event["message_type"] = "private"

        # ========== 类型分发 ==========
        # 文本
        if content_type == 0:
            is_at_me = self.is_at_me(_msg) if group_id else False
            detail = _msg["content"].get("detail", []) or []
            def _fmt_cq(guid, infos):
                out = []
                for it in infos:
                    t = it.get("type")
                    v = it.get("value") or ""
                    if t == "str":
                        out.append(v)
                    elif t == "at":
                        qq = 0
                        uid0 = self._resolve_member_uid_by_name(guid, v) if guid else None
                        if uid0:
                            # 如果是机器人自己
                            if uid0 == self.my_account.get("UserName"):
                                qq = self.self_id
                            else:
                                try:
                                    qq = wx_client.get_client_qq_by_uid(uid0)
                                except Exception:
                                    qq = 0
                        if v in ("所有人", "全体成员"):
                            out.append("[CQ:at,qq=all]")
                        elif qq:
                            out.append(f"[CQ:at,qq={qq}]")
                        else:
                            out.append(f"@{v}")
                return "".join(out)
            question = _fmt_cq(group_uid, detail) or (_msg["content"].get("desc") or _msg["content"].get("data") or "")

            # 解析文本引用消息
            try:
                # 预处理：去除首尾空白及常见不可见字符
                clean_question = question.strip()
                # [Fix] 针对网页版/UOS协议，处理 HTML 换行符 artifact (如 br/)
                # 必须先于不可见字符清理，或者一起处理
                # 特别注意: 某些情况下会出现 <\n> 这种奇怪的组合 (Hex: 3c 0a 3e)
                for char in ['br/', '<br/>', '<br>', '<br />', '<\n>', '>\n<']:
                    clean_question = clean_question.replace(char, '\n')
                    
                for char in ['\u2005', '\u200b', '\u200c', '\u200d', '\u200e', '\u200f', '\ufeff']:
                    clean_question = clean_question.replace(char, '')
                
                # 策略1：整体正则匹配（最严格）
                # 使用 unicode 转义避免编码问题: \u002d(-), \u2013(en dash), \u2014(em dash)
                # 恢复安全字符集 (不包含 br/ 等)
                quote_match = re.search(r"^\s*「(.+?)[：:](.+?)」\s*[\u002d\u2013\u2014\s]{3,}(.+)$", clean_question, re.DOTALL)
                
                # 策略2：分步特征匹配（更鲁棒）
                if not quote_match:
                    # 查找特征：收尾引号 + 可能的空白 + 至少3个分隔符
                    # 恢复安全字符集
                    split_pattern = r"」\s*[\u002d\u2013\u2014\s]{3,}"
                    split_match = re.search(split_pattern, clean_question, re.DOTALL)
                    if split_match:
                        reply_start = split_match.end()
                        q_reply = clean_question[reply_start:]
                        
                        # 提取引用部分
                        pre_separator = clean_question[:split_match.start() + 1] # 到 」 为止
                        last_open_quote = pre_separator.rfind("「")
                        
                        if last_open_quote != -1:
                            quote_body = pre_separator[last_open_quote:] # 「......」
                            inner = quote_body[1:-1] # 去掉引号
                            
                            # 尝试分割昵称和内容
                            nm_match = re.match(r"^(.+?)[：:](.+)$", inner, re.DOTALL)
                            if nm_match:
                                q_name = nm_match.group(1)
                                q_content = nm_match.group(2)
                                quote_match = True 
                            else:
                                # 即使无法解析内部格式，只要确定是引用格式，就强制通过
                                q_name = ""
                                q_content = inner
                                quote_match = True
                        else:
                             pass
                    else:
                        pass
                        
                        # 策略3：最后尝试直接按分隔符行分割（不依赖引号）
                        # 匹配连续的 3个以上 dash (允许中间有空格)
                        # 此时我们假设分隔符是一整行，或者至少前后有换行
                        strategy3_pattern = r"[\r\n]+\s*[\u002d\u2013\u2014]{1}[\u002d\u2013\u2014\s]{2,}\s*[\r\n]+"
                        s3_match = re.search(strategy3_pattern, clean_question, re.DOTALL)
                        if s3_match:
                            reply_start = s3_match.end()
                            q_reply = clean_question[reply_start:]
                            # 前半部分作为引用，尝试找引号
                            pre_part = clean_question[:s3_match.start()]
                            # 简单粗暴：认为前半部分就是引用
                            # 尝试提取名字
                            nm_match = re.search(r"「(.+?)[：:](.+?)」", pre_part, re.DOTALL)
                            if nm_match:
                                q_name = nm_match.group(1)
                                q_content = nm_match.group(2)
                            else:
                                q_name = ""
                                q_content = pre_part
                            quote_match = True
                        else:
                             pass
                             
                             # 策略4：终极兜底 - 只要发现一行只有 - 和空格
                             lines = clean_question.split('\n')
                             for i, line in enumerate(lines):
                                 s_line = line.strip()
                                 # 检查是否全是 dash 和 space，且至少有 3 个 dash
                                 if len(s_line) > 3 and re.match(r"^[\u002d\u2013\u2014\s]+$", s_line) and s_line.count('-') + s_line.count('—') + s_line.count('–') >= 3:
                                      # 这一行是分隔符
                                      # 后面的是回复，前面的是引用
                                      q_reply = '\n'.join(lines[i+1:])
                                      pre_part = '\n'.join(lines[:i])
                                      
                                      nm_match = re.search(r"「(.+?)[：:](.+?)」", pre_part, re.DOTALL)
                                      if nm_match:
                                          q_name = nm_match.group(1)
                                          q_content = nm_match.group(2)
                                      else:
                                          q_name = ""
                                          q_content = pre_part
                                      quote_match = True
                                      break



                if quote_match:
                    # 如果是策略1匹配的，从 group 中取值；如果是策略2，变量已在外部定义
                    if isinstance(quote_match, re.Match):
                        q_name = quote_match.group(1)
                        q_content = quote_match.group(2)
                        q_reply = quote_match.group(3)
                    
                    found_msg_id = None
                    # 只有当有内容时才尝试匹配缓存（避免空匹配）
                    if q_content and len(q_content) > 1:
                        for mid in reversed(list(self._msg_cache.keys())):
                            # Skip current message to avoid self-match
                            if mid == message_id:
                                continue

                            cache_item = self._msg_cache[mid]
                            c_name = cache_item.get("sender_name", "")
                            c_content = cache_item.get("content", "")
                            
                            # [Fix] 增加名字匹配校验，防止匹配到其他人的相似消息 (如引用机器人的话却匹配到用户的消息)
                            # 如果 q_name 存在，必须要求名字有一定的相似性
                            if q_name and c_name:
                                # 移除特殊字符对比
                                n1 = q_name.strip()
                                n2 = c_name.strip()
                                if n1 != n2 and n1 not in n2 and n2 not in n1:
                                    continue

                            # 尝试精确匹配或宽松匹配
                            # [Fix] Ensure content is string before 'in' check
                            c_content_str = str(c_content) if not isinstance(c_content, str) else c_content
                            
                            if (c_name == q_name and c_content_str == q_content) or \
                               (len(c_content_str) > 1 and c_content_str in q_content) or \
                               (len(q_content) > 5 and q_content in c_content_str):
                                found_msg_id = mid
                                break
                    
                    # 准备 AT 成分
                    cq_at = ""
                    
                    if found_msg_id:
                        # 场景A：找到原消息ID -> 尝试通过缓存中的 UID 获取 QQ
                        sender_uid = self._msg_cache[found_msg_id].get("from_uid")
                        # [Fix] 私聊不艾特 (只在群聊 group_uid 存在时尝试 AT)
                        if sender_uid:
                            if group_uid:
                                # [Fix] 如果发送者是机器人自己，强制使用 self_id
                                if sender_uid == self.my_account.get("UserName"):
                                    if self.self_id:
                                        cq_at = f"[CQ:at,qq={self.self_id}]"
                                else:
                                    try:
                                        qq = wx_client.get_client_qq_by_uid(sender_uid)
                                        if qq:
                                            cq_at = f"[CQ:at,qq={qq}]"
                                    except Exception:
                                        pass
                            else:
                                # 私聊：使用纯文本 @名字
                                target_name = self._msg_cache[found_msg_id].get("sender_name") or q_name
                                if target_name:
                                    cq_at = f"@{target_name} "
                        
                        # 构造最终消息：引用 + AT + 回复
                        if group_uid:
                            # 群聊：带引用 CQ 码
                            question = f"[CQ:reply,id={found_msg_id}]{cq_at}{q_reply}"
                        else:
                            # 私聊：不带引用 CQ 码，仅 AT 文本 + 回复
                            question = f"{cq_at}{q_reply}"
                            
                    
                    else:
                        # 场景B：没找到原消息ID -> 尝试通过名字反查 UID 获取 QQ (Best Effort)
                        if q_name:
                            try:
                                # 1. 优先判断是否是引用机器人自己 (仅群聊)
                                my_nick = self.my_account.get("NickName", "")
                                if group_uid and (q_name == my_nick or q_name == "我"):
                                    if self.self_id:
                                        cq_at = f"[CQ:at,qq={self.self_id}]"
                                    else:
                                        cq_at = "" 

                                # 2. 如果不是引用机器人，尝试查找其他成员
                                elif not cq_at:
                                    if group_uid:
                                        # 群聊：在群成员中查找
                                        target_uid = self._resolve_member_uid_by_name(group_uid, q_name)
                                        if target_uid:
                                            if target_uid == self.my_account.get("UserName"):
                                                if self.self_id:
                                                    cq_at = f"[CQ:at,qq={self.self_id}]"
                                            else:
                                                qq = wx_client.get_client_qq_by_uid(target_uid)
                                                if qq:
                                                    cq_at = f"[CQ:at,qq={qq}]"
                                    else:
                                        # 私聊：不加艾特
                                        pass
                            except Exception as e:
                                print(f"[robot] resolve at error: {e}")
                        
                        # 如果还没找到真实 AT，降级为文本 AT (群聊和私聊都支持)
                        if not cq_at and q_name:
                            cq_at = f"@{q_name} "
                        
                        # 构造最终消息：(无引用ID) + AT + 回复
                        # 关键点：一定要去除引用原文，只保留 AT 和回复
                        question = f"{cq_at}{q_reply}"

                else:
                    # 调试：如果不匹配，打印前100个字符看看为什么
                    pass
            except Exception as e:
                print(f"[robot] parse quote error: {e}")

            # Plugin Hook
            if hasattr(self, "plugin_manager"):
                try:
                    # Strip CQ codes for simple matching if needed, or pass as is
                    plugin_context = {
                        "content": question,
                        "sender": name,
                        "group": group_name,
                        "group_id": group_id, 
                        "user_id": user_id,
                        "msg_type": "group" if group_id else "private",
                        "bot": self
                    }
                    reply, should_block = self.plugin_manager.process(plugin_context)
                    if reply:
                        target_uid = group_uid if group_id else client_uid
                        if target_uid:
                            self.send_msg_by_uid(reply, target_uid)
                    
                    if should_block:
                        if self.config.get("debug"):
                            print(f"[gateway] Event blocked by plugin")
                        return

                except Exception as e:
                    print(f"[robot] plugin process error: {e}")

            event = dict(base_event)
            event.update({
                "post_type": "message",
                "sub_type": "normal",
                "raw_message": _msg["content"].get("data") or question,
                "message": question,
                "font": 0
            })
            # 推送给 C#
            self._push_onebot_event(event)
        # 图片 / 语音 / 推荐 / 视频 / 分享
        elif content_type in (3, 4, 5, 6, 7):
            type_map = {3: "image", 4: "voice", 5 : "contact", 6: "video", 7: "share"}
            kind = type_map[content_type]
            content = _msg.get("content", {})
            detail = content.get("detail", []) or []
            elems = []
            for d in detail:
                t = d.get("type")
                if t == "image":
                    u = d.get("value") or d.get("url")
                    if u:
                        elems.append({"type": "image", "data": {"file": u}})
                elif t == "voice":
                    u = d.get("value") or d.get("url")
                    if u:
                        elems.append({"type": "record", "data": {"file": u}})
                elif t == "contact":
                    u = d.get("value") or d.get("url")
                    if u:
                        elems.append({"type": "contact", "data": {"file": u}})
                elif t == "video":
                    u = d.get("value") or d.get("url")
                    if u:
                        elems.append({"type": "video", "data": {"file": u}})
                elif t == "share":
                    title = d.get("title") or ""
                    u = d.get("url") or ""
                    # print(f"[robot] share title={title} url={u} data={d}")
                    elems.append({"type": "json", "data": {"data": json.dumps({"title": title, "url": u}, ensure_ascii=False)}})
            if not elems:
                if content_type == 3:
                    u = content.get("data")
                    if u:
                        elems = [{"type": "image", "data": {"file": u}}]
                elif content_type == 4:
                    u = content.get("data")
                    if u:
                        elems = [{"type": "record", "data": {"file": u}}]
                elif content_type == 5:
                    u = content.get("data")
                    if u:
                        elems = [{"type": "contact", "data": {"file": u}}]
                elif content_type == 6:
                    u = content.get("data")
                    if u:
                        elems = [{"type": "video", "data": {"file": u}}]
                elif content_type == 7:
                    dct = content.get("data") or {}
                    app_msg_type = dct.get("app_msg_type")
                    
                    if str(app_msg_type) == "19": # Merged Forward
                        title = dct.get("title") or "合并转发消息"
                        desc = dct.get("desc") or ""
                        json_data = {
                            "type": "merged_forward",
                            "title": title,
                            "desc": desc,
                            "url": dct.get("url"),
                            # "xml": dct.get("content") 
                        }
                        elems = [{"type": "json", "data": {"data": json.dumps(json_data, ensure_ascii=False)}}]
                        
                    elif str(app_msg_type) == "57": # Quote Reply
                        title = dct.get("title") or "引用回复"
                        # raw_xml = dct.get("content") or ""
                        json_data = {
                            "type": "quote_reply",
                            "title": title, # This is usually the user's reply text
                            "desc": dct.get("desc") or ""
                        }
                        elems = [{"type": "json", "data": {"data": json.dumps(json_data, ensure_ascii=False)}}]
                        
                    else:
                        title = dct.get("title") or ""
                        u = dct.get("url") or ""
                        if title or u or dct:
                            elems = [{"type": "json", "data": {"data": json.dumps({"title": title, "url": u, "desc": dct.get("desc") or ""}, ensure_ascii=False)}}]
            event = dict(base_event)
            event.update({
                "post_type": "message",
                "sub_type": "normal",
                "raw_message": json.dumps(content, ensure_ascii=False)
            })
            event["message"] = elems if elems else [{"type": "json", "data": {"data": json.dumps({"kind": kind}, ensure_ascii=False)}}]
            self._push_onebot_event(event)

        # 撤回
        elif content_type == 10:
            recall_raw = _msg.get("content", {}).get("data")
            message_id_recalled = 0
            if isinstance(recall_raw, dict):
                try:
                    message_id_recalled = int(recall_raw.get("id", 0))
                except Exception:
                    message_id_recalled = 0
            elif isinstance(recall_raw, str):
                try:
                    m = re.search(r"<oldmsgid>(\d+)</oldmsgid>", recall_raw)
                    if not m:
                        m = re.search(r"<msgid>(\d+)</msgid>", recall_raw)
                    if m:
                        message_id_recalled = int(m.group(1))
                except Exception:
                    message_id_recalled = 0
            event = {
                "self_id": self.self_id,
                "time": int(time.time()),
                "post_type": "notice",
                "notice_type": "group_recall" if group_id else "friend_recall",
                "group_id": group_id or None,
                "group_name": group_name or "",
                "user_id": user_id,
                "operator_id": user_id,
                "message_id": message_id_recalled
            }
            self._push_onebot_event(event)

        # 系统提示：入群/邀请/扫码/游戏中心 等
        elif content_type == 12:
            text = _msg["content"].get("data", "") or ""
            # 匹配入群（邀请、扫码、游戏中心）
            if any([
                common.is_match(common.regex_invitation, text),
                common.is_match(common.regex_join_group_by_qrcode, text),
                common.is_match(common.regex_join_group_by_game_center, text),
            ]):
                invited_name = ""
                invit_name = ""
                if common.is_match(common.regex_invitation, text):
                    m = re.compile(common.regex_invitation).match(text)
                    invit_name = m.group("invit_name")
                    invited_name = m.group("invited_name")
                elif common.is_match(common.regex_join_group_by_qrcode, text):
                    m = re.compile(common.regex_join_group_by_qrcode).match(text)
                    invit_name = m.group("invit_name")
                    invited_name = m.group("invited_name")
                elif common.is_match(common.regex_join_group_by_game_center, text):
                    m = re.compile(common.regex_join_group_by_game_center).match(text)
                    invited_name = m.group("client_name")

                invited_name = common.removeEmoji(invited_name or "")
                event = {
                    "self_id": self.self_id,
                    "time": int(time.time()),
                    "post_type": "notice",
                    "notice_type": "group_increase",
                    "sub_type": "approve",              # 你也可以根据解析细化 sub_type
                    "group_id": group_id,
                    "group_name": group_name,
                    "user_id": 0,                        # 可在拿到被邀请人唯一ID后填充
                    "operator_id": user_id,            # 邀请者
                    "inviter_name": invit_name,
                    "member": {
                        "user_id": 0,
                        "nickname": invited_name,
                        "card": invited_name
                    }
                }
                self._push_onebot_event(event)

            # 踢出/退群 检测
            elif "移出群聊" in text or "退出了群聊" in text:
                kicked_name = ""
                operator_name = ""
                is_kick_me = False
                sub_type = "kick" # kick, kick_me, leave
                
                # 1. 他人被踢: "被踢者"被"操作者"移出群聊
                m = re.search(r'"(.*?)"\s*被\s*"(.*?)"\s*移出群聊', text)
                if m:
                    kicked_name = m.group(1)
                    operator_name = m.group(2)
                    sub_type = "kick"
                
                # 2. 自己被踢: 你被"操作者"移出群聊
                if not kicked_name:
                    m = re.search(r'你被\s*"(.*?)"\s*移出群聊', text)
                    if m:
                        is_kick_me = True
                        kicked_name = "我" 
                        operator_name = m.group(1)
                        sub_type = "kick_me"
                
                # 3. 主动退群: "退群者"退出了群聊
                if not kicked_name:
                    m = re.search(r'"(.*?)"\s*退出了群聊', text)
                    if m:
                        kicked_name = m.group(1)
                        operator_name = kicked_name # 主动退群，操作者是自己
                        sub_type = "leave"

                if kicked_name:
                    # 尝试反查 ID
                    operator_uid = ""
                    kicked_uid = ""
                    
                    if group_uid:
                        # 如果是 "我"，使用 self.my_account['UserName']
                        if kicked_name == "我":
                            kicked_uid = self.my_account.get('UserName', '')
                        else:
                            kicked_uid = self._find_member_uid_by_name(group_uid, kicked_name)
                            
                        if operator_name == "我":
                            operator_uid = self.my_account.get('UserName', '')
                        elif operator_name == kicked_name:
                            operator_uid = kicked_uid
                        else:
                            operator_uid = self._find_member_uid_by_name(group_uid, operator_name)
                    
                    op_qq = wx_client.get_client_qq_by_uid(operator_uid) if operator_uid else 0
                    kicked_qq = wx_client.get_client_qq_by_uid(kicked_uid) if kicked_uid else 0
                    
                    event = {
                        "self_id": self.self_id,
                        "time": int(time.time()),
                        "post_type": "notice",
                        "notice_type": "group_decrease",
                        "sub_type": sub_type,
                        "group_id": group_id,
                        "group_name": group_name,
                        "user_id": kicked_qq,      # 离开者/被踢者
                        "operator_id": op_qq,      # 操作者
                        "raw_info": {
                            "kicked_name": kicked_name,
                            "operator_name": operator_name
                        }
                    }
                    self._push_onebot_event(event)

            # 拍一拍检测
            elif "拍了拍" in text:
                # 尝试解析 "A" 拍了拍 "B" 或 "A" 拍了拍我
                # 注意：微信Web端显示的系统消息可能包含引号也可能不包含，视版本而定
                # 常见格式： "NickName" 拍了拍 "TargetName"
                # 或者： "NickName" 拍了拍我
                
                # 简单的正则匹配
                sender_nick = ""
                target_nick = ""
                
                # 尝试匹配带引号的
                m = re.search(r'"(.*?)"\s*拍了拍\s*"(.*?)"', text)
                if m:
                    sender_nick = m.group(1)
                    target_nick = m.group(2)
                else:
                    m = re.search(r'"(.*?)"\s*拍了拍我', text)
                    if m:
                        sender_nick = m.group(1)
                        target_nick = "我" # 表示自己
                
                if not sender_nick:
                    # 尝试不带引号的匹配（某些情况）
                    if "拍了拍我" in text:
                        sender_nick = text.split("拍了拍我")[0].strip()
                        target_nick = "我"
                    else:
                        parts = text.split("拍了拍")
                        if len(parts) >= 2:
                            sender_nick = parts[0].strip()
                            target_nick = parts[1].strip()
                
                # 尝试反查 UID 和 QQ
                final_user_id = 0
                final_target_id = 0
                
                if group_uid:
                    s_uid = self._find_member_uid_by_name(group_uid, sender_nick)
                    if s_uid:
                        final_user_id = wx_client.get_client_qq_by_uid(s_uid)
                    
                    t_uid = self._find_member_uid_by_name(group_uid, target_nick)
                    
                    # 如果精确/模糊查找失败，尝试前缀匹配（处理 "拍了拍 Bob 的头像" 等后缀情况）
                    if not t_uid and target_nick and target_nick != "我":
                        candidates = self.group_members.get(group_uid, [])
                        # 收集所有可能的名称 (name, uid)，按长度倒序排列以优先匹配长名字
                        potential_matches = []
                        for m in candidates:
                            uid = m['UserName']
                            d_name = m.get('DisplayName')
                            n_name = m.get('NickName')
                            if d_name: potential_matches.append((d_name, uid))
                            if n_name: potential_matches.append((n_name, uid))
                            # 同时也加入去 Emoji 的版本
                            if d_name: potential_matches.append((msg.remove_Emoji(d_name), uid))
                            if n_name: potential_matches.append((msg.remove_Emoji(n_name), uid))
                        
                        potential_matches.sort(key=lambda x: len(x[0]), reverse=True)
                        
                        clean_target = msg.remove_Emoji(target_nick)
                        for name, uid in potential_matches:
                            # 尝试对原始 target_nick 和 去Emoji后的 target_nick 进行前缀匹配
                            if target_nick.startswith(name) or clean_target.startswith(name):
                                t_uid = uid
                                break

                    if t_uid:
                        final_target_id = wx_client.get_client_qq_by_uid(t_uid)
                else:
                    # 私聊
                    if target_nick == "我":
                        final_user_id = user_id
                        final_target_id = self.self_id
                    elif sender_nick == "我":
                        final_user_id = self.self_id
                        final_target_id = user_id
                
                # 如果没找到，对于群消息，只能留0或者使用原来的 user_id (虽然可能是群ID)
                # 还是留0比较安全，表示未知用户

                # 确保群名存在
                if group_uid and not group_name:
                    group_name = self._find_group_name_direct(group_uid)
                
                event = {
                    "self_id": self.self_id,
                    "time": int(time.time()),
                    "post_type": "notice",
                    "notice_type": "notify",
                    "sub_type": "poke",
                    "group_id": group_id or None,
                    "group_name": group_name, # 移除 if group_id else "" 限制，确保有名字就传
                    "user_id": final_user_id,
                    "target_id": final_target_id,
                    "sender_id": final_user_id,
                    "raw_info": text
                }
                self._push_onebot_event(event)
            else:
                # 其他系统提示可映射到 notice.notify 自定义
                event = {
                    "self_id": self.self_id,
                    "time": int(time.time()),
                    "post_type": "notice",
                    "notice_type": "notify",
                    "group_id": group_id or None,
                    "group_name": group_name,
                    "user_id": user_id,
                    "operator_id": user_id,
                    "title": "system_tips",
                    "content": text
                }
                self._push_onebot_event(event)

        # 其他类型可按需扩展
        elif content_type in (11, 99):
            # 11=system, 99=videomsg（如果和 6 的视频不同可单独处理）
            event = dict(base_event)
            event.update({
                "post_type": "message",
                "sub_type": "normal",
                "raw_message": json.dumps(_msg.get("content", {}), ensure_ascii=False),
                "message": "[system]" if content_type == 11 else "[video]"
            })
            #self._push_onebot_event(event)

        # 未知类型
        else:
            # print("[robot] 未知 content_type:", content_type)
            pass
            
    def get_attr_status(self, gid, uid):
        for group in self.group_members:
            for member in self.group_members[group]:
                if group == gid and member["UserName"] == uid:
                    return member["AttrStatus"]            

    # ========= 可选：群/私聊封装（来自你的老代码） =========
    def send_message_uid(self, send_uid, robot_qq, _group_id, group_id, group_name, client_qq, client_name, question, answer_id, message, is_send=True):
        robot_name = u"指路天使"
        # print(robot_name, robot_qq, time.strftime(" %H:%M:%S"))
        if message:
            if is_send:
                self.send_msg_by_uid(message, send_uid)
            else:
                message = u"[未发送]" + message
            Color.print_green_text(message)
        else:
            # print(u"[未发送]无回复")
            pass
        # print("\n")

    def send_message(self, robot_qq, _group_id, group_id, group_name, client_qq, client_name, question, answer_id, message, is_send=True):
        send_uid = self.get_send_uid(group_id, group_name, client_qq, client_name)
        self.send_message_uid(send_uid, robot_qq, _group_id, group_id, group_name, client_qq, client_name, question, answer_id, message, is_send)

    def get_send_uid(self, group_id, group_name, client_qq, client_name):
        if group_id:
            uid = self._group_map_uid_by_id.get(group_id)
            if uid:
                return uid
            return wx_group.get_group_uid(group_id)
        else:
            return wx_client.get_client_uid(client_qq)

    def load_contacts_cache(self):
        # 捕获是否使用了缓存加载联系人
        success = super().load_contacts_cache()
        self._is_cached_contacts = success
        return success

    def sync_all_group_members(self):
        if getattr(self, '_is_cached_contacts', False):
            if self.DEBUG:
                print("[onebot] Contacts loaded from cache, skipping full sync (fast start).")
            return

        if self.DEBUG:
            print("[onebot] Start syncing all group members...")
        count = 0
        robot_qq = getattr(common, 'default_robot_qq', 0)
        
        # Ensure self.group_members exists
        if not hasattr(self, 'group_members'):
            return

        for gid, members in self.group_members.items():
            # Get group name
            gname = ""
            for g in getattr(self, 'group_list', []):
                if g.get('UserName') == gid:
                    gname = msg.remove_Emoji(g.get('NickName') or "")
                    break
            
            # Ensure group exists and get group_id
            try:
                group_id = wx_group.get_wx_group(robot_qq, gid, gname, 0, "")
            except Exception as e:
                print(f"[onebot] sync group error {gid}: {e}")
                continue
                
            if not group_id:
                continue
                
            # print(f"[onebot] Syncing members for group {gname} ({group_id})...")
            
            for m in members:
                try:
                    uid = m.get('UserName')
                    nickname = msg.remove_Emoji(m.get('NickName') or "")
                    display_name = msg.remove_Emoji(m.get('DisplayName') or "")
                    remark_name = msg.remove_Emoji(m.get('RemarkName') or "")
                    attr_status = m.get('AttrStatus') or ""
                    
                    # Trigger the sync logic in wx_client
                    # Signature: get_client_qq(robot_qq, group_id, client_uid, client_name, display_name, remark_name, nick_name, attr_status)
                    wx_client.get_client_qq(robot_qq, group_id, uid, nickname, display_name, remark_name, nickname, attr_status)
                    count += 1
                except Exception as e:
                    # print(f"[onebot] sync member error: {e}")
                    pass
                    
        if self.DEBUG:
            print(f"[onebot] Finished syncing {count} members.")

    def proc_msg(self):
        self.sync_all_group_members()
        super().proc_msg()


class onebot_work(WXWorkBot):
    """
    企业微信机器人 OneBot 适配层
    """
    def __init__(self, gateway, corpid, corpsecret, agentid):
        super().__init__(corpid, corpsecret, agentid)
        self.gateway = gateway
        self.gateway.add_bot(self, self.self_id)
        
    def _mk_msg_id(self):
        return int(time.time() * 1000)

    def _push_onebot_event(self, event):
        loop = getattr(self.gateway, "_loop", None)
        if loop:
            try:
                asyncio.run_coroutine_threadsafe(self.gateway._event_queue.put(event), loop)
                return
            except Exception as e:
                print(f"[work-bot] schedule failed: {e}")
        try:
            loop = asyncio.get_running_loop()
            loop.create_task(self.gateway._event_queue.put(event))
        except RuntimeError:
            pass

class onebot_dingtalk(DingTalkBot):
    """
    钉钉机器人 OneBot 适配层
    """
    def __init__(self, gateway, access_token, secret):
        super().__init__(access_token, secret)
        self.gateway = gateway
        self.gateway.add_bot(self, self.self_id)
        
    def _mk_msg_id(self):
        return int(time.time() * 1000)

    def _push_onebot_event(self, event):
         # DingTalk Webhook 模式主要是发送，接收需要独立 HTTP Server，此处暂留空或实现简单的 loop 注入
         pass


class onebot_feishu(FeishuBot):
    """
    飞书机器人 OneBot 适配层
    """
    def __init__(self, gateway, webhook_url, secret):
        super().__init__(webhook_url, secret)
        self.gateway = gateway
        self.gateway.add_bot(self, self.self_id)
        
    def _mk_msg_id(self):
        return int(time.time() * 1000)

    def _push_onebot_event(self, event):
        pass

class onebot_telegram(TelegramBot):
    """
    Telegram 机器人 OneBot 适配层
    """
    def __init__(self, gateway, token):
        super().__init__(token)
        self.gateway = gateway
        self.gateway.add_bot(self, self.self_id)
        
    def _mk_msg_id(self):
        return int(time.time() * 1000)

    def _push_onebot_event(self, event):
        # 实际应从 getUpdates 轮询中获取并推送
        pass


class BotManager:
    def __init__(self):
        self.config = self.load_config()
        self.gateways = {}  # port -> OneBotGateway
        self.bots = []
        self._loop = None
        
        # Default config
        ws_conf = self.config.get("network", {}).get("ws_server", {})
        self.default_host = ws_conf.get("host", "0.0.0.0")
        self.default_port = ws_conf.get("port", 3001)
        self.default_config = ws_conf

        # Create default gateway
        self._get_or_create_gateway(self.default_port)
        
        # Start gateway loop
        t = threading.Thread(target=self._start_event_loop, daemon=True)
        t.start()
        
    def load_config(self):
        if os.path.exists(CONFIG_FILE):
            try:
                with open(CONFIG_FILE, 'r', encoding='utf-8') as f:
                    return json.load(f)
            except Exception as e:
                print(f"[BotManager] Load config failed: {e}")
        return DEFAULT_CONFIG.copy()

    def save_config(self, new_config):
        try:
            self.config.update(new_config)
            with open(CONFIG_FILE, 'w', encoding='utf-8') as f:
                json.dump(self.config, f, indent=4, ensure_ascii=False)
            return True
        except Exception as e:
            print(f"[BotManager] Save config failed: {e}")
            return False

    def _get_or_create_gateway(self, port):
        if port in self.gateways:
            return self.gateways[port]
        
        print(f"[BotManager] Creating gateway on port {port}")
        gw_conf = self.default_config.copy()
        gw_conf['port'] = port
        gw = OneBotGateway(host=self.default_host, port=port, config=gw_conf)
        self.gateways[port] = gw
        
        # If loop is running, start it
        if self._loop and self._loop.is_running():
            gw._loop = self._loop
            asyncio.run_coroutine_threadsafe(gw.start(), self._loop)
            
        return gw

    def _start_event_loop(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        self._loop = loop
        
        tasks = []
        # Use list() to avoid "dictionary changed size during iteration"
        for gw in list(self.gateways.values()):
            gw._loop = loop
            tasks.append(loop.create_task(gw.start()))
            
        if tasks:
            loop.run_until_complete(asyncio.gather(*tasks))
        else:
            loop.run_forever()
        
    @property
    def gateway(self):
        # Compatibility for single-port access (e.g. from WebUI)
        # Returns the default port gateway
        return self.gateways.get(self.default_port)

    def add_bot(self, self_id=None, port=None):
        target_port = port or self.default_port
        gw = self._get_or_create_gateway(target_port)
        
        bot = onebot(gw, self_id)
        bot.global_config = self.config
        bot.DEBUG = False
        if not hasattr(bot, "conf"):
            bot.conf = {}
        bot.conf["qr"] = "png"
        
        self.bots.append(bot)
        
        # Start bot in thread
        t = threading.Thread(target=bot.run, daemon=True)
        t.start()
        return bot

    def add_work_bot(self, corpid, corpsecret, agentid, port=None):
        target_port = port or self.default_port
        gw = self._get_or_create_gateway(target_port)
        
        bot = onebot_work(gw, corpid, corpsecret, agentid)
        self.bots.append(bot)
        t = threading.Thread(target=bot.run, daemon=True)
        t.start()
        return bot

    def add_dingtalk_bot(self, access_token, secret, port=None):
        target_port = port or self.default_port
        gw = self._get_or_create_gateway(target_port)
        
        bot = onebot_dingtalk(gw, access_token, secret)
        self.bots.append(bot)
        t = threading.Thread(target=bot.run, daemon=True)
        t.start()
        return bot

    def add_feishu_bot(self, webhook_url, secret, port=None):
        target_port = port or self.default_port
        gw = self._get_or_create_gateway(target_port)
        
        bot = onebot_feishu(gw, webhook_url, secret)
        self.bots.append(bot)
        t = threading.Thread(target=bot.run, daemon=True)
        t.start()
        return bot

    def add_telegram_bot(self, token, port=None):
        target_port = port or self.default_port
        gw = self._get_or_create_gateway(target_port)
        
        bot = onebot_telegram(gw, token)
        self.bots.append(bot)
        t = threading.Thread(target=bot.run, daemon=True)
        t.start()
        return bot

# ========== 入口 ==========

def main():
    manager = BotManager()
    
    # Load bots from config
    bots_config = manager.config.get("bots", [])
    
    if not bots_config:
        # Default behavior: Add one personal WeChat bot
        print("[main] No 'bots' config found, starting default personal WeChat bot...")
        manager.add_bot(self_id=1098299491)
    else:
        print(f"[main] Found {len(bots_config)} bots in config")
        for bot_conf in bots_config:
            btype = bot_conf.get("type", "wechat")
            port = bot_conf.get("port")
            
            if btype == "wechat":
                sid = bot_conf.get("self_id")
                print(f"[main] Adding WeChat bot self_id={sid} port={port or 'default'}")
                manager.add_bot(self_id=sid, port=port)
            elif btype == "wxwork":
                corpid = bot_conf.get("corpid")
                secret = bot_conf.get("corpsecret")
                agentid = bot_conf.get("agentid")
                if corpid and secret and agentid:
                    print(f"[main] Adding WXWork bot agentid={agentid} port={port or 'default'}")
                    manager.add_work_bot(corpid, secret, agentid, port=port)
                else:
                    print(f"[main] Invalid wxwork config: {bot_conf}")
            elif btype == "dingtalk":
                token = bot_conf.get("access_token")
                secret = bot_conf.get("secret")
                if token:
                    print(f"[main] Adding DingTalk bot token={token[:6]}... port={port or 'default'}")
                    manager.add_dingtalk_bot(token, secret, port=port)
                else:
                    print(f"[main] Invalid dingtalk config: {bot_conf}")
            elif btype == "feishu":
                webhook = bot_conf.get("webhook_url")
                secret = bot_conf.get("secret")
                if webhook:
                    print(f"[main] Adding Feishu bot... port={port or 'default'}")
                    manager.add_feishu_bot(webhook, secret, port=port)
                else:
                    print(f"[main] Invalid feishu config: {bot_conf}")
            elif btype == "telegram":
                token = bot_conf.get("token")
                if token:
                    print(f"[main] Adding Telegram bot... port={port or 'default'}")
                    manager.add_telegram_bot(token, port=port)
                else:
                    print(f"[main] Invalid telegram config: {bot_conf}")

    # 启动 WebUI
    try:
        start_web_ui(manager, port=5000)
    except Exception as e:
        print(f"[WebUI] Startup failed: {e}")

    # Keep main thread alive
    while True:
        time.sleep(1)

if __name__ == "__main__":
    main()
