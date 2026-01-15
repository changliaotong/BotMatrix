using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models;
using BotWorker.Infrastructure.Utils.Schema;
using Microsoft.Extensions.Logging;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "points",
        Name = "金融级积分账务系统",
        Version = "1.1.0",
        Author = "BotMatrix Financial",
        Description = "采用复式记账法的金融级积分系统，确保有进必有出，借贷必相等。",
        Category = "Financial"
    )]
    public class PointsService : IPlugin
    {
        private readonly ILogger<PointsService>? _logger;
        private IRobot? _robot;
        private const string SYSTEM_RESERVE = "0"; // 系统发行账户 (使用原有数据库保留ID)
        private const string SYSTEM_REVENUE = "1"; // 系统回收账户 (使用原有数据库保留ID)

        private string NormalizeAccountId(string accountId)
        {
            if (accountId == "SYSTEM_RESERVE") return SYSTEM_RESERVE;
            if (accountId == "SYSTEM_REVENUE") return SYSTEM_REVENUE;
            return accountId;
        }

        public PointsService() { }

        public PointsService(ILogger<PointsService> logger)
        {
            _logger = logger;
        }

        public List<Intent> Intents => [
            new() { Name = "积分查询", Keywords = ["积分", "余额", "balance"] },
            new() { Name = "签到", Keywords = ["签到", "sign"] },
            new() { Name = "财务报表", Keywords = ["财务报表", "报表", "账单"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;

            // 积分系统不再新建表，直接使用原有 User 和 Credit 表
            // await EnsureTablesCreatedAsync(); 

            // 初始化系统账户 (确保 User 表中有这些记录)
            await EnsureSystemAccountAsync(SYSTEM_RESERVE, "系统积分发行储备");
            await EnsureSystemAccountAsync(SYSTEM_REVENUE, "系统积分回收收益");

            // 注册指令处理
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "积分财务系统",
                Commands = ["积分", "余额", "balance", "签到", "财务报表"],
                Description = "金融级积分管理：【积分】查询余额；【签到】获取奖励；【财务报表】系统审计"
            }, HandleCommandAsync);

            // 注册跨插件调用接口 (Skill API)
            await robot.RegisterSkillAsync(new SkillCapability { Name = "points.transfer" }, async (ctx, args) => {
                if (args == null || args.Length < 3) return "❌ 错误：缺少转账参数。格式：[FromId, ToId, Amount, Reason]";
                
                string fromId = args[0];
                string toId = args[1];
                if (!long.TryParse(args[2], out long amount)) return "❌ 错误：金额格式不正确。";
                string reason = args.Length > 3 ? args[3] : "系统调用";

                if (string.IsNullOrEmpty(fromId) || string.IsNullOrEmpty(toId) || amount <= 0)
                {
                    return "❌ 错误：转账参数 incomplete 或金额错误。";
                }

                // 执行转账逻辑 (贷记 fromId, 借记 toId)
                bool success = await TransferAsync(toId, fromId, amount, reason, ctx);
                return success ? "✅ 转账成功" : "❌ 转账失败：余额不足或系统错误";
            });
        }

        public Task StopAsync() => Task.CompletedTask;

        private Task EnsureTablesCreatedAsync() => Task.CompletedTask;

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim().Split(' ')[0].TrimStart('!', '！', '/', ' ');
            
            return cmd switch
            {
                "积分" or "余额" or "balance" => await GetBalanceMsgAsync(ctx),
                "签到" => await SignMsgAsync(ctx),
                "财务报表" => await GetSystemReportMsgAsync(ctx),
                _ => "未知指令"
            };
        }

        #region 核心账务逻辑 (对接原有数据库)

        public async Task<bool> TransferAsync(string debitId, string creditId, long amount, string description, IPluginContext? ctx = null)
        {
            if (amount <= 0) return false;

            try
            {
                debitId = NormalizeAccountId(debitId);
                creditId = NormalizeAccountId(creditId);

                long debitQQ = long.Parse(debitId);
                long creditQQ = long.Parse(creditId);

                long botUin = ctx != null ? long.Parse(ctx.BotId) : 0;
                long groupId = ctx != null && !string.IsNullOrEmpty(ctx.GroupId) ? long.Parse(ctx.GroupId) : 0;
                string groupName = ctx?.GroupName ?? "系统";
                string debitName = (debitId == SYSTEM_RESERVE || debitId == SYSTEM_REVENUE) ? "系统账户" : (ctx?.UserName ?? debitId);
                string creditName = (creditId == SYSTEM_RESERVE || creditId == SYSTEM_REVENUE) ? "系统账户" : (ctx?.UserName ?? creditId);

                // 1. 检查付款方余额 (系统发行方除外)
                if (creditId != SYSTEM_RESERVE)
                {                    
                    long currentBalance = await UserInfo.GetCreditAsync(botUin, groupId, creditQQ);
                    if (currentBalance < amount)
                    {
                        _logger?.LogWarning($"转账失败：账户 {creditId} 余额不足 ({currentBalance} < {amount})");
                        return false;
                    }
                }

                // 2. 使用原有事务逻辑执行转账
                var result = await UserInfo.TransferCreditAsync(
                    botUin, groupId, groupName,
                    creditQQ, creditName,
                    debitQQ, debitName,
                    amount, amount, description);

                if (result.Result != 0) return false;

                // 3. 发布交易事件 (保持新系统的事件能力)
                if (_robot != null)
                {
                    if (amount >= 1000)
                    {
                        _ = _robot.Events.PublishAsync(new SystemAuditEvent {
                            Level = "Warning",
                            Source = "Points",
                            Message = $"检测到大额交易: {creditName} -> {debitName} | 金额: {amount}",
                            TargetUser = debitId
                        });
                    }

                    _ = _robot.Events.PublishAsync(new PointTransactionEvent
                    {
                        UserId = debitId,
                        AccountType = (debitId == SYSTEM_RESERVE || debitId == SYSTEM_REVENUE) ? "System" : "User",
                        Amount = amount,
                        Description = description,
                        TransactionType = "Income"
                    });

                    _ = _robot.Events.PublishAsync(new PointTransactionEvent
                    {
                        UserId = creditId,
                        AccountType = (creditId == SYSTEM_RESERVE || creditId == SYSTEM_REVENUE) ? "System" : "User",
                        Amount = -amount,
                        Description = description,
                        TransactionType = "Expense"
                    });
                }

                _logger?.LogInformation($"[原有库转账] {description}: {creditName} -> {debitName} | 金额: {amount}");
                return true;
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "执行原有库转账时发生异常");
                return false;
            }
        }

        #endregion

        #region 指令逻辑

        private async Task<string> GetBalanceMsgAsync(IPluginContext ctx)
        {
            return "🏅 积分总览\n💎 {积分类型}：{积分} \n🏦 已存积分：{已存积分}\n📈 积分总额：{积分总额}\n🌐 全球排名：第{积分总排名}名 ✨";
        }

        private async Task<string> SignMsgAsync(IPluginContext ctx)
        {
            // 获取用户等级以计算加成
            var userLevel = await UserLevel.GetByUserIdAsync(ctx.UserId);
            int level = userLevel?.Level ?? 1;
            
            long baseReward = 100;
            double multiplier = 1.0 + (level * 0.02);
            double globalBuff = _robot?.Events.GetActiveBuff(BuffType.PointsMultiplier) ?? 1.0;
            long finalReward = (long)(baseReward * multiplier * globalBuff);

            bool success = await TransferAsync(ctx.UserId, SYSTEM_RESERVE, finalReward, $"每日签到奖励 (等级加成 x{multiplier:F2}, 全服 Buff x{globalBuff:F2})", ctx);
            
            if (success)
            {
                long groupId = !string.IsNullOrEmpty(ctx.GroupId) ? long.Parse(ctx.GroupId) : 0;
                long botUin = long.Parse(ctx.BotId);
                long balance = await UserInfo.GetCreditAsync(botUin, groupId, long.Parse(ctx.UserId));
                string planeInfo = userLevel != null ? $" [{GetPlaneName(level)}]" : "";
                string buffNotice = globalBuff > 1.0 ? $"🔥 全服翻倍 x{globalBuff:F1}\n" : "";
                return $"✅ 签到成功！\n" +
                       $"{buffNotice}" +
                       $"您的等级：Lv.{level}{planeInfo}\n" +
                       $"获得奖励：{finalReward} 积分 (含 {((multiplier * globalBuff - 1) * 100):F0}% 复合加成)\n" +
                       $"当前总额：{balance}";
            }
            return "签到失败，请稍后再试。";
        }

        private string GetPlaneName(int level)
        {
            if (level < 10) return "原质";
            if (level < 30) return "构件";
            if (level < 60) return "逻辑";
            if (level < 90) return "协议";
            if (level < 120) return "矩阵";
            return "奇点";
        }

        private async Task<string> GetSystemReportMsgAsync(IPluginContext ctx)
        {
            long botUin = long.Parse(ctx.BotId);
            long reserveBalance = await UserInfo.GetCreditAsync(botUin, 0, long.Parse(SYSTEM_RESERVE));
            long revenueBalance = await UserInfo.GetCreditAsync(botUin, 0, long.Parse(SYSTEM_REVENUE));
            
            return $"📊 系统财务简报 (原有数据库)：\n" +
                   $"----------------\n" +
                   $"积分发行总量：{-reserveBalance}\n" +
                   $"系统回收收益：{revenueBalance}\n" +
                   $"流通中总量：{(-reserveBalance) - revenueBalance}\n" +
                   $"----------------\n" +
                   $"会计准则：借贷必相等";
        }

        #endregion

        #region 私有辅助方法

        private async Task EnsureSystemAccountAsync(string accountId, string name)
        {
            long qq = long.Parse(accountId);
            if (!await UserInfo.ExistsAsync(qq))
            {
                var user = new UserInfo
                {
                    Id = qq,
                    Name = name,
                    Credit = 0,
                    InsertDate = DateTime.Now
                };
                await user.InsertAsync();
            }
        }

        // 跨插件调用接口 (通过 IRobot 注册的逻辑需要符合特定的委托签名)
        public async Task<object> TransferSkillAsync(object args)
        {
            if (args is Dictionary<string, object> dict &&
                dict.TryGetValue("to", out var to) &&
                dict.TryGetValue("from", out var from) &&
                dict.TryGetValue("amount", out var amountObj) &&
                dict.TryGetValue("desc", out var desc))
            {
                long amount = Convert.ToInt64(amountObj);
                return await TransferAsync(to.ToString()!, from.ToString()!, amount, desc.ToString()!);
            }
            return false;
        }

        public async Task<object> GetBalanceSkillAsync(object args)
        {
            if (args is string userId)
            {
                return await UserInfo.GetCreditAsync(long.Parse(userId));
            }
            return 0L;
        }

        #endregion
    }
}
