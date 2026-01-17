using System.Data;

namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        //卖出积分
        public async Task<string> GetSellCreditAsync()
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (string.IsNullOrEmpty(CmdPara))
                return "📄 命令格式：卖分 + 数值\n📌 使用示例：卖分 1000\n💎 超级积分：10,000→4R\n🎁 普通积分：10,000→1R\n📦 您的{积分类型}：{积分}";

            if (await BotInfo.GetIsCreditAsync(SelfId))
                return "本机积分不能兑换余额";

            if (await GroupInfo.GetIsCreditAsync(GroupId))
                return "本群积分不能兑换余额";

            if (!CmdPara.IsNum())
                return "数量不正确！";

            long creditMinus = CmdPara.AsLong();
            if (creditMinus < 1000)
                return "至少需要1000分";

            long creditValue = await UserService.GetCreditAsync(SelfId, GroupId, UserId);
            if (creditValue < creditMinus)
                return $"您只有{creditValue:N0}分";

            return $"✅ 卖出成功！\n💎 {{积分类型}}：-{creditMinus:N0}→{creditValue - creditMinus:N0}\n💳 余额：...";
        }

        public async Task<string> GetSaveCreditResAsync()
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (string.IsNullOrEmpty(CmdPara))
                return "格式:存分 + 积分数\n取分 + 积分数\n例如：存分 100";

            if (!CmdPara.IsNum())
                return "参数不正确";

            long credit_oper = CmdPara.AsLong();
            string originalCmdName = CmdName;
            CmdName = CmdName.ToLower();
            if (CmdName.StartsWith("存") || CmdName.StartsWith("c"))
                CmdName = "存分";
            else if (CmdName.StartsWith("取") || CmdName.StartsWith("q"))
                CmdName = "取分";

            string res = "";

            if (CmdName == "存分")
            {
                credit_oper = credit_oper == 0 ? await UserService.GetCreditAsync(SelfId, GroupId, UserId) : credit_oper;
                if (credit_oper == 0)
                    return "您没有积分可存";

                var saveRes = await DoSaveCreditAsync(credit_oper);
                res = saveRes.Res;
            }
            else if (CmdName == "取分")
            {
                credit_oper = credit_oper == 0 ? await UserService.GetSaveCreditAsync(SelfId, GroupId, UserId) : credit_oper;
                if (credit_oper == 0)
                    return "您没有积分可取";

                var saveRes = await DoSaveCreditAsync(-credit_oper);
                res = saveRes.Res;
            }
            else
            {
                // 如果 CmdName 不是存分或取分，但匹配了正则（可能是因为 regex 比较宽泛），则尝试根据 originalCmdName 再次判断
                if (originalCmdName.Contains("取"))
                {
                    var saveRes = await DoSaveCreditAsync(-credit_oper);
                    res = saveRes.Res;
                }
                else if (originalCmdName.Contains("存"))
                {
                    var saveRes = await DoSaveCreditAsync(credit_oper);
                    res = saveRes.Res;
                }
            }
            return res;
        }

        //存取分 (异步重构版)
        public async Task<(int Result, long CreditValue, long CreditSave, string Res)> DoSaveCreditAsync(long creditOper)
        {
            var res = await UserService.SaveCreditAsync(SelfId, GroupId, GroupName, UserId, Name, creditOper);
            
            if (res.Result == -2)
                return (-1, res.CreditValue, res.SaveCreditValue, $"您只有{res.CreditValue:N0}分");
            if (res.Result == -3)
                return (-1, res.CreditValue, res.SaveCreditValue, $"您已存分只有{res.SaveCreditValue:N0}");
            if (res.Result == -1)
                return (-1, 0, 0, RetryMsg);

            string cmdName = creditOper > 0 ? "存分" : "取分";
            long absOper = Math.Abs(creditOper);

            string response = $"✅ {cmdName}：{absOper:N0}\n" +
                $"💰 {{积分类型}}：{res.CreditValue:N0}\n" +
                $"🏦 已存积分：{res.SaveCreditValue:N0}\n" +
                $"📈 积分总额：{res.CreditValue + res.SaveCreditValue:N0}";
            
            return (0, res.CreditValue, res.SaveCreditValue, response);
        } 

        public async Task<(int Result, long CreditValue)> AddCreditAsync(long creditAdd, string creditInfo, IDbTransaction? trans = null)
        {
            if (trans != null)
            {
                var res = await UserService.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, creditAdd, creditInfo, trans);
                return (res.Result, res.CreditValue);
            }
            else
            {
                var res = await UserService.AddCreditTransAsync(SelfId, GroupId, GroupName, UserId, Name, creditAdd, creditInfo);
                return (res.Result, res.CreditValue);
            }
        }

        public async Task<(int Result, long CreditValue)> MinusCreditAsync(long creditMinus, string creditInfo, IDbTransaction? trans = null)
        {
            return await AddCreditAsync(-creditMinus, creditInfo, trans);
        }

        //打赏
        public async Task<string> GetRewardCreditAsync()
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            string regex_reward;
            if (CmdPara.IsMatch(Regexs.CreditParaAt))
                regex_reward = Regexs.CreditParaAt;
            else if (CmdPara.IsMatch(Regexs.CreditParaAt2))
                regex_reward = Regexs.CreditParaAt2;
            else if (CmdPara.IsMatch(Regexs.CreditPara))
                regex_reward = Regexs.CreditPara;
            else
                return $"🎉 打赏格式：\n打赏 [QQ号] [积分]\n📌 例如：\n打赏 51437810 100";
            long rewardQQ = CmdPara.RegexGetValue(regex_reward, "UserId").AsLong();
            long rewardCredit = CmdPara.RegexGetValue(regex_reward, "credit").AsLong();

            if (rewardCredit < 10)
                return "至少打赏10分";

            long creditMinus = rewardCredit * 12 / 10;
            bool isSell = UserId.In(BotInfo.AdminUin, BotInfo.AdminUin2) && (GroupId == 0 || IsPublic);

            bool isSuper = User.IsSuper;
            bool isPartner = Partner.IsPartner(UserId);
            if (isSuper || isPartner)
                creditMinus = rewardCredit;

            long senderCredit = await UserService.GetCreditAsync(SelfId, GroupId, UserId);
            if (senderCredit < creditMinus && !isSell)
                return $"您的积分{senderCredit:N0}不足{creditMinus:N0}。";

            int i;
            long receiverCredit = 0;
            if (isSell)
            {
                var addRes = await UserService.AddCreditTransAsync(SelfId, GroupId, GroupName, rewardQQ, "", rewardCredit, $"打赏加分:{UserId}");
                i = addRes.Result;
                receiverCredit = addRes.CreditValue;
            }
            else if (Group.IsCredit)
            {
                // 使用异步事务版本
                var res = await GroupMember.TransferCoinsAsync(SelfId, GroupId, UserId, Name, rewardQQ, "", (int)CoinsLog.CoinsType.groupCredit, creditMinus, rewardCredit, "打赏");
                i = res.Result;
                senderCredit = res.SenderCoins;
                receiverCredit = res.ReceiverCoins;
            }
            else
            {
                // 使用我们新重写的异步事务版本！
                var result = await UserService.TransferCreditAsync(SelfId, GroupId, GroupName, UserId, Name, rewardQQ, "", creditMinus, rewardCredit, "打赏");
                i = result.Result;
                senderCredit = result.SenderCredit;
                receiverCredit = result.ReceiverCredit;
            }

            string transferFee = isPartner || isSuper ? "" : $"\n💸 服务费：{rewardCredit * 2 / 10:N0}";

            return i == -1
                ? RetryMsg
                : $"✅ 打赏成功！\n🎉 打赏{{积分类型}}：{rewardCredit:N0}{transferFee:N0}\n🎯 对方{{积分类型}}：{receiverCredit:N0}\n🙋 您的{{积分类型}}：{senderCredit:N0}";
        }

        //游戏扣分 (异步重构版)
        public async Task<string> MinusCreditResAsync(long creditMinus, string creditInfo)
        {
            if (!Group.IsCreditSystem) return "";
            if (!IsBlackSystem && (IsPublic || IsGuild || IsRealProxy)) return "";
            
            var res = await UserService.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, -creditMinus, creditInfo);
            return res.Result == -1 ? "" : $"\n💎 {{积分类型}}：-{creditMinus}，累计：{res.CreditValue}";
        }

        public async Task GetCreditMoreAsync()
        {
            CmdPara = "领积分";
            await GetAnswerAsync();
        }

        public async Task<string> GetCreditListAllAsync(long qq, long top = 10)
        {
            var format = !IsRealProxy && (IsMirai || IsQQ) ? "{i} [@:{0}]：{1}\n" : "{i} {0} {1}\n";
            string res = SelfInfo.IsCredit
                ? await FriendRepository.GetCreditRankingAsync(SelfId, GroupId, (int)top, format)
                : await UserRepository.GetCreditRankingAsync(GroupId, (int)top, format);
            if (!res.Contains(qq.ToString()))
                res += $"\n{{积分总排名}} {qq}：{{积分}}";
            return res;
        }

        public async Task<string> GetCreditListAsync(long top = 10)
        {
            var format = !IsRealProxy && (IsMirai || IsQQ) ? "第{i}名[@:{0}] 💎{1:N0}\n" : "第{i}名{0} 💎{1:N0}\n";
            string res = Group.IsCredit
                ? await GroupMemberRepository.GetCreditRankingAsync(GroupId, (int)top, format)
                : SelfInfo.IsCredit
                    ? await FriendRepository.GetCreditRankingAsync(SelfId, GroupId, (int)top, format)
                    : await UserRepository.GetCreditRankingAsync(GroupId, (int)top, format);
            if (!res.Contains(UserId.ToString()))
                res += $"{{积分排名}} [@:{UserId}] 💎{{积分}}\n";
            res = ReplaceRankWithIcon(res);
            return $"🏆 积分排行榜\n{res}";
        }

        static string ReplaceRankWithIcon(string text)
        {
            // 直接用正则替换，匹配“第N名”，用MatchEvaluator决定替换内容
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
}
