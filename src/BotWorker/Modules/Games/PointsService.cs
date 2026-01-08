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
        private const string SYSTEM_RESERVE = "SYSTEM_RESERVE"; // 系统发行账户
        private const string SYSTEM_REVENUE = "SYSTEM_REVENUE"; // 系统回收账户

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

            // 自动同步表结构
            await EnsureTablesCreatedAsync();

            // 初始化系统账户
            await EnsureSystemAccountAsync(SYSTEM_RESERVE, "系统积分发行储备", AccountType.SystemReserve);
            await EnsureSystemAccountAsync(SYSTEM_REVENUE, "系统积分回收收益", AccountType.SystemRevenue);

            // 注册指令处理
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "积分财务系统",
                Commands = ["积分", "余额", "balance", "签到", "财务报表"],
                Description = "金融级积分管理：【积分】查询余额；【签到】获取奖励；【财务报表】系统审计"
            }, HandleCommandAsync);

            // 注册跨插件调用接口 (Skill API)
            // 注意：跨插件调用通常使用 object 参数
            await robot.RegisterSkillAsync(new SkillCapability { Name = "points.transfer" }, async (ctx, args) => {
                // 这里是作为指令的回调，但我们也需要它作为 Skill 被调用
                return "Skill: points.transfer registered";
            });
            
            // 为了支持传统的 robot.CallSkillAsync("points.transfer", dict)
            // 我们需要确保 PointsService 实例能被找到并调用其方法
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task EnsureTablesCreatedAsync()
        {
            try
            {
                var checkTable = await PointAccount.QueryScalarAsync<int>("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'PointAccounts'");
                if (checkTable == 0)
                {
                    await PointAccount.ExecAsync(SchemaSynchronizer.GenerateCreateTableSql<PointAccount>());
                }

                var checkLedger = await PointLedger.QueryScalarAsync<int>("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'PointLedgers'");
                if (checkLedger == 0)
                {
                    await PointLedger.ExecAsync(SchemaSynchronizer.GenerateCreateTableSql<PointLedger>());
                }
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "PointsService 数据库初始化失败");
            }
        }

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

        #region 核心账务逻辑 (会计分录)

        public async Task<bool> TransferAsync(string debitId, string creditId, long amount, string description)
        {
            if (amount <= 0) return false;

            try
            {
                var debitAccount = await GetOrCreateAccountAsync(debitId);
                var creditAccount = await GetOrCreateAccountAsync(creditId);

                if (creditAccount.Type != AccountType.SystemReserve && creditAccount.Balance < amount)
                {
                    _logger?.LogWarning($"转账失败：账户 {creditId} 余额不足 ({creditAccount.Balance} < {amount})");
                    return false;
                }

                debitAccount.Balance += amount;
                creditAccount.Balance -= amount;
                debitAccount.LastUpdateTime = DateTime.Now;
                creditAccount.LastUpdateTime = DateTime.Now;

                await debitAccount.UpdateAsync();
                await creditAccount.UpdateAsync();

                var ledger = new PointLedger
                {
                    TransactionId = Guid.NewGuid().ToString("N"),
                    DebitAccountId = debitId,
                    DebitAccountName = debitAccount.AccountName,
                    CreditAccountId = creditId,
                    CreditAccountName = creditAccount.AccountName,
                    Amount = amount,
                    Description = description,
                    TransactionTime = DateTime.Now
                };
                await ledger.InsertAsync();

                // 发布交易事件
                if (_robot != null)
                {
                    _ = _robot.Events.PublishAsync(new PointTransactionEvent
                    {
                        UserId = debitId,
                        AccountType = debitAccount.Type.ToString(),
                        Amount = amount,
                        Description = description,
                        TransactionType = "Income"
                    });

                    _ = _robot.Events.PublishAsync(new PointTransactionEvent
                    {
                        UserId = creditId,
                        AccountType = creditAccount.Type.ToString(),
                        Amount = -amount,
                        Description = description,
                        TransactionType = "Expense"
                    });
                }

                _logger?.LogInformation($"[会计分录] {description}: {creditAccount.AccountName} -> {debitAccount.AccountName} | 金额: {amount}");
                return true;
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "执行会计转账时发生异常");
                return false;
            }
        }

        #endregion

        #region 指令逻辑

        private async Task<string> GetBalanceMsgAsync(IPluginContext ctx)
        {
            var account = await GetOrCreateAccountAsync(ctx.UserId, ctx.UserName);
            return $"💰 您的积分账户：\n余额：{account.Balance}\n账户：{ctx.UserId}";
        }

        private async Task<string> SignMsgAsync(IPluginContext ctx)
        {
            long reward = 100;
            bool success = await TransferAsync(ctx.UserId, SYSTEM_RESERVE, reward, "每日签到奖励");
            
            if (success)
            {
                var account = await GetOrCreateAccountAsync(ctx.UserId);
                return $"✅ 签到成功！\n获得奖励：{reward} 积分\n当前总额：{account.Balance}\n[分录：系统储备 -> 用户账户]";
            }
            return "❌ 签到失败，请稍后再试。";
        }

        private async Task<string> GetSystemReportMsgAsync(IPluginContext ctx)
        {
            var reserve = await PointAccount.GetByAccountIdAsync(SYSTEM_RESERVE);
            var revenue = await PointAccount.GetByAccountIdAsync(SYSTEM_REVENUE);
            
            return $"📊 系统财务简报：\n" +
                   $"----------------\n" +
                   $"积分发行总量：{-(reserve?.Balance ?? 0)}\n" +
                   $"系统回收收益：{revenue?.Balance ?? 0}\n" +
                   $"流通中总量：{(-(reserve?.Balance ?? 0)) - (revenue?.Balance ?? 0)}\n" +
                   $"----------------\n" +
                   $"会计准则：借贷必相等";
        }

        #endregion

        #region 私有辅助方法

        private async Task<PointAccount> GetOrCreateAccountAsync(string accountId, string name = "")
        {
            var account = await PointAccount.GetByAccountIdAsync(accountId);
            if (account == null)
            {
                account = new PointAccount
                {
                    AccountId = accountId,
                    AccountName = string.IsNullOrEmpty(name) ? accountId : name,
                    Type = AccountType.User,
                    Balance = 0
                };
                await account.InsertAsync();
            }
            return account;
        }

        private async Task EnsureSystemAccountAsync(string accountId, string name, AccountType type)
        {
            var account = await PointAccount.GetByAccountIdAsync(accountId);
            if (account == null)
            {
                account = new PointAccount
                {
                    AccountId = accountId,
                    AccountName = name,
                    Type = type,
                    Balance = 0
                };
                await account.InsertAsync();
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
                var account = await PointAccount.GetByAccountIdAsync(userId);
                return account?.Balance ?? 0L;
            }
            return 0L;
        }

        #endregion
    }
}
