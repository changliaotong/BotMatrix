using BotWorker.Groups;
using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;
using BotWorker.Bots.Users;
using BotWorker.Bots.Entries;

namespace BotWorker.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    { 

        public string TrySignIn(bool isAuto = false)
        {
            if (isAuto && !Group.IsAutoSignin)
                return "";

            if (AddGroupMember() == -1)
                return RetryMsg;            

            var member = GroupMember.GetDict(GroupId, UserId, "SignTimes", "SignLevel", "SignTimesAll");
            var signTimes = member?["SignTimes"].AsInt() ?? 0;
            var signLevel = member?["SignLevel"].AsInt() ?? 0;
            var signTimesAll = member?["SignTimesAll"].AsInt() ?? 0;

            bool isSignedToday = GroupMember.IsSignIn(GroupId, UserId);
            if (isSignedToday)
                return isAuto ? "" : BuildSignedMessage(signTimes, signLevel, signTimesAll, true);
                                    
            if (!isSignedToday && GroupSignIn.Append(SelfId, GroupId, UserId, CmdPara) == -1)
                return $"系统繁忙，{RetryMsg}";

            int nextLevelDays;
            int creditAdd = 50;

            if (GroupMember.GetSignDateDiff(GroupId, UserId) <= 1)
            {
                if (!isSignedToday)
                {
                    signTimes++;
                    signLevel = signTimes switch
                    {
                        >= 230 => 10,
                        >= 170 => 9,
                        >= 120 => 8,
                        >= 80 => 7,
                        >= 50 => 6,
                        >= 30 => 5,
                        >= 14 => 4,
                        >= 7 => 3,
                        >= 3 => 2,
                        _ => 1,
                    };
                }

                nextLevelDays = signLevel switch
                {
                    10 => 0,
                    9 => 230 - signTimes,
                    8 => 170 - signTimes,
                    7 => 120 - signTimes,
                    6 => 80 - signTimes,
                    5 => 50 - signTimes,
                    4 => 30 - signTimes,
                    3 => 14 - signTimes,
                    2 => 7 - signTimes,
                    1 => 3 - signTimes,
                    _ => 0,
                };

                creditAdd = signLevel * 50;
            }
            else
            {
                signTimes = 1;
                signLevel = 1;
            }

            string? result;
            if (isSignedToday)
            {
                result = BuildSignedMessage(signTimes, signLevel, signTimesAll, alreadySigned: true);
            }
            else
            {
                if (User.IsSuper)
                    creditAdd *= 2;

                var creditValue = UserInfo.GetCredit(GroupId, UserId) + creditAdd;
                var tokensAdd = creditAdd;

                var sqls = new[]
                {
                    GroupMember.SqlUpdateSignInfo(GroupId, UserId, signTimes, signLevel),
                    Group.IsCreditSystem ? UserInfo.SqlAddCredit(SelfId, GroupId, UserId, creditAdd): ("", []),
                    Group.IsCreditSystem ? CreditLog.SqlHistory(SelfId, GroupId, GroupName, UserId, Name, creditAdd, "签到加分") : ("", []),
                    UserInfo.SqlPlus("tokens", tokensAdd, UserId),
                    TokensLog.SqlLog(SelfId, GroupId, GroupName, UserId, Name, tokensAdd, "签到加算力")
                };

                if (ExecTrans(sqls) == -1)
                    return $"系统繁忙，{RetryMsg}";

                result = $"{GetHeadCQ()}✅ {(isAuto ? "自动" : "")}签到成功！\n";
                result += Group.IsCreditSystem ? $"💎 {{积分类型}}：+{creditAdd}→{creditValue:N0}\n" : "";

                result += BuildSignedMessage(signTimes, signLevel, signTimesAll + 1);
            }

            return result;
        }

        private string BuildSignedMessage(int signTimes = 0, int signLevel = 1, int signTimesAll = 0, bool alreadySigned = false)
        {
            var res = alreadySigned ? $"{GetHeadCQ()}✅ 今天签过了，明天再来！\n{(Group.IsCreditSystem ? $"💎 {{积分类型}}：{{积分}}\n" : "")}" : "";
            var nextLevelDays = signLevel switch
            {
                10 => 0,
                9 => 230 - signTimes,
                8 => 170 - signTimes,
                7 => 120 - signTimes,
                6 => 80 - signTimes,
                5 => 50 - signTimes,
                4 => 30 - signTimes,
                3 => 14 - signTimes,
                2 => 7 - signTimes,
                1 => 3 - signTimes,
                _ => 0,
            };

            res += Group.IsCreditSystem ? $"🏆 积分排名：本群{{积分排名}} 世界{{积分总排名}}\n" : "";
            res += $"📅 签到天数：连签{signTimes} 累计{signTimesAll} ✨\n" +
                   $"🗣️ 发言次数：今天{{今日发言次数}} 昨天{{昨日发言次数}}\n" +
                   $"👥 签到人次：今天{{今日签到人数}} 昨天{{昨日签到人数}}";

            return res;
        }
    }

}
