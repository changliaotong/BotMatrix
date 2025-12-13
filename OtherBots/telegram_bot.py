# coding: utf-8
import requests
import json
import time
import logging

class TelegramBot:
    """
    Telegram Bot (Polling Mode / Webhook Mode)
    目前简单实现发送功能
    """
    def __init__(self, token):
        self.token = token
        self.base_url = f"https://api.telegram.org/bot{token}"
        # Telegram Bot Token 格式通常是 123456:ABC-DEF...，前半部分是 ID
        try:
            self.self_id = int(token.split(':')[0])
        except:
            self.self_id = 0
            
        # 兼容性字段
        self.group_list = [] 
        self.contact_list = []
        self._last_send_info = {}

    def _send_group_message_via_wechat(self, group_id, message):
        """
        发送群消息
        group_id 必须是 Telegram Chat ID (通常是负数)
        """
        return self._send_msg(group_id, message)

    def _send_private_message_via_wechat(self, user_id, message):
        """
        发送私聊消息
        user_id 必须是 Telegram User ID (正数)
        """
        return self._send_msg(user_id, message)

    def _send_msg(self, chat_id, text):
        url = f"{self.base_url}/sendMessage"
        payload = {
            "chat_id": chat_id,
            "text": text
        }
        try:
            # 需要设置代理？Telegram 在国内通常需要代理
            # 这里简单实现，不带代理配置，如果运行环境无法访问 api.telegram.org 会失败
            r = requests.post(url, json=payload, timeout=10)
            data = r.json()
            self._last_send_info = {"status": r.status_code, "ret": data}
            if data.get("ok"):
                return str(data.get("result", {}).get("message_id"))
            else:
                logging.error(f"[Telegram] 发送失败: {data}")
                return None
        except Exception as e:
            logging.error(f"[Telegram] 发送异常: {e}")
            return None

    def run(self):
        print(f"[Telegram] Bot initialized (Sender Mode) self_id={self.self_id}")
        # 可以在这里启动 getUpdates 轮询
