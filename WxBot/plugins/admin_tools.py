import gc
import time
from SQLConn import SQLConn

def check_admin(bot, user_id):
    # 1. Config based check
    if bot and hasattr(bot, "global_config"):
        admins = bot.global_config.get("admins", [])
        if str(user_id) in admins:
            return True
            
    # 2. DB based check
    try:
        bot_uin = getattr(bot, "self_id", 0)
        bot_wxid = ""
        if hasattr(bot, "my_account") and isinstance(bot.my_account, dict):
            bot_wxid = bot.my_account.get("UserName", "")
            
        sql = """
          SELECT count(1) 
          FROM [Member] m
          INNER JOIN [User] u ON m.AdminId = u.Id
          WHERE (m.BotUin = ? OR (m.UserName IS NOT NULL AND m.UserName <> '' AND m.UserName = ?))
            AND u.WxId = ?
        """
        params = (bot_uin, bot_wxid, user_id)
        res = SQLConn.Query(sql, params)
        if res and int(res) > 0:
            return True
    except:
        pass
    return False

def handle(context):
    content = context.get("content", "").strip()
    bot = context.get("bot")
    user_id = str(context.get("user_id", ""))
    
    # Only process commands starting with #
    if not content.startswith("#"):
        return None

    # Check Admin
    if not check_admin(bot, user_id):
        return None

    if content == "#gc":
        n = gc.collect()
        return {"reply": f"System GC executed. Collected {n} objects.", "block": True}
        
    if content.startswith("#broadcast "):
        msg = content[11:].strip()
        if not msg:
            return {"reply": "Usage: #broadcast <message>", "block": True}
            
        # Get active groups from DB (last 3 days)
        try:
            sql = "SELECT DISTINCT group_id FROM ChatLog WHERE create_time > DATEADD(day, -3, GETDATE()) AND group_id <> '' AND group_id IS NOT NULL"
            rows = SQLConn.QueryDict(sql)
            count = 0
            for row in rows:
                gid = row.get("group_id")
                if gid:
                    # Delay slightly to avoid flood
                    # But call_api_nowait is fire-and-forget.
                    # We rely on gateway buffering?
                    # Gateway doesn't buffer outgoing actions yet, but asyncio handles it.
                    bot.call_api_nowait("send_group_msg", {"group_id": gid, "message": msg})
                    count += 1
            return {"reply": f"Broadcast command sent for {count} groups.", "block": True}
        except Exception as e:
            return {"reply": f"Broadcast failed: {e}", "block": True}
            
    if content.startswith("#echo "):
        msg = content[6:].strip()
        return {"reply": msg, "block": True}

    return None
