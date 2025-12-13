# coding: utf-8
import time
import psutil
import os
import sys

def get_sys_status():
    mem = psutil.virtual_memory()
    cpu = psutil.cpu_percent(interval=None)
    uptime = int(time.time()) - psutil.boot_time()
    
    # Calculate uptime in h:m:s
    m, s = divmod(uptime, 60)
    h, m = divmod(m, 60)
    
    return (
        f"【系统状态】\n"
        f"CPU使用率: {cpu}%\n"
        f"内存使用: {mem.percent}%\n"
        f"运行时间: {int(h)}小时{int(m)}分\n"
        f"Python版本: {sys.version.split()[0]}"
    )

def handle(context):
    """
    Gateway Service Plugin
    Provides system monitoring and maintenance commands.
    """
    content = context.get("content", "").strip()
    bot = context.get("bot")
    
    # Security check: DB based
    user_id = str(context.get("user_id", ""))
    
    is_admin = False
    
    # 1. Config based check (Fast & Fallback)
    if bot and hasattr(bot, "global_config"):
        admins = bot.global_config.get("admins", [])
        if admins and user_id in admins:
            is_admin = True
            
    # 2. DB based check (if not yet admin)
    if not is_admin:
        try:
            from SQLConn import SQLConn
            
            # Get Bot Identifiers
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
            
            # Parameters: BotUin, BotUserName, UserWxId
            params = (bot_uin, bot_wxid, user_id)
            
            res = SQLConn.Query(sql, params)
            if res and int(res) > 0:
                is_admin = True
                
        except Exception as e:
            # DB check failed, ignore
            pass

    if not is_admin:
        # Silently ignore if not admin
        return None
    
    if content == "#status" or content == "#状态":
        gateway_info = ""
        if bot and hasattr(bot, "gateway"):
            # The bot passed here is likely OneBot instance which IS the gateway
            # Or it has .gateway attribute?
            # In onebot.py: "bot": self. And self is OneBot. 
            # OneBot has self.clients and self.ws_bots
            client_count = len(bot.clients) + len(bot.ws_bots)
            gateway_info = f"\n网关连接数: {client_count}"
            
        return {"reply": get_sys_status() + gateway_info, "block": True}

    if content == "#reload" or content == "#重载":
        if bot and hasattr(bot, "plugin_manager"):
            bot.plugin_manager.load_plugins()
            return {"reply": "Plugins reloaded.", "block": True}
        return {"reply": "Plugin manager not found.", "block": True}

    if content == "#help" or content == "#菜单":
        help_msg = """
[Gateway Service]
#status - 查看系统状态
#reload - 重载插件
#help - 显示帮助
""".strip()
        return {"reply": help_msg, "block": True}

    return None
