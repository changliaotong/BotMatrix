#encoding:utf-8
from MetaData import *
from SQLConn import *

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
    def get_client_qq(robot_qq, group_id, client_uid, client_name, display_name, remark_name, nick_name, attr_status):
        is_new = False
        client_qq = wx_client.get_client_qq_by_uid(client_uid)
        if not client_qq:
            # Fallback: Try to find by Name/AttrStatus if UID changed (e.g. restart)
            client_qq = wx_client.get_client_qq_by_name(robot_qq, group_id, client_uid, client_qq, nick_name, display_name, attr_status)
            if client_qq:
                # print(f"[wxclient] Recovered client_qq={client_qq} for uid={client_uid} by name")
                pass
            else:
                is_new = True
                wx_client.append(robot_qq, group_id, client_uid, client_qq, client_name, display_name, remark_name, nick_name, attr_status)
                client_qq = wx_client.get_client_qq_by_uid(client_uid)                       
        if not is_new:
            # Always update UID to keep it fresh
            wx_client.update(robot_qq, group_id, client_uid, client_qq, client_name, display_name, remark_name, nick_name, attr_status)
        
        # Ensure it exists in robot_client mapping (if needed)
        try:
            if client_qq > 0:
                if not User.exists(client_qq):
                    # print(f"[wxclient] User {client_qq} not found in User table, inserting...")
                    User.append(robot_qq, group_id, client_qq, client_name, group_id)
                else:
                    pass
                    # print(f"[wxclient] User {client_qq} already exists.")
        except Exception as e:
            print(f"[wxclient] User sync error: {e}")
        
        return client_qq    

    @staticmethod
    def append(robot_qq, group_id, client_uid, client_qq, client_name, display_name, remark_name, nick_name, attr_status):
        client_name = client_name.replace("'", "''")
        display_name = display_name.replace("'", "''")
        remark_name = remark_name.replace("'", "''")
        nick_name = nick_name.replace("'", "''")
        # Ensure attr_status is not None
        attr_status = attr_status or ""
        sql = str.format("insert into wx_client(client_uid, client_name, display_name, remark_name, nick_name, robot_qq, attr_status) "\
            "values('{0}','{1}','{2}','{3}','{4}','{5}','{6}')", client_uid, client_name, display_name, remark_name, nick_name, robot_qq, attr_status)
        SQLConn.Exec(sql)
        # Assuming client_oid is the auto-increment primary key, map it to client_qq
        # Note: This logic assumes 'client_oid' exists. If not, this will fail. 
        # But keeping original logic for safety.
        sql = str.format("update wx_client set client_qq = client_oid + 90000000000 where client_uid = '{0}'", client_uid)
        return SQLConn.Exec(sql)

    @staticmethod
    def update(robot_qq, group_id, client_uid, client_qq, client_name, display_name, remark_name, nick_name, attr_status):   
        client_name = client_name.replace("'", "''")
        display_name = display_name.replace("'", "''")
        remark_name = remark_name.replace("'", "''")
        nick_name = nick_name.replace("'", "''")
        attr_status = attr_status or ""
        sql = str.format("update wx_client set client_uid = '{0}', client_name = '{1}', display_name = '{2}', remark_name = '{3}', nick_name = '{4}', " \
            "robot_qq='{5}' , attr_status = '{7}', update_date = getdate() where client_qq = '{6}'", client_uid, client_name, display_name, 
            remark_name, nick_name, robot_qq, client_qq, attr_status)
        res = SQLConn.Exec(sql)
        return res

    @staticmethod
    def get_client_qq_by_uid(client_uid):
        if not client_uid: return 0
        sql = str.format("select top 1 client_qq from wx_client where client_uid = '{0}'", client_uid)
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
            sql = str.format("select top 1 client_qq from wx_client where nick_name = '{0}' and attr_status = '{1}'", nick_name, attr_status)
            client_qq = SQLConn.Query(sql)
            if client_qq:
                return int(client_qq)
                
        # Fallback: DisplayName (if stored in wx_client, though usually nick_name is better)
        if display_name:
             sql = str.format("select top 1 client_qq from wx_client where display_name = '{0}' and attr_status = '{1}'", display_name, attr_status)
             client_qq = SQLConn.Query(sql)
             if client_qq:
                 return int(client_qq)
                 
        return 0


if __name__ == '__main__':
    common.is_debug = True
    common.is_debug_sql = True   