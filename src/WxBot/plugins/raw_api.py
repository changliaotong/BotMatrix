import json
try:
    from plugins.admin_tools import check_admin
except ImportError:
    def check_admin(bot, user_id):
        if bot and hasattr(bot, "global_config"):
            admins = bot.global_config.get("admins", [])
            return str(user_id) in admins
        return False

def handle(context):
    content = context.get("content", "").strip()
    bot = context.get("bot")
    user_id = str(context.get("user_id", ""))
    
    if not content.startswith("#api "):
        return None
        
    if not check_admin(bot, user_id):
        return None
        
    json_str = content[5:].strip()
    try:
        # Support loose JSON (e.g. single quotes) by replacing? No, strict JSON.
        data = json.loads(json_str)
        action = data.get("action")
        params = data.get("params", {})
        
        if not action:
            return {"reply": "Missing 'action' field", "block": True}
            
        bot.call_api_nowait(action, params)
        return {"reply": f"API {action} called.", "block": True}
    except Exception as e:
        return {"reply": f"JSON Parse Error: {e}", "block": True}
