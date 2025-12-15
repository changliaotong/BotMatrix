#!/usr/bin/env python
# coding: utf-8

import os
import sys
import traceback
import webbrowser
import mimetypes
import json
import xml.dom.minidom
import urllib
import time
import re
import random
import hashlib
from traceback import format_exc

import pyqrcode
import requests
import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from requests.exceptions import ConnectionError, ReadTimeout
import urllib.parse
import html

from SQLConn import *
from wxgroup import *
from msg import *


UNKONWN = 'unkonwn'
SUCCESS = '200'
SCANED = '201'
TIMEOUT = '408'


def show_image(file_path):
    """
    跨平台显示图片文件
    :param file_path: 图片文件路径
    """
    if sys.version_info >= (3, 3):
        from shlex import quote
    else:
        from pipes import quote

    if sys.platform == "darwin":
        command = "open -a /Applications/Preview.app %s&" % quote(file_path)
        os.system(command)
    else:
        webbrowser.open(os.path.join(os.getcwd(),'temp',file_path))


class SafeSession(requests.Session):
    def request(self, method, url, params=None, data=None, headers=None, cookies=None, files=None, auth=None,
                timeout=None, allow_redirects=True, proxies=None, hooks=None, stream=None, verify=None, cert=None,
                json=None):
        for i in range(3):
            try:
                return super(SafeSession, self).request(method, url, params, data, headers, cookies, files, auth,
                                                        timeout,
                                                        allow_redirects, proxies, hooks, stream, verify, cert, json)
            except Exception as e:
                #print(e.message, traceback.format_exc())
                continue


class WXBot:
    """WXBot功能类"""

    def __init__(self):
        self.DEBUG = False
        self.uuid = ''
        self.base_uri = ''
        self.base_host = ''
        self.redirect_uri = ''
        self.uin = ''
        self.sid = ''
        self.skey = ''
        self.pass_ticket = ''
        self.device_id = 'e' + repr(random.random())[2:17]
        self.base_request = {}
        self.sync_key_str = ''
        self.sync_key = []
        self.sync_host = ''

        #文件缓存目录
        self.temp_pwd  =  os.path.join(os.getcwd(),'temp')
        if os.path.exists(self.temp_pwd) == False:
            os.makedirs(self.temp_pwd)

        self.session = SafeSession()
        self.session.headers.update({'User-Agent': 'Mozilla/5.0 (X11; Linux i686; U;) Gecko/20070322 Kazehakase/0.4.5'})
        self.conf = {'qr': 'png'}
        self.cache_file = os.path.join(self.temp_pwd, 'session.json')
        self.qr_file_path = os.path.join(self.temp_pwd, 'wxqr.png')

        self.is_ready = False # Flag to indicate if bot is fully initialized (logged in & contacts loaded)

        self.my_account = {}  # 当前账户

        # 所有相关账号: 联系人, 公众号, 群组, 特殊账号
        self.member_list = []

        # 所有群组的成员, {'group_id1': [member1, member2, ...], ...}
        self.group_members = {}

        # 所有账户, {'group_member':{'id':{'type':'group_member', 'info':{}}, ...}, 'normal_member':{'id':{}, ...}}
        self.account_info = {'group_member': {}, 'normal_member': {}}

        self.contact_list = []  # 联系人列表
        self.public_list = []  # 公众账号列表
        self.group_list = []  # 群聊列表
        self.special_list = []  # 特殊账号列表
        self.encry_chat_room_id_list = []  # 存储群聊的EncryChatRoomId，获取群内成员头像时需要用到

        self.file_index = 0


    def get_contact(self):
        """获取当前账户的所有相关账号(包括联系人、公众号、群聊、特殊账号)"""
        url = self.base_uri + '/webwxgetcontact?pass_ticket=%s&skey=%s&r=%s' \
                              % (self.pass_ticket, self.skey, int(time.time()))
        r = self.session.post(url, data='{}')
        r.encoding = 'utf-8'
        if self.DEBUG:
            with open(os.path.join(self.temp_pwd,'contacts.json'), 'w') as f:
                f.write(r.text.encode('utf-8'))
        dic = json.loads(r.text)
        self.member_list = dic['MemberList']

        special_users = ['newsapp', 'fmessage', 'filehelper', 'weibo', 'qqmail',
                         'fmessage', 'tmessage', 'qmessage', 'qqsync', 'floatbottle',
                         'lbsapp', 'shakeapp', 'medianote', 'qqfriend', 'readerapp',
                         'blogapp', 'facebookapp', 'masssendapp', 'meishiapp',
                         'feedsapp', 'voip', 'blogappweixin', 'weixin', 'brandsessionholder',
                         'weixinreminder', 'wxid_novlwrv3lqwv11', 'gh_22b87fa7cb3c',
                         'officialaccounts', 'notification_messages', 'wxid_novlwrv3lqwv11',
                         'gh_22b87fa7cb3c', 'wxitil', 'userexperience_alarm', 'notification_messages']

        self.contact_list = []
        self.public_list = []
        self.special_list = []
        self.group_list = []

        for contact in self.member_list:
            if contact['VerifyFlag'] & 8 != 0:  # 公众号
                self.public_list.append(contact)
                self.account_info['normal_member'][contact['UserName']] = {'type': 'public', 'info': contact}
            elif contact['UserName'] in special_users:  # 特殊账户
                self.special_list.append(contact)
                self.account_info['normal_member'][contact['UserName']] = {'type': 'special', 'info': contact}
            elif contact['UserName'].find('@@') != -1:  # 群聊
                self.group_list.append(contact)
                self.account_info['normal_member'][contact['UserName']] = {'type': 'group', 'info': contact}
            elif contact['UserName'] == self.my_account['UserName']:  # 自己
                self.account_info['normal_member'][contact['UserName']] = {'type': 'self', 'info': contact}
            else:
                self.contact_list.append(contact)
                self.account_info['normal_member'][contact['UserName']] = {'type': 'contact', 'info': contact}

        self.batch_get_group_members()

        for group in self.group_members:
            for member in self.group_members[group]:
                if member['UserName'] not in self.account_info:
                    self.account_info['group_member'][member['UserName']] = \
                        {'type': 'group_member', 'info': member, 'group': group}

        if self.DEBUG:
            with open(os.path.join(self.temp_pwd,'contact_list.json'), 'w') as f:
                f.write(json.dumps(self.contact_list))
            with open(os.path.join(self.temp_pwd,'special_list.json'), 'w') as f:
                f.write(json.dumps(self.special_list))
            with open(os.path.join(self.temp_pwd,'group_list.json'), 'w') as f:
                f.write(json.dumps(self.group_list))
            with open(os.path.join(self.temp_pwd,'public_list.json'), 'w') as f:
                f.write(json.dumps(self.public_list))
            with open(os.path.join(self.temp_pwd,'member_list.json'), 'w') as f:
                f.write(json.dumps(self.member_list))
            with open(os.path.join(self.temp_pwd,'group_users.json'), 'w') as f:
                f.write(json.dumps(self.group_members))
            with open(os.path.join(self.temp_pwd,'account_info.json'), 'w') as f:
                f.write(json.dumps(self.account_info))
        self.save_contacts_cache()
        return True

    def batch_get_group_members(self):
        """批量获取所有群聊成员信息"""
        
        group_members = {}
        encry_chat_room_id = {}
        
        # 分批获取，每批 50 个
        CHUNK_SIZE = 50
        total_groups = len(self.group_list)
        
        for i in range(0, total_groups, CHUNK_SIZE):
            chunk = self.group_list[i:i + CHUNK_SIZE]
            
            url = self.base_uri + '/webwxbatchgetcontact?type=ex&r=%s&pass_ticket=%s' % (int(time.time()), self.pass_ticket)
            params = {
                'BaseRequest': self.base_request,
                "Count": len(chunk),
                "List": [{"UserName": group['UserName'], "EncryChatRoomId": group.get('EncryChatRoomId', "")} for group in chunk]
            }
            
            try:
                r = self.session.post(url, data=json.dumps(params))
                r.encoding = 'utf-8'
                dic = json.loads(r.text)
                
                # 检查 RetCode
                ret = dic.get('BaseResponse', {}).get('Ret')
                if ret != 0:
                    print(f"[wxbot] batch_get_group_members error in batch {i}: Ret={ret}")
                    continue

                for group in dic.get('ContactList', []):
                    gid = group['UserName']
                    members = group['MemberList']
                    group_members[gid] = members
                    encry_chat_room_id[gid] = group['EncryChatRoomId']
                    # print(f"[wxbot] Fetched {len(members)} members for group {group.get('NickName', 'Unknown')} ({gid})")

                    
                # 避免请求过快
                time.sleep(0.5)
                
            except Exception as e:
                print(f"[wxbot] batch_get_group_members exception in batch {i}: {e}")

        self.group_members = group_members
        self.encry_chat_room_id_list = encry_chat_room_id

    def get_group_member_name(self, gid, uid):
        """
        获取群聊中指定成员的名称信息
        :param gid: 群id
        :param uid: 群聊成员id
        :return: 名称信息，类似 {"display_name": "test_user", "nickname": "test", "remark_name": "for_test" }
        """
        if gid not in self.group_members:
            return None
        group = self.group_members[gid]
        for member in group:
            if member['UserName'] == uid:
                names = {}
                if 'RemarkName' in member and member['RemarkName']:
                    names['remark_name'] = member['RemarkName']
                if 'NickName' in member and member['NickName']:
                    names['nickname'] = member['NickName']
                if 'DisplayName' in member and member['DisplayName']:
                    names['display_name'] = member['DisplayName']
                if 'attr_status' in member and member['AttrStatus']:
                    names['attr_status'] = member['attr_status']    
                return names
        return None

    def set_group_member_name(self, gid, uid, names):
        if gid not in self.group_members:
            return None
        group = self.group_members[gid]
        for member in group:
            if member['UserName'] == uid:
                if 'RemarkName' in names and names['RemarkName']:
                    member['RemarkName'] =names['remark_name'] 
                if 'NickName' in names and names['NickName']:
                    member['nickname'] = names['NickName']
                if 'DisplayName' in names and names['DisplayName']:
                    member['display_name'] = names['DisplayName']
        return None

    def get_contact_info(self, uid):
        return self.account_info['normal_member'].get(uid)


    def get_group_member_info(self, uid):
        return self.account_info['group_member'].get(uid)

    def get_contact_name(self, uid):
        info = self.get_contact_info(uid)
        if info is None:
            return None
        info = info['info']
        name = {}
        if 'RemarkName' in info and info['RemarkName']:
            name['remark_name'] = info['RemarkName']
        if 'NickName' in info and info['NickName']:
            name['nickname'] = info['NickName']
        if 'DisplayName' in info and info['DisplayName']:
            name['display_name'] = info['DisplayName']
        if len(name) == 0:
            return None
        else:
            return name

    @staticmethod
    def get_contact_prefer_name(name):
        if name is None:
            return None
        if 'remark_name' in name:
            return name['remark_name']
        if 'nickname' in name:
            return name['nickname']
        if 'display_name' in name:
            return name['display_name']
        return None

    @staticmethod
    def get_group_member_prefer_name(name):
        if name is None:
            return None
        if 'remark_name' in name:
            return name['remark_name']
        if 'display_name' in name:
            return name['display_name']
        if 'nickname' in name:
            return name['nickname']
        return None

    def load_session_cache(self):
        try:
            if not os.path.exists(self.cache_file):
                print(f"[wxbot] Session file not found: {self.cache_file}")
                return False
            with open(self.cache_file, 'r') as f:
                content = f.read()
                if not content:
                    print("[wxbot] Session file is empty")
                    return False
                data = json.loads(content)
            self.uin = data.get('uin', '')
            self.sid = data.get('sid', '')
            self.skey = data.get('skey', '')
            self.pass_ticket = data.get('pass_ticket', '')
            self.base_uri = data.get('base_uri', '')
            self.base_host = data.get('base_host', '')
            self.device_id = data.get('device_id', self.device_id)
            self.base_request = {
                'Uin': self.uin,
                'Sid': self.sid,
                'Skey': self.skey,
                'DeviceID': self.device_id,
            }
            cookies_dict = data.get('cookies') or {}
            try:
                self.session.cookies = requests.utils.cookiejar_from_dict(cookies_dict, cookiejar=None, overwrite=True)
            except Exception:
                pass
            
            valid = all([self.uin, self.sid, self.skey, self.pass_ticket, self.base_uri])
            if valid:
                print(f"[wxbot] Session loaded successfully. Uin={self.uin}")
            else:
                print(f"[wxbot] Session loaded but invalid (missing fields).")
            return valid
        except Exception as e:
            print(f"[wxbot] Load session cache failed: {e}")
            return False

    def save_session_cache(self):
        try:
            data = {
                'uin': self.uin,
                'sid': self.sid,
                'skey': self.skey,
                'pass_ticket': self.pass_ticket,
                'base_uri': self.base_uri,
                'base_host': self.base_host,
                'device_id': self.device_id,
                'cookies': requests.utils.dict_from_cookiejar(self.session.cookies)
            }
            with open(self.cache_file, 'w') as f:
                f.write(json.dumps(data))
            return True
        except Exception as e:
            print(f"[ERROR] Save session failed: {e}")
            return False

    def save_contacts_cache(self):
        try:
            data = {
                'member_list': self.member_list,
                'group_members': self.group_members,
                'encry_chat_room_id_list': self.encry_chat_room_id_list,
                'contact_list': self.contact_list,
                'group_list': self.group_list,
                'public_list': self.public_list,
                'special_list': self.special_list,
                'account_info': self.account_info
            }
            with open(os.path.join(self.temp_pwd, 'contacts_cache.json'), 'w', encoding='utf-8') as f:
                json.dump(data, f, ensure_ascii=False)
            return True
        except Exception as e:
            print(f"[ERROR] Save contacts cache failed: {e}")
            return False

    def load_contacts_cache(self):
        try:
            path = os.path.join(self.temp_pwd, 'contacts_cache.json')
            if not os.path.exists(path):
                return False
            with open(path, 'r', encoding='utf-8') as f:
                data = json.load(f)
            self.member_list = data.get('member_list', [])
            self.group_members = data.get('group_members', {})
            self.encry_chat_room_id_list = data.get('encry_chat_room_id_list', {})
            self.contact_list = data.get('contact_list', [])
            self.group_list = data.get('group_list', [])
            self.public_list = data.get('public_list', [])
            self.special_list = data.get('special_list', [])
            self.account_info = data.get('account_info', {'group_member': {}, 'normal_member': {}})
            return True
        except Exception as e:
            print(f"[ERROR] Load contacts cache failed: {e}")
            return False

    def get_user_type(self, wx_user_id):
        """
        获取特定账号与自己的关系
        :param wx_user_id: 账号id:
        :return: 与当前账号的关系
        """
        for account in self.contact_list:
            if wx_user_id == account['UserName']:
                return 'contact'
        for account in self.public_list:
            if wx_user_id == account['UserName']:
                return 'public'
        for account in self.special_list:
            if wx_user_id == account['UserName']:
                return 'special'
        for account in self.group_list:
            if wx_user_id == account['UserName']:
                return 'group'
        for group in self.group_members:
            for member in self.group_members[group]:
                if member['UserName'] == wx_user_id:
                    return 'group_member'
        return 'unknown'

    def is_contact(self, uid):
        for account in self.contact_list:
            if uid == account['UserName']:
                return True
        return False

    def is_public(self, uid):
        for account in self.public_list:
            if uid == account['UserName']:
                return True
        return False

    def is_special(self, uid):
        for account in self.special_list:
            if uid == account['UserName']:
                return True
        return False

    def handle_msg_all(self, msg):
        """
        处理所有消息，请子类化后覆盖此函数
        msg:
            msg_id  ->  消息id
            msg_type_id  ->  消息类型id
            user  ->  发送消息的账号id
            content  ->  消息内容
        :param msg: 收到的消息
        """
        pass

    @staticmethod
    def proc_at_info(msg):
        if not msg:
            return '', []
        segs = msg.split(u'\u2005')
        str_msg_all = ''
        str_msg = ''
        infos = []
        if len(segs) > 1:
            for i in range(0, len(segs) - 1):
                segs[i] += u'\u2005'
                if not segs[i]:
                    continue
                match = re.search(u'@.*\u2005', segs[i])
                if match:
                    pm = match.group()
                    name = pm[1:-1]
                    string = segs[i].replace(pm, '')
                    str_msg_all += string + '@' + name + ' '
                    str_msg += string
                    if string:
                        infos.append({'type': 'str', 'value': string})
                    infos.append({'type': 'at', 'value': name})
                else:
                    infos.append({'type': 'str', 'value': segs[i]})
                    str_msg_all += segs[i]
                    str_msg += segs[i]
            str_msg_all += segs[-1]
            str_msg += segs[-1]
            infos.append({'type': 'str', 'value': segs[-1]})
        else:
            infos.append({'type': 'str', 'value': segs[-1]})
            str_msg_all = msg
            str_msg = msg
        return str_msg_all.replace(u'\u2005', ''), str_msg.replace(u'\u2005', ''), infos

    def extract_msg_content(self, msg_type_id, msg):
        """
        content_type_id:
            0 -> Text
            1 -> Location
            3 -> Image
            4 -> Voice
            5 -> Recommend
            6 -> Animation
            7 -> Share
            8 -> Video
            9 -> VideoCall
            10 -> Redraw
            11 -> Empty
            99 -> Unknown
        :param msg_type_id: 消息类型id
        :param msg: 消息结构体
        :return: 解析的消息
        """
        mtype = msg['MsgType']
        content = html.unescape(msg['Content'])
        msg_id = msg['MsgId']

        msg_content = {}
        if msg_type_id == 0:
            return {'type': 11, 'data': ''}
        elif msg_type_id == 2:  # File Helper
            return {'type': 0, 'data': content.replace('<br/>', '\n')}
        elif msg_type_id == 3:  # 群聊
            sp = content.find('<br/>')
            if sp > -1:
                uid = content[:sp]
                content = content[sp:]
                content = content.replace('<br/>', '')
                uid = uid[:-1]
                name = self.get_contact_prefer_name(self.get_contact_name(uid))
                if not name:
                    name = self.get_group_member_prefer_name(self.get_group_member_name(msg['FromUserName'], uid))
                if not name:
                    name = 'unknown'
                msg_content['user'] = {'id': uid, 'name': name}
            else:
                # System message or Event (e.g. Tickle, Privacy Warning)
                # Content has no uid prefix
                msg_content['user'] = {'id': 'system', 'name': 'System'}
                
                # Parse System Events
                # 1. Invite/Join: "Inviter"邀请"Invitee"加入了
                match_invite = re.search(r'"(.*?)"邀请"(.*?)"加入了', content)
                match_qrcode = re.search(r'"(.*?)"通过扫描"(.*?)"分享的二维码加入群聊', content)
                
                if match_invite:
                    msg_content['system_event'] = {
                        'type': 'group_increase',
                        'sub_type': 'invite',
                        'operator_name': match_invite.group(1),
                        'target_name': match_invite.group(2)
                    }
                elif match_qrcode:
                    msg_content['system_event'] = {
                        'type': 'group_increase',
                        'sub_type': 'invite', # Treat QR join as invite by the sharer
                        'operator_name': match_qrcode.group(2),
                        'target_name': match_qrcode.group(1)
                    }
                
                # 2. Group Rename: "Operator"修改群名为“NewName
                elif '修改群名为' in content:
                    match_rename = re.search(r'"(.*?)"修改群名为“(.*)', content)
                    if match_rename:
                        new_name = match_rename.group(2)
                        if new_name.endswith('”'): new_name = new_name[:-1]
                        msg_content['system_event'] = {
                            'type': 'group_update',
                            'sub_type': 'name',
                            'operator_name': match_rename.group(1),
                            'new_name': new_name
                        }

                # 3. Tickle: "Operator" 拍了拍 "Target"
                elif '拍了拍' in content:
                    match_tickle = re.search(r'"(.*?)" 拍了拍 "(.*?)"', content)
                    if match_tickle:
                        msg_content['system_event'] = {
                            'type': 'poke',
                            'operator_name': match_tickle.group(1),
                            'target_name': match_tickle.group(2)
                        }
                
                # 4. QR Code Join: "Inviter"分享的二维码加入群聊 (Implicitly usually "Invitee" joined via...)
                # Pattern: "Invitee"通过扫描"Inviter"分享的二维码加入群聊
                elif '二维码加入群聊' in content:
                    match_qr = re.search(r'"(.*?)"通过扫描"(.*?)"分享的二维码加入群聊', content)
                    if match_qr:
                        msg_content['system_event'] = {
                            'type': 'group_increase',
                            'sub_type': 'qrcode', # or scan
                            'target_name': match_qr.group(1),
                            'operator_name': match_qr.group(2)
                        }
                         
                # 5. Remove Top/Pin: "Operator"移除了一条置顶消息
                # Or: "Operator"置顶了一条消息 (Assumption)
                elif '置顶' in content:
                    match_pin = re.search(r'"(.*?)"(.*?)了(.*?)置顶', content)
                    if match_pin:
                        action = 'unset' if '移除' in match_pin.group(2) else 'set'
                        msg_content['system_event'] = {
                            'type': 'group_update',
                            'sub_type': 'pin',
                            'action': action,
                            'operator_name': match_pin.group(1)
                        }
                        
                # 6. Owner Change: 你已成为新群主
                elif '成为新群主' in content:
                    # This usually means SELF became owner, or someone else?
                    # Pattern: "User"已成为新群主
                    match_owner = re.search(r'"(.*?)"已成为新群主', content)
                    if match_owner:
                       msg_content['system_event'] = {
                            'type': 'group_update',
                            'sub_type': 'owner',
                            'operator_name': match_owner.group(1)
                       }
                    elif '你已成为新群主' in content:
                       msg_content['system_event'] = {
                            'type': 'group_update',
                            'sub_type': 'owner',
                            'operator_name': 'self' # Handle self resolution later
                       }
                
                # 7. Recall: "Operator" 撤回了一条消息
                elif '撤回了一条消息' in content:
                    match_recall = re.search(r'"(.*?)" 撤回了一条消息', content)
                    if match_recall:
                        msg_content['system_event'] = {
                            'type': 'group_recall',
                            'operator_name': match_recall.group(1)
                        }
                        
                # 8. Red Packet: 收到红包
                elif '收到红包' in content:
                    msg_content['system_event'] = {
                        'type': 'red_packet',
                        'content': content
                    }
        else:  # Self, Contact, Special, Public, Unknown
            pass

        msg_prefix = (msg_content['user']['name'] + ':') if 'user' in msg_content else ''

        if mtype == 1:
            if content.find('http://weixin.qq.com/cgi-bin/redirectforward?args=') != -1:
                r = self.session.get(content)
                r.encoding = 'gbk'
                data = r.text
                pos = self.search_content('title', data, 'xml')
                msg_content['type'] = 1
                msg_content['data'] = pos
                msg_content['detail'] = data
            else:
                msg_content['type'] = 0
                if msg_type_id == 3 or (msg_type_id == 1 and msg['ToUserName'][:2] == '@@'):  # Group text message
                    msg_infos = self.proc_at_info(content)
                    str_msg_all = msg_infos[0]
                    str_msg = msg_infos[1]
                    detail = msg_infos[2]
                    msg_content['data'] = str_msg_all
                    msg_content['detail'] = detail
                    msg_content['desc'] = str_msg
                else:
                    msg_content['data'] = content
        elif mtype == 3:
            msg_content['type'] = 3
            msg_content['data'] = self.get_msg_img_url(msg_id)
            msg_content['img'] = None #self.session.get(msg_content['data']).content.encode('hex')
                                      #edited by derlin 2016-12-16
        elif mtype == 34:
            msg_content['type'] = 4
            msg_content['data'] = self.get_voice_url(msg_id)
            msg_content['voice'] = None #self.session.get(msg_content['data']).content.encode('hex')
                                        #edited by derlin 2016-12-16
        elif mtype == 37:
            msg_content['type'] = 37
            msg_content['data'] = msg['RecommendInfo']
        elif mtype == 42:
            msg_content['type'] = 5
            info = msg['RecommendInfo']
            msg_content['data'] = {'nickname': info['NickName'],
                                   'alias': info['Alias'],
                                   'province': info['Province'],
                                   'city': info['City'],
                                   'gender': ['unknown', 'male', 'female'][info['Sex']]}
        elif mtype == 47:
            msg_content['type'] = 6
            msg_content['data'] = self.search_content('cdnurl', content)
        elif mtype == 49:
            msg_content['type'] = 7
            if msg['AppMsgType'] == 3:
                app_msg_type = 'music'
            elif msg['AppMsgType'] == 5:
                app_msg_type = 'link'
            elif msg['AppMsgType'] == 7:
                app_msg_type = 'weibo'
            elif msg['AppMsgType'] == 19:
                app_msg_type = 'merged_forward'
            elif msg['AppMsgType'] == 57:
                app_msg_type = 'quote'
            else:
                app_msg_type = 'unknown_%s' % msg['AppMsgType']
            
            msg_content['data'] = {'type': app_msg_type,
                                   'title': msg['FileName'],
                                   'desc': self.search_content('des', content, 'xml'),
                                   'url': msg['Url'],
                                   'from': self.search_content('appname', content, 'xml'),
                                   'content': msg.get('Content'),  # 有的公众号会发一次性3 4条链接一个大图,如果只url那只能获取第一条,content里面有所有的链接
                                   'app_msg_type': msg['AppMsgType']
                                   }

        elif mtype == 62:
            msg_content['type'] = 8
            msg_content['data'] = content
        elif mtype == 53:
            msg_content['type'] = 9
            msg_content['data'] = content
        elif mtype == 10002:
            msg_content['type'] = 10
            msg_content['data'] = content
        elif mtype == 10000:  # unknown, maybe red packet, or group invite
            msg_content['type'] = 12
            msg_content['data'] = msg['Content']
        else:
            msg_content['type'] = 99
            msg_content['data'] = content
        return msg_content

    def handle_msg(self, r):
        """
        处理原始微信消息的内部函数
        msg_type_id:
            0 -> Init
            1 -> Self
            2 -> FileHelper
            3 -> Group
            4 -> Contact
            5 -> Public
            6 -> Special
            99 -> Unknown
        :param r: 原始微信消息
        """
        for msg in r['AddMsgList']:
            user = {'id': msg['FromUserName'], 'name': 'unknown'}
            if msg['MsgType'] == 51:  # init message
                msg_type_id = 0
                user['name'] = 'system'
            elif msg['MsgType'] == 37:  # friend request
                msg_type_id = 37
                pass
                # content = msg['Content']
                # username = content[content.index('fromusername='): content.index('encryptusername')]
                # username = username[username.index('"') + 1: username.rindex('"')]
                # print u'[Friend Request]'
                # print u'       Nickname：' + msg['RecommendInfo']['NickName']
                # print u'       附加消息：'+msg['RecommendInfo']['Content']
                # # print u'Ticket：'+msg['RecommendInfo']['Ticket'] # Ticket添加好友时要用
                # print u'       微信号：'+username #未设置微信号的 腾讯会自动生成一段微信ID 但是无法通过搜索 搜索到此人
            elif msg['FromUserName'] == self.my_account['UserName']:  # Self
                msg_type_id = 1
                user['name'] = 'self'
            elif msg['ToUserName'] == 'filehelper':  # File Helper
                msg_type_id = 2
                user['name'] = 'file_helper'
            elif msg['FromUserName'][:2] == '@@':  # Group
                msg_type_id = 3
                user['name'] = self.get_contact_prefer_name(self.get_contact_name(user['id']))
            elif self.is_contact(msg['FromUserName']):  # Contact
                msg_type_id = 4
                user['name'] = self.get_contact_prefer_name(self.get_contact_name(user['id']))
            elif self.is_public(msg['FromUserName']):  # Public
                msg_type_id = 5
                user['name'] = self.get_contact_prefer_name(self.get_contact_name(user['id']))
            elif self.is_special(msg['FromUserName']):  # Special
                msg_type_id = 6
                user['name'] = self.get_contact_prefer_name(self.get_contact_name(user['id']))
            else:
                msg_type_id = 99
                user['name'] = 'unknown'
            if not user['name']:
                user['name'] = 'unknown'
            user['name'] = html.unescape(user['name'])

            content = self.extract_msg_content(msg_type_id, msg)
            
            # If extract_msg_content identified a system user, override the user info ONLY if not group
            # For groups, we want 'user' to remain the Group info, and 'content.user' to be the Sender info
            if msg_type_id != 3 and isinstance(content, dict) and 'user' in content:
                user = content['user']

            message = {'msg_type_id': msg_type_id,
                       'msg_id': msg['MsgId'],
                       'content': content,
                       'to_user_id': msg['ToUserName'],
                       'user': user,
                       'create_time': msg.get('CreateTime')}
            self.handle_msg_all(message)

    def schedule(self):
        """
        做任务型事情的函数，如果需要，可以在子类中覆盖此函数
        此函数在处理消息的间隙被调用，请不要长时间阻塞此函数
        """
        pass

    def proc_msg(self):
        self.test_sync_check()
        while True:
            check_time = time.time()
            try:
                [retcode, selector] = self.sync_check()
                # print([DEBUG] sync_check:', retcode, selector
                if retcode == '1100':  # 从微信客户端上登出
                    break
                elif retcode == '1101':  # 从其它设备上登了网页微信
                    break
                elif retcode == '0':
                    if selector == '2':  # 有新消息
                        r = self.sync()
                        if r is not None:
                            self.handle_msg(r)
                    elif selector == '3':  # 未知
                        r = self.sync()
                        if r is not None:
                            self.handle_msg(r)
                    elif selector == '4':  # 通讯录更新
                        r = self.sync()
                        if r is not None:
                            self.get_contact()
                    elif selector == '6':  # 可能是红包
                        r = self.sync()
                        if r is not None:
                            self.handle_msg(r)
                    elif selector == '7':  # 在手机上操作了微信
                        r = self.sync()
                        if r is not None:
                            self.handle_msg(r)
                    elif selector == '0':  # 无事件
                        pass
                    else:
                        r = self.sync()
                        if r is not None:
                            self.handle_msg(r)
                else:
                    time.sleep(10)
                self.schedule()
            except:
                print('[ERROR] Except in proc_msg')
                print(format_exc())
            check_time = time.time() - check_time
            if check_time < 0.8:
                time.sleep(1 - check_time)

    def apply_useradd_requests(self,RecommendInfo):
        url = self.base_uri + '/webwxverifyuser?r='+str(int(time.time()))+'&lang=zh_CN'
        params = {
            "BaseRequest": self.base_request,
            "Opcode": 3,
            "VerifyUserListSize": 1,
            "VerifyUserList": [
                {
                    "Value": RecommendInfo['UserName'],
                    "VerifyUserTicket": RecommendInfo['Ticket']             }
            ],
            "VerifyContent": "",
            "SceneListCount": 1,
            "SceneList": [
                33
            ],
            "skey": self.skey
        }
        headers = {'content-type': 'application/json; charset=UTF-8'}
        data = json.dumps(params, ensure_ascii=False).encode('utf8')
        try:
            r = self.session.post(url, data=data, headers=headers)
        except (ConnectionError, ReadTimeout):
            return False
        dic = r.json()
        return dic['BaseResponse']['Ret'] == 0

    def add_groupuser_to_friend_by_uid(self,uid,VerifyContent):
        """
        主动向群内人员打招呼，提交添加好友请求
        uid-群内人员得uid   VerifyContent-好友招呼内容
        慎用此接口！封号后果自负！慎用此接口！封号后果自负！慎用此接口！封号后果自负！
        """
        if self.is_contact(uid):
            return True
        url = self.base_uri + '/webwxverifyuser?r='+str(int(time.time()))+'&lang=zh_CN'
        params ={
            "BaseRequest": self.base_request,
            "Opcode": 2,
            "VerifyUserListSize": 1,
            "VerifyUserList": [
                {
                    "Value": uid,
                    "VerifyUserTicket": ""
                }
            ],
            "VerifyContent": VerifyContent,
            "SceneListCount": 1,
            "SceneList": [
                33
            ],
            "skey": self.skey
        }
        headers = {'content-type': 'application/json; charset=UTF-8'}
        data = json.dumps(params, ensure_ascii=False).encode('utf8')
        try:
            r = self.session.post(url, data=data, headers=headers)
        except (ConnectionError, ReadTimeout):
            return False
        dic = r.json()
        return dic['BaseResponse']['Ret'] == 0

    '''
    def add_friend_to_group(self,uid,group_name):
        """
        将好友加入到群聊中
        """
        gid = ''
        #通过群名获取群id,群没保存到通讯录中的话无法添加哦
        for group in self.group_list:
            if group['NickName'] == group_name:
                gid = group['UserName']
        if gid == '':
            return False
        #通过群id判断uid是否在群中
        for user in self.group_members[gid]:
            if user['UserName'] == uid:
                #已经在群里面了,不用加了
                return True
        url = self.base_uri + '/webwxupdatechatroom?fun=addmember&pass_ticket=%s' % self.pass_ticket
        params ={
            "AddMemberList": uid,
            "ChatRoomName": gid,
            "BaseRequest": self.base_request
        }
        headers = {'content-type': 'application/json; charset=UTF-8'}
        data = json.dumps(params, ensure_ascii=False).encode('utf8')
        try:
            r = self.session.post(url, data=data, headers=headers)
        except (ConnectionError, ReadTimeout):
            return False
        dic = r.json()
        return dic['BaseResponse']['Ret'] == 0

    def invite_friend_to_group(self,uid,group_name):
       """
       将好友加入到群中。对人数多的群，需要调用此方法。
       拉人时，可以先尝试使用add_friend_to_group方法，当调用失败(Ret=1)时，再尝试调用此方法。
       """
       gid = ''
       # 通过群名获取群id,群没保存到通讯录中的话无法添加哦
       for group in self.group_list:
           if group['NickName'] == group_name:
               gid = group['UserName']
       if gid == '':
           return False
       # 通过群id判断uid是否在群中
       for user in self.group_members[gid]:
           if user['UserName'] == uid:
               # 已经在群里面了,不用加了
               return True
       url = self.base_uri '/webwxupdatechatroom?fun=invitemember&pass_ticket=%s' % self.pass_ticket
       params = {
           "InviteMemberList": uid,
           "ChatRoomName": gid,
           "BaseRequest": self.base_request
       }
       headers = {'content-type': 'application/json; charset=UTF-8'}
       data = json.dumps(params, ensure_ascii=False).encode('utf8')
       try:
           r = self.session.post(url, data=data, headers=headers)
       except (ConnectionError, ReadTimeout):
           return False
       dic = r.json()
       return dic['BaseResponse']['Ret'] == 0
    '''


    #edited by derlin 2016/12/29
    def add_friend_to_group(self, gid, uid):
        """
        将好友加入到群聊中
        """
        if gid == '':
            return False
        #通过群id判断uid是否在群中
        for user in self.group_members[gid]:
            if user['UserName'] == uid:
                #已经在群里面了,不用加了
                return True
        url = self.base_uri + '/webwxupdatechatroom?fun=addmember&pass_ticket=%s' % self.pass_ticket
        params ={
            "AddMemberList": uid,
            "ChatRoomName": gid,
            "BaseRequest": self.base_request
        }
        headers = {'content-type': 'application/json; charset=UTF-8'}
        data = json.dumps(params, ensure_ascii=False).encode('utf8')
        try:
            r = self.session.post(url, data=data, headers=headers)
        except (ConnectionError, ReadTimeout):
            return False
        dic = r.json()
        return dic['BaseResponse']['Ret'] == 0

    def delete_user_from_group(self, gid, uid):
        """
        将群用户从群中剔除，只有群管理员有权限
        """
        if uid == "":
            return False
        url = self.base_uri + '/webwxupdatechatroom?fun=delmember&pass_ticket=%s' % self.pass_ticket
        params ={
            "DelMemberList": uid,
            "ChatRoomName": gid,
            "BaseRequest": self.base_request
        }
        headers = {'content-type': 'application/json; charset=UTF-8'}
        data = json.dumps(params, ensure_ascii=False).encode('utf8')
        try:
            r = self.session.post(url, data=data, headers=headers)
        except (ConnectionError, ReadTimeout):
            return False
        dic = r.json()
        try:
            self._last_api_ret = dic
        except Exception:
            pass
        if self.DEBUG:
            print(f"[delete_user] gid={gid} uid={uid} ret={dic}")
        ok = dic.get('BaseResponse', {}).get('Ret') == 0
        if ok:
            return True
        try:
            alt_url = self.base_uri + '/webwxupdatechatroom?fun=delmember&f=json&lang=zh_CN'
            alt_params = dict(params)
            alt_params["Scene"] = 1
            alt_headers = {'content-type': 'application/json; charset=UTF-8', 'accept': 'application/json'}
            alt_data = json.dumps(alt_params, ensure_ascii=False).encode('utf8')
            r2 = self.session.post(alt_url, data=alt_data, headers=alt_headers)
            dic2 = r2.json()
            self._last_api_ret = dic2
            if self.DEBUG:
                print(f"[delete_user][fallback] gid={gid} uid={uid} ret={dic2}")
            return dic2.get('BaseResponse', {}).get('Ret') == 0
        except Exception:
            return False
        #end of edited by derlin 2016/12/29

    def is_group_owner(self, gid):
        try:
            me = self.my_account.get('UserName') or ''
        except Exception:
            me = ''
        owner = ''
        # 从缓存的 group_list 查找
        for g in self.group_list:
            if g.get('UserName') == gid:
                owner = g.get('ChatRoomOwner') or ''
                break
        # 如果未命中或为空，尝试刷新单群信息
        if not owner:
            info = self._refresh_group_info(gid)
            owner = (info.get('ChatRoomOwner') if isinstance(info, dict) else '') or owner
        # 如果仍未知，则不阻断（返回 True 以允许尝试实际接口）
        if not owner or not me:
            return True
        return owner == me

    def _refresh_group_info(self, gid):
        try:
            url = self.base_uri + '/webwxbatchgetcontact?type=ex&r=%s&pass_ticket=%s' % (int(time.time()), self.pass_ticket)
            params = {
                'BaseRequest': self.base_request,
                "Count": 1,
                "List": [{"UserName": gid, "EncryChatRoomId": ""}]
            }
            r = self.session.post(url, data=json.dumps(params))
            r.encoding = 'utf-8'
            dic = json.loads(r.text)
            items = dic.get('ContactList') or []
            if items:
                info = items[0]
                # 更新 group_list 中的对应项
                for i, g in enumerate(self.group_list):
                    if g.get('UserName') == gid:
                        try:
                            self.group_list[i]['ChatRoomOwner'] = info.get('ChatRoomOwner', '')
                            self.group_list[i]['NickName'] = info.get('NickName', self.group_list[i].get('NickName', ''))
                        except Exception:
                            pass
                        break
                return info
        except Exception:
            pass
        return {}

    def quit_group(self, gid):
        """
        退出群聊
        """
        return self.delete_user_from_group(gid, self.my_account['UserName'])

    def set_group_name(self, gid, gname):
        """
        设置群聊名称
        """
        url = self.base_uri + '/webwxupdatechatroom?fun=modtopic&pass_ticket=%s' % self.pass_ticket
        params ={
            "NewTopic": gname,
            "ChatRoomName": gid,
            "BaseRequest": self.base_request
        }
        headers = {'content-type': 'application/json; charset=UTF-8'}
        data = json.dumps(params, ensure_ascii=False).encode('utf8')
        try:
            r = self.session.post(url, data=data, headers=headers)
            r.encoding = 'utf-8'
        except (ConnectionError, ReadTimeout):
            return False
        
        try:
            dic = r.json()
            # print(f"[wxbot] set_group_name gid={gid} name={gname} response={dic}")
            ret = dic['BaseResponse']['Ret']
            if ret == 0:
                # 注意：这里不再手动更新本地缓存，而是等待服务器推送 ModContact 更新
                # 这样可以避免"本地修改成功但服务器未生效"的假象
                is_owner = self.is_group_owner(gid)
                if self.DEBUG:
                    print(f"[wxbot] set_group_name success (server ret=0). Am I owner? {is_owner}")
                if not is_owner:
                    if self.DEBUG:
                        print(f"[wxbot] WARNING: Not group owner. Name change might fail silently or only apply locally.")
                return True
            return False
        except Exception as e:
            print(f"[wxbot] set_group_name exception: {e}")
            return False

    def set_group_remark(self, gid, remark):
        """
        设置群聊备注
        """
        url = self.base_uri + '/webwxoplog?lang=zh_CN&pass_ticket=%s' % self.pass_ticket
        params = {
            'BaseRequest': self.base_request,
            'CmdId': 2,
            'RemarkName': remark,
            'UserName': gid
        }
        headers = {'content-type': 'application/json; charset=UTF-8'}
        data = json.dumps(params, ensure_ascii=False).encode('utf8')
        try:
            r = self.session.post(url, data=data, headers=headers)
            r.encoding = 'utf-8'
            dic = r.json()
            if self.DEBUG:
                print(f"[wxbot] set_group_remark gid={gid} remark={remark} response={dic}")
            return dic['BaseResponse']['Ret'] == 0
        except Exception as e:
            print(f"[wxbot] set_group_remark exception: {e}")
            return False

    def send_msg_by_uid(self, word, dst='filehelper'):
        url = self.base_uri + '/webwxsendmsg?pass_ticket=%s' % self.pass_ticket
        msg_id = str(int(time.time() * 1000)) + str(random.random())[:5].replace('.', '')
        params = {
            'BaseRequest': self.base_request,
            'Msg': {
                "Type": 1,
                "Content": word,
                "FromUserName": self.my_account['UserName'],
                "ToUserName": dst,
                "LocalID": msg_id,
                "ClientMsgId": msg_id
            }
        }
        headers = {'content-type': 'application/json; charset=UTF-8'}
        data = json.dumps(params, ensure_ascii=False).encode('utf8')
        try:
            r = self.session.post(url, data=data, headers=headers)
        except (ConnectionError, ReadTimeout):
            print('[send] post error Connection/Timeout')
            return False
        try:
            dic = r.json()
        except Exception:
            print(f"[send] parse json error status={r.status_code}")
            return False
        ret = (dic.get('BaseResponse') or {}).get('Ret')
        if self.DEBUG:
            print(f"[send] webwxsendmsg dst={dst} ret={ret} status={r.status_code}")
        try:
            self._last_send_info = {'ret': ret, 'status': r.status_code}
        except Exception:
            pass
            
        if ret == 0:
            # 补充 LocalID 以便撤回使用（如果服务器未返回则使用发送时生成的）
            if 'LocalID' not in dic or not dic['LocalID']:
                dic['LocalID'] = msg_id
            return dic
        return False

    def revoke_msg(self, client_msg_id, svr_msg_id, to_user_name):
        """
        撤回消息
        :param client_msg_id: 本地消息ID (LocalID)
        :param svr_msg_id: 服务器消息ID (MsgID)
        :param to_user_name: 接收者 UserName
        """
        url = self.base_uri + '/webwxrevokemsg?pass_ticket=%s' % self.pass_ticket
        data = {
            'BaseRequest': self.base_request,
            'ClientMsgId': client_msg_id,
            'SvrMsgId': svr_msg_id,
            'ToUserName': to_user_name
        }
        headers = {'content-type': 'application/json; charset=UTF-8'}
        data = json.dumps(data, ensure_ascii=False).encode('utf8')
        try:
            r = self.session.post(url, data=data, headers=headers)
            res = json.loads(r.text)
            if self.DEBUG:
                print(f"[revoke] ret={res.get('BaseResponse', {}).get('Ret')} res={res}")
            return res.get('BaseResponse', {}).get('Ret') == 0
        except Exception as e:
            print(f"[revoke] error: {e}")
            return False

    def send_poke(self, to_user_name, member_uid=None):
        """
        发送拍一拍（Web端不支持真实拍一拍，使用文本模拟）
        :param to_user_name: 接收者 UserName (群或个人)
        :param member_uid: 如果在群里，被拍者的 UserName
        """
        target_name = "你"
        if member_uid:
            # 尝试获取群成员昵称
            # self.group_members 是 {gid: [members]}
            if hasattr(self, 'group_members') and to_user_name in self.group_members:
                for m in self.group_members[to_user_name]:
                    if m['UserName'] == member_uid:
                        target_name = m.get('NickName') or m.get('DisplayName') or "群友"
                        break
        else:
             # 私聊，尝试获取对方昵称
             name_dict = self.get_contact_name(to_user_name)
             target_name = self.get_contact_prefer_name(name_dict) or "你"

        content = f"*[拍一拍] {target_name}*"
        return self.send_msg_by_uid(content, to_user_name)

    def upload_media(self, fpath, is_img=False):
        if not os.path.exists(fpath):
            print('[ERROR] File not exists.')
            return None
            
        url_1 = 'https://file.'+self.base_host+'/cgi-bin/mmwebwx-bin/webwxuploadmedia?f=json'
        url_2 = 'https://file2.'+self.base_host+'/cgi-bin/mmwebwx-bin/webwxuploadmedia?f=json'
        
        # safely get webwx_data_ticket
        ticket = ''
        try:
            ticket = self.session.cookies['webwx_data_ticket']
        except:
            for c in self.session.cookies:
                if c.name == 'webwx_data_ticket':
                    ticket = c.value
                    break
                    
        flen = os.path.getsize(fpath)
        ftype = mimetypes.guess_type(fpath)[0] or 'application/octet-stream'
        filename = os.path.basename(fpath)
        
        # 分片逻辑：大文件（例如 > 10MB）进行分片上传
        # 微信 Web 端对大文件通常要求分片，每片 512KB (524288 bytes) 是个常见值
        CHUNK_SIZE = 524288
        chunks = (flen + CHUNK_SIZE - 1) // CHUNK_SIZE
        
        client_media_id = int(time.time() * 1000)
        
        if chunks > 1 and self.DEBUG:
            print(f"[upload] Large file detected ({flen} bytes), uploading in {chunks} chunks...")
        
        with open(fpath, 'rb') as f:
            for chunk_idx in range(chunks):
                chunk_data = f.read(CHUNK_SIZE)
                # 重新计算实际读取的长度（最后一片可能小于 CHUNK_SIZE）
                current_chunk_len = len(chunk_data)
                
                files = {
                    'id': (None, 'WU_FILE_%s' % str(self.file_index)),
                    'name': (None, filename),
                    'type': (None, ftype),
                    'lastModifiedDate': (None, time.strftime('%m/%d/%Y, %H:%M:%S GMT+0800 (CST)')),
                    'size': (None, str(flen)),
                    'mediatype': (None, 'pic' if is_img else 'doc'),
                    'uploadmediarequest': (None, json.dumps({
                        'BaseRequest': self.base_request,
                        'ClientMediaId': client_media_id,
                        'TotalLen': flen,
                        'StartPos': chunk_idx * CHUNK_SIZE,
                        'DataLen': current_chunk_len,
                        'MediaType': 4,
                        'FileMd5': hashlib.md5(chunk_data).hexdigest() if chunks > 1 else None # 尝试补充 md5，虽然不一定必须
                    })),
                    'webwx_data_ticket': (None, ticket),
                    'pass_ticket': (None, self.pass_ticket),
                    'filename': (filename, chunk_data, ftype.split('/')[1]),
                }
                
                # 如果是分片，除了最后一片，其他可能不返回 MediaId，或者需要持续上传
                # 微信Web协议中，最后一片上传成功才会返回 MediaId
                
                try:
                    r = self.session.post(url_1, files=files)
                    res_json = json.loads(r.text)
                    if res_json['BaseResponse']['Ret'] != 0:
                        # 尝试备用服务器
                        r = self.session.post(url_2, files=files)
                        res_json = json.loads(r.text)
                    
                    if res_json['BaseResponse']['Ret'] != 0:
                        print(f'[ERROR] Upload media failure at chunk {chunk_idx+1}/{chunks}. Ret={res_json["BaseResponse"]["Ret"]}')
                        return None
                        
                    # 只有最后一片或单片上传才会有 MediaId
                    if chunk_idx == chunks - 1:
                        mid = res_json.get('MediaId')
                        if mid:
                            self.file_index += 1
                            return mid
                        else:
                            print('[ERROR] No MediaId in response after last chunk.')
                            return None
                            
                except Exception as e:
                    print(f"[upload] error at chunk {chunk_idx}: {e}")
                    return None
                    
        return None

    def send_file_msg_by_uid(self, fpath, uid):
        try:
            mid = self.upload_media(fpath)
            if mid is None or not mid:
                print("[send_file] upload_media failed")
                return False
            
            import html
            url = self.base_uri + '/webwxsendappmsg?fun=async&f=json&pass_ticket=' + self.pass_ticket
            msg_id = str(int(time.time() * 1000)) + str(random.random())[:5].replace('.', '')
            
            filename = os.path.basename(fpath)
            filesize = str(os.path.getsize(fpath))
            _, fileext = os.path.splitext(fpath)
            fileext = fileext.replace('.', '')
            
            # 构造 Content XML
            # 注意: json.dumps 需要字符串，不能是 bytes
            # 使用 html.escape 防止文件名包含特殊字符破坏 XML
            content_xml = (
                "<appmsg appid='wxeb7ec651dd0aefa9' sdkver=''>"
                f"<title>{html.escape(filename)}</title>"
                "<des></des>"
                "<action></action>"
                "<type>6</type>"
                "<content></content>"
                "<url></url>"
                "<lowurl></lowurl>"
                "<appattach>"
                f"<totallen>{filesize}</totallen>"
                f"<attachid>{mid}</attachid>"
                f"<fileext>{fileext}</fileext>"
                "</appattach>"
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
                        'ClientMsgId': msg_id, }, }
            
            headers = {'content-type': 'application/json; charset=UTF-8'}
            data = json.dumps(data, ensure_ascii=False).encode('utf8')
            r = self.session.post(url, data=data, headers=headers)
            res = json.loads(r.text)
            if res['BaseResponse']['Ret'] == 0:
                if 'LocalID' not in res or not res['LocalID']:
                    res['LocalID'] = msg_id
                return res
            else:
                print(f"[send_file] server return error: {res}")
                return False
        except Exception as e:
            print(f"[send_file] exception: {e}")
            return False

    def send_img_msg_by_uid(self, fpath, uid):
        mid = self.upload_media(fpath, is_img=True)
        if mid is None:
            return False
        url = self.base_uri + '/webwxsendmsgimg?fun=async&f=json'
        msg_id = str(int(time.time() * 1000)) + str(random.random())[:5].replace('.', '')
        data = {
                'BaseRequest': self.base_request,
                'Msg': {
                    'Type': 3,
                    'MediaId': mid,
                    'FromUserName': self.my_account['UserName'],
                    'ToUserName': uid,
                    'LocalID': msg_id,
                    'ClientMsgId': msg_id, }, }
        if fpath[-4:] == '.gif':
            url = self.base_uri + '/webwxsendemoticon?fun=sys'
            data['Msg']['Type'] = 47
            data['Msg']['EmojiFlag'] = 2
        try:
            headers = {'content-type': 'application/json; charset=UTF-8'}
            data = json.dumps(data, ensure_ascii=False).encode('utf8')
            r = self.session.post(url, data=data, headers=headers)
            res = json.loads(r.text)
            if res['BaseResponse']['Ret'] == 0:
                if 'LocalID' not in res or not res['LocalID']:
                    res['LocalID'] = msg_id
                return res
            else:
                print(f"[send_img] server return error: {res}")
                return False
        except Exception as e:
            return False

    def get_user_id(self, name):
        if name == '':
            return None
        for contact in self.contact_list:
            if 'RemarkName' in contact and contact['RemarkName'] == name:
                return contact['UserName']
            elif 'NickName' in contact and contact['NickName'] == name:
                return contact['UserName']
            elif 'DisplayName' in contact and contact['DisplayName'] == name:
                return contact['UserName']
        for group in self.group_list:
            if 'RemarkName' in group and group['RemarkName'] == name:
                return group['UserName']
            if 'NickName' in group and group['NickName'] == name:
                return group['UserName']
            if 'DisplayName' in group and group['DisplayName'] == name:
                return group['UserName']

        return ''

    def send_msg(self, name, word, isfile=False):
        uid = self.get_user_id(name)
        if uid is not None:
            if isfile:
                with open(word, 'r') as f:
                    result = True
                    for line in f.readlines():
                        line = line.replace('\n', '')
                        print('-> ' + name + ': ' + line)
                        if self.send_msg_by_uid(line, uid):
                            pass
                        else:
                            result = False
                        time.sleep(1)
                    return result
            else:
                if self.send_msg_by_uid(word, uid):
                    return True
                else:
                    return False
        else:
            if self.DEBUG:
                print('[ERROR] This user does not exist .')
            return True

    @staticmethod
    def search_content(key, content, fmat='attr'):
        if fmat == 'attr':
            pm = re.search(key + r'\s?=\s?"([^"<]+)"', content)
            if pm:
                return pm.group(1)
        elif fmat == 'xml':
            pm = re.search('<{0}>([^<]+)</{0}>'.format(key), content)
            if pm:
                return pm.group(1)
        return 'unknown'

    def run(self):
        if self.load_session_cache():
            if self.DEBUG:
                print('[INFO] Try cached session .')
            if self.init():
                if self.DEBUG:
                    print('[INFO] Cached session valid .')
                self.status_notify()
                # 强制刷新联系人，解决群信息 unknown 问题
                # if self.load_contacts_cache():
                #     if self.DEBUG:
                #         print('[INFO] Loaded contacts from cache. Skip fetching.')
                # else:
                if True:
                    if self.DEBUG:
                        print('[INFO] Getting Contacts & groups (Force Refresh)')
                    self.get_contact()
                if self.DEBUG:
                    print('[INFO] Get %d contacts' % len(self.contact_list))
                    print('[INFO] Get %d groups' % len(self.group_list))
                
                print(f"[wxbot] Processing {len(self.group_list)} groups...")
                for wxgroup in self.group_list:
                    try:
                        group_uid = wxgroup['UserName']
                        group_name = wxgroup['NickName']
                        group_name = msg.replace_emoji(group_name)
                        group_id = 0
                        client_qq = 0
                        robot_qq = common.default_robot_qq
                        client_name = ""
                        group_id = wx_group.get_wx_group(robot_qq, group_uid, group_name, client_qq, client_name)
                        #if self.DEBUG:
                        print(f"[GroupInfo] {group_name} => {group_id} (uid={group_uid})")
                    except Exception as e:
                        print(f"[GroupInfo] Error processing group {wxgroup.get('NickName')}: {e}")
                print('[INFO] Start to process messages .')
                self.is_ready = True
                self.proc_msg()
                return
            else:
                if self.DEBUG:
                    print('[INFO] Cached session invalid, fallback to QR .')

        while not self.get_uuid():
            print('[ERROR] Web WeChat get uuid failed. Retry in 3s...')
            time.sleep(3)

        self.gen_qr_code(self.qr_file_path)
        print('[INFO] Please use WeChat to scan the QR code .')

        result = self.wait4login()
        if result != SUCCESS:
            print('[ERROR] Web WeChat login failed. failed code=%s' % (result,))
            return

        if self.login():
            print('[INFO] Web WeChat login succeed .')
        else:
            print('[ERROR] Web WeChat login failed .')
            return

        if self.init():
            if self.DEBUG:
                print('[INFO] Web WeChat init succeed .')
            self.save_session_cache()
        else:
            print('[INFO] Web WeChat init failed')
            return
        self.status_notify()
        if self.DEBUG:
            print('[INFO] Getting Contacts & groups')
        self.get_contact()
        if self.DEBUG:
            print('[INFO] Get %d contacts' % len(self.contact_list))
        #added by derlin 添加或更新群信息到数据库中！ 
        if self.DEBUG:
            print('[INFO] Get %d groups' % len(self.group_list))
        
        print(f"[wxbot] Processing {len(self.group_list)} groups (fresh login)...")
        for wxgroup in self.group_list:
            try:
                group_uid = wxgroup['UserName']
                group_name = wxgroup['NickName']
                group_name = msg.replace_emoji(group_name)
                group_id = 0
                client_qq = 0
                robot_qq = common.default_robot_qq
                client_name = ""
                group_id = wx_group.get_wx_group(robot_qq, group_uid, group_name, client_qq, client_name)
                #if self.DEBUG:
                print(f"[GroupInfo] {group_name} => {group_id} (uid={group_uid})")
            except Exception as e:
                print(f"[GroupInfo] Error processing group {wxgroup.get('NickName')}: {e}")
        #end added by derlin
        print('[INFO] Start to process messages .')
        self.is_ready = True
        self.proc_msg()

    def get_uuid(self):
        url = 'https://login.weixin.qq.com/jslogin'
        params = {
            'appid': 'wx782c26e4c19acffb',
            'fun': 'new',
            'lang': 'zh_CN',
            '_': int(time.time()) * 1000 + random.randint(1, 999),
        }
        r = self.session.get(url, params=params)
        r.encoding = 'utf-8'
        data = r.text
        regx = r'window.QRLogin.code = (\d+); window.QRLogin.uuid = "(\S+?)"'
        pm = re.search(regx, data)
        if pm:
            code = pm.group(1)
            self.uuid = pm.group(2)
            return code == '200'
        return False

    def gen_qr_code(self, qr_file_path):
        string = 'https://login.weixin.qq.com/l/' + self.uuid
        qr = pyqrcode.create(string)
        
        # Always save PNG for WebUI
        try:
            qr.png(qr_file_path, scale=8)
        except Exception as e:
            print(f'[ERROR] Failed to save QR code to {qr_file_path}: {e}')

        # Support TTY mode for Docker logs
        if os.environ.get("WX_QR_MODE") == "tty":
            print(qr.terminal(quiet_zone=1))

        if self.conf['qr'] == 'png':
            try:
                show_image(qr_file_path)
            except Exception:
                pass
        elif self.conf['qr'] == 'tty':
            print(qr.terminal(quiet_zone=1))

    def do_request(self, url):
        r = self.session.get(url)
        r.encoding = 'utf-8'
        data = r.text
        param = re.search(r'window.code=(\d+);', data)
        code = param.group(1)
        return code, data

    def wait4login(self):
        """
        http comet:
        tip=1, 等待用户扫描二维码,
               201: scaned
               408: timeout
        tip=0, 等待用户确认登录,
               200: confirmed
        """
        LOGIN_TEMPLATE = 'https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login?tip=%s&uuid=%s&_=%s'
        tip = 1

        try_later_secs = 1
        MAX_RETRY_TIMES = 10

        code = UNKONWN

        retry_time = MAX_RETRY_TIMES
        while retry_time > 0:
            url = LOGIN_TEMPLATE % (tip, self.uuid, int(time.time()))
            code, data = self.do_request(url)
            if code == SCANED:
                print('[INFO] Please confirm to login .')
                tip = 0
            elif code == SUCCESS:  # 确认登录成功
                param = re.search(r'window.redirect_uri="(\S+?)";', data)
                redirect_uri = param.group(1) + '&fun=new'
                self.redirect_uri = redirect_uri
                self.base_uri = redirect_uri[:redirect_uri.rfind('/')]
                temp_host = self.base_uri[8:]
                self.base_host = temp_host[:temp_host.find("/")]
                return code
            elif code == TIMEOUT:
                print('[ERROR] WeChat login timeout. retry in %s secs later...' % (try_later_secs,))

                tip = 1  # 重置
                retry_time -= 1
                time.sleep(try_later_secs)
            else:
                print ('[ERROR] WeChat login exception return_code=%s. retry in %s secs later...' %
                       (code, try_later_secs))
                tip = 1
                retry_time -= 1
                time.sleep(try_later_secs)
        
        # 增加逻辑：如果重试次数用尽，尝试刷新 UUID 并重新生成二维码
        print('[INFO] QR code expired, refreshing...')
        if self.get_uuid():
            self.gen_qr_code(self.qr_file_path)
            retry_time = MAX_RETRY_TIMES # 重置重试次数
            tip = 1
            # 递归调用或循环继续？
            # 简单起见，这里修改为无限循环重试模式，直到登录成功
            # 更好的方式是修改外层循环，或者在这里重置循环条件
            return self.wait4login()
            
        return code

    def login(self):
        if len(self.redirect_uri) < 4:
            print('[ERROR] Login failed due to network problem, please try again.')
            return False
        r = self.session.get(self.redirect_uri)
        r.encoding = 'utf-8'
        data = r.text
        doc = xml.dom.minidom.parseString(data)
        root = doc.documentElement

        for node in root.childNodes:
            if node.nodeName == 'skey':
                self.skey = node.childNodes[0].data
            elif node.nodeName == 'wxsid':
                self.sid = node.childNodes[0].data
            elif node.nodeName == 'wxuin':
                self.uin = node.childNodes[0].data
            elif node.nodeName == 'pass_ticket':
                self.pass_ticket = node.childNodes[0].data

        if '' in (self.skey, self.sid, self.uin, self.pass_ticket):
            return False

        self.base_request = {
            'Uin': self.uin,
            'Sid': self.sid,
            'Skey': self.skey,
            'DeviceID': self.device_id,
        }
        return True

    def init(self):
        url = self.base_uri + '/webwxinit?r=%i&lang=en_US&pass_ticket=%s' % (int(time.time()), self.pass_ticket)
        params = {
            'BaseRequest': self.base_request
        }
        r = self.session.post(url, data=json.dumps(params))
        r.encoding = 'utf-8'
        dic = json.loads(r.text)

        # Check return code first
        if dic['BaseResponse']['Ret'] != 0:
            self.my_account = {}  # Clear account info on failure
            return False

        self.sync_key = dic['SyncKey']
        self.my_account = dic['User']
        self.sync_key_str = '|'.join([str(keyVal['Key']) + '_' + str(keyVal['Val'])
                                      for keyVal in self.sync_key['List']])
        return dic['BaseResponse']['Ret'] == 0

    def status_notify(self):
        url = self.base_uri + '/webwxstatusnotify?lang=zh_CN&pass_ticket=%s' % self.pass_ticket
        self.base_request['Uin'] = int(self.base_request['Uin'])
        params = {
            'BaseRequest': self.base_request,
            "Code": 3,
            "FromUserName": self.my_account['UserName'],
            "ToUserName": self.my_account['UserName'],
            "ClientMsgId": int(time.time())
        }
        r = self.session.post(url, data=json.dumps(params))
        r.encoding = 'utf-8'
        dic = json.loads(r.text)
        return dic['BaseResponse']['Ret'] == 0

    def test_sync_check(self):
        for host1 in ['webpush.', 'webpush2.']:
            self.sync_host = host1+self.base_host
            try:
                retcode = self.sync_check()[0]
            except:
                retcode = -1
            if retcode == '0':
                return True
        return False

    def sync_check(self):
        params = {
            'r': int(time.time()),
            'sid': self.sid,
            'uin': self.uin,
            'skey': self.skey,
            'deviceid': self.device_id,
            'synckey': self.sync_key_str,
            '_': int(time.time()),
        }
        url = 'https://' + self.sync_host + '/cgi-bin/mmwebwx-bin/synccheck?' + urllib.parse.urlencode(params)
        try:
            r = self.session.get(url, timeout=60)
            r.encoding = 'utf-8'
            data = r.text
            pm = re.search(r'window.synccheck=\{retcode:"(\d+)",selector:"(\d+)"\}', data)
            retcode = pm.group(1)
            selector = pm.group(2)
            return [retcode, selector]
        except:
            return [-1, -1]

    def sync(self):
        url = self.base_uri + '/webwxsync?sid=%s&skey=%s&lang=en_US&pass_ticket=%s' \
                              % (self.sid, self.skey, self.pass_ticket)
        params = {
            'BaseRequest': self.base_request,
            'SyncKey': self.sync_key,
            'rr': ~int(time.time())
        }
        try:
            r = self.session.post(url, data=json.dumps(params), timeout=60)
            r.encoding = 'utf-8'
            dic = json.loads(r.text)
            if dic['BaseResponse']['Ret'] == 0:
                self.sync_key = dic['SyncKey']
                self.sync_key_str = '|'.join([str(keyVal['Key']) + '_' + str(keyVal['Val'])
                                              for keyVal in self.sync_key['List']])
            return dic
        except:
            return None

    def get_icon(self, uid, gid=None):
        """
        获取联系人或者群聊成员头像
        :param uid: 联系人id
        :param gid: 群id，如果为非None获取群中成员头像，如果为None则获取联系人头像
        """
        if gid is None:
            url = self.base_uri + '/webwxgeticon?username=%s&skey=%s' % (uid, self.skey)
        else:
            url = self.base_uri + '/webwxgeticon?username=%s&skey=%s&chatroomid=%s' % (
            uid, self.skey, self.encry_chat_room_id_list[gid])
        r = self.session.get(url)
        data = r.content
        fn = 'icon_' + uid + '.jpg'
        with open(os.path.join(self.temp_pwd,fn), 'wb') as f:
            f.write(data)
        return fn

    def get_head_img(self, uid):
        """
        获取群头像
        :param uid: 群uid
        """
        url = self.base_uri + '/webwxgetheadimg?username=%s&skey=%s' % (uid, self.skey)
        r = self.session.get(url)
        data = r.content
        fn = 'head_' + uid + '.jpg'
        with open(os.path.join(self.temp_pwd,fn), 'wb') as f:
            f.write(data)
        return fn

    def get_msg_img_url(self, msgid):
        return self.base_uri + '/webwxgetmsgimg?MsgID=%s&skey=%s' % (msgid, self.skey)

    def get_msg_img(self, msgid):
        """
        获取图片消息，下载图片到本地
        :param msgid: 消息id
        :return: 保存的本地图片文件路径
        """
        url = self.base_uri + '/webwxgetmsgimg?MsgID=%s&skey=%s' % (msgid, self.skey)
        r = self.session.get(url)
        data = r.content
        fn = 'img_' + msgid + '.jpg'
        with open(os.path.join(self.temp_pwd,fn), 'wb') as f:
            f.write(data)
        return fn

    def get_voice_url(self, msgid):
        return self.base_uri + '/webwxgetvoice?msgid=%s&skey=%s' % (msgid, self.skey)

    def get_voice(self, msgid):
        """
        获取语音消息，下载语音到本地
        :param msgid: 语音消息id
        :return: 保存的本地语音文件路径
        """
        url = self.base_uri + '/webwxgetvoice?msgid=%s&skey=%s' % (msgid, self.skey)
        r = self.session.get(url)
        data = r.content
        fn = 'voice_' + msgid + '.mp3'
        with open(os.path.join(self.temp_pwd,fn), 'wb') as f:
            f.write(data)
        return fn
        
    def set_remarkname(self,uid,remarkname):#设置联系人的备注名
        url = self.base_uri + '/webwxoplog?lang=zh_CN&pass_ticket=%s' \
                              % (self.pass_ticket)
        params = {
            'BaseRequest': self.base_request,
            'CmdId': 2,
            'RemarkName': remarkname,
            'UserName': uid
        }
        try:
            r = self.session.post(url, data=json.dumps(params), timeout=60)
            r.encoding = 'utf-8'
            dic = json.loads(r.text)
            return dic['BaseResponse']['ErrMsg']
        except:
            return None
