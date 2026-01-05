using BotMatrix.SDK;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using System.Text.Json;

namespace PointsSystem
{
    public class MarketOrder
    {
        public string Id { get; set; } = Guid.NewGuid().ToString("n").Substring(0, 8);
        public string UserId { get; set; } = string.Empty;
        public string Side { get; set; } = "buy"; // "buy" or "sell"
        public long Amount { get; set; }
        public double Price { get; set; } // Price in Global points (G)
        public DateTime CreatedAt { get; set; } = DateTime.Now;
    }

    /// <summary>
    /// 权限与激活管理器
    /// 明确“消费总额”统计口径：
    /// 1. 收入表 (Income)：记录机器人主人购买服务、积分、算力的真实货币支出。
    /// 2. 积分审计日志 (PointsLogs)：记录群成员在插件内消耗通用积分的虚拟支出。
    /// 
    /// 本群积分激活逻辑：
    /// - 必须由【机器人主人】(Robot Owner) 手动开通。
    /// - 开通后，该群默认进入“本群积分模式”，影响签到、任务等默认产出。
    /// </summary>
    public static class PrivilegeManager
    {
        // 默认激活门槛配置 (如果配置文件中没有则使用这些默认值)
        private static long _todayThreshold = 500;
        private static long _rollingThreshold = 1000;

        public static void Initialize(JsonElement? config)
        {
            if (config != null && config.Value.TryGetProperty("config", out var cfg))
            {
                if (cfg.TryGetProperty("thresholds", out var th))
                {
                    if (th.TryGetProperty("today_total", out var today)) _todayThreshold = today.GetInt64();
                    if (th.TryGetProperty("rolling_12m", out var rolling)) _rollingThreshold = rolling.GetInt64();
                }
            }
        }

        public static long TODAY_TOTAL_THRESHOLD => _todayThreshold;
        public static long ROLLING_12M_THRESHOLD => _rollingThreshold;

        /// <summary>
        /// 检查是否为机器人主人 (Robot Owner)
        /// </summary>
        public static async Task<bool> IsRobotOwner(Context ctx, string userId)
        {
            if (string.IsNullOrEmpty(userId)) return false;
            
            // 从数据库 group 表中查询 RobotOwner
            // 映射到 table:groups:id:{groupId}:robot_owner
            string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
            if (!string.IsNullOrEmpty(groupId))
            {
                string owner = await ctx.Session.GetAsync<string>($"table:groups:id:{groupId}:robot_owner");
                if (owner == userId) return true;
            }

            // 备选方案：从全局管理员列表中检查
            string admins = await ctx.Session.GetAsync<string>("config:global:admins") ?? "";
            return admins.Split(',').Contains(userId) || userId == "admin";
        }

        /// <summary>
        /// 检查群积分模式是否已激活
        /// </summary>
        public static async Task<bool> IsGroupModeActive(Context ctx, string groupId)
        {
            if (string.IsNullOrEmpty(groupId)) return false;
            return await ctx.Session.GetAsync<bool>($"config:group:{groupId}:points_mode_active");
        }

        /// <summary>
        /// 个人激活状态检查 (用于本机积分等个人特权功能)
        /// </summary>
        public static async Task<(bool isActive, long currentRollingTotal, long currentTodayTotal, long totalThreshold, long todayThreshold)> CheckPersonalActivation(Context ctx)
        {
            string userId = ctx.Event.Payload["from"]?.ToString() ?? "";
            DateTime now = DateTime.Now;
            
            string todayKey = $"stats:user:{userId}:spent:date:{now:yyyyMMdd}";
            long todayTotal = await ctx.Session.GetAsync<long>(todayKey);

            long rollingTotal = 0;
            for (int i = 0; i < 12; i++)
            {
                string monthKey = $"stats:user:{userId}:spent:month:{now.AddMonths(-i):yyyyMM}";
                rollingTotal += await ctx.Session.GetAsync<long>(monthKey);
            }

            bool isActive = rollingTotal >= ROLLING_12M_THRESHOLD || todayTotal >= TODAY_TOTAL_THRESHOLD;
            return (isActive, rollingTotal, todayTotal, ROLLING_12M_THRESHOLD, TODAY_TOTAL_THRESHOLD);
        }

        /// <summary>
        /// 检查是否为超级积分用户 (不收打赏手续费)
        /// </summary>
        public static async Task<bool> IsSuperPointsUser(Context ctx, string userId)
        {
            if (string.IsNullOrEmpty(userId)) return false;
            // 从 table:users:id:{userId}:is_super_points 获取
            return await ctx.Session.GetAsync<bool>($"table:users:id:{userId}:is_super_points");
        }
    }

    class Program
    {
        private static BotMatrixPlugin _plugin = null!;

        static async Task Main(string[] args)
        {
            _plugin = new BotMatrixPlugin();

            // 初始化激活门槛
            PrivilegeManager.Initialize(_plugin.Config);

            // 1. 处理通用积分 (Global Points) - 官方严控
            // 存储在 [Users] 表
            _plugin.OnAction("transfer_global", async ctx => {
                string callerId = ctx.Event.Payload.ContainsKey("caller_id") ? ctx.Event.Payload["caller_id"]?.ToString() ?? "unknown" : "unknown";
                string operatorId = ctx.Event.Payload.ContainsKey("from") ? ctx.Event.Payload["from"]?.ToString() ?? "" : "";
                
                // 强制安全校验：
                // 1. 必须是官方插件 (IsOfficialPlugin)
                // 2. 如果是增分操作(amount > 0)，严禁通过普通用户指令触发（即 from 必须为空或系统级账号）
                // 3. 即使是群主/机器人主人，也无权通过指令直接增减通用积分
                long amount = ctx.Event.Payload.ContainsKey("amount") ? Convert.ToInt64(ctx.Event.Payload["amount"]?.ToString() ?? "0") : 0;
                
                if (!IsOfficialPlugin(callerId)) {
                    ctx.Reply("❌ 安全错误：非官方插件严禁操作通用积分。");
                    return;
                }

                if (amount > 0 && !string.IsNullOrEmpty(operatorId)) {
                    // 只有系统内部行为（如充值回调、签到赠送）可加分，用户指令不可直接加分
                    ctx.Reply("❌ 安全错误：禁止通过用户指令直接增加通用积分。");
                    return;
                }

                string userId = ctx.Event.Payload.ContainsKey("user_id") ? ctx.Event.Payload["user_id"]?.ToString() ?? "" : "";

                // 映射到 Users 表
                string key = $"table:users:id:{userId}:global_points";
                long current = await ctx.Session.GetAsync<long>(key);
                await ctx.Session.SetAsync(key, current + amount);

                // 更新消费统计 (如果 amount < 0 说明是消费/支出)
                if (amount < 0) {
                    long spent = Math.Abs(amount);
                    DateTime now = DateTime.Now;
                    
                    // 1. 更新本月累计消费 (用于滚动12个月判定)
                    string monthKey = $"stats:user:{userId}:spent:month:{now:yyyyMM}";
                    long monthlyTotal = await ctx.Session.GetAsync<long>(monthKey);
                    await ctx.Session.SetAsync(monthKey, monthlyTotal + spent);

                    // 2. 更新当日累计消费
                    string todayKey = $"stats:user:{userId}:spent:date:{now:yyyyMMdd}";
                    long todayTotal = await ctx.Session.GetAsync<long>(todayKey);
                    await ctx.Session.SetAsync(todayKey, todayTotal + spent);

                    // 3. 同时更新当日最高单次消费 (保留作为参考)
                    string maxKey = $"stats:user:{userId}:max_spent:date:{now:yyyyMMdd}";
                    long todayMax = await ctx.Session.GetAsync<long>(maxKey);
                    if (spent > todayMax) {
                        await ctx.Session.SetAsync(maxKey, spent);
                    }

                    // 4. 记录积分审计日志 (Internal Points Audit Log)
                    // 注意：这里记录的是虚拟积分的变动，不计入 real-money 收入表 (Income Table)
                    var logEntry = new Dictionary<string, object> {
                        { "user_id", userId },
                        { "type", "global" },
                        { "amount", amount },
                        { "balance_after", current + amount },
                        { "caller_id", callerId },
                        { "reason", ctx.Event.Payload.ContainsKey("reason") ? ctx.Event.Payload["reason"]?.ToString() ?? "插件消费" : "插件消费" },
                        { "group_id", ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "0" : "0" },
                        { "created_at", now.ToString("yyyy-MM-dd HH:mm:ss") }
                    };
                    await ctx.Session.SetAsync("table:points_logs:insert", logEntry);
                }

                // Console.WriteLine($"[Global] {callerId} transferred {amount} to User {userId}");
            });

            // 6. 自动路由积分 Action (transfer_auto)
            _plugin.OnAction("transfer_auto", async ctx => {
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                bool isGroupActive = await PrivilegeManager.IsGroupModeActive(ctx, groupId);
                
                string action = isGroupActive ? "transfer_group" : "transfer_global";
                await _plugin.EmitAction(action, ctx.Event.Payload);
            });

            // 5. 激活与状态查询
            _plugin.Command(new[] { "activate", "jh", "jihuo", "激活", "status", "zt", "zhuangtai", "状态" }, async ctx => {
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                bool isGroupActive = await PrivilegeManager.IsGroupModeActive(ctx, groupId);
                var (isPersonalActive, rollingTotal, todayTotal, totalThreshold, todayThreshold) = await PrivilegeManager.CheckPersonalActivation(ctx);

                var resp = $"🛠️ 积分系统状态查询：\n" +
                           $"------------------\n" +
                           $"🌍 通用积分：✅ 默认开启\n" +
                           $"🏘️ 本群积分模式：{(isGroupActive ? "✅ 已由主人开启" : "❌ 未开启 (默认使用通用积分)")}\n" +
                           $"🤖 个人特权功能：{(isPersonalActive ? "✅ 已激活" : "❌ 未激活")}\n" +
                           $"------------------\n" +
                           $"📊 个人特权进度 (满足其一即可)：\n" +
                           $"1. 今日累计消费：{todayTotal} / {todayThreshold} G\n" +
                           $"2. 滚动12个月累计：{rollingTotal} / {totalThreshold} G\n" +
                           $"------------------\n" +
                           $"💡 提示：本群积分模式需由【机器人主人】执行 /activate_group 开启。";
                
                ctx.Reply(resp);
            });

            // 6. 群积分模式控制 (仅限机器人主人)
            _plugin.Command(new[] { "activate_group", "qjh", "qunjihuo", "群激活" }, async ctx => {
                string userId = ctx.Event.Payload["from"]?.ToString() ?? "";
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";

                if (string.IsNullOrEmpty(groupId)) {
                    ctx.Reply("⚠️ 请在群聊中执行此指令。");
                    return;
                }

                if (!await PrivilegeManager.IsRobotOwner(ctx, userId)) {
                    ctx.Reply("❌ 权限不足：只有【机器人主人】有权决定是否开启本群积分模式。");
                    return;
                }

                bool currentState = await PrivilegeManager.IsGroupModeActive(ctx, groupId);
                bool newState = !currentState;

                await ctx.Session.SetAsync($"config:group:{groupId}:points_mode_active", newState);
                
                string modeName = newState ? "【本群积分模式】" : "【通用积分模式】";
                ctx.Reply($"✅ 操作成功！当前群组已切换至 {modeName}。\n" +
                          (newState ? "💡 提示：签到、任务等功能将优先发放/消耗本群积分。" : "💡 提示：所有功能已恢复使用通用积分。"));
            });

            // 2. 处理本群积分 (Group Points) - 自由模式
            // 存储在 [GroupMembers] 表
            _plugin.OnAction("transfer_group", async ctx => {
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                
                // 检查群模式是否开启
                if (!await PrivilegeManager.IsGroupModeActive(ctx, groupId)) {
                    ctx.Reply("⚠️ 本群尚未开启【本群积分模式】，无法进行本群积分操作。");
                    return;
                }

                string userId = ctx.Event.Payload.ContainsKey("user_id") ? ctx.Event.Payload["user_id"]?.ToString() ?? "" : "";
                long amount = ctx.Event.Payload.ContainsKey("amount") ? Convert.ToInt64(ctx.Event.Payload["amount"]?.ToString() ?? "0") : 0;
                string pluginId = ctx.Event.Payload.ContainsKey("caller_id") ? ctx.Event.Payload["caller_id"]?.ToString() ?? "unknown" : "unknown";
                string operatorId = ctx.Event.Payload.ContainsKey("from") ? ctx.Event.Payload["from"]?.ToString() ?? "" : "";

                if (string.IsNullOrEmpty(groupId)) {
                    ctx.Reply("❌ 错误：无法识别当前群组 ID。");
                    return;
                }

                // 权限检查：只有【官方插件】或【机器人主人】可以操作本群积分
                bool isOwner = await PrivilegeManager.IsRobotOwner(ctx, operatorId);
                bool isOfficial = IsOfficialPlugin(pluginId);

                if (!isOwner && !isOfficial) {
                    ctx.Reply("❌ 权限不足：只有官方插件或机器人主人可操作本群积分。");
                    return;
                }

                // 更新成员积分 (已取消金库预算限制)
                string memberKey = $"table:group_members:group:{groupId}:user:{userId}:points";
                long current = await ctx.Session.GetAsync<long>(memberKey);
                await ctx.Session.SetAsync(memberKey, current + amount);

                // 记录本群积分变动日志
                var logEntry = new Dictionary<string, object> {
                    { "user_id", userId },
                    { "group_id", groupId },
                    { "type", "group" },
                    { "amount", amount },
                    { "balance_after", current + amount },
                    { "caller_id", pluginId },
                    { "operator_id", operatorId },
                    { "reason", ctx.Event.Payload.ContainsKey("reason") ? ctx.Event.Payload["reason"]?.ToString() ?? "系统操作" : "系统操作" },
                    { "created_at", DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss") }
                };
                await ctx.Session.SetAsync("table:points_logs:insert", logEntry);

                string logMsg = isOwner ? $"[Owner Action] {operatorId} adjusted {amount} Q for User {userId} in Group {groupId}" :
                                          $"[Official Plugin] {pluginId} transferred {amount} Q in Group {groupId} to User {userId}";
                // Console.Error.WriteLine(logMsg);
            });

            // 3. 处理本机积分 (Local Points)
            // 存储在 [Friend] 表
            _plugin.OnAction("transfer_local", async ctx => {
                var (isPersonalActive, _, _, _, _) = await PrivilegeManager.CheckPersonalActivation(ctx);
                if (!isPersonalActive) {
                    ctx.Reply($"⚠️ 本机积分功能尚未激活。\n激活条件：个人今日累计消费满 {PrivilegeManager.TODAY_TOTAL_THRESHOLD} G 或最近12个月累计消费满 {PrivilegeManager.ROLLING_12M_THRESHOLD} G。");
                    return;
                }

                string botId = ctx.Event.Payload.ContainsKey("bot_id") ? ctx.Event.Payload["bot_id"]?.ToString() ?? "" : "";
                string userId = ctx.Event.Payload.ContainsKey("user_id") ? ctx.Event.Payload["user_id"]?.ToString() ?? "" : "";
                long amount = ctx.Event.Payload.ContainsKey("amount") ? Convert.ToInt64(ctx.Event.Payload["amount"]?.ToString() ?? "0") : 0;
                string pluginId = ctx.Event.Payload.ContainsKey("caller_id") ? ctx.Event.Payload["caller_id"]?.ToString() ?? "unknown" : "unknown";
                
                // 映射到 Friend 表
                string key = $"table:bot_friends:bot:{botId}:user:{userId}:local_points";
                long current = await ctx.Session.GetAsync<long>(key);
                await ctx.Session.SetAsync(key, current + amount);

                // Console.Error.WriteLine($"[Local] {pluginId} transferred {amount} for Bot {botId} to User {userId}");
            });

            // 4. 查询余额意图 & 特定查询指令
            _plugin.OnIntent("check_points", async ctx => {
                string text = ctx.Event.Payload.ContainsKey("text") ? ctx.Event.Payload["text"]?.ToString() ?? "" : "";
                string userId = ctx.Event.Payload.ContainsKey("from") ? ctx.Event.Payload["from"]?.ToString() ?? "" : "";
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                string botId = ctx.Event.Payload.ContainsKey("bot_id") ? ctx.Event.Payload["bot_id"]?.ToString() ?? "" : "";

                bool isGroupActive = await PrivilegeManager.IsGroupModeActive(ctx, groupId);
                string groupPointsName = await ctx.Session.GetAsync<string>($"table:groups:id:{groupId}:points_name") ?? "本群积分";
                string localPointsName = await ctx.Session.GetAsync<string>($"table:bot_friends:bot:{botId}:user:{userId}:points_name") ?? "本机积分";

                long globalPoints = await ctx.Session.GetAsync<long>($"table:users:id:{userId}:global_points");
                long groupPoints = await ctx.Session.GetAsync<long>($"table:group_members:group:{groupId}:user:{userId}:points");
                long localPoints = await ctx.Session.GetAsync<long>($"table:bot_friends:bot:{botId}:user:{userId}:local_points");

                // 根据指令内容决定显示侧重
                if (text == "通用积分") {
                    ctx.Reply($"🌍 您的【通用积分】余额为：{globalPoints} G");
                    return;
                }

                if (text == groupPointsName || text == "本群积分") {
                    ctx.Reply($"🏘️ 您在当前群的【{groupPointsName}】余额为：{groupPoints} Q");
                    return;
                }

                if (text == localPointsName || text == "本机积分") {
                    ctx.Reply($"🤖 您在当前机器人的【{localPointsName}】余额为：{localPoints} L");
                    return;
                }

                // 默认“积分”指令
                string resp = string.Empty;
                if (isGroupActive) {
                    // 如果开启了本群积分模式，且只发“积分”，则重点显示本群积分
                    resp = $"💰 您的本群资产：\n" +
                           $"------------------\n" +
                           $"🏘️ {groupPointsName}: {groupPoints} Q\n" +
                           $"🌍 通用积分: {globalPoints} G\n" +
                           $"------------------\n" +
                           $"💡 当前已开启【本群积分模式】，日常签到与游戏将优先使用{groupPointsName}。";
                } else {
                    resp = $"💰 您的资产概览：\n" +
                           $"------------------\n" +
                           $"🌍 通用积分 (G): {globalPoints}\n" +
                           $"🏘️ {groupPointsName} (Q): {groupPoints}\n" +
                           $"🤖 {localPointsName} (L): {localPoints}\n" +
                           $"------------------\n" +
                           $"💡 当前处于【通用积分模式】。";
                }
                
                ctx.Reply(resp);
            });

            // 添加显式命令支持
            _plugin.Command(new[] { "global_points", "tyjf", "tongyongjifen", "通用积分" }, async ctx => await _plugin.EmitIntent("check_points", ctx.Event.Payload));
            _plugin.Command(new[] { "group_points", "bqjf", "benqunjifen", "本群积分" }, async ctx => await _plugin.EmitIntent("check_points", ctx.Event.Payload));
            // 7. 查询积分指令
            _plugin.Command(new[] { "points", "jf", "jifen", "积分", "balance", "ye", "yue", "余额" }, async ctx => {
                string userId = ctx.Event.Payload["from"]?.ToString() ?? "";
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                
                // 获取通用积分
                long globalPoints = await ctx.Session.GetAsync<long>($"table:users:id:{userId}:global_points");
                
                string resp = $"💰 您的积分资产：\n" +
                              $"------------------\n" +
                              $"🌐 通用积分：{globalPoints} G\n";

                // 如果在群里，获取本群积分
                if (!string.IsNullOrEmpty(groupId)) {
                    string groupPointsName = await ctx.Session.GetAsync<string>($"table:groups:id:{groupId}:points_name") ?? "本群积分";
                    long groupPoints = await ctx.Session.GetAsync<long>($"table:group_members:group:{groupId}:user:{userId}:points");
                    bool isActive = await PrivilegeManager.IsGroupModeActive(ctx, groupId);
                    
                    resp += $"群 【{groupPointsName}】：{groupPoints} Q {(isActive ? "" : "(未开启)")}\n";
                    
                    // 获取冻结积分
                    long frozenGroup = await ctx.Session.GetAsync<long>($"frozen:group:{groupId}:user:{userId}");
                    if (frozenGroup > 0) resp += $"❄️ 冻结({groupPointsName})：{frozenGroup} Q\n";
                }

                long frozenGlobal = await ctx.Session.GetAsync<long>($"frozen:global:user:{userId}");
                if (frozenGlobal > 0) resp += $"❄️ 冻结(通用)：{frozenGlobal} G\n";

                ctx.Reply(resp);
            });

            // 7. 积分互转引导 (移除固定比例转换)
            _plugin.Command(new[] { "convert", "dh", "duihuan", "兑换" }, async ctx => {
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                string groupPointsName = await ctx.Session.GetAsync<string>($"table:groups:id:{groupId}:points_name") ?? "本群积分";

                ctx.Reply($"🔄 积分转换已升级为【市场化定价】交易系统。\n" +
                          $"------------------\n" +
                          $"本群已不再支持固定比例互转。请使用以下指令在市场中与其他用户进行兑换：\n\n" +
                          $"📈 查看当前市场价：/market list\n" +
                          $"💰 买入{groupPointsName}：/market buy Q <数量> <价格>\n" +
                          $"💵 卖出{groupPointsName}：/market sell Q <数量> <价格>\n\n" +
                          $"💡 提示：兑换比例由市场竞争决定，群主无法直接干预价格。");
            });

            // 4. 自定义积分名称
            _plugin.Command(new[] { "set_points_name", "szjfmc", "shezhijifenmingcheng", "设置积分名称" }, async ctx => {
                if (ctx.Args.Length < 2) {
                    ctx.Reply("📝 使用方法：/set_points_name <group|local> <新名称>");
                    return;
                }

                string type = ctx.Args[0].ToLower();
                string newName = ctx.Args[1];
                string userId = ctx.Event.Payload["from"]?.ToString() ?? "";
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                string botId = ctx.Event.Payload.ContainsKey("bot_id") ? ctx.Event.Payload["bot_id"]?.ToString() ?? "" : "";

                if (type == "group") {
                    // 群积分更名需要激活群模式且是主人
                    if (!await PrivilegeManager.IsGroupModeActive(ctx, groupId)) {
                        ctx.Reply("⚠️ 本群尚未开启【本群积分模式】，无法修改名称。");
                        return;
                    }
                    if (!await PrivilegeManager.IsRobotOwner(ctx, userId)) {
                        ctx.Reply("❌ 权限不足：只有【机器人主人】可以修改本群积分名称。");
                        return;
                    }
                    await ctx.Session.SetAsync($"table:groups:id:{groupId}:points_name", newName);
                    ctx.Reply($"✅ 本群积分已更名为：【{newName}】");
                } else {
                    // 本机积分更名需要个人激活
                    var (isPersonalActive, _, _, _, _) = await PrivilegeManager.CheckPersonalActivation(ctx);
                    if (!isPersonalActive) {
                        ctx.Reply($"⚠️ 自定义本机积分功能尚未激活。");
                        return;
                    }
                    await ctx.Session.SetAsync($"table:bot_friends:bot:{botId}:user:{userId}:points_name", newName);
                    ctx.Reply($"✅ 本机积分已更名为：【{newName}】");
                }
            });

            // 5. 积分交易市场 (Exchange Market)
            _plugin.Command(new[] { "market", "sc", "shichang", "市场" }, async ctx => {
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                
                // 市场必须在群积分模式开启后才可用
                if (!await PrivilegeManager.IsGroupModeActive(ctx, groupId)) {
                    ctx.Reply("⚠️ 本群尚未开启【本群积分模式】，交易市场暂不开放。");
                    return;
                }

                string subCmd = ctx.Args.Length > 0 ? ctx.Args[0].ToLower() : "list";
                switch (subCmd) {
                    case "list":
                        await ShowMarketOverview(ctx);
                        break;
                    case "buy":
                    case "sell":
                        await HandleTradeOrder(ctx, subCmd == "buy");
                        break;
                    case "cancel":
                        await CancelOrder(ctx);
                        break;
                    default:
                        ctx.Reply("📈 积分交易市场指令：\n" +
                                  "/market list - 查看交易对\n" +
                                  "/market buy Q <数量> <价格> - 挂单买入\n" +
                                  "/market sell Q <数量> <价格> - 挂单卖出\n" +
                                  "/market cancel <订单ID> - 撤单\n\n" +
                                  "币种说明：G(通用), Q(本群)");
                        break;
                }
            });

            // 8. 机器人主人特权：直接调整积分 (铸币/销毁)
            _plugin.Command(new[] { "adjust_points", "tzjf", "tiaozhengjifen", "调整积分" }, async ctx => {
                string operatorId = ctx.Event.Payload["from"]?.ToString() ?? "";
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                
                if (!await PrivilegeManager.IsRobotOwner(ctx, operatorId)) {
                    ctx.Reply("❌ 权限不足：只有【机器人主人】可以使用此指令直接干预积分。");
                    return;
                }

                if (ctx.Args.Length < 2) {
                    ctx.Reply("📝 使用方法：/adjust_points <@用户> <增减数量>\n示例：/adjust_points @张三 1000 (给张三加1000分)");
                    return;
                }

                // 解析目标用户 (简单处理，实际应解析 At 消息)
                string targetMention = ctx.Args[0];
                string targetUserId = targetMention.Replace("@", "").Trim(); // 简化处理
                if (!long.TryParse(ctx.Args[1], out long amount)) {
                    ctx.Reply("⚠️ 请输入有效的增减数量。");
                    return;
                }

                // 直接调用 transfer_group Action
                var payload = new Dictionary<string, object> {
                    { "group_id", groupId },
                    { "user_id", targetUserId },
                    { "amount", amount },
                    { "from", operatorId },
                    { "reason", "机器人主人手动调整" }
                };

                await _plugin.EmitAction("transfer_group", payload);
                
                string groupPointsName = await ctx.Session.GetAsync<string>($"table:groups:id:{groupId}:points_name") ?? "本群积分";
                ctx.Reply($"✅ 调整成功！已为用户 {targetUserId} {(amount > 0 ? "增加" : "减少")} {Math.Abs(amount)} {groupPointsName}。\n" +
                          $"💡 提示：此操作不消耗金库预算，直接影响市场流通量。");
            });

            // 9. 打赏积分功能 (/tip)
            _plugin.Command("/tip", async ctx => {
                if (ctx.Args.Length < 2) {
                    ctx.Reply("🎁 打赏积分使用方法：/tip <@用户> <数量>\n💡 提示：打赏将扣除 20% 的系统手续费。");
                    return;
                }

                string operatorId = ctx.Event.Payload["from"]?.ToString() ?? "";
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                string targetMention = ctx.Args[0];
                string targetUserId = targetMention.Replace("@", "").Trim();

                if (!long.TryParse(ctx.Args[1], out long amount) || amount <= 0) {
                    ctx.Reply("⚠️ 请输入有效的打赏数量。");
                    return;
                }

                if (operatorId == targetUserId) {
                    ctx.Reply("❌ 错误：不能打赏给自己。");
                    return;
                }

                bool isGroupActive = await PrivilegeManager.IsGroupModeActive(ctx, groupId);
                string pointsName = "通用积分";
                string pointsKey = $"table:users:id:{operatorId}:global_points";
                string targetPointsKey = $"table:users:id:{targetUserId}:global_points";
                string actionName = "transfer_global";

                if (isGroupActive) {
                    pointsName = await ctx.Session.GetAsync<string>($"table:groups:id:{groupId}:points_name") ?? "本群积分";
                    pointsKey = $"table:group_members:group:{groupId}:user:{operatorId}:points";
                    targetPointsKey = $"table:group_members:group:{groupId}:user:{targetUserId}:points";
                    actionName = "transfer_group";
                }

                // 1. 检查打赏者余额
                long balance = await ctx.Session.GetAsync<long>(pointsKey);
                if (balance < amount) {
                    ctx.Reply($"❌ 余额不足：打赏需要 {amount} {pointsName}，当前余额 {balance}。");
                    return;
                }

                // 2. 检查是否为超级积分用户 (超级积分免手续费)
                bool isSuperUser = await PrivilegeManager.IsSuperPointsUser(ctx, operatorId);
                
                // 3. 计算手续费 (20%，超级积分用户免除)
                long fee = isSuperUser ? 0 : (long)Math.Ceiling(amount * 0.2);
                long netAmount = amount - fee;

                // 4. 执行扣费
                await ctx.Session.SetAsync(pointsKey, balance - amount);

                // 5. 执行到账 (通过 EmitAction 以触发日志记录)
                await _plugin.EmitAction(actionName, new Dictionary<string, object> {
                    { "group_id", groupId },
                    { "user_id", targetUserId },
                    { "amount", netAmount },
                    { "from", operatorId },
                    { "reason", $"来自 {operatorId} 的打赏" }
                });

                string superTip = isSuperUser ? "✨ 您是超级积分用户，已免除手续费！\n" : "";
                ctx.Reply($"✅ 打赏成功！\n" +
                          superTip +
                          $"👤 目标：{targetUserId}\n" +
                          $"💰 总额：{amount} {pointsName}\n" +
                          $"📉 手续费：{fee} {(isSuperUser ? "(0%)" : "(20%)")}\n" +
                          $"🎁 实际到账：{netAmount} {pointsName}");
            });

            // 10. 存积分/取积分
            _plugin.Command(new[] { "deposit", "c", "cun", "存" }, async ctx => {
                await HandleBankOperation(ctx, true);
            });

            _plugin.Command(new[] { "withdraw", "q", "qu", "取" }, async ctx => {
                await HandleBankOperation(ctx, false);
            });

            // 11. 冻结与解冻积分 (仅限机器人主人)
            _plugin.Command(new[] { "freeze", "dj", "dongjie", "冻结" }, async ctx => {
                string operatorId = ctx.Event.Payload["from"]?.ToString() ?? "";
                if (!await PrivilegeManager.IsRobotOwner(ctx, operatorId)) {
                    ctx.Reply("❌ 权限不足：只有【机器人主人】可以冻结积分。");
                    return;
                }

                if (ctx.Args.Length < 2) {
                    ctx.Reply("❄️ 冻结积分使用方法：/freeze <@用户> <数量>");
                    return;
                }

                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                string targetUserId = ctx.Args[0].Replace("@", "").Trim();
                if (!long.TryParse(ctx.Args[1], out long amount)) {
                    ctx.Reply("⚠️ 请输入有效的冻结数量。");
                    return;
                }

                bool isGroupActive = await PrivilegeManager.IsGroupModeActive(ctx, groupId);
                string freezeKey = isGroupActive ? $"frozen:group:{groupId}:user:{targetUserId}" : $"frozen:global:user:{targetUserId}";
                string pointsKey = isGroupActive ? $"table:group_members:group:{groupId}:user:{targetUserId}:points" : $"table:users:id:{targetUserId}:global_points";

                long balance = await ctx.Session.GetAsync<long>(pointsKey);
                if (balance < amount) {
                    ctx.Reply($"⚠️ 该用户余额不足，无法冻结 {amount}。");
                    return;
                }

                await ctx.Session.SetAsync(pointsKey, balance - amount);
                long currentFrozen = await ctx.Session.GetAsync<long>(freezeKey);
                await ctx.Session.SetAsync(freezeKey, currentFrozen + amount);

                ctx.Reply($"✅ 已成功冻结用户 {targetUserId} 的 {amount} {(isGroupActive ? "本群" : "通用")}积分。");
            });

            _plugin.Command(new[] { "unfreeze", "jd", "jiedong", "解冻" }, async ctx => {
                string operatorId = ctx.Event.Payload["from"]?.ToString() ?? "";
                if (!await PrivilegeManager.IsRobotOwner(ctx, operatorId)) {
                    ctx.Reply("❌ 权限不足：只有【机器人主人】可以解冻积分。");
                    return;
                }

                if (ctx.Args.Length < 2) {
                    ctx.Reply("🔥 解冻积分使用方法：/unfreeze <@用户> <数量>");
                    return;
                }

                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                string targetUserId = ctx.Args[0].Replace("@", "").Trim();
                if (!long.TryParse(ctx.Args[1], out long amount)) {
                    ctx.Reply("⚠️ 请输入有效的解冻数量。");
                    return;
                }

                bool isGroupActive = await PrivilegeManager.IsGroupModeActive(ctx, groupId);
                string freezeKey = isGroupActive ? $"frozen:group:{groupId}:user:{targetUserId}" : $"frozen:global:user:{targetUserId}";
                string pointsKey = isGroupActive ? $"table:group_members:group:{groupId}:user:{targetUserId}:points" : $"table:users:id:{targetUserId}:global_points";

                long frozen = await ctx.Session.GetAsync<long>(freezeKey);
                if (frozen < amount) {
                    ctx.Reply($"⚠️ 该用户冻结资产不足，无法解冻 {amount}。");
                    return;
                }

                await ctx.Session.SetAsync(freezeKey, frozen - amount);
                long currentBalance = await ctx.Session.GetAsync<long>(pointsKey);
                await ctx.Session.SetAsync(pointsKey, currentBalance + amount);

                ctx.Reply($"✅ 已成功解冻用户 {targetUserId} 的 {amount} {(isGroupActive ? "本群" : "通用")}积分。");
            });

            // 12. 积分排名功能
            _plugin.Command(new[] { "rank", "ph", "phb", "paihangbang", "排行", "排行榜" }, async ctx => {
                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                bool isGroupActive = await PrivilegeManager.IsGroupModeActive(ctx, groupId);
                string groupPointsName = await ctx.Session.GetAsync<string>($"table:groups:id:{groupId}:points_name") ?? "本群积分";

                // 这里假设 Session.GetAsync 可以处理简单的聚合查询或预存的 Top 列表
                // 在分布式存储中，通常会有定时任务更新排行榜
                string rankType = isGroupActive ? "group" : "global";
                string rankKey = isGroupActive ? $"rank:group:{groupId}" : "rank:global";
                
                // 模拟获取前 10 名 (实际应从数据库查询)
                var topUsers = await ctx.Session.GetAsync<List<Dictionary<string, object>>>(rankKey);
                
                if (topUsers == null || !topUsers.Any()) {
                    ctx.Reply($"📊 暂无 { (isGroupActive ? groupPointsName : "通用积分") } 排名数据，请稍后再试。");
                    return;
                }

                var resp = $"🏆 { (isGroupActive ? groupPointsName : "通用积分") } 财富榜 (TOP 10)\n" +
                           $"------------------\n";
                
                for (int i = 0; i < topUsers.Count; i++) {
                    resp += $"{i + 1}. {topUsers[i]["user_id"]} - {topUsers[i]["points"]} {(isGroupActive ? "Q" : "G")}\n";
                }
                
                ctx.Reply(resp);
            });

            // Console.Error.WriteLine("PointsSystem (Central Bank) with Exchange Market started...");
            // 7. 打赏功能 (/tip, ds, dashang)
            _plugin.Command(new[] { "tip", "ds", "dashang", "打赏" }, async ctx => {
                if (ctx.Args.Length < 2) {
                    ctx.Reply("💡 使用方法：/tip @用户 金额 [留言]\n" +
                              "示例：/tip 123456 100 给你点个赞！\n" +
                              "注意：系统将额外收取 20% 作为手续费。");
                    return;
                }

                string fromUserId = ctx.Event.Payload["from"]?.ToString() ?? "";
                string targetUserId = ctx.Args[0].Replace("@", "").Trim(); // 支持 @123 或 123
                if (!long.TryParse(ctx.Args[1], out long amount) || amount <= 0) {
                    ctx.Reply("❌ 错误：请输入有效的打赏金额（必须大于0）。");
                    return;
                }

                if (fromUserId == targetUserId) {
                    ctx.Reply("❌ 错误：不能给自己打赏。");
                    return;
                }

                string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
                bool isGroupMode = await PrivilegeManager.IsGroupModeActive(ctx, groupId);
                
                // 1. 获取余额
                string pointType = isGroupMode ? "本群积分" : "通用积分";
                string balanceKey = isGroupMode ? 
                    $"table:member_cache:id:{groupId}:{fromUserId}:points" : 
                    $"table:users:id:{fromUserId}:global_points";
                
                long balance = await ctx.Session.GetAsync<long>(balanceKey);
                
                // 超级积分用户免收手续费
                bool isSuperUser = await PrivilegeManager.IsSuperPointsUser(ctx, fromUserId);
                long fee = isSuperUser ? 0 : (long)Math.Ceiling(amount * 0.2);
                long totalRequired = amount + fee;

                if (balance < totalRequired) {
                    string feeMsg = isSuperUser ? "" : $"需要额外支付 {fee} 手续费，";
                    ctx.Reply($"❌ 余额不足：打赏 {amount} {feeMsg}共计 {totalRequired} {pointType}。您当前余额为 {balance}。");
                    return;
                }

                // 2. 执行扣费（打赏者）
                await _plugin.EmitAction("transfer_auto", new Dictionary<string, object> {
                    { "user_id", fromUserId },
                    { "group_id", groupId },
                    { "amount", -totalRequired },
                    { "reason", $"打赏支出 (给 {targetUserId})" }
                });

                // 3. 执行增加（接收者）
                await _plugin.EmitAction("transfer_auto", new Dictionary<string, object> {
                    { "user_id", targetUserId },
                    { "group_id", groupId },
                    { "amount", amount },
                    { "reason", $"收到打赏 (来自 {fromUserId})" }
                });

                string message = ctx.Args.Length > 2 ? string.Join(" ", ctx.Args.Skip(2)) : "给大佬递茶！";
                string feeText = isSuperUser ? "免除 (超级积分用户)" : $"{fee} {pointType}";
                ctx.Reply($"✅ 打赏成功！\n" +
                          $"------------------\n" +
                          $"👤 接收者：{targetUserId}\n" +
                          $"💰 打赏金额：{amount} {pointType}\n" +
                          $"📈 手续费(20%)：{feeText}\n" +
                          $"💬 留言：{message}\n" +
                          $"------------------\n" +
                          $"💡 剩余余额：{balance - totalRequired} {pointType}");
            });

            await _plugin.RunAsync();
        }

        private static async Task ShowMarketOverview(Context ctx)
        {
            string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
            if (string.IsNullOrEmpty(groupId)) {
                ctx.Reply("⚠️ 请在群聊中使用此指令查看本群市场。");
                return;
            }

            string groupPointsName = await ctx.Session.GetAsync<string>($"table:groups:id:{groupId}:points_name") ?? "本群积分";
            var orders = await ctx.Session.GetAsync<List<MarketOrder>>($"market:book:group:{groupId}") ?? new List<MarketOrder>();

            var buys = orders.Where(o => o.Side == "buy").OrderByDescending(o => o.Price).Take(5).ToList();
            var sells = orders.Where(o => o.Side == "sell").OrderBy(o => o.Price).Take(5).ToList();

            var resp = $"📊 {groupPointsName} 交易市场 (对通用积分 G)\n" +
                       "------------------\n" +
                       "🔴 卖盘 (Sell Orders):\n";
            
            if (!sells.Any()) resp += "暂无挂单\n";
            else foreach (var s in sells) resp += $"  {s.Price:F2} G | {s.Amount} {groupPointsName}\n";

            resp += "🟢 买盘 (Buy Orders):\n";
            if (!buys.Any()) resp += "暂无挂单\n";
            else foreach (var b in buys) resp += $"  {b.Price:F2} G | {b.Amount} {groupPointsName}\n";

            resp += "------------------\n" +
                    $"输入 /market buy Q <数量> <价格> 参与竞争。";
            
            ctx.Reply(resp);
        }

        private static async Task HandleBankOperation(Context ctx, bool isDeposit)
        {
            string fromUserId = ctx.Event.Payload["from"]?.ToString() ?? "";
            string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
            
            if (ctx.Args.Length == 0) {
                string text = ctx.Event.Payload.ContainsKey("text") ? ctx.Event.Payload["text"]?.ToString() ?? "" : "";
                if (!text.StartsWith("/")) return; 

                ctx.Reply($"🏦 {(isDeposit ? "存" : "取")}积分使用方法：/{(isDeposit ? "deposit" : "withdraw")} <金额|0|.>\n" +
                          $"💡 0 或 . 表示全部。");
                return;
            }

            string amountStr = ctx.Args[0];
            long amount = 0;
            bool isAll = amountStr == "0" || amountStr == ".";

            bool isGroupMode = await PrivilegeManager.IsGroupModeActive(ctx, groupId);
            string pointType = isGroupMode ? "本群积分" : "通用积分";
            string walletKey = isGroupMode ? 
                $"table:member_cache:id:{groupId}:{fromUserId}:points" : 
                $"table:users:id:{fromUserId}:global_points";
            string bankKey = isGroupMode ? 
                $"table:member_cache:id:{groupId}:{fromUserId}:bank_points" : 
                $"table:users:id:{fromUserId}:bank_points";

            long walletBalance = await ctx.Session.GetAsync<long>(walletKey);
            long bankBalance = await ctx.Session.GetAsync<long>(bankKey);

            if (isAll) {
                amount = isDeposit ? walletBalance : bankBalance;
            } else {
                if (!long.TryParse(amountStr, out amount) || amount <= 0) {
                    string text = ctx.Event.Payload.ContainsKey("text") ? ctx.Event.Payload["text"]?.ToString() ?? "" : "";
                    if (!text.StartsWith("/")) return;
                    
                    ctx.Reply("⚠️ 请输入有效的金额。");
                    return;
                }
            }

            if (amount <= 0) {
                ctx.Reply($"❌ 余额不足，无法{(isDeposit ? "存入" : "取出")}。");
                return;
            }

            if (isDeposit) {
                if (walletBalance < amount) {
                    ctx.Reply($"❌ 余额不足：您只有 {walletBalance} {pointType}。");
                    return;
                }
                await ctx.Session.SetAsync(walletKey, walletBalance - amount);
                await ctx.Session.SetAsync(bankKey, bankBalance + amount);
                ctx.Reply($"✅ 存入成功！\n💰 存入：{amount} {pointType}\n🏦 银行余额：{bankBalance + amount}");
            } else {
                if (bankBalance < amount) {
                    ctx.Reply($"❌ 银行余额不足：您只有 {bankBalance} {pointType} 在银行中。");
                    return;
                }
                await ctx.Session.SetAsync(walletKey, walletBalance + amount);
                await ctx.Session.SetAsync(bankKey, bankBalance - amount);
                ctx.Reply($"✅ 取出成功！\n💰 取出：{amount} {pointType}\n👛 钱包余额：{walletBalance + amount}");
            }
        }

        private static async Task HandleTradeOrder(Context ctx, bool isBuy)
        {
            if (ctx.Args.Length < 4) {
                ctx.Reply($"📝 使用方法：/market {(isBuy ? "buy" : "sell")} Q <数量> <价格>");
                return;
            }

            string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
            if (string.IsNullOrEmpty(groupId)) {
                ctx.Reply("❌ 错误：只能在群聊中进行市场交易。");
                return;
            }

            string userId = ctx.Event.Payload["from"]?.ToString() ?? "";
            long amount = Convert.ToInt64(ctx.Args[2]);
            double price = Convert.ToDouble(ctx.Args[3]);
            string groupPointsName = await ctx.Session.GetAsync<string>($"table:groups:id:{groupId}:points_name") ?? "本群积分";

            if (amount <= 0 || price <= 0) {
                ctx.Reply("⚠️ 数量和价格必须大于 0。");
                return;
            }

            // 1. 资产检查与扣除 (挂单即冻结)
            string globalKey = $"table:users:id:{userId}:global_points";
            string groupKey = $"table:group_members:group:{groupId}:user:{userId}:points";

            if (isBuy) {
                long totalCostG = (long)(amount * price);
                long balanceG = await ctx.Session.GetAsync<long>(globalKey);
                if (balanceG < totalCostG) {
                    ctx.Reply($"❌ 余额不足：挂单需要 {totalCostG} 通用积分，当前余额 {balanceG}。");
                    return;
                }
                await ctx.Session.SetAsync(globalKey, balanceG - totalCostG);
            } else {
                long balanceQ = await ctx.Session.GetAsync<long>(groupKey);
                if (balanceQ < amount) {
                    ctx.Reply($"❌ 余额不足：挂单需要 {amount} {groupPointsName}，当前余额 {balanceQ}。");
                    return;
                }
                await ctx.Session.SetAsync(groupKey, balanceQ - amount);
            }

            // 2. 撮合逻辑
            var order = new MarketOrder { UserId = userId, Side = isBuy ? "buy" : "sell", Amount = amount, Price = price };
            var book = await ctx.Session.GetAsync<List<MarketOrder>>($"market:book:group:{groupId}") ?? new List<MarketOrder>();

            bool fullyMatched = false;
            long remainingAmount = amount;

            var opposites = book.Where(o => o.Side == (isBuy ? "sell" : "buy"))
                                .OrderBy(o => isBuy ? o.Price : -o.Price)
                                .ToList();

            foreach (var opp in opposites) {
                if ((isBuy && opp.Price <= price) || (!isBuy && opp.Price >= price)) {
                    long matchAmount = Math.Min(remainingAmount, opp.Amount);
                    
                    // 执行交易
                    // 买方得到 Q，卖方得到 G
                    string sellerId = isBuy ? opp.UserId : userId;
                    string buyerId = isBuy ? userId : opp.UserId;
                    long totalG = (long)(matchAmount * opp.Price);

                    // 给买方 Q
                    string buyerGroupKey = $"table:group_members:group:{groupId}:user:{buyerId}:points";
                    long currentQ = await ctx.Session.GetAsync<long>(buyerGroupKey);
                    await ctx.Session.SetAsync(buyerGroupKey, currentQ + matchAmount);

                    // 给卖方 G
                    string sellerGlobalKey = $"table:users:id:{sellerId}:global_points";
                    long currentG = await ctx.Session.GetAsync<long>(sellerGlobalKey);
                    await ctx.Session.SetAsync(sellerGlobalKey, currentG + totalG);

                    // 如果是买单，且撮合价低于挂单价，返还差价给买方
                    if (isBuy && price > opp.Price) {
                        long refundG = (long)(matchAmount * (price - opp.Price));
                        long currentBuyerG = await ctx.Session.GetAsync<long>(globalKey);
                        await ctx.Session.SetAsync(globalKey, currentBuyerG + refundG);
                    }

                    opp.Amount -= matchAmount;
                    remainingAmount -= matchAmount;

                    if (opp.Amount <= 0) book.Remove(opp);
                    if (remainingAmount <= 0) {
                        fullyMatched = true;
                        break;
                    }
                }
            }

            if (!fullyMatched) {
                order.Amount = remainingAmount;
                book.Add(order);
                ctx.Reply($"📝 挂单成功！剩余 {remainingAmount} {groupPointsName} 已进入订单簿等待撮合。\n订单 ID: {order.Id}");
            } else {
                ctx.Reply($"✅ 交易成功！您的订单已全部撮合完成。");
            }

            await ctx.Session.SetAsync($"market:book:group:{groupId}", book);
        }

        private static async Task CancelOrder(Context ctx)
        {
            if (ctx.Args.Length < 2) {
                ctx.Reply("📝 使用方法：/market cancel <订单ID>");
                return;
            }

            string orderId = ctx.Args[1];
            string groupId = ctx.Event.Payload.ContainsKey("group_id") ? ctx.Event.Payload["group_id"]?.ToString() ?? "" : "";
            string userId = ctx.Event.Payload["from"]?.ToString() ?? "";

            if (string.IsNullOrEmpty(groupId)) {
                ctx.Reply("❌ 错误：只能在群聊中撤回本群市场订单。");
                return;
            }

            var book = await ctx.Session.GetAsync<List<MarketOrder>>($"market:book:group:{groupId}") ?? new List<MarketOrder>();
            var order = book.FirstOrDefault(o => o.Id == orderId);

            if (order == null) {
                ctx.Reply($"⚠️ 找不到订单 {orderId}，可能已成交或 ID 错误。");
                return;
            }

            if (order.UserId != userId) {
                ctx.Reply("❌ 权限不足：您只能撤回自己的订单。");
                return;
            }

            // 退还资产
            if (order.Side == "buy") {
                long refundG = (long)(order.Amount * order.Price);
                string globalKey = $"table:users:id:{userId}:global_points";
                long currentG = await ctx.Session.GetAsync<long>(globalKey);
                await ctx.Session.SetAsync(globalKey, currentG + refundG);
            } else {
                string groupKey = $"table:group_members:group:{groupId}:user:{userId}:points";
                long currentQ = await ctx.Session.GetAsync<long>(groupKey);
                await ctx.Session.SetAsync(groupKey, currentQ + order.Amount);
            }

            book.Remove(order);
            await ctx.Session.SetAsync($"market:book:group:{groupId}", book);

            ctx.Reply($"✅ 订单 {orderId} 已成功撤回，冻结资产已原路返还。");
        }

        private static bool IsOfficialPlugin(string callerId)
        {
            // 简单实现，正式环境应从配置加载
            return callerId == "com.botmatrix.official.bank" || 
                   callerId == "com.botmatrix.official.mall" || 
                   callerId == "com.botmatrix.official.sign";
        }
    }
}
