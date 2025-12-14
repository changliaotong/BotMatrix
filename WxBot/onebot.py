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
import xml.etree.ElementTree as ET

import os
from bots.wxbot import WXBot
from SQLConn import *
from color import *
from wxgroup import wx_group
from wxclient import wx_client
from common import common
from msg import msg
from web_ui import WebUI

CONFIG_FILE = "config.json"
DEFAULT_CONFIG = {
    "network": {
        "ws_server": {
            "name": "test",
            "host": "0.0.0.0",
            "port": 3001,
            "heartbeat_interval": 30000,
            "token": ""
        }
    },
    "bots": []
}

class ManagerAdapter:
    def __init__(self, bot, config):
        self.bots = [bot]
        self.config = config
    
    def save_config(self, data):
        # Update config in memory and save to file
        try:
            self.config.update(data)
            with open(CONFIG_FILE, 'w', encoding='utf-8') as f:
                json.dump(self.config, f, indent=4, ensure_ascii=False)
            return True
        except Exception as e:
            print(f"[ManagerAdapter] Save config error: {e}")
            return False

    def add_bot(self, self_id):
        # We don't support adding bots dynamically in this mode
        print("[ManagerAdapter] add_bot not supported in single-bot mode")
        pass

# ========== 你的机器人（融合） ==========

class onebot(WXBot):
    """
    - 继承你原来的 WXBot
    - 在 __init__ 中启动 OneBot WS 服务器（另起线程）
    - 接收到微信消息时，组装 OneBot 事件，通过 gateway 推送给 C#
    - 保留 super().handle_msg_all(_msg)
    """

    def __init__(self, self_id: int = None):
        super().__init__()
        
        # Override QR config: default to 'tty' for Docker environments
        self.conf['qr'] = os.getenv("WX_QR_MODE", "tty")
        print(f"[onebot] Configured QR Code Mode: {self.conf['qr']}")

        self.self_id = int(self_id) if self_id else 0
        
        if self.self_id:
            self.cache_file = os.path.join(self.temp_pwd, f'session_{self.self_id}.json')
            self.qr_file_path = os.path.join(self.temp_pwd, f'wxqr_{self.self_id}.png')
            
            default_session = os.path.join(self.temp_pwd, 'session.json')
            if not os.path.exists(self.cache_file) and os.path.exists(default_session):
                print(f"[onebot] Migrating default session to {self.cache_file}")
                try:
                    import shutil
                    shutil.copy(default_session, self.cache_file)
                except Exception as e:
                    print(f"[onebot] Migration failed: {e}")

        self.boot_ts = int(time.time())
        self._seen_msg_ids = set()
        self._last_active_group_uid = None
        self._last_active_ts = 0
        self._group_map_uid_by_id = {}
        self._group_map_id_by_uid = {}
        self._name_index_by_group = {}
        self._msg_cache = collections.OrderedDict()
        
        self.driver = None

    def set_driver(self, driver):
        self.driver = driver

    def _ensure_group_cache(self):
        if self._group_map_id_by_uid:
            return

        try:
            print("[onebot] Loading group cache...")
            conn = SQLConn.conn()
            if not conn: return
            
            cursor = conn.cursor()
            cursor.execute("SELECT group_uid, group_id FROM wx_group")
            
            rows = cursor.fetchall()
            
            count = 0
            for row in rows:
                try:
                    uid = None
                    gid = None
                    if isinstance(row, dict):
                         uid = row['group_uid']
                         gid = row['group_id']
                    else:
                         uid = row[0]
                         gid = row[1]
                    
                    if uid and gid:
                        self._group_map_uid_by_id[gid] = uid
                        self._group_map_id_by_uid[uid] = gid
                        count += 1
                except:
                    pass
            
            conn.close()
            print(f"[onebot] Loaded {count} groups into cache")
            
        except Exception as e:
            print(f"[onebot] Cache load error: {e}")

    def _find_user_id_by_name(self, group_uid, group_id, name):
        """
        Find user_id by name (NickName or DisplayName) in a group.
        """
        if not group_uid or not name:
            return 0
            
        if not hasattr(self, 'group_members') or group_uid not in self.group_members:
            return 0

        members = self.group_members[group_uid]
        target_m = None
        for m in members:
            dname = m.get('DisplayName')
            nname = m.get('NickName')
            
            # Check exact match
            if (dname and dname == name) or (nname and nname == name):
                target_m = m
                break
        
        if target_m:
            target_uid = target_m.get('UserName')
            
            # Check if target is self
            my_uid = getattr(self, 'my_account', {}).get('UserName')
            if my_uid and target_uid == my_uid:
                 return self.self_id

            try:
                nickname = target_m.get('NickName') or ""
                display_name = target_m.get('DisplayName') or ""
                
                user_id = wx_client.get_client_qq(self.self_id, group_id, target_uid, nickname, display_name, "", nickname, "")
                if user_id:
                    return user_id
            except:
                pass
        
        return 0

    def _convert_at_to_cq(self, message, group_uid, group_id):
        """
        Convert @Name in message to [CQ:at,qq=ID]
        """
        if not group_uid or not message or "@" not in message:
            return message
            
        def replace(match):
            original_text = match.group(0)
            name = match.group(1)
            
            if not name:
                return original_text
                
            if name == "所有人" or name == "All":
                return "[CQ:at,qq=all]"
                
            user_id = self._find_user_id_by_name(group_uid, group_id, name)
            if user_id:
                return f"[CQ:at,qq={user_id}]"
            
            return original_text

        # Pattern matches @Name followed by space, \u2005, or end of string.
        pattern = re.compile(r'@(.*?)(?:\s|\u2005|$)')
        return pattern.sub(replace, message)

    def _parse_cq_code(self, message, group_id=None):
        """
        Parse CQ codes in message and convert to WeChat friendly format.
        Currently supports: [CQ:at,qq=...]
        """
        if not message:
            return ""

        # 1. Handle [CQ:at,qq=...]
        # Regex to find all CQ:at
        at_pattern = re.compile(r'\[CQ:at,qq=([^,\]]+)(?:,name=[^,\]]*)?\]')
        
        def replace_at(match):
            qq = match.group(1)
            if qq == "all":
                return "@所有人"
            
            # Try to resolve QQ to Nickname
            nickname = str(qq)
            try:
                # 1. Get UID from QQ
                uid = wx_client.get_client_uid(int(qq))
                if uid:
                    # 2. If group_id provided, look in group members
                    if group_id:
                        gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                        if gid_uid and hasattr(self, 'group_members') and gid_uid in self.group_members:
                             for m in self.group_members[gid_uid]:
                                 if m.get('UserName') == uid:
                                     nickname = m.get('NickName') or m.get('DisplayName') or str(qq)
                                     break
                    
                    # 3. If still just QQ or not found in group, try contact list
                    if nickname == str(qq):
                        for c in getattr(self, 'contact_list', []) or []:
                             if c.get('UserName') == uid:
                                 nickname = c.get('NickName') or c.get('RemarkName') or str(qq)
                                 break
            except Exception:
                pass
            
            return f"@{msg.remove_Emoji(nickname)}"

        message = at_pattern.sub(replace_at, message)
        
        # 2. Handle [CQ:image,file=...] -> [图片]
        # For now, just replace with text indicating image
        img_pattern = re.compile(r'\[CQ:image,[^\]]*\]')
        message = img_pattern.sub(r'[图片]', message)
        
        # 3. Handle [CQ:face,id=...] -> WeChat Face
        message = msg.cq_to_wx_face(message)
        
        # 4. Handle [CQ:reply,id=...] -> [回复] (Strip it or keep it)
        reply_pattern = re.compile(r'\[CQ:reply,[^\]]*\]')
        message = reply_pattern.sub(r'', message) # Reply is usually implicit or handled by app

        return message

    # 你的自定义函数 ...
    def execute_onebot_action(self, action: str, params: dict = None) -> dict:
        """
        Execute a OneBot action locally on this bot instance.
        Returns a dict with keys: status, retcode, data, msg (optional)
        """
        if params is None:
            params = {}
        
        name = action
        result = {"status": "ok", "retcode": 0, "data": {}}
        
        # if name in ["send_group_msg", "send_private_msg"]:
        #      print(f"[onebot] Execute Action: {name}, params={str(params)[:100]}...")
        
        try:
            if name == "send_group_msg":
                group_id = params.get("group_id")
                raw_message = params.get("message", "")
                
                # Parse CQ Code
                message = self._parse_cq_code(str(raw_message), group_id)
                
                # Try to get group name for logging
                group_name = "Unknown"
                try:
                    gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                    if gid_uid:
                        for g in getattr(self, 'group_list', []) or []:
                            if g.get('UserName') == gid_uid:
                                group_name = msg.remove_Emoji(g.get('NickName') or '')
                                break
                except:
                    pass

                print(f"[SEND] [Group: {group_name}({group_id})] {message}")
                mid = self._send_group_message_via_wechat(group_id, message)
                info = getattr(self, '_last_send_info', {})
                result["data"] = {"message_id": mid, "debug_ret": info.get('ret'), "http_status": info.get('status')}

            elif name == "send_private_msg":
                user_id = params.get("user_id")
                raw_message = params.get("message", "")
                
                # Parse CQ Code
                message = self._parse_cq_code(str(raw_message))
                
                # Try to get user nickname
                nickname = "Unknown"
                try:
                    uid = wx_client.get_client_uid(user_id)
                    if uid:
                        for c in getattr(self, 'contact_list', []) or []:
                             if c.get('UserName') == uid:
                                 nickname = msg.remove_Emoji(c.get('NickName') or '')
                                 break
                except:
                    pass

                print(f"[SEND] [Private: {nickname}({user_id})] {message}")
                mid = self._send_private_message_via_wechat(user_id, message)
                info = getattr(self, '_last_send_info', {})
                result["data"] = {"message_id": mid, "debug_ret": info.get('ret'), "http_status": info.get('status')}

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

            elif name == "get_group_list":
                self._ensure_group_cache()
                items = []
                for g in getattr(self, 'group_list', []) or []:
                    group_uid = g.get('UserName')
                    group_id = self._group_map_id_by_uid.get(group_uid, 0)
                    
                    group_name = msg.remove_Emoji(g.get('NickName') or '')
                    
                # Update group name in DB if needed
                    if group_id:
                        try:
                            # Update group info to ensure latest name is saved
                            wx_group.update(self.self_id, group_uid, group_id, group_name, 0, "")
                        except Exception as e:
                            print(f"[onebot] wx_group.update error: {e}")

                    if not group_id:
                        try:
                            # Use get_wx_group to find or create/update
                            # Pass 0 for client_qq, relying on wxgroup.update to preserve it
                            group_id = wx_group.get_wx_group(self.self_id, group_uid, group_name, 0, "")
                            
                            if group_id:
                                self._group_map_id_by_uid[group_uid] = group_id
                                self._group_map_uid_by_id[group_id] = group_uid
                        except Exception as e:
                            print(f"[onebot] get_wx_group error: {e}")
                            group_id = 0
                            
                    member_count = 0
                    if hasattr(self, 'group_members') and group_uid in self.group_members:
                        member_count = len(self.group_members[group_uid])
                        
                    items.append({
                        "group_id": group_id, 
                        "group_name": group_name,
                        "member_count": member_count,
                        "max_member_count": member_count
                    })
                result["data"] = items

            elif name == "get_group_member_list":
                group_id = params.get("group_id")
                gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
                items = []
                if gid_uid and hasattr(self, 'group_members') and gid_uid in self.group_members:
                     for m in self.group_members[gid_uid]:
                        uid = m.get('UserName')
                        
                        # Extract names first to pass to get_client_qq
                        nickname = msg.remove_Emoji(m.get('NickName') or m.get('RemarkName') or '')
                        card = msg.remove_Emoji(m.get('DisplayName') or '')
                        remark = msg.remove_Emoji(m.get('RemarkName') or '')
                        
                        user_id = 0
                        try:
                            # Pass extracted names to create/update client
                            user_id = wx_client.get_client_qq(self.self_id, group_id, uid, nickname, card, remark, nickname, "")
                        except Exception:
                            user_id = 0
                        
                        # Parse Sex
                        sex_val = m.get('Sex', 0)
                        sex = "unknown"
                        if sex_val == 1: sex = "male"
                        elif sex_val == 2: sex = "female"
                        
                        # Default fields
                        role = "member"
                        # Try to detect owner if possible (not easily available in simple WXBot structure without extra calls)
                        
                        items.append({
                            "group_id": group_id,
                            "user_id": user_id, 
                            "nickname": nickname, 
                            "card": card,
                            "role": role,
                            "sex": sex,
                            "age": 0,
                            "area": "",
                            "join_time": 0,
                            "last_sent_time": 0,
                            "level": "1",
                            "unfriendly": False,
                            "title": "",
                            "title_expire_time": 0,
                            "card_changeable": False
                        })
                result["data"] = items
            
            elif name == "get_login_info":
                 info = getattr(self, 'my_account', {})
                 nickname = msg.remove_Emoji(info.get('NickName') or "Unknown")
                 result["data"] = {"user_id": self.self_id, "nickname": nickname}

            else:
                result.update({"status": "failed", "retcode": 10002, "msg": f"unsupported action {name}"})
                
        except Exception as e:
            print(f"[gateway] action error name={name} params={params} msg={e}")
            result.update({"status": "failed", "retcode": 10002, "msg": str(e)})
        return result

    def _send_group_message_via_wechat(self, group_id, message):
        """
        Helper to send group message
        """
        gid_uid = (self._group_map_uid_by_id.get(group_id) if hasattr(self, '_group_map_uid_by_id') else None) or wx_group.get_group_uid(group_id)
        if not gid_uid:
            print(f"[onebot] Error: Unknown group_id {group_id}")
            return 0
        
        try:
            self.send_msg_by_uid(message, gid_uid)
            return int(time.time()) # Fake ID
        except Exception as e:
            print(f"[onebot] Send Error: {e}")
            return 0

    def _send_private_message_via_wechat(self, user_id, message):
        """
        Helper to send private message
        """
        uid = wx_client.get_client_uid(user_id)
        if not uid:
             print(f"[onebot] Error: Unknown user_id {user_id}")
             return 0
        
        try:
            self.send_msg_by_uid(message, uid)
            return int(time.time())
        except Exception as e:
            print(f"[onebot] Send Private Error: {e}")
            return 0
            
    def _fetch_group_owner_uid(self, gid_uid):
        # Need to fetch group details if not available
        # This is hard because WXBot doesn't always have owner info cached
        # Try to find in group_members if 'AttrStatus' or 'IsOwner' is available?
        # WXBot 'group_members' list usually contains basic info.
        # This is a placeholder.
        return ""

    def _push_onebot_event(self, event):
        if self.driver:
            if event.get("post_type") == "message":
                 mid = event.get("message_id")
                 print(f"[onebot] Event: Received message {mid} from WeChat. Pushing to Driver.")
            self.driver.broadcast_event(event)

    def _download_to_temp(self, file_or_url: str) -> str:
        # Simplified download
        return file_or_url

    def handle_msg_all(self, _msg):
        # Override to hook message processing
        
        # Filter history messages
        create_time = _msg.get('create_time', 0)
        if create_time and int(create_time) < self.boot_ts:
            # print(f"[onebot] Ignored history message from {create_time}")
            return

        # 1. Convert to OneBot Event and Push
        try:
            # print(f"[onebot] DEBUG: Received msg: {_msg}")
            
            msg_type_id = _msg.get('msg_type_id')
            # Filter unsupported types if needed, but let's try to handle most text/media
            
            # Initialize OneBot Event
            event = {
                "self_id": self.self_id,
                "time": int(time.time()),
                "post_type": "message",
            }

            is_group = False
            user_data = _msg.get('user', {}) # The "From" user/group
            from_uid = user_data.get('id', '')
            
            # Determine if group message
            # msg_type_id 3 is Group
            if msg_type_id == 3 or (from_uid and from_uid.startswith('@@')):
                is_group = True
                event["message_type"] = "group"
                event["sub_type"] = "normal"
                group_uid = from_uid
                
                # Get Group ID (int)
                group_id = 0
                try:
                    group_id = wx_group.get_group_id(group_uid)
                    if group_id:
                        self._group_map_uid_by_id[group_id] = group_uid
                except:
                    pass
                event["group_id"] = group_id
                
                # Sender Info
                content = _msg.get('content')
                sender_uid = ""
                sender_name = "Unknown"
                text = ""
                
                if isinstance(content, dict) and 'user' in content:
                    sender_uid = content['user'].get('id', '')
                    sender_name = content['user'].get('name', 'Unknown')
                    text = content.get('data', '')
                else:
                    # Fallback
                    sender_uid = ""
                    text = str(content) if content else ""
                    
                # Map Sender UID to User ID (int)
                user_id = 0
                if sender_uid:
                    try:
                        # self.self_id is robot_qq
                        user_id = wx_client.get_client_qq(self.self_id, group_id, sender_uid, sender_name, "", "", "", "")
                    except:
                        pass
                event["user_id"] = user_id
                
                # Try to enrich sender info from group members cache
                card = ""
                role = "member"
                if hasattr(self, 'group_members') and group_uid in self.group_members:
                    for m in self.group_members[group_uid]:
                         if m.get('UserName') == sender_uid:
                             # Enrich nickname if unknown or simple
                             real_nick = msg.remove_Emoji(m.get('NickName') or '')
                             if sender_name == "Unknown" or not sender_name:
                                 sender_name = real_nick
                             
                             card = msg.remove_Emoji(m.get('DisplayName') or '')
                             break

                event["sender"] = {
                    "user_id": user_id,
                    "nickname": sender_name,
                    "card": card,
                    "role": role,
                    "sex": "unknown",
                    "age": 0
                }
                
            else:
                # Private Message
                event["message_type"] = "private"
                event["sub_type"] = "friend"
                
                sender_uid = from_uid
                sender_name = user_data.get('name', 'Unknown')
                
                # Map Sender UID to User ID (int)
                user_id = 0
                if sender_uid:
                    try:
                        # Use get_client_qq to ensure user exists in DB
                        # robot_qq, group_id, client_uid, client_name, display_name, remark_name, nick_name, attr_status
                        user_id = wx_client.get_client_qq(self.self_id, 0, sender_uid, sender_name, sender_name, "", sender_name, "")
                    except Exception as e:
                        print(f"[onebot] Failed to map private user_id: {e}")
                        pass
                event["user_id"] = user_id
                
                # Content
                content = _msg.get('content')
                text = ""
                if isinstance(content, dict):
                    text = content.get('data', '')
                else:
                    text = str(content) if content else ""
                    
                # Try to enrich sender info
                remark = ""
                for c in getattr(self, 'contact_list', []) or []:
                     if c.get('UserName') == sender_uid:
                         if sender_name == "Unknown":
                             sender_name = msg.remove_Emoji(c.get('NickName') or '')
                         remark = msg.remove_Emoji(c.get('RemarkName') or '')
                         break

                event["sender"] = {
                    "user_id": user_id,
                    "nickname": sender_name,
                    "card": remark,
                    "sex": "unknown",
                    "age": 0
                }

            # 2. Content Handling (Text, Image, etc.)
            msg_content_type = 0
            if isinstance(_msg.get('content'), dict):
                msg_content_type = _msg['content'].get('type', 0)
            
            final_message = text
            
            # Handle Text Quote/Reply format: "「...」\n- - - - - - - - - - - - - - -\nReply"
            if msg_content_type == 0 and "- - - - - - - - - - - - - - -" in final_message:
                parts = final_message.split("- - - - - - - - - - - - - - -")
                if len(parts) > 1:
                    quote_part = parts[0].strip()
                    # The last part is the reply content
                    final_message = parts[-1].strip()
                    # Clean up strange characters
                    final_message = final_message.replace('<\n>', '').replace('<br/>', '\n')
                    
                    # Try to extract sender from quote_part and add At
                    # Format: 「Name：Content」
                    if is_group and quote_part.startswith("「") and "：" in quote_part:
                        try:
                            # Remove starting 「
                            temp = quote_part[1:]
                            # Split by first full-width colon
                            q_parts = temp.split("：", 1)
                            if len(q_parts) >= 2:
                                q_name = q_parts[0]
                                q_uid = self._find_user_id_by_name(group_uid, group_id, q_name)
                                if q_uid:
                                    final_message = f"[CQ:at,qq={q_uid}] {final_message}"
                        except Exception as e:
                            print(f"[onebot] Failed to parse quote sender: {e}")

            # Convert @ to CQ code for group text messages
            if is_group and msg_content_type == 0:
                final_message = self._convert_at_to_cq(final_message, group_uid, group_id)
            
            # WXBot content types: 0=Text, 3=Image, 4=Voice, etc.
            if msg_content_type == 3: # Image
                # data might be url or id
                # WXBot usually needs fetching, but let's assume we can put a placeholder or url
                img_data = _msg['content'].get('data', '')
                final_message = f"[CQ:image,file={img_data}]"
            elif msg_content_type == 4: # Voice
                voice_data = _msg['content'].get('data', '')
                final_message = f"[CQ:record,file={voice_data}]"
            elif msg_content_type == 5: # Card
                final_message = f"[Card: {text}]"
            
            event["message"] = final_message
            event["raw_message"] = final_message
            event["font"] = 0
            event["message_id"] = _msg.get('MsgId', int(time.time()))
            
            # Get Group Name if available
            group_name = "Unknown"
            if is_group:
                try:
                    # Try to find group name from group_list
                    for g in getattr(self, 'group_list', []) or []:
                        if g.get('UserName') == group_uid:
                            group_name = msg.remove_Emoji(g.get('NickName') or '')
                            break
                    if not group_name and group_uid:
                         # Fallback if not found in list
                         group_name = "Group"
                except:
                    pass
                
                # ADDED: Include group_name in event
                event["group_name"] = group_name
            
            # Log received message (Compact)
            log_time = time.strftime("%H:%M:%S", time.localtime(event.get('time')))
            if is_group:
                 print(f"[RECV] [{log_time}] [Group: {group_name}({event.get('group_id')})] [User: {sender_name}({event.get('user_id')})] {final_message}")
            else:
                 print(f"[RECV] [{log_time}] [Private] [User: {sender_name}({event.get('user_id')})] {final_message}")
            
            # Print Raw JSON
            print(f"[RAW] {json.dumps(event, ensure_ascii=False)}")

            # Ignore self messages (if sent by me on another device)
            my_uid = getattr(self, 'my_account', {}).get('UserName')
            if sender_uid == my_uid:
                 # It's a sync message
                 pass
            else:
                 # Push Event
                 if final_message or msg_content_type != 0:
                     self._push_onebot_event(event)

        except Exception as e:
            print(f"[onebot] handle_msg_all error: {e}")
            import traceback
            traceback.print_exc()
            
        super().handle_msg_all(_msg)

    def _convert_and_push(self, msg_data):
        # This function converts WXBot msg_data to OneBot event
        try:
            msg_type_id = msg_data.get('msg_type_id')
            
            # Filter supported types: 1=Text, 3=Image, 34=Voice, 47=Emoji, 49=AppMsg(Quote/File)
            if msg_type_id not in [1, 3, 34, 47, 49]:
                return

            user = msg_data.get('user', {})
            # user dict: {'id': '...uid...', 'name': '...'}
            uid = user.get('id')
            
            # Determine if group
            # In WXBot, if msg_type_id is group message, user['id'] is group_uid usually?
            # Or msg_data has 'user' as sender?
            # WXBot logic: 
            # if group msg: msg['user']['id'] is group_uid. msg['content']['user']['id'] is sender_uid?
            
            # Let's use the provided content structure from WXBot
            content = msg_data.get('content')
            
            # Initialize Event
            event = {
                "self_id": self.self_id,
                "time": int(time.time()),
                "post_type": "message",
            }
            
            is_group = False
            group_uid = ""
            sender_uid = ""
            
            # Check message source type
            # WXBot usually sets msg_data['msg_type_id'] based on content analysis
            # But the raw type is msg_data['MsgType']
            
            # Check if it is from a group
            # WXBot's handle_msg_all logic:
            # if msg['msg_type_id'] == 4 (Init) ...
            # Typically WXBot structure:
            # msg['user']['name'] -> sender name (or group name)
            # msg['user']['id'] -> sender uid (or group uid)
            
            # If from group, the 'id' starts with '@@'
            if uid and uid.startswith('@@'):
                is_group = True
                group_uid = uid
                # Sender inside content
                if isinstance(content, dict):
                    sender_uid = content.get('user', {}).get('id')
                    text = content.get('data', '')
                else:
                    # Parse text for sender? WXBot usually parses it.
                    # If content is string, maybe it's already stripped?
                    text = str(content)
            else:
                is_group = False
                sender_uid = uid
                if isinstance(content, dict):
                     text = content.get('data', '')
                else:
                     text = str(content)

            # Resolve numeric IDs
            if is_group:
                event["message_type"] = "group"
                # Map group_uid to group_id (int)
                # We need a consistent mapping. 
                # WXGroup.get_group_id uses hash or DB.
                try:
                    group_id = wx_group.get_group_id(group_uid)
                    self._group_map_uid_by_id[group_id] = group_uid
                except:
                    group_id = 0
                event["group_id"] = group_id
                
                # Sender
                try:
                    user_id = wx_client.get_client_qq(self.self_id, group_id, sender_uid, "", "", "", "", "")
                except:
                    user_id = 0
                event["user_id"] = user_id
                
            else:
                event["message_type"] = "private"
                try:
                    user_id = wx_client.get_client_qq_by_uid(sender_uid)
                except:
                    user_id = 0
                event["user_id"] = user_id

            # Ignore self messages
            if sender_uid == getattr(self, 'my_account', {}).get('UserName'):
                return

            # Construct Message Segment
            # Simple text for now
            # Handle Image/Voice later
            
            # Handle Quotes/AppMsg (Type 49)
            if msg_type_id == 49:
                try:
                    # Clean up XML string (sometimes has garbage at start)
                    xml_content = text
                    # If it doesn't start with <, find the first <
                    idx = xml_content.find('<msg>')
                    if idx != -1:
                        xml_content = xml_content[idx:]
                    
                    root = ET.fromstring(xml_content)
                    appmsg = root.find('appmsg')
                    if appmsg is not None:
                        app_type = appmsg.find('type')
                        if app_type is not None and app_type.text == '57': # Refer/Quote
                            # It is a quote message
                            title = appmsg.find('title')
                            reply_text = title.text if title is not None else ""
                            
                            refer = appmsg.find('refermsg')
                            if refer is not None:
                                svrid = refer.find('svrid')
                                ref_id = svrid.text if svrid is not None else ""
                                
                                displayname = refer.find('displayname')
                                ref_name = displayname.text if displayname is not None else ""
                                
                                refer_content = refer.find('content')
                                ref_content = refer_content.text if refer_content is not None else ""
                                
                                # Construct CQ Code
                                # [CQ:reply,id=...] [CQ:at,qq=...] Text
                                prefix = ""
                                if ref_id:
                                    prefix += f"[CQ:reply,id={ref_id}]"
                                
                                if is_group and ref_name:
                                     ref_uid = self._find_user_id_by_name(group_uid, group_id, ref_name)
                                     if ref_uid:
                                          prefix += f" [CQ:at,qq={ref_uid}]"

                                # WeChat quote usually doesn't include @Name automatically in the text content
                                # But sometimes it does.
                                # Let's just use the reply text.
                                text = f"{prefix} {reply_text}"
                                # Clean up strange characters
                                text = text.replace('<\n>', '').replace('<br/>', '\n')
                                
                                # Try to convert @ in reply_text as well
                                if is_group:
                                    text = self._convert_at_to_cq(text, group_uid, group_id)
                except Exception as e:
                    print(f"[onebot] Failed to parse Type 49 XML: {e}")
                    pass

            event["message"] = text
            event["raw_message"] = text
            event["font"] = 0
            event["message_id"] = msg_data.get('MsgId', int(time.time()))
            
            # Sender Info
            event["sender"] = {
                "user_id": event["user_id"],
                "nickname": "Unknown", # Fill if possible
                "sex": "unknown",
                "age": 0
            }
            
            self._push_onebot_event(event)
            
        except Exception as e:
            print(f"[onebot] conversion error: {e}")

    def schedule(self):
        # Periodic tasks
        # Sync Group Members
        # WXBot doesn't automatically sync members details unless we ask.
        # We can do it periodically.
        now = int(time.time())
        if now - getattr(self, '_last_sync_ts', 0) > 3600: # Every hour
            self._last_sync_ts = now
            threading.Thread(target=self.sync_all_group_members, daemon=True).start()
            
        super().schedule()
        
    def sync_all_group_members(self):
        # Iterate all groups and call wx_client.get_client_qq to populate DB/Cache
        # print("[onebot] Syncing group members...")
        robot_qq = self.self_id
        
        count = 0
        for g in getattr(self, 'group_list', []) or []:
            gid = g.get('UserName')
            members = g.get('MemberList', [])
            if not members:
                continue
                
            # Get Group ID
            try:
                gname = msg.remove_Emoji(g.get('NickName') or "")
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

# ========== 入口 ==========

def main():
    print("[main] Starting WxBot (OneBot Driver Mode)...")
    
    # Import Driver locally to avoid circular import if driver imports onebot
    try:
        from driver import OneBotDriver
    except ImportError:
        import sys
        sys.path.append(os.path.dirname(os.path.abspath(__file__)))
        from driver import OneBotDriver

    # Load config
    config = DEFAULT_CONFIG.copy()
    if os.path.exists(CONFIG_FILE):
        try:
            with open(CONFIG_FILE, 'r', encoding='utf-8') as f:
                loaded_config = json.load(f)
                config.update(loaded_config)
        except Exception as e:
            print(f"[main] Error loading config: {e}")

    # Config default host to localhost if 0.0.0.0 (since we are client now)
    ws_conf = config.get("network", {}).get("ws_server", {})
    host = ws_conf.get("host", "127.0.0.1")
    if host == "0.0.0.0":
        host = "127.0.0.1" # Connect to localhost if configured as server default
    port = ws_conf.get("port", 3001)

    # Determine self_id
    bots_config = config.get("bots", [])
    self_id = 0
    
    # Priority 1: Environment Variable
    env_self_id = os.getenv("BOT_SELF_ID")
    if env_self_id:
        try:
            self_id = int(env_self_id)
            print(f"[main] Using BOT_SELF_ID={self_id} from environment")
        except:
            print(f"[main] Invalid BOT_SELF_ID: {env_self_id}")

    # Priority 2: Config file
    if not self_id and bots_config:
        # Find first wechat bot
        target_conf = next((b for b in bots_config if b.get("type", "wechat") == "wechat"), None)
        if target_conf:
            self_id = target_conf.get("self_id")
            print(f"[main] Using self_id={self_id} from config")
    
    # Priority 3: Default fallback
    if not self_id:
        self_id = 1098299491 
        print(f"[main] Using default fallback self_id={self_id}")
    
    print(f"[main] Initializing WXBot for self_id={self_id}...")
    
    # Initialize Driver
    # Default to 3001, but allow env var override (e.g. for local testing to avoid conflict)
    if os.getenv("WS_PORT"):
        print(f"[main] Using WS_PORT={os.getenv('WS_PORT')} from environment")
    
    driver = OneBotDriver()
    
    # Enable internal WS server (Positive WS)
    driver.ws_enable = True
    
    # Configure Driver with BotNexus URL (Reverse WS)
    # If MANAGER_URL is set, driver picks it up automatically. 
    # Otherwise we add a default one pointing to localhost:3001 (BotNexus default)
    if not os.getenv("MANAGER_URL") and not os.getenv("WS_URLS"):
        nexus_host = "127.0.0.1"
        nexus_port = 3001
        # If we are in Docker, we might want "botnexus" as host, but let's stick to localhost default for now
        # or allow override via BOTNEXUS_HOST
        if os.getenv("BOTNEXUS_HOST"):
            nexus_host = os.getenv("BOTNEXUS_HOST")
            
        nexus_url = f"ws://{nexus_host}:{nexus_port}/"
        print(f"[main] Adding default BotNexus URL: {nexus_url}")
        driver.ws_urls.append(nexus_url)
    
    # Initialize Bot
    bot = onebot(self_id=self_id)
    bot.global_config = config
    
    # Link Bot and Driver
    driver.add_bot(bot, self_id)
    bot.set_driver(driver)
    
    # Start Bot (Thread)
    print("[main] Starting WXBot thread...")
    bot_thread = threading.Thread(target=bot.run, daemon=True)
    bot_thread.start()
    
    # Start WebUI (Thread)
    try:
        print("[main] Starting WebUI on port 5001...")
        adapter = ManagerAdapter(bot, config)
        web_ui = WebUI(adapter)
        
        def run_webui():
            # Disable flask banner
            import logging
            log = logging.getLogger('werkzeug')
            log.setLevel(logging.ERROR)
            web_ui.app.run(host='0.0.0.0', port=5001, debug=False, use_reloader=False)
            
        webui_thread = threading.Thread(target=run_webui, daemon=True)
        webui_thread.start()
    except Exception as e:
        print(f"[main] Failed to start WebUI: {e}")

    # Start Driver (Blocking Asyncio Loop)
    print("[main] Starting Driver Loop...")
    try:
        asyncio.run(driver.start())
    except KeyboardInterrupt:
        print("[main] Stopped by user")


if __name__ == "__main__":
    main()
