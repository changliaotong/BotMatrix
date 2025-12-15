#encoding:utf-8

import re

class common():
    is_debug = False
    is_debug_sql = False

    default_robot_qq = 1098299491 #数据库相关 

    retry_msg = "操作失败，请稍后重试"

    regex_emoji_span = r'<span class="emoji emoji([0-9a-fA-F]+)"></span>'
    regex_invitation = r'"(?P<invite_name>[\S\s]*?)"邀请"(?P<invited_name>[\S\s]*?)"加入了群聊'
    regex_group_right = r'"(?P<group_owner>[\S\s]*?)"已成为新群主'
    regex_join_group_by_game_center = r"(?P<client_name>[\S\s]*?)通过游戏中心加入群聊"
    regex_join_group_by_qrcode = r'"(?P<invited_name>[\S\s]*?)"通过扫描"(?P<invite_name>[\S\s]*?)"分享的二维码加入群聊' 

    #emoji
    @staticmethod
    def replace_emoji(message):
        if not message:
            return message
        
        def _replace(match):
            try:
                return chr(int(match.group(1), 16))
            except:
                return match.group(0)
                
        return re.sub(common.regex_emoji_span, _replace, message)

    @staticmethod
    def is_match(regex, s):
        if not s:
            return False
        else:
            p = re.compile(regex, re.I)
            m = p.match(s)
            return bool(m)
            