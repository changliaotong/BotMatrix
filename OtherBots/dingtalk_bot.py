# coding: utf-8
import requests
import json
import time
import hmac
import hashlib
import base64
import urllib.parse
import logging

class DingTalkBot:
    """
    钉钉机器人 (自定义机器人/Webhook模式)
    """
    def __init__(self, access_token, secret=None):
        self.access_token = access_token
        self.secret = secret
        # 钉钉 Webhook 不需要 agentid，但为了兼容 OneBot 结构，可以使用 token 的一部分或 hash 作为 self_id
        self.self_id = int(hashlib.md5(access_token.encode('utf-8')).hexdigest()[:8], 16)
        
        # 兼容性字段
        self.group_list = [] 
        self.contact_list = []
        self._last_send_info = {}

    def _get_webhook_url(self):
        url = f"https://oapi.dingtalk.com/robot/send?access_token={self.access_token}"
        if self.secret:
            timestamp = str(round(time.time() * 1000))
            secret_enc = self.secret.encode('utf-8')
            string_to_sign = '{}\n{}'.format(timestamp, self.secret)
            string_to_sign_enc = string_to_sign.encode('utf-8')
            hmac_code = hmac.new(secret_enc, string_to_sign_enc, digestmod=hashlib.sha256).digest()
            sign = urllib.parse.quote_plus(base64.b64encode(hmac_code))
            url = f"{url}&timestamp={timestamp}&sign={sign}"
        return url

    def _send_group_message_via_wechat(self, group_id, message):
        """
        发送群消息
        对于钉钉 Webhook，group_id 其实没有意义，因为 Webhook 本身就绑定了一个群。
        但为了接口一致性，我们忽略 group_id 或将其作为校验。
        """
        url = self._get_webhook_url()
        
        # 钉钉 Markdown 消息支持
        payload = {
            "msgtype": "text",
            "text": {
                "content": message
            }
        }
        
        try:
            r = requests.post(url, json=payload)
            data = r.json()
            self._last_send_info = {"status": r.status_code, "ret": data}
            if data.get("errcode") == 0:
                return "ok"
            else:
                logging.error(f"[DingTalk] 发送失败: {data}")
                return None
        except Exception as e:
            logging.error(f"[DingTalk] 发送异常: {e}")
            return None

    def _send_private_message_via_wechat(self, user_id, message):
        """
        钉钉 Webhook 不支持私聊发送。
        这里可以通过 @用户 实现伪私聊提醒，或者直接忽略。
        """
        # 尝试通过 @手机号 提醒
        url = self._get_webhook_url()
        payload = {
            "msgtype": "text",
            "text": {
                "content": message
            },
            "at": {
                "atUserIds": [str(user_id)], # 假设 user_id 是钉钉 userid
                # "atMobiles": [str(user_id)] # 或者假设是手机号
            }
        }
        
        try:
            r = requests.post(url, json=payload)
            data = r.json()
            if data.get("errcode") == 0:
                return "ok"
            return None
        except Exception:
            return None

    def run(self):
        print(f"[DingTalk] Bot initialized (Webhook Mode) self_id={self.self_id}")
