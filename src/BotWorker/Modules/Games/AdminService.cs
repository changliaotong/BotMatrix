using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Extensions.Text;
using BotWorker.Modules.Plugins;
using BotWorker.Infrastructure.Communication.OneBot;
using System.Text;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.admin.v2",
        Name = "超级群管",
        Version = "1.1.0",
        Author = "Matrix",
        Description = "全方位的群组管理系统：踢人禁言、黑名单、欢迎语及高级治理功能（刷屏、脏话、广告拦截等）",
        Category = "Admin"
    )]
    public class AdminService : IPlugin
    {
        public List<Intent> Intents => [
            new() { Name = "基础管理", Keywords = ["踢", "禁言", "取消禁言", "设置头衔"] },
            new() { Name = "群组配置", Keywords = ["开机", "关机", "设置欢迎语", "改名提示", "设置管理权限", "设置使用权限"] },
            new() { Name = "名单管理", Keywords = ["拉黑", "取消拉黑", "黑名单", "被踢拉黑", "退群拉黑", "敏感词系统"] },
            new() { Name = "高级治理", Keywords = ["刷屏", "脏话", "广告", "图片", "网址", "推荐群", "推荐好友", "合并转发", "撤回词", "扣分词", "警告词", "禁言词", "踢出词", "拉黑词"] }
        ];

        public async Task StopAsync() => await Task.CompletedTask;

        private IRobot? _robot;

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;
            await EnsureTablesCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "超级群管",
                Commands = [
                    "踢", "禁言", "取消禁言", "设置头衔", 
                    "开机", "关机", "设置欢迎语", "欢迎语", "改名提示",
                    "拉黑", "取消拉黑", "黑名单", "清空黑名单", "被踢拉黑", "退群拉黑", "敏感词系统",
                    "治理设置", "设置", "开启", "关闭", "撤回词", "扣分词", "警告词", "禁言词", "踢出词", "拉黑词"
                ],
                Description = "【超级群管】提供全方位的群组管理及高级治理功能。发送“超级群管 帮助”查看详细指令。"
            }, HandleCommandAsync);
        }

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            // 优先处理会话响应
            if (!string.IsNullOrEmpty(ctx.SessionAction))
            {
                return await HandleMenuAsync(ctx);
            }

            var cmd = ctx.RawMessage.Trim().Split(' ')[0];
            var groupId = long.Parse(ctx.GroupId ?? "0");
            var userId = long.Parse(ctx.UserId ?? "0");
            var botId = long.Parse(ctx.BotId ?? "0");
            var cmdPara = string.Join(" ", args);

            return cmd switch
            {
                // 1. 核心开关逻辑 (复用 GroupInfo)
                "开机" or "关机" => GroupInfo.SetPowerOnOff(botId, groupId, userId, cmd),
                "设置欢迎语" or "欢迎语" => GroupInfo.SetWelcomeMsg(groupId, cmdPara),
                "改名提示" => GroupInfo.SetChangHint(groupId, cmdPara),

                // 2. 自动化策略开关
                "被踢拉黑" or "退群拉黑" or "敏感词系统" => await HandlePolicyToggleAsync(groupId, cmd, cmdPara),

                // 3. 高级治理 (直接复用系统内置的 WarnSetup 逻辑)
                "治理设置" => await HandleMenuAsync(ctx),
                "刷屏" or "脏话" or "广告" or "图片" or "网址" or "推荐群" or "推荐好友" or "合并转发" or "设置" or "开启" or "关闭" => await HandleAdvancedWarnAsync(ctx),
                "撤回词" or "扣分词" or "警告词" or "禁言词" or "踢出词" or "拉黑词" => GroupWarn.GetEditKeyword(groupId, ctx.RawMessage),

                // 4. 帮助指令
                "帮助" => "【超级群管】提供全方位的群组管理功能。\n指令列表：开机/关机、欢迎语、拉黑/取消拉黑、踢/禁言、被踢拉黑等。",

                // 5. 名单管理 (复用 BlackList 和系统内置方法)
                "拉黑" or "取消拉黑" or "黑名单" or "清空黑名单" => await HandleBlacklistAsync(ctx, cmd, args),

                // 4. 成员操作
                "踢" => await HandleKickAsync(ctx, args),
                "禁言" => await HandleMuteAsync(ctx, args, true),
                "取消禁言" => await HandleMuteAsync(ctx, args, false),
                "设置头衔" => await HandleSetTitleAsync(ctx, args),
                
                _ => "未知管理指令"
            };
        }

        private async Task<string> HandleMenuAsync(IPluginContext ctx)
        {
            var userId = ctx.UserId;
            var groupId = ctx.GroupId;
            var input = ctx.RawMessage.Trim();

            // 1. 处理二级菜单响应
            if (ctx.SessionAction == "AdminMenu_Root")
            {
                if (input == "1")
                {
                    await _robot!.Sessions.SetSessionAsync(userId, groupId, "game.admin.v2", "AdminMenu_Spam");
                    return "【刷屏拦截设置】\n1. 开启拦截\n2. 关闭拦截\n3. 设置阈值\n回复数字进行设置，或发送“返回”回到主菜单。";
                }
                else if (input == "2")
                {
                    return "广告拦截设置功能开发中...\n发送“治理设置”重新开始。";
                }
                else if (input == "取消")
                {
                    await _robot!.Sessions.ClearSessionAsync(userId, groupId);
                    return "✅ 已退出菜单。";
                }
            }
            
            if (ctx.SessionAction == "AdminMenu_Spam")
            {
                if (input == "返回")
                {
                    return await StartMainMenuAsync(ctx);
                }
                // 处理刷屏设置逻辑...
                return "✅ 设置成功（示例）。";
            }

            // 2. 初始进入主菜单
            return await StartMainMenuAsync(ctx);
        }

        private async Task<string> StartMainMenuAsync(IPluginContext ctx)
        {
            await _robot!.Sessions.SetSessionAsync(ctx.UserId ?? "0", ctx.GroupId ?? "0", "game.admin.v2", "AdminMenu_Root");
            return "【高级治理设置】\n1. 刷屏拦截设置\n2. 广告拦截设置\n3. 脏话拦截设置\n请回复数字选择要配置的项目，或发送“取消”退出。";
        }

        private async Task<string> HandleAdvancedWarnAsync(IPluginContext ctx)
        {
            // 权限检查：机器人主人或系统管理员
            var botId = long.Parse(ctx.BotId ?? "0");
            var userId = long.Parse(ctx.UserId ?? "0");
            if (botId != userId && !BotInfo.IsAdmin(botId, userId))
            {
                return "❌ 只有机器人主人或系统管理员可以执行此操作。";
            }

            // 获取底层的 BotMessage 实例并调用原有的 GetWarnSetup 逻辑
            if (ctx is PluginContext pctx && pctx.Event is BotMessageEvent botMsgEvent)
            {
                var botMsg = botMsgEvent.BotMessage;
                botMsg.GetWarnSetup(Regexs.WarnCmd);
                return botMsg.Answer;
            }

            return "❌ 系统繁忙，请稍后重试。";
        }

        private async Task<string> HandlePolicyToggleAsync(long groupId, string cmd, string para)
        {
            // 映射到 GroupInfo 的字段名
            string field = cmd switch
            {
                "被踢拉黑" => "IsBlackKick",
                "退群拉黑" => "IsBlackExit",
                "敏感词系统" => "IsWarn",
                _ => ""
            };

            if (string.IsNullOrEmpty(field)) return "指令错误";

            bool? targetStatus = null;
            if (para.In("开启", "打开", "on")) targetStatus = true;
            else if (para.In("关闭", "停用", "off")) targetStatus = false;

            if (targetStatus == null)
            {
                var current = GroupInfo.GetBool(field, groupId);
                return $"📌 {cmd} 当前状态：{(current ? "开启" : "关闭")}\n使用“{cmd} 开启/关闭”来设置。";
            }

            int res = GroupInfo.SetValue(field, targetStatus.Value, groupId);
            return res == -1 ? "❌ 设置失败，请稍后重试" : $"✅ {cmd} 已{(targetStatus.Value ? "开启" : "关闭")}";
        }

        private async Task<string> HandleBlacklistAsync(IPluginContext ctx, string cmd, string[] args)
        {
            if (cmd == "清空黑名单")
            {
                // 权限检查
                var botId = long.Parse(ctx.BotId);
                var userId = long.Parse(ctx.UserId);
                if (botId != userId && !BotInfo.IsAdmin(botId, userId))
                {
                    return "❌ 只有机器人主人或系统管理员可以清空黑名单。";
                }

                if (ctx.RawMessage.Trim() == (string?)ctx.SessionData)
                {
                    int res = BlackList.ClearGroupBlacklist(long.Parse(ctx.GroupId ?? "0"));
                    return res >= 0 ? $"✅ 已成功清空本群黑名单（共影响 {res} 条记录）。" : "❌ 清空失败，请稍后重试。";
                }
                else
                {
                    var code = _robot!.Sessions.GenerateConfirmationCode();
                    await _robot!.Sessions.SetSessionAsync(ctx.UserId ?? "0", ctx.GroupId ?? "0", "game.admin.v2", "ClearBlacklist", null, 60, code);
                    return $"⚠️ 【危险操作确认】\n您正在尝试清空本群所有黑名单记录，此操作不可撤销。\n\n请输入验证码【{code}】以确认执行，或发送“取消”退出。";
                }
            }

            // 复用主程序中已有的逻辑，或者直接调用 BlackList 实体
            var targetId = ctx.RawMessage.ExtractAt();
            if (targetId == 0 && args.Length > 0) long.TryParse(args[0], out targetId);

            if (targetId == 0) return "请输入正确的QQ号或艾特用户。";

            if (cmd == "拉黑")
            {
                int res = BlackList.AddBlackList(
                    long.Parse(ctx.BotId ?? "0"), 
                    long.Parse(ctx.GroupId ?? "0"), 
                    ctx.GroupName ?? string.Empty, 
                    long.Parse(ctx.UserId ?? "0"), 
                    ctx.UserName ?? string.Empty, 
                    targetId, 
                    "管理员手动拉黑");
                return res > 0 ? $"🚫 已将 {targetId} 加入本群黑名单" : "该用户已在黑名单中或操作失败";
            }
            else if (cmd == "取消拉黑")
            {
                int res = BlackList.Delete(long.Parse(ctx.GroupId ?? "0"), targetId);
                return res > 0 ? $"✅ 已将 {targetId} 移出黑名单" : "该用户不在黑名单中或操作失败";
            }
            
            return "黑名单指令错误";
        }

        private async Task EnsureTablesCreatedAsync()
        {
            // AdminService 主要是复用现有的 GroupInfo, BlackList, GroupWarn 等表
            // 这些表通常在系统初始化或各自的实体类中确保创建
            await Task.CompletedTask;
        }

        private async Task<string> HandleKickAsync(IPluginContext ctx, string[] args)
        {
            var targetId = ctx.RawMessage.ExtractAt();
            if (targetId == 0) return "请艾特要踢出的人。";
            
            await _robot!.CallSkillAsync("KickMember", ctx, [targetId.ToString()]);
            return $"✅ 已下达移除指令。";
        }

        private async Task<string> HandleMuteAsync(IPluginContext ctx, string[] args, bool isMute)
        {
            var targetId = ctx.RawMessage.ExtractAt();
            if (targetId == 0) return "请艾特要禁言的人。";
            
            int duration = 600;
            if (args.Length > 0 && int.TryParse(args.Last(), out var d)) duration = d * 60;

            await _robot!.CallSkillAsync("MuteMember", ctx, [isMute ? "Mute" : "Unmute", targetId.ToString(), duration.ToString()]);
            return isMute ? $"🔇 已禁言 {targetId}。" : $"🔊 已解除 {targetId} 禁言。";
        }

        private async Task<string> HandleSetTitleAsync(IPluginContext ctx, string[] args)
        {
            var targetId = ctx.RawMessage.ExtractAt();
            var title = string.Join(" ", args.Where(a => !a.Contains("@")));
            if (targetId == 0) return "请艾特要设置头衔的人。";

            await _robot!.CallSkillAsync("SetMemberTitle", ctx, ["Set", targetId.ToString(), title]);
            return $"✅ 头衔设置指令已发出。";
        }
    }
}
