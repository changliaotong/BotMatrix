using System.Text.RegularExpressions;
using BotWorker.Modules.Office;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Infrastructure.Persistence.Database;
using BotWorker.Infrastructure.Utils;
using BotWorker.Domain.Entities;
using BotWorker.Common;

namespace BotWorker.Application.Services
{
    public interface IUserService
    {
        Task<string> HandleBlacklistAsync(BotMessage botMsg);
        string GetSaveCreditRes(BotMessage botMsg);
        string GetRewardCredit(BotMessage botMsg);
        string GetCreditList(BotMessage botMsg, long top = 10);
        string GetSellCredit(BotMessage botMsg);
        Task<string> HandleSaveCreditAsync(BotMessage botMsg);
        Task<string> HandleRewardCreditAsync(BotMessage botMsg);
        Task<string> GetCreditRankAsync(BotMessage botMsg);
        Task<string> ExchangeCoinsAsync(BotMessage botMsg);
        Task<string> ExchangeCoinsAsync(BotMessage botMsg, string cmdPara, string cmdPara2);
    }

    public class UserService : IUserService
    {
        private readonly IBotApiService _apiService;

        public UserService(IBotApiService apiService)
        {
            _apiService = apiService;
        }

        #region 黑名单逻辑 (复刻自 BlackMessage.cs)

        public async Task<string> HandleBlacklistAsync(BotMessage botMsg)
        {
            botMsg.IsCancelProxy = true;

            if (botMsg.CmdName == "清空黑名单")
                return await GetClearBlackAsync(botMsg);

            if (botMsg.CmdPara.IsNull())
                return GetGroupBlackList(botMsg);

            //一次加多个号码进入黑名单
            string res = "";
            var cmdName = botMsg.CmdName.Replace("解除", "取消").Replace("删除", "取消");
            foreach (Match match in botMsg.CmdPara.Matches(Regexs.Users))
            {
                long blackUserId = match.Groups["UserId"].Value.AsLong();
                if (cmdName == "拉黑")
                {
                    res += GetAddBlack(botMsg, blackUserId);
                    await _apiService.KickMemberAsync(botMsg.SelfId, botMsg.GroupId, blackUserId);
                }
                else if (cmdName == "取消拉黑")
                    res += GetCancelBlack(botMsg, blackUserId);
            }
            return res;
        }

        private string GetGroupBlackList(BotMessage botMsg)
        {
            return SQLConn.QueryRes($"SELECT {MetaData.SqlTop(10)} BlackId FROM {BlackList.FullName} WHERE GroupId = {botMsg.GroupId} ORDER BY Id DESC {MetaData.SqlLimit(10)}",
                            "{i} {0}\n") +
                   "已拉黑人数：" + BlackList.CountWhere($"GroupId = {botMsg.GroupId}") +
                   "\n拉黑 + QQ\n删黑 + QQ";
        }

        private async Task<string> GetClearBlackAsync(BotMessage botMsg)
        {
            if (!botMsg.IsRobotOwner())
                return C.OwnerOnlyMsg;

            long blackCount = BlackList.CountKey2(botMsg.GroupId.ToString());
            if (blackCount == 0)
                return "黑名单已为空，无需清空";

            if (!botMsg.IsConfirm && blackCount > 10)
                return await botMsg.ConfirmMessage($"清空黑名单 人数{blackCount}");

            return BlackList.DeleteAll(botMsg.GroupId) == -1
                ? C.RetryMsg
                : "✅ 黑名单已清空";
        }

        private string GetAddBlack(BotMessage botMsg, long qqBlack)
        {
            string res = "";

            //加入黑名单
            if (BlackList.Exists(botMsg.GroupId, qqBlack))
                return $"[@:{qqBlack}] 已被拉黑，无需再次加入\n";

            if (qqBlack == botMsg.UserId)
                return "不能拉黑你自己";

            if (BotInfo.IsRobot(qqBlack))
                return "不能拉黑机器人";

            if (botMsg.Group.RobotOwner == qqBlack)
                return "不能拉黑我主人";

            if (WhiteList.Exists(botMsg.GroupId, qqBlack))
            {
                if (botMsg.Group.RobotOwner != botMsg.UserId && !BotInfo.IsAdmin(botMsg.SelfId, botMsg.UserId))
                    return $"您无权拉黑白名单成员";
                res += WhiteList.Delete(botMsg.GroupId, qqBlack) == -1
                    ? $"未能将[@:{qqBlack}]从白名单删除"
                    : $"✅ 已将[@:{qqBlack}]从白名单删除！\n";
            }
            res += BlackList.AddBlackList(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name, qqBlack, "") == -1
                ? $"[@:{qqBlack}]{C.RetryMsg}"
                : $"✅ 已拉黑！";
            return res;
        }

        private string GetCancelBlack(BotMessage botMsg, long userId)
        {
            string res;

            if (BlackList.Exists(botMsg.GroupId, userId))
                res = BlackList.Delete(botMsg.GroupId, userId) == -1
                    ? $"[@:{userId}]{C.RetryMsg}\n"
                    : $"[@:{userId}]已解除拉黑\n";
            else
                res = $"[@:{userId}]不在黑名单，无需解除\n";

            if (BlackList.IsSystemBlack(userId))
                res += $"[@:{userId}]已被列入官方黑名单\n";
            return res;
        }

        #endregion

        #region 积分逻辑 (复刻自 CreditMessage.cs)

        public async Task<string> GetSaveCreditResAsync(BotMessage botMsg)
        {
            botMsg.IsCancelProxy = true;

            if (!botMsg.Group.IsCreditSystem)
                return C.CreditSystemClosed;

            if (botMsg.CmdPara == "")
                return "格式:存分 + 积分数\n取分 + 积分数\n例如：存分 100";

            if (!botMsg.CmdPara.IsNum())
                return "参数不正确";

            long credit_oper = botMsg.CmdPara.AsLong();
            var cmdName = botMsg.CmdName.ToLower();
            if (cmdName.StartsWith('存') | cmdName.StartsWith('c'))
                cmdName = "存分";

            if (cmdName.StartsWith('取') | cmdName.StartsWith('q'))
                cmdName = "取分";

            string res = "";

            if (cmdName == "存分")
            {
                credit_oper = credit_oper == 0 ? await UserInfo.GetCreditAsync(botMsg.SelfId, botMsg.GroupId, botMsg.UserId) : credit_oper;
                if (credit_oper == 0)
                    return "您没有积分可存";

                var saveRes = await DoSaveCreditAsync(botMsg, credit_oper);
                res = saveRes.Res;
            }
            else if (cmdName == "取分")
            {
                credit_oper = credit_oper == 0 ? await UserInfo.GetSaveCreditAsync(botMsg.SelfId, botMsg.GroupId, botMsg.UserId) : credit_oper;
                if (credit_oper == 0)
                    return "您没有积分可取";

                var saveRes = await DoSaveCreditAsync(botMsg, -credit_oper);
                res = saveRes.Res;
            }
            return res;
        }

        public string GetSaveCreditRes(BotMessage botMsg)
        {
            return GetSaveCreditResAsync(botMsg).GetAwaiter().GetResult();
        }

        private (int Result, long CreditValue, long CreditSave, string Res) DoSaveCredit(BotMessage botMsg, long creditOper)
        {
            return DoSaveCreditAsync(botMsg, creditOper).GetAwaiter().GetResult();
        }

        private async Task<(int Result, long CreditValue, long CreditSave, string Res)> DoSaveCreditAsync(BotMessage botMsg, long creditOper)
        {
            long creditValue = await UserInfo.GetCreditAsync(botMsg.SelfId, botMsg.GroupId, botMsg.UserId);
            long creditSave = await UserInfo.GetSaveCreditAsync(botMsg.SelfId, botMsg.GroupId, botMsg.UserId);
            long credit_oper2 = creditOper;
            string cmdName = "存分";
            string res = "";
            if (creditOper > 0)
            {
                if (creditValue < credit_oper2)
                {
                    res = $"您只有{creditValue:N0}分";
                    return (-1, creditValue, creditSave, res);
                }
            }
            else
            {
                credit_oper2 = -creditOper;
                if (creditSave < credit_oper2)
                {
                    res = $"您已存分只有{creditSave:N0}";
                    return (-1, creditValue, creditSave, res);
                }
                cmdName = "取分";
            }

            using var trans = await UserInfo.BeginTransactionAsync();
            try
            {
                // 1. 记录日志 (自动支持事务)
                await CreditLog.AddLogAsync(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name, -creditOper, cmdName, trans);

                // 2. 更新存分 (自动支持事务)
                var (sql, paras) = UserInfo.SqlSaveCredit(botMsg.SelfId, botMsg.GroupId, botMsg.UserId, creditOper);
                await UserInfo.ExecAsync(sql, trans, paras);

                await trans.CommitAsync();

                creditSave += creditOper;
                creditValue -= creditOper;

                // 同步缓存
                UserInfo.SyncCacheField(botMsg.UserId, botMsg.GroupId, "Credit", creditValue);
                UserInfo.SyncCacheField(botMsg.UserId, botMsg.GroupId, "SaveCredit", creditSave);

                res = $"✅ {cmdName}：{credit_oper2}\n" +
                    $"💰 {{积分类型}}：{{积分}}\n" +
                    $"🏦 已存积分：{{已存存分}}\n" +
                    $"📈 积分总额：{{积分总额}}";

                return (0, creditValue, creditSave, res);
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[DoSaveCredit Error] {ex.Message}");
                res = C.RetryMsg;
                return (-1, creditValue, creditSave, res);
            }
        }

        public async Task<string> GetRewardCreditAsync(BotMessage botMsg)
        {
            botMsg.IsCancelProxy = true;

            if (!botMsg.Group.IsCreditSystem)
                return C.CreditSystemClosed;

            string regex_reward;
            if (botMsg.CmdPara.IsMatch(Regexs.CreditParaAt))
                regex_reward = Regexs.CreditParaAt;
            else if (botMsg.CmdPara.IsMatch(Regexs.CreditParaAt2))
                regex_reward = Regexs.CreditParaAt2;
            else if (botMsg.CmdPara.IsMatch(Regexs.CreditPara))
                regex_reward = Regexs.CreditPara;
            else
                return $"🎉 打赏格式：\n打赏 [QQ号] [积分]\n📌 例如：\n打赏 51437810 100";
            long rewardQQ = botMsg.CmdPara.RegexGetValue(regex_reward, "UserId").AsLong();
            long rewardCredit = botMsg.CmdPara.RegexGetValue(regex_reward, "credit").AsLong();

            if (rewardCredit < 10)
                return "至少打赏10分";

            long creditMinus = rewardCredit * 12 / 10;
            bool isSell = botMsg.UserId.In(BotInfo.AdminUin, BotInfo.AdminUin2) && (botMsg.GroupId == 0 || botMsg.IsPublic);

            bool isSuper = botMsg.User.IsSuper;
            bool isPartner = Partner.IsPartner(botMsg.UserId);
            if (isSuper || isPartner)
                creditMinus = rewardCredit;

            long creditValue = UserInfo.GetCredit(botMsg.GroupId, botMsg.UserId);
            if (creditValue < creditMinus && !isSell)
                return $"您的积分{creditValue:N0}不足{creditMinus:N0}。";

            long creditValue2 = UserInfo.GetCredit(botMsg.GroupId, rewardQQ);
            int i;
            if (isSell)
            {
                var addRes = await UserInfo.AddCreditAsync(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, rewardQQ, "", rewardCredit, $"打赏加分:{botMsg.UserId}");
                i = addRes.Result;
                creditValue2 = addRes.CreditValue;
            }
            else if (botMsg.Group.IsCredit)
            {
                var res = await GroupMember.TransferCoinsAsync(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name, rewardQQ, "", (int)CoinsLog.CoinsType.groupCredit, creditMinus, rewardCredit, "打赏");
                i = res.Result;
                creditValue = res.SenderCoins;
                creditValue2 = res.ReceiverCoins;
            }
            else
            {
                var res = await UserInfo.TransferCreditAsync(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name, rewardQQ, "", creditMinus, rewardCredit, "打赏");
                i = res.Result;
                creditValue = res.SenderCredit;
                creditValue2 = res.ReceiverCredit;
            }

            string transferFee = isPartner || isSuper ? "" : $"\n💸 服务费：{rewardCredit * 2 / 10:N0}";

            return i == -1
                ? C.RetryMsg
                : $"✅ 打赏成功！\n🎉 打赏积分：{rewardCredit:N0}{transferFee:N0}\n🎯 对方积分：{creditValue2:N0}\n🙋 您的积分：{creditValue:N0}";
        }

        public string GetRewardCredit(BotMessage botMsg)
        {
            return GetRewardCreditAsync(botMsg).GetAwaiter().GetResult();
        }

        public string GetCreditList(BotMessage botMsg, long top = 10)
        {
            var format = !botMsg.IsRealProxy && (botMsg.IsMirai || botMsg.IsQQ) ? "第{i}名[@:{0}] 💎{1:N0}\n" : "第{i}名{0} 💎{1:N0}\n";
            string res = botMsg.Group.IsCredit
                ? GroupMember.QueryWhere($"{MetaData.SqlTop(top)} UserId, GroupCredit", $"groupId = {botMsg.GroupId}", $"GroupCredit desc {MetaData.SqlLimit(top)}", format)
                : botMsg.SelfInfo.IsCredit
                    ? Friend.QueryWhere($"{MetaData.SqlTop(top)} UserId, credit", $"UserId in (select UserId from {GroupMember.FullName} where GroupId = {botMsg.GroupId})",
                                        $"credit desc {MetaData.SqlLimit(top)}", format)
                    : UserInfo.QueryWhere($"{MetaData.SqlTop(top)} Id, Credit", $"Id in (select UserId from {CreditLog.FullName} where GroupId = {botMsg.GroupId})",
                                 $"credit desc {MetaData.SqlLimit(top)}", format);
            if (!res.Contains(botMsg.UserId.ToString()))
                res += $"{{积分排名}} [@:{botMsg.UserId}] 💎{{积分}}\n";
            
            res = ReplaceRankWithIcon(res);
            
            // 替换占位符
            res = res.Replace("{积分}", UserInfo.GetCredit(botMsg.GroupId, botMsg.UserId).ToString("N0"));
            
            return $"🏆 积分排行榜\n{res}";
        }

        private static string ReplaceRankWithIcon(string text)
        {
            return text.RegexReplace(@"第(\d+)名", match =>
            {
                int rank = int.Parse(match.Groups[1].Value);
                string icon = rank switch
                {
                    1 => "🥇",
                    2 => "🥈",
                    3 => "🥉",
                    4 => "4️⃣",
                    5 => "5️⃣",
                    6 => "6️⃣",
                    7 => "7️⃣",
                    8 => "8️⃣",
                    9 => "9️⃣",
                    10 => "🔟",
                    _ => ""
                };
                return icon;
            });
        }

        public string GetSellCredit(BotMessage botMsg)
        {
            botMsg.IsCancelProxy = true;

            if (!botMsg.Group.IsCreditSystem)
                return C.CreditSystemClosed;

            if (botMsg.CmdPara == "")
                return "📄 命令格式：卖分 + 数值\n📌 使用示例：卖分 1000\n💎 超级积分：10,000→4R\n🎁 普通积分：10,000→1R\n📦 您的{积分类型}：{积分}";

            if (BotInfo.GetIsCredit(botMsg.SelfId))
                return "本机积分不能兑换余额";

            if (GroupInfo.GetIsCredit(botMsg.GroupId))
                return "本群积分不能兑换余额";

            if (!botMsg.CmdPara.IsNum())
                return "数量不正确！";

            long creditMinus = botMsg.CmdPara.AsLong();
            if (creditMinus < 1000)
                return "至少需要1000分";

            long creditValue = UserInfo.GetCredit(botMsg.GroupId, botMsg.UserId);
            if (creditValue < creditMinus)
                return $"您只有{creditValue}分";

            return "您无权使用此命令";
        }

        public async Task<string> HandleSaveCreditAsync(BotMessage botMsg)
        {
            return await GetSaveCreditResAsync(botMsg);
        }

        public async Task<string> HandleRewardCreditAsync(BotMessage botMsg)
        {
            return await GetRewardCreditAsync(botMsg);
        }

        public async Task<string> GetCreditRankAsync(BotMessage botMsg)
        {
            return await Task.Run(() => GetCreditList(botMsg));
        }

        public async Task<string> ExchangeCoinsAsync(BotMessage botMsg)
        {
            // 这里可以根据 CmdPara 决定逻辑
            return "暂不支持";
        }

        public async Task<string> ExchangeCoinsAsync(BotMessage botMsg, string cmdPara, string cmdPara2)
        {
            botMsg.CmdPara = $"{cmdPara} {cmdPara2}";
            return await ExchangeCoinsAsync(botMsg);
        }

        #endregion
    }
}
