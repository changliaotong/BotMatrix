using System.Text.RegularExpressions;
using sz84.Bots.Entries;
using sz84.Bots.Groups;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        private readonly string[] kickMessages =
            [
                "拜拜了您呐~",
                "请大家自觉遵守群规",
                "我送你离开，千里之外~",
            ];

        //群员被踢
        public void GetBeKicked()
        {
            //邀请统计
            SubInviteCount();

            var isRobot = UserId == SelfId;
            if (isRobot)
            {
                BotEventLog.Append(this, "被踢出群");
                GroupInfo.SetInvalid(GroupId, GroupName);
                return;
            }

            IsSend = !isRobot && UserId != 0 && Group.IsExitHint;

            Answer = $"({UserId}) 已被{(Operater == 0 ? "" : $"[@:{Operater}]({Operater})")}T飞！\n{kickMessages.RandomOne()}";

            //10秒内踢过人的不提示
            IsSend = IsSend && GroupInfo.GetLastHintTime(GroupId) >= 10;
            if (!IsGuild)
                IsCancelProxy = true;

            if (Group.IsBlackKick && (UserId != 1000000) && !BotInfo.IsRobot(UserId))
            {
                int i = BlackList.AddBlackList(SelfId, GroupId, GroupName, Operater, OperaterName, UserId, "被踢拉黑");
                if (i != -1)
                    Answer += "\n已拉黑！";
            }

            //更新最后踢人提示时间
            if (Answer != "")
                GroupInfo.SetHintDate(GroupId);
        }

        public async Task<string> GetKickOutAsync()
        {
            IsCancelProxy = true;

            // 判断权限            
            if (!HaveSetupRight())            
                return "您没有踢人权限";            

            // 判断是否是踢人命令
            if (!Regex.IsMatch(Message, Regexs.KickCommandPrefixPattern, RegexOptions.IgnoreCase))
                return "";

            // 提取所有 QQ 号
            var qqMatches = Regex.Matches(Message, Regexs.QqNumberPattern);
            var kicked = new List<long>();
            var isBlackKick = Group.IsBlackKick;   
            var answer = "";
            foreach (Match match in qqMatches)
            {
                var qq = match.Value.AsLong();

                // 拉黑逻辑
                if (isBlackKick)    
                {
                    if (AddBlack(qq, "被踢拉黑") != -1)
                        answer += $"\n{qq} 已拉黑！";                    
                }               

                await KickOutAsync(SelfId, RealGroupId, qq);

                kicked.Add(qq);
            }

            if (kicked.Count == 0)
                return "未识别到有效的目标 QQ";
            else if (string.IsNullOrWhiteSpace(Answer))
                return $"✅ 已T飞 {kicked.Count} 个成员";     
            
            return answer;
        }


    }
}
