namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
        //卖出积分
        public string GetSellCredit()
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (CmdPara == "")
                return "📄 命令格式：卖分 + 数值\n📌 使用示例：卖分 1000\n💎 超级积分：10,000→4R\n🎁 普通积分：10,000→1R\n📦 您的{积分类型}：{积分}";

            if (BotInfo.GetIsCredit(SelfId))
                return "本机积分不能兑换余额";

            if (GroupInfo.GetIsCredit(GroupId))
                return "本群积分不能兑换余额";

            if (!CmdPara.IsNum())
                return "数量不正确！";

            long creditMinus = CmdPara.AsLong();
            if (creditMinus < 1000)
                return "至少需要1000分";

            long creditValue = UserInfo.GetCredit(GroupId, UserId);
            if (creditValue < creditMinus)
                return $"您只有{creditValue}分";

            return "您无权使用此命令";

            //creditValue -= creditMinus;
            //decimal balanceValue = GetBalance(userId);
            //decimal xCredit = GetIsSuper(userId) ? 0.04m : 0.01m;
            //decimal banalceAdd = creditMinus * xCredit / 100;
            //decimal balanceNew = balanceValue + banalceAdd;

            //扣分、加余额
            //var sql = SqlAddCredit(botUin, groupId, userId, -creditMinus);
            //var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, userId, name, -creditMinus, "卖分");
            //var sql3 = SqlAddBalance(userId, banalceAdd);
            //var sql4 = BalanceLog.SqlLog(botUin, groupId, groupName, userId, name, banalceAdd, "卖分");
            //int i = ExecTrans(sql, sql2, sql3, sql4);

            //return i == -1
            //  ? RetryMsg
            //: $"✅ 卖出成功！\n💎 积分：-{creditMinus:N0}→{creditValue:N0}\n💳 余额：+{banalceAdd:N}→{balanceNew:N}";
        }

        //取分
        public int WithdrawCredit(long creditOper, ref long creditValue, ref long creditSave, ref string res)
            => DoSaveCreditInternal(-creditOper, ref creditValue, ref creditSave, ref res);

        private int DoSaveCreditInternal(long creditOper, ref long creditValue, ref long creditSave, ref string res)
        {
            var result = DoSaveCreditAsync(creditOper).GetAwaiter().GetResult();
            creditValue = result.CreditValue;
            creditSave = result.CreditSave;
            res = result.Res;
            return result.Result;
        }

        public async Task<string> GetSaveCreditResAsync()
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (CmdPara == "")
                return "格式:存分 + 积分数\n取分 + 积分数\n例如：存分 100";

            if (!CmdPara.IsNum())
                return "参数不正确";

            long credit_oper = CmdPara.AsLong();
            CmdName = CmdName.ToLower();
            if (CmdName.StartsWith('存') | CmdName.StartsWith('c'))
                CmdName = "存分";

            if (CmdName.StartsWith('取') | CmdName.StartsWith('q'))
                CmdName = "取分";

            string res = "";

            if (CmdName == "存分")
            {
                credit_oper = credit_oper == 0 ? await UserInfo.GetCreditAsync(GroupId, UserId) : credit_oper;
                if (credit_oper == 0)
                    return "您没有积分可存";

                var saveRes = await DoSaveCreditAsync(credit_oper);
                res = saveRes.Res;
            }
            else if (CmdName == "取分")
            {
                credit_oper = credit_oper == 0 ? await UserInfo.GetSaveCreditAsync(GroupId, UserId) : credit_oper;
                if (credit_oper == 0)
                    return "您没有积分可取";

                var saveRes = await DoSaveCreditAsync(-credit_oper);
                res = saveRes.Res;
            }
            return res;
        }

        //存取分 (异步重构版)
        public async Task<(int Result, long CreditValue, long CreditSave, string Res)> DoSaveCreditAsync(long creditOper)
        {
            long creditValue = await UserInfo.GetCreditAsync(GroupId, UserId);
            long creditSave = await UserInfo.GetSaveCreditAsync(GroupId, UserId);
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

            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 记录日志 (自动支持事务)
                await CreditLog.AddLogAsync(SelfId, GroupId, GroupName, UserId, Name, -creditOper, cmdName, trans);

                // 2. 更新存分 (自动支持事务)
                var (sql, paras) = UserInfo.SqlSaveCredit(SelfId, GroupId, UserId, creditOper);
                await ExecAsync(sql, trans, paras);

                await trans.CommitAsync();

                creditSave += creditOper;
                creditValue -= creditOper;

                // 同步缓存
                UserInfo.SyncCacheField(UserId, GroupId, "Credit", creditValue);
                UserInfo.SyncCacheField(UserId, GroupId, "SaveCredit", creditSave);

                res = $"✅ {cmdName}：{credit_oper2}\n" +
                    $"💰 {{积分类型}}：{creditValue:N0}\n" +
                    $"🏦 已存积分：{creditSave:N0}\n" +
                    $"📈 积分总额：{creditValue + creditSave:N0}";
                return (0, creditValue, creditSave, res);
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[DoSaveCredit Error] {ex.Message}");
                res = RetryMsg;
                return (-1, creditValue, creditSave, res);
            }
        }

        
        //增加算力
        public int AddTokens(long tokensAdd, string tokensInfo)
        {
            return UserInfo.AddTokens(SelfId, GroupId, GroupName, UserId, Name, tokensAdd, tokensInfo);
        }

        //减少算力
        public int MinusTokens(long tokensMinus, string tokensInfo)
        {
            return AddTokens(-tokensMinus, tokensInfo);
        }

        //增加积分
        public async Task<(int, long)> AddCreditAsync(long creditAdd, string creditInfo)
        {
            var res = await UserInfo.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, creditAdd, creditInfo);
            return (res.Result, res.CreditValue);
        }

        //增加积分
        public (int, long) AddCredit(long creditAdd, string creditInfo)
        {
            return UserInfo.AddCredit(SelfId, GroupId, GroupName, UserId, Name, creditAdd, creditInfo);
        }

        //减少积分
        public async Task<(int, long)> MinusCreditAsync(long creditMinus, string creditInfo)
        {
            return await AddCreditAsync(-creditMinus, creditInfo);
        }

        //减少积分
        public (int, long) MinusCredit(long creditMinus, string creditInfo)
        {
            return AddCredit(-creditMinus, creditInfo);
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

            long senderCredit = UserInfo.GetCredit(GroupId, UserId);
            if (senderCredit < creditMinus && !isSell)
                return $"您的积分{senderCredit:N0}不足{creditMinus:N0}。";

            int i;
            long receiverCredit = 0;
            if (isSell)
            {
                var addRes = await UserInfo.AddCreditAsync(SelfId, GroupId, GroupName, rewardQQ, "", rewardCredit, $"打赏加分:{UserId}");
                i = addRes.Result;
                receiverCredit = addRes.CreditValue;
            }
            else if (Group.IsCredit)
            {
                // 使用异步事务版本
                var res = await GroupMember.TransferCoinsAsync(SelfId, GroupId, GroupName, UserId, Name, rewardQQ, "", (int)CoinsLog.CoinsType.groupCredit, creditMinus, rewardCredit, "打赏");
                i = res.Result;
                senderCredit = res.SenderCoins;
                receiverCredit = res.ReceiverCoins;
            }
            else
            {
                // 使用我们新重写的异步事务版本！
                var result = await UserInfo.TransferCreditAsync(SelfId, GroupId, GroupName, UserId, Name, rewardQQ, "", creditMinus, rewardCredit, "打赏");
                i = result.Result;
                senderCredit = result.SenderCredit;
                receiverCredit = result.ReceiverCredit;
            }

            string transferFee = isPartner || isSuper ? "" : $"\n💸 服务费：{rewardCredit * 2 / 10:N0}";

            return i == -1
                ? RetryMsg
                : $"✅ 打赏成功！\n🎉 打赏积分：{rewardCredit:N0}{transferFee:N0}\n🎯 对方积分：{receiverCredit:N0}\n🙋 您的积分：{senderCredit:N0}";
        }

        public long GetCredit()
        {
            return UserInfo.GetCredit(GroupId, UserId);
        }

        //游戏扣分 (异步重构版)
        public async Task<string> MinusCreditResAsync(long creditMinus, string creditInfo)
        {
            if (!Group.IsCreditSystem) return "";
            if (!IsBlackSystem && (IsPublic || IsGuild || IsRealProxy)) return "";
            
            var res = await UserInfo.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, -creditMinus, creditInfo);
            return res.Result == -1 ? "" : $"\n💎 积分：-{creditMinus}，累计：{res.CreditValue}";
        }

        public string MinusCreditRes(long creditMinus, string creditInfo)
        {
            return MinusCreditResAsync(creditMinus, creditInfo).GetAwaiter().GetResult();
        }

        public async Task GetCreditMoreAsync()
        {
            CmdPara = "领积分";
            await GetAnswerAsync();
        }

        public string GetCreditListAll(long qq, long top = 10)
        {
            var format = !IsRealProxy && (IsMirai || IsQQ) ? "{i} [@:{0}]：{1}\n" : "{i} {0} {1}\n";
            string res = SelfInfo.IsCredit
                ? QueryRes($"select top {top} UserId, credit from {Friend.FullName} where BotUin = {SelfId} order by Credit desc", format)
                : QueryRes($"select top {top} Id, credit from {UserInfo.FullName} order by Credit desc", format);
            if (!res.Contains(qq.ToString()))
                res += $"{{积分总排名}} {qq}：{{积分}}\n";
            return res;
        }

        public string GetCreditList(long top = 10)
        {
            var format = !IsRealProxy && (IsMirai || IsQQ) ? "第{i}名[@:{0}] 💎{1:N0}\n" : "第{i}名{0} 💎{1:N0}\n";
            string res = Group.IsCredit
                ? GroupMember.QueryWhere($"top {top} UserId, GroupCredit", $"groupId = {GroupId}", "GroupCredit desc", format)
                : SelfInfo.IsCredit
                    ? Friend.QueryWhere($"top {top} UserId, credit", $"UserId in (select UserId from {GroupMember.FullName} where GroupId = {GroupId})",
                                        $"credit desc", format)
                    : UserInfo.QueryWhere($"top {top} Id, Credit", $"Id in (select UserId from {CreditLog.FullName} where GroupId = {GroupId})",
                                 $"credit desc", format);
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
