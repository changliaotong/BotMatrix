# coding: utf-8
import requests
import json
import time
import hmac
import hashlib
import base64
import logging

class FeishuBot:
    """
    飞书机器人 (自定义机器人/Webhook模式)
    """
    def __init__(self, webhook_url, secret=None):
        self.webhook_url = webhook_url
        self.secret = secret
        # 使用 webhook url 的 hash 作为 self_id
        self.self_id = int(hashlib.md5(webhook_url.encode('utf-8')).hexdigest()[:8], 16)
        
        # 兼容性字段
        self.group_list = [] 
        self.contact_list = []
        self._last_send_info = {}

    def _gen_sign(self, timestamp):
        # 拼接timestamp和secret
        string_to_sign = '{}\n{}'.format(timestamp, self.secret)
        hmac_code = hmac.new(string_to_sign.encode("utf-8"), digestmod=hashlib.sha256).digest()
        # 对结果进行base64处理
        sign = base64.b64encode(hmac_code).decode('utf-8')
        return sign

    def _send_group_message_via_wechat(self, group_id, message):
        """
        发送群消息
        """
        payload = {
            "msg_type": "text",
            "content": {
                "text": message
            }
        }
        
        if self.secret:
            timestamp = str(int(time.time()))
            sign = self._gen_sign(timestamp)
            payload["timestamp"] = timestamp
            payload["sign"] = sign
            
        try:
            r = requests.post(self.webhook_url, json=payload)
            data = r.json()
            self._last_send_info = {"status": r.status_code, "ret": data}
            if data.get("code") == 0:
                return "ok"
            else:
                logging.error(f"[Feishu] 发送失败: {data}")
                return None
        except Exception as e:
            logging.error(f"[Feishu] 发送异常: {e}")
            return None

    def _send_private_message_via_wechat(self, user_id, message):
        """
        飞书 Webhook 同样主要针对群，可通过 @用户 提醒
        user_id 应为 open_id 或 user_id
        """
        text = f"<at user_id=\"{user_id}\"></at> {message}"
        payload = {
            "msg_type": "text",
            "content": {
                "text": text
            }
        }

        if self.secret:
            timestamp = str(int(time.time()))
            sign = self._gen_sign(timestamp)
            payload["timestamp"] = timestamp
            payload["sign"] = sign
            
        try:
            r = requests.post(self.webhook_url, json=payload)
            data = r.json()
            if data.get("code") == 0:
                return "ok"
            return None
        except Exception:
            return None

    def run(self):
        print(f"[Feishu] Bot initialized (Webhook Mode) self_id={self.self_id}")
