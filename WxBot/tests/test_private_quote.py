import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
import time
from unittest.mock import MagicMock

# Mock dependencies BEFORE importing onebot
sys.modules['wxbot'] = MagicMock()
sys.modules['wxclient'] = MagicMock()
sys.modules['wxgroup'] = MagicMock()
sys.modules['common'] = MagicMock()
sys.modules['common'].default_robot_qq = 123456
sys.modules['SQLConn'] = MagicMock()
sys.modules['color'] = MagicMock()
sys.modules['msg'] = MagicMock()
sys.modules['web_ui'] = MagicMock()
sys.modules['wxwork_bot'] = MagicMock()
sys.modules['dingtalk_bot'] = MagicMock()
sys.modules['feishu_bot'] = MagicMock()
sys.modules['telegram_bot'] = MagicMock()

# Mock WXBot class specifically
class MockWXBot:
    def __init__(self):
        self.temp_pwd = "."
        pass
    def get_contact_name(self, uid):
        return None
    def get_contact_prefer_name(self, info):
        return None
    def get_group_member_name(self, guid, uid):
        return {}
    def get_attr_status(self, guid, uid):
        return 0

sys.modules['wxbot'].WXBot = MockWXBot

# Add project root to path
sys.path.append('d:\\projects\\wxBot3')

# Now import onebot
from onebot import onebot

def test_private_quote():
    # Mock Gateway
    gateway = MagicMock()
    gateway.add_bot = MagicMock()
    
    # Instantiate bot
    bot = onebot(gateway, self_id=10000)
    bot.my_account = {"UserName": "robot_uid", "NickName": "Robot"}
    
    # Mock methods
    bot._push_onebot_event = MagicMock()
    bot._mk_msg_id = MagicMock(return_value=999)
    bot._resolve_member_uid_by_name = MagicMock(return_value=None)
    bot.get_contact_name = MagicMock(return_value=None)
    bot._find_group_name_direct = MagicMock(return_value=None)
    
    # Mock cache
    bot._msg_cache = {
        100: {
            "msg_id": 100,
            "content": "Hello",
            "sender_name": "UserA",
            "from_uid": "usera_uid",
            "time": time.time()
        }
    }
    
    # Mock wxclient
    import wxclient
    wxclient.wx_client.get_client_qq_by_uid.return_value = 11111
    
    # Mock wxgroup
    import wxgroup
    wxgroup.wx_group.get_member.return_value = 22222
    wxgroup.wx_group.get_wx_group.return_value = 0 # No group
    
    # Case 1: Private chat, quote exists in cache
    print("\n--- Case 1: Private chat, quote in cache ---")
    msg_private = {
        "msg_type_id": 1, # Text
        "content": {
            "type": 0, # Text
            "data": "「UserA：Hello」\n- - - - - - - - - - - - - - -\nReply",
            "user": {"id": "usera_uid", "name": "UserA"}
        },
        "user": {"id": "usera_uid", "name": "UserA"},
        "to_user_id": "robot_uid"
    }
    
    # Ensure msg module mock works for remove_Emoji
    import msg
    msg.msg.remove_Emoji.side_effect = lambda x: x
    
    bot.handle_msg_all(msg_private)
    
    # Check output
    call_args = bot._push_onebot_event.call_args
    if call_args:
        event = call_args[0][0]
        print(f"Message: {event['message']}")
        if "[CQ:at" in event['message']:
            print("FAILURE: Found CQ:at in private chat")
        elif "@UserA" in event['message']:
            print("SUCCESS: Found text @UserA")
        else:
            print("WARNING: Found neither")
    else:
        print("FAILURE: No event pushed")

    # Case 2: Private chat, quote NOT in cache
    print("\n--- Case 2: Private chat, quote NOT in cache ---")
    msg_private_no_cache = {
        "msg_type_id": 1,
        "content": {
            "type": 0,
            "data": "「UserB：Hi」\n- - - - - - - - - - - - - - -\nReply",
            "user": {"id": "userb_uid", "name": "UserB"}
        },
        "user": {"id": "userb_uid", "name": "UserB"},
        "to_user_id": "robot_uid"
    }
    
    bot.handle_msg_all(msg_private_no_cache)
    
    call_args = bot._push_onebot_event.call_args
    if call_args:
        event = call_args[0][0]
        print(f"Message: {event['message']}")
        if "[CQ:at" in event['message']:
            print("FAILURE: Found CQ:at in private chat")
        elif "@UserB" in event['message']:
            print("SUCCESS: Found text @UserB")
        else:
            print("WARNING: Found neither")

if __name__ == "__main__":
    test_private_quote()
