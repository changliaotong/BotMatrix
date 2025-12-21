from SQLConn import SQLConn
try:
    from plugins.admin_tools import check_admin
except ImportError:
    # Fallback if import fails
    def check_admin(bot, user_id):
        if bot and hasattr(bot, "global_config"):
            admins = bot.global_config.get("admins", [])
            return str(user_id) in admins
        return False

def handle(context):
    content = context.get("content", "").strip()
    bot = context.get("bot")
    user_id = str(context.get("user_id", ""))
    
    if not content.startswith("#db_"):
        return None
        
    if not check_admin(bot, user_id):
        return None

    if content == "#db_status":
        try:
            res = []
            tables = ["ChatLog", "User", "Member"]
            for t in tables:
                c = SQLConn.Query(f"SELECT COUNT(1) FROM [{t}]")
                res.append(f"{t}: {c}")
            return {"reply": "DB Status:\n" + "\n".join(res), "block": True}
        except Exception as e:
            return {"reply": f"DB Check Error: {e}", "block": True}

    if content.startswith("#db_clean "):
        try:
            days_str = content[10:].strip()
            if not days_str.isdigit():
                 return {"reply": "Usage: #db_clean <days>", "block": True}
            
            days = int(days_str)
            if days < 1:
                return {"reply": "Days must be >= 1", "block": True}
                
            sql = f"DELETE FROM ChatLog WHERE create_time < DATEADD(day, -{days}, GETDATE())"
            # We can't get affected rows easily with SQLConn.Exec unless we change it.
            # But we can run it.
            SQLConn.Exec(sql)
            
            return {"reply": f"Clean command executed for logs older than {days} days.", "block": True}
        except Exception as e:
            return {"reply": f"DB Clean Error: {e}", "block": True}

    return None
