#encoding:utf-8
from MetaData import *
from SQLConn import *
from wxclient import User
import hashlib

class wx_client_v2(MetaData):
    """改进版微信客户端识别逻辑"""
    def __init__(self, client_qq):
        self.database = "sz84_robot"
        self.table_name = "wx_client"
        self.key_field = "client_qq"
        self.key_value = client_qq

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
        client_qq = wx_client_v2.get_client_qq_by_uid(client_uid)

        # Step 2: 若找不到，尝试匹配旧记录
        if not client_qq:
            client_qq = wx_client_v2.match_existing_client(robot_qq, group_id, remark_name, display_name, nick_name, attr_status)
            if client_qq:
                # 同一个人，但 UID 改变了
                SQLConn.Exec(f"update wx_client set client_uid='{client_uid}', update_date=getdate() where client_qq='{client_qq}'")
            else:
                # Step 3: 若名字都 unknown，则生成 anon_key
                if not client_name or client_name.lower() == "unknown":
                    anon_key = wx_client_v2._make_anon_key(robot_qq, group_id, client_uid)
                    client_name = anon_key
                    display_name = anon_key
                    remark_name = anon_key
                    nick_name = anon_key
                is_new = True
                wx_client_v2.append(robot_qq, group_id, client_uid, 0, client_name, display_name, remark_name, nick_name, attr_status)
                client_qq = wx_client_v2.get_client_qq_by_uid(client_uid)

        # Step 4: 更新常规信息
        if not is_new:
            wx_client_v2.update(robot_qq, group_id, client_uid, client_qq, client_name, display_name, remark_name, nick_name, attr_status)

        # Step 5: 确保 client 表中存在
        if not User.exists(client_qq):
            User.append(robot_qq, group_id, client_qq, client_name, group_id)

        return client_qq

    @staticmethod
    def match_existing_client(robot_qq, group_id, remark_name, display_name, nick_name, attr_status):
        """匹配旧 UID 用户"""
        fields = [remark_name, display_name, nick_name]
        for name_field in fields:
            if not name_field or name_field.lower() == "unknown":
                continue
            name_field = name_field.replace("'", "''")
            sql = (
        "select top 1 a.UserId "
        "from GroupMember a inner join wx_client b on a.UserId = b.client_qq "
        f"where a.GroupId = {group_id} and "
        f"(a.NickName = '{name_field}' or a.DisplayName = '{name_field}' or b.remark_name = '{name_field}') "
        f"and b.attr_status = '{attr_status}' "
        "order by a.InsertDate desc"
    )
            client_qq = SQLConn.Query(sql)
            if client_qq:
                return int(client_qq)
        return 0

    @staticmethod
    def append(robot_qq, group_id, client_uid, client_qq, client_name, display_name, remark_name, nick_name, attr_status):
        """插入用户（带未知名修正）"""
        def safe(x): return x.replace("'", "''") if x else ""
        client_name = safe(client_name)
        display_name = safe(display_name)
        remark_name = safe(remark_name)
        nick_name = safe(nick_name)
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
        sql = (
            "update wx_client set client_uid='{0}', client_name='{1}', display_name='{2}', "
            "remark_name='{3}', nick_name='{4}', robot_qq='{5}', attr_status='{7}', "
            "update_date=getdate() where client_qq='{6}'"
        ).format(client_uid, client_name, display_name, remark_name, nick_name, robot_qq, client_qq, attr_status)
        return SQLConn.Exec(sql)

    @staticmethod
    def get_client_qq_by_uid(client_uid):
        """查找 UID 对应 client_qq"""
        sql = f"select top 1 client_qq from wx_client where client_uid = '{client_uid}'"
        res = SQLConn.Query(sql)
        return int(res) if res else 0

