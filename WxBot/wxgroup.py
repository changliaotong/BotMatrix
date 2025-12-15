#encoding:utf-8
from MetaData import *
from SQLConn import *
from wxclient import *

class wx_group(MetaData):

    def __init__(self, group_id):
        self.database = "sz84_robot"
        self.table_name = "wx_group"
        self.key_field = "group_id"
        self.key_value = group_id

    @staticmethod
    def get_group_uid(group_id):
        return wx_group.get_value("group_uid", group_id)

    @staticmethod
    def get_group_id(group_uid):
        if not group_uid: return 0
        sql = f"SELECT TOP 1 group_id FROM wx_group WHERE group_uid = '{group_uid}'"
        res = SQLConn.Query(sql)
        # Debug: check if we found it
        # if not res and len(group_uid) > 50:
        #    print(f"[wxgroup] get_group_id FAILED for long uid ({len(group_uid)}). SQL={sql}")
        return int(res) if res else 0  

    @staticmethod
    def update_name(group_id, new_name):
        """Update group name only"""
        new_name = new_name.replace("'", "''")
        sql = f"UPDATE wx_group SET group_name = N'{new_name}', update_date = getdate() WHERE group_id = {group_id}"
        return SQLConn.Exec(sql)

    @staticmethod
    def update(robot_qq, group_uid, group_id, group_name, client_qq, client_name):
        if not group_id: return False
        group_name = group_name.replace("'", "''")
        
        # If client_qq is 0, try to preserve existing value
        if client_qq == 0:
            existing_client_qq = wx_group.get_int("client_qq", group_id)
            if existing_client_qq > 0:
                client_qq = existing_client_qq
        
        sql = f"UPDATE wx_group SET group_uid = '{group_uid}', group_name = '{group_name}', robot_qq = {robot_qq}, client_qq = {client_qq}, update_date = getdate() WHERE group_id = {group_id}"
        return SQLConn.Exec(sql)

    @staticmethod
    def exists(robot_qq, group_uid, group_id, group_name):
        group_name = group_name.replace("'", "''")
        # Ensure parameters are handled safely
        sql = f"SELECT TOP 1 group_id FROM wx_group WHERE group_uid = '{group_uid}' OR group_id = {group_id or 0} OR (robot_qq = {robot_qq} AND group_name = '{group_name}')"
        res = SQLConn.Query(sql)
        if res is None or res == "":
            return 0
        else:
            return int(res)

    @staticmethod
    def get_member(robot_qq, group_id, client_uid, client_name, display_name, remark_name, nick_name, attr_status):
        group_name = ""
        client_qq = wx_client.get_client_qq(robot_qq, group_id, client_uid, client_name, display_name, remark_name, nick_name, attr_status)
        return client_qq

    @staticmethod
    def get_wx_group(robot_qq, group_uid, group_name, client_qq, client_name):
        is_new = False
        group_id = wx_group.get_group_id(group_uid)
        if not group_id:
            group_id = wx_group.exists(robot_qq, group_uid, group_id, group_name)
            if group_id:
                # print(f"[wxgroup] Recovered group_id={group_id} for uid={group_uid} by name={group_name}")
                pass
            else:
                is_new = True
                group_id = wx_group.append(robot_qq, group_uid, group_id, group_name, client_qq, client_name)
        if not is_new:
            wx_group.update(robot_qq, group_uid, group_id, group_name, client_qq, client_name)
        return group_id    

    @staticmethod
    def append(robot_qq, group_uid, group_id, group_name, client_qq, client_name):
        # print(f"[wxgroup] Appending group: uid_len={len(group_uid)} name={group_name}")
        group_name = group_name.replace("'", "''")
        
        sql = f"INSERT INTO wx_group (group_uid, group_name, robot_qq, client_qq) VALUES ('{group_uid}', N'{group_name}', {robot_qq}, {client_qq}); SELECT @@IDENTITY"
        
        conn = None
        try:
            conn = SQLConn.conn()
            cursor = conn.cursor()
            cursor.execute(sql)
            
            # Fetch the result of the SELECT @@IDENTITY
            # Note: iterate to the last result set if there are multiple (e.g. from triggers)
            row = None
            try:
                # Some drivers return the INSERT result first, then the SELECT result
                # We need to handle potential multiple result sets
                while True:
                    try:
                        row = cursor.fetchone()
                        if row: break # Found a result
                    except:
                        pass
                    if not cursor.nextset():
                        break
            except Exception:
                # Fallback for simple drivers
                pass
                
            # If logic above didn't catch it (simple execute), try fetchone directly
            if not row:
                row = cursor.fetchone()

            conn.commit()
            if row and row[0]:
                return int(row[0])
            return 0
        except Exception as e:
            print(f"[wxgroup] Append Failed: {e}")
            if conn: conn.rollback()
            return 0
        finally:
            if conn:
                try:
                    conn.close()
                except:
                    pass

    @staticmethod
    def is_admin(group_id):
        sql = str.format("select top 1 is_admin from wx_group where group_id = {0}", group_id)
        res = SQLConn.Query(sql)
        return bool(res)

    @staticmethod
    def sync_group_uid(robot_qq, group_name, new_uid):
        """
        Find group by name (prioritizing latest update) and update its UID.
        Returns group_id if found and updated, 0 otherwise.
        """
        group_name = group_name.replace("'", "''")
        # Find the most recently updated group with this name for this robot
        sql = f"SELECT TOP 1 group_id FROM wx_group WHERE robot_qq = {robot_qq} AND group_name = N'{group_name}' ORDER BY update_date DESC"
        res = SQLConn.Query(sql)
        
        if res:
            group_id = int(res)
            # Update the UID for this group
            update_sql = f"UPDATE wx_group SET group_uid = '{new_uid}', update_date = getdate() WHERE group_id = {group_id}"
            if SQLConn.Exec(update_sql):
                return group_id
        return 0

