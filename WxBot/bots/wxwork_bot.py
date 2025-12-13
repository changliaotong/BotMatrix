# coding: utf-8
import requests
import json
import time
import threading
import logging

class WXWorkBot:
    """
    企业微信机器人 (自建应用模式)
    """
    def __init__(self, corpid, corpsecret, agentid):
        self.corpid = corpid
        self.corpsecret = corpsecret
        self.agentid = agentid
        self.self_id = agentid # 使用 agentid 作为 bot 的标识
        self.access_token = None
        self.token_expires_time = 0
        
        # 为了兼容 OneBotGateway，需要维护一些属性
        self.group_list = [] 
        self.contact_list = []
        self._last_send_info = {}

    def get_access_token(self):
        """获取或刷新 Access Token"""
        if self.access_token and time.time() < self.token_expires_time:
            return self.access_token
        
        url = f"https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid={self.corpid}&corpsecret={self.corpsecret}"
        try:
            r = requests.get(url)
            data = r.json()
            if data.get("errcode") == 0:
                self.access_token = data.get("access_token")
                # 提前 200 秒过期，防止边界情况
                self.token_expires_time = time.time() + data.get("expires_in", 7200) - 200
                return self.access_token
            else:
                logging.error(f"[WXWork] 获取 Token 失败: {data}")
                return None
        except Exception as e:
            logging.error(f"[WXWork] 获取 Token 异常: {e}")
            return None

    def _send_group_message_via_wechat(self, group_id, message):
        """
        发送群消息
        注意：企业微信 API 中，group_id 对应 chatid
        """
        token = self.get_access_token()
        if not token:
            return None
            
        url = f"https://qyapi.weixin.qq.com/cgi-bin/appchat/send?access_token={token}"
        payload = {
            "chatid": str(group_id),
            "msgtype": "text",
            "text": {
                "content": message
            },
            "safe": 0
        }
        
        try:
            r = requests.post(url, json=payload)
            data = r.json()
            self._last_send_info = {"status": r.status_code, "ret": data}
            if data.get("errcode") == 0:
                return "ok" # 返回一个伪 ID
            else:
                logging.error(f"[WXWork] 发送群消息失败: {data}")
                return None
        except Exception as e:
            logging.error(f"[WXWork] 发送群消息异常: {e}")
            return None

    def _send_private_message_via_wechat(self, user_id, message):
        """
        发送私聊消息
        """
        token = self.get_access_token()
        if not token:
            return None
            
        url = f"https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token={token}"
        payload = {
            "touser": str(user_id),
            "agentid": self.agentid,
            "msgtype": "text",
            "text": {
                "content": message
            },
            "safe": 0
        }
        
        try:
            r = requests.post(url, json=payload)
            data = r.json()
            self._last_send_info = {"status": r.status_code, "ret": data}
            if data.get("errcode") == 0:
                return "ok"
            else:
                logging.error(f"[WXWork] 发送私聊消息失败: {data}")
                return None
        except Exception as e:
            logging.error(f"[WXWork] 发送私聊消息异常: {e}")
            return None

    def run(self):
        """
        启动逻辑，可以在这里启动接收消息的 HTTP Server
        目前仅作为发送端使用
        """
        print(f"[WXWork] Bot {self.agentid} initialized (Sender Mode)")
