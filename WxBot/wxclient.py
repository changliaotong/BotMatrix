#encoding:utf-8
from MetaData import *
from SQLConn import *
import hashlib

class User(MetaData):
    database = "sz84_robot"
    table_name = "[User]"
    key_field = "Id"

    def __init__(self, client_qq):
        self.key_value = client_qq
    
    @staticmethod
    def append(robot_qq, group_id, client_qq, client_name, ref_qq):
        client_name = client_name.replace("'", "''")
        sql = str.format("insert into [User](BotUin, GroupId, Id, Name, RefUserId) \
            values({0}, {1}, {2}, '{3}', {4})", robot_qq, group_id, client_qq, client_name, ref_qq)
        return SQLConn.Exec(sql)

class wx_client(MetaData):   
    def __init__(self, client_qq):
        self.database = "sz84_robot"
        self.table_name = "wx_client"
        self.key_field = "client_qq"
        self.key_value = client_qq
    
    @staticmethod
    def isBinded(robot_qq, client_qq):
        res = SQLConn.Query(str.format("select top 1 1 from wx_client where client_qq = '{0}'", client_qq))
        return bool(res)

    @staticmethod
    def _make_anon_key(robot_qq, group_id, client_uid):
        """为 unknown 用户生成稳定匿名标识"""
        text = f"{robot_qq}_{group_id}_{client_uid}"
        md5 = hashlib.md5(text.encode("utf-8")).hexdigest()
        return f"anon_{md5[:8]}"

    @staticmethod
    def get_client_qq(robot_qq, group_id, client_uid, client_name, display_name, remark_name, nick_name, attr_status):
        """
        改进版：
        1️⃣ 先按 UID 查找；
        2️⃣ 未找到则尝试匹配旧 UID 用户；
        3️⃣ 若全为 unknown 则使用 anon_key；
        4️⃣ 自动更新 UID；
        """
        is_new = False

        # Step 1: 尝试按 UID 查
        client_qq = wx_client.get_client_qq_by_uid(robot_qq, client_uid)

        # Step 2: 若找不到，尝试匹配旧记录
        if not client_qq:
            client_qq = wx_client.match_existing_client(robot_qq, group_id, remark_name, display_name, nick_name, attr_status)
            if client_qq:
                # 同一个人，但 UID 改变了
                def safe(s): return s.replace("'", "''") if s else ""
                
                # 1. Update the matched client with new UID
                SQLConn.Exec(f"update wx_client set client_uid='{client_uid}', update_date=getdate() where client_qq='{client_qq}'")
                
                # 2. Clear names for OTHER clients with the same name to avoid future conflicts
                # This prevents "zombie" records from being matched again
                conditions = []
                if nick_name and nick_name.lower() != "unknown":
                    conditions.append(f"nick_name = '{safe(nick_name)}'")
                if display_name and display_name.lower() != "unknown":
                    conditions.append(f"display_name = '{safe(display_name)}'")
                if remark_name and remark_name.lower() != "unknown":
                    conditions.append(f"remark_name = '{safe(remark_name)}'")
                
                if conditions:
                    where_name = " OR ".join(conditions)
                    # Update wx_client: Clear names for other records with same name
                    sql_clean = f"UPDATE wx_client SET nick_name='', display_name='', remark_name='' WHERE robot_qq={robot_qq} AND client_qq != {client_qq} AND ({where_name})"
                    SQLConn.Exec(sql_clean)
            else:
                # Step 3: 若名字都 unknown，则生成 anon_key
                if not client_name or client_name.lower() == "unknown":
                    anon_key = wx_client._make_anon_key(robot_qq, group_id, client_uid)
                    client_name = anon_key
                    display_name = anon_key
                    remark_name = anon_key
                    nick_name = anon_key
                is_new = True
                wx_client.append(robot_qq, group_id, client_uid, 0, client_name, display_name, remark_name, nick_name, attr_status)
                client_qq = wx_client.get_client_qq_by_uid(robot_qq, client_uid)

        # Step 4: 更新常规信息
        if not is_new:
            wx_client.update(robot_qq, group_id, client_uid, client_qq, client_name, display_name, remark_name, nick_name, attr_status)

        # Step 5: 确保 client 表中存在
        if not User.exists(client_qq):
            User.append(robot_qq, group_id, client_qq, client_name, group_id)

        return client_qq

    @staticmethod
    def match_existing_client(robot_qq, group_id, remark_name, display_name, nick_name, attr_status):
        """匹配旧 UID 用户"""
        # Optimized: Directly query wx_client instead of relying on unmaintained GroupMember table
        
        conditions = []
        def safe(s): return s.replace("'", "''") if s else ""
        
        if nick_name and nick_name.lower() != "unknown":
            conditions.append(f"nick_name = '{safe(nick_name)}'")
        if display_name and display_name.lower() != "unknown":
            conditions.append(f"display_name = '{safe(display_name)}'")
        if remark_name and remark_name.lower() != "unknown":
            conditions.append(f"remark_name = '{safe(remark_name)}'")
            
        if not conditions:
            return 0
            
        where_clause = " OR ".join(conditions)
        
        # Filter by robot_qq to scope to current bot instance
        attr_filter = f" AND attr_status = '{attr_status}'" if attr_status else ""
        
        # Prioritize most recently updated record
        sql = f"SELECT TOP 1 client_qq FROM wx_client WHERE robot_qq = {robot_qq} AND ({where_clause}) {attr_filter} ORDER BY update_date DESC"
        
        res = SQLConn.Query(sql)
        if res:
            return int(res)
        return 0
    
    @staticmethod
    def append(robot_qq, group_id, client_uid, client_qq, client_name, display_name, remark_name, nick_name, attr_status):
        """插入用户（带未知名修正）"""
        # Guard: Prevent system messages or junk data from being inserted as users
        # 1. Length check: Standard IDs are usually < 60 chars
        if len(client_uid) > 60:
            return False
            
        # 2. Keyword check: Filter out system messages, but allow valid IDs (starting with @, wxid_, gh_)
        is_standard_id = client_uid.startswith('@') or client_uid.startswith('wxid_') or client_uid.startswith('gh_') or client_uid in ['filehelper', 'weixin']
        
        if not is_standard_id and any(k in client_uid for k in ['邀请', '修改群名为', '拍了拍', '撤回', '红包']):
            return False

        def safe(x): return x.replace("'", "''") if x else ""
        client_name = safe(client_name)
        display_name = safe(display_name)
        remark_name = safe(remark_name)
        nick_name = safe(nick_name)
        # Ensure attr_status is not None
        attr_status = attr_status or ""

        sql = (
            "insert into wx_client(client_uid, client_name, display_name, remark_name, nick_name, robot_qq, attr_status) "
            f"values('{client_uid}','{client_name}','{display_name}','{remark_name}','{nick_name}','{robot_qq}','{attr_status}')"
        )
        SQLConn.Exec(sql)
        sql = f"update wx_client set client_qq = client_oid + 90000000000 where client_uid = '{client_uid}'"
        return SQLConn.Exec(sql)

    @staticmethod
    def update(robot_qq, group_id, client_uid, client_qq, client_name, display_name, remark_name, nick_name, attr_status):   
        """更新用户"""
        def safe(x): return x.replace("'", "''") if x else ""
        client_name = safe(client_name)
        display_name = safe(display_name)
        remark_name = safe(remark_name)
        nick_name = safe(nick_name)
        attr_status = attr_status or ""

        sql = (
            "update wx_client set client_uid='{0}', client_name='{1}', display_name='{2}', "
            "remark_name='{3}', nick_name='{4}', robot_qq='{5}', attr_status='{7}', "
            "update_date=getdate() where client_qq='{6}'"
        ).format(client_uid, client_name, display_name, remark_name, nick_name, robot_qq, client_qq, attr_status)
        return SQLConn.Exec(sql)

    @staticmethod
    def get_client_qq_by_uid(robot_qq, client_uid):
        if not client_uid: return 0
        sql = str.format("select top 1 client_qq from wx_client where robot_qq = {0} and client_uid = '{1}'", robot_qq, client_uid)
        client_qq = SQLConn.Query(sql)
        if not client_qq:
            return 0
        else:
            return int(client_qq)

    @staticmethod
    def get_client_uid(client_qq):
        sql = str.format("select top 1 client_uid from wx_client where client_qq = '{0}'", client_qq)
        client_uid = SQLConn.Query(sql)
        return client_uid       

    @staticmethod
    def get_client_name_by_qq(client_qq):
        sql = str.format("select top 1 nick_name from wx_client where client_qq = '{0}'", client_qq)
        nick = SQLConn.Query(sql)
        if nick: return nick
        
        sql = str.format("select top 1 display_name from wx_client where client_qq = '{0}'", client_qq)
        return SQLConn.Query(sql)

    @staticmethod
    def get_client_qq_by_name(robot_qq, group_id, client_uid, client_qq, nick_name, display_name, attr_status):
        # Optimized: Removed dead code querying 'robot_group_member'
        # Fallback to checking wx_client by NickName and AttrStatus
        if not attr_status:
            attr_status = ""
        if not nick_name:
            nick_name = ""
        if not display_name:
            display_name = ""    
            
        nick_name = nick_name.replace("'", "''")
        display_name = display_name.replace("'", "''")
        
        # Priority: NickName + AttrStatus
        if nick_name:
            sql = str.format("select top 1 client_qq from wx_client where robot_qq = {0} and nick_name = '{1}' and attr_status = '{2}'", robot_qq, nick_name, attr_status)
            client_qq = SQLConn.Query(sql)
            if client_qq:
                return int(client_qq)
                
        # Fallback: DisplayName (if stored in wx_client, though usually nick_name is better)
        if display_name:
             sql = str.format("select top 1 client_qq from wx_client where robot_qq = {0} and display_name = '{1}' and attr_status = '{2}'", robot_qq, display_name, attr_status)
             client_qq = SQLConn.Query(sql)
             if client_qq:
                 return int(client_qq)
                 
        return 0

    @staticmethod
    def sync_client_uid(robot_qq, new_uid, nick_name, remark_name, display_name):
        """
        Find client by names (prioritizing latest update) and update UID.
        Returns client_qq if found/updated, 0 otherwise.
        """
        nick_name = (nick_name or "").replace("'", "''")
        remark_name = (remark_name or "").replace("'", "''")
        display_name = (display_name or "").replace("'", "''")
        
        # Build where clause conditions
        conditions = []
        if nick_name:
            conditions.append(f"nick_name = N'{nick_name}'")
        if remark_name:
            conditions.append(f"remark_name = N'{remark_name}'")
        if display_name:
            conditions.append(f"display_name = N'{display_name}'")
            
        if not conditions:
            return 0
            
        where_clause = " OR ".join(conditions)
        sql = f"SELECT TOP 1 client_qq FROM wx_client WHERE robot_qq = {robot_qq} AND ({where_clause}) ORDER BY update_date DESC"
        
        res = SQLConn.Query(sql)
        if res:
            client_qq = int(res)
            # Update UID
            update_sql = f"UPDATE wx_client SET client_uid = '{new_uid}', update_date = getdate() WHERE client_qq = {client_qq}"
            if SQLConn.Exec(update_sql):
                return client_qq
        return 0

    @staticmethod
    def find_client_by_name(robot_qq, name):
        """Find client ID by any name field (nick/remark/display)"""
        name = (name or "").replace("'", "''")
        if not name: return 0
        sql = f"SELECT TOP 1 client_qq FROM wx_client WHERE robot_qq = {robot_qq} AND (nick_name = N'{name}' OR remark_name = N'{name}' OR display_name = N'{name}') ORDER BY update_date DESC"
        res = SQLConn.Query(sql)
        return int(res) if res else 0

if __name__ == '__main__':
    common.is_debug = True
    common.is_debug_sql = True   
