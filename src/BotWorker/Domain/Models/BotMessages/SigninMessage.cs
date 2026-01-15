using System.Reflection;

namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{ 
        public async Task<string> TrySignInAsync(bool isAuto = false)
        {
            if (isAuto && !Group.IsAutoSignin)
                return "";

            if (await AddGroupMemberAsync() == -1)
                return RetryMsg;            

            var member = await GroupMember.LoadAsync(GroupId, UserId);
            var signTimes = member.SignTimes;
            var signLevel = member.SignLevel;
            var signTimesAll = member.SignTimesAll;

            bool isSignedToday = member.SignDate.Date == DateTime.Today;           
            if (isSignedToday)
                return isAuto ? "" : BuildSignedMessage(signTimes, signLevel, signTimesAll, true);
                                    
            int dateDiff = (DateTime.Today - member.SignDate.Date).Days;
            if (dateDiff == 1)
            {
                // 昨天签到过，连签天数+1
                signTimes++;
            }
            else
            {
                // 昨天没签到（断签或首次签到），连签天数重置为1
                signTimes = 1;
            }

            // 计算等级
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

            // 计算积分：等级 * 50
            int creditAdd = signLevel * 50;

            if (User.IsSuper)
                creditAdd *= 2;

            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 记录签到流水 (group_signin)
                await SignInRepository.AddSignInAsync(SelfId, GroupId, UserId, CmdPara, trans);

                // 2. 更新 GroupMember 签到信息
                await GroupMember.UpdateSignInfoAsync(GroupId, UserId, signTimes, signLevel, trans);

                // 3. 增加积分 (UserInfo/GroupMember/Friend)
                var res = await UserInfo.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, creditAdd, "签到加分", trans);

                // 4. 增加算力
                var resTokens = await UserInfo.AddTokensAsync(SelfId, GroupId, GroupName, UserId, Name, creditAdd, "签到加算力", trans);

                await trans.CommitAsync();

                // 5. 同步缓存
                await UserInfo.SyncCreditCacheAsync(SelfId, GroupId, UserId, res.CreditValue);
                await UserInfo.SyncTokensCacheAsync(UserId, resTokens.TokensValue);

                await GroupMember.InvalidateAllCachesAsync(GroupId, UserId);

                var result = $"{GetHeadCQ()}✅ {(isAuto ? "自动" : "")}签到成功！\n";
                result += Group.IsCreditSystem ? $"💎 {{积分类型}}：+{creditAdd}→{{积分}}\n" : "";
                result += BuildSignedMessage(signTimes, signLevel, signTimesAll + 1);
                return result;
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Logger.Error($"[TrySignIn Error] {ex.Message}");
                return $"系统繁忙，{RetryMsg}";
            }
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
