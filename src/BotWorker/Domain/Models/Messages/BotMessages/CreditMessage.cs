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
        {
            return DoSaveCredit(-creditOper, ref creditValue, ref creditSave, ref res);
        }

        public string GetSaveCreditRes()
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
            long creditValue = 0;
            long saveCredit = 0;

            if (CmdName == "存分")
            {
                credit_oper = credit_oper == 0 ? UserInfo.GetCredit(GroupId, UserId) : credit_oper;
                if (credit_oper == 0)
                    return "您没有积分可存";

                DoSaveCredit(credit_oper, ref creditValue, ref saveCredit, ref res);
            }
            else if (CmdName == "取分")
            {
                credit_oper = credit_oper == 0 ? UserInfo.GetSaveCredit(GroupId, UserId) : credit_oper;
                if (credit_oper == 0)
                    return "您没有积分可取";

                WithdrawCredit(credit_oper, ref creditValue, ref saveCredit, ref res);
            }
            return res;
        }

        //存取分
        public int DoSaveCredit(long creditOper, ref long creditValue, ref long creditSave, ref string res)
        {
            creditValue = UserInfo.GetCredit(GroupId, UserId);
            creditSave = UserInfo.GetSaveCredit(GroupId, UserId);
            long credit_oper2 = creditOper;
            string cmdName = "存分";
            if (creditOper > 0)
            {
                if (creditValue < credit_oper2)
                {
                    res = $"您只有{creditValue:N0}分";
                    return -1;
                }
            }
            else
            {
                credit_oper2 = -creditOper;
                if (creditSave < credit_oper2)
                {
                    res = $"您已存分只有{creditSave:N0}";
                    return -1;
                }
                cmdName = "取分";
            }
            creditSave += creditOper;
            creditValue -= creditOper;
            var sql = CreditLog.SqlHistory(SelfId, GroupId, GroupName, UserId, Name, -creditOper, cmdName);
            var sql2 = UserInfo.SqlSaveCredit(SelfId, GroupId, UserId, creditOper);
            int i = ExecTrans(sql, sql2);
            if (i == -1)
            {
                res = RetryMsg;
                return i;
            }
            res = $"✅ {cmdName}：{credit_oper2}\n" +
                $"💰 {{积分类型}}：{creditValue:N0}\n" +
                $"🏦 已存积分：{creditSave:N0}\n" +
                $"📈 积分总额：{creditValue + creditSave:N0}";
            return i;
        }

        public string GetFreeCredit()
        {
            //领积分
            //if (!ClientPublic.IsBind(QQ))
            //return $"TOKEN:MP{ClientPublic.GetBindToken(robotKey, clientKey)}\n复制此消息发给QQ机器人即可得分";
            return $"";
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
        public (int, long) AddCredit(long creditAdd, string creditInfo)
        {
            return UserInfo.AddCredit(SelfId, GroupId, GroupName, UserId, Name, creditAdd, creditInfo);
        }

        //减少积分
        public (int, long) MinusCredit(long creditMinus, string creditInfo)
        {
            return AddCredit(-creditMinus, creditInfo);
        }

        //打赏
        public string GetRewardCredit()
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

            long creditValue = UserInfo.GetCredit(GroupId, UserId);
            if (creditValue < creditMinus && !isSell)
                return $"您的积分{creditValue:N0}不足{creditMinus:N0}。";

            long creditValue2 = UserInfo.GetCredit(GroupId, rewardQQ);
            int i;
            if (isSell)
            {                
                i = UserInfo.AddCredit(SelfId, GroupId, GroupName, rewardQQ, "", rewardCredit, $"打赏加分:{UserId}").Item1;
                creditValue2 += rewardCredit;
            }
            else if (Group.IsCredit)
                i = GroupMember.TransferCoins(SelfId, GroupId, GroupName, UserId, Name, rewardQQ, (int)CoinsLog.CoinsType.groupCredit, creditMinus, rewardCredit, ref creditValue, ref creditValue2);
            else 
                i = UserInfo.TransferCredit(SelfId, GroupId, GroupName, UserId, Name, rewardQQ, "", creditMinus, rewardCredit, ref creditValue, ref creditValue2, "打赏");

            string transferFee = isPartner || isSuper ? "" : $"\n💸 服务费：{rewardCredit * 2 / 10:N0}";

            return i == -1
                ? RetryMsg
                : $"✅ 打赏成功！\n🎉 打赏积分：{rewardCredit:N0}{transferFee:N0}\n🎯 对方积分：{creditValue2:N0}\n🙋 您的积分：{creditValue:N0}";
        }

        public long GetCredit()
        {
            return UserInfo.GetCredit(GroupId, UserId);
        }

        //游戏扣分
        public string MinusCreditRes(long creditMinus, string creditInfo)
        {
            if (!Group.IsCreditSystem) return "";
            if (!IsBlackSystem && (IsPublic || IsGuild || IsRealProxy)) return "";
            (int i, long creditValue) = MinusCredit(creditMinus, creditInfo);
            return i == -1 ? "" : $"\n💎 积分：-{creditMinus}，累计：{creditValue}";
        }

        public async Task GetCreditMoreAsync()
        {
            CmdPara = "领积分";
            await GetAnswerAsync();
        }

        public string GetCreditListAll(long qq, long top = 10)
        {
            var format = !IsRealProxy && (IsMirai || IsNapCat) ? "{i} [@:{0}]：{1}\n" : "{i} {0} {1}\n";
            string res = SelfInfo.IsCredit
                ? QueryRes($"select top {top} UserId, credit from {Friend.FullName} where BotUin = {SelfId} order by Credit desc", format)
                : QueryRes($"select top {top} Id, credit from {UserInfo.FullName} order by Credit desc", format);
            if (!res.Contains(qq.ToString()))
                res += $"{{积分总排名}} {qq}：{{积分}}\n";
            return res;
        }

        public string GetCreditList(long top = 10)
        {
            var format = !IsRealProxy && (IsMirai || IsNapCat) ? "第{i}名[@:{0}] 💎{1:N0}\n" : "第{i}名{0} 💎{1:N0}\n";
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
