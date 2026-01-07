using System.Text.RegularExpressions;
using sz84.Bots.Entries;
using sz84.Bots.Groups;
using BotWorker.Common;
using BotWorker.Common.Exts;
using sz84.Core.MetaDatas;

namespace sz84.Bots.BotMessages
{
    //黑名单 blacklist
    public partial class BotMessage : MetaData<BotMessage>
    {        
        // 解除黑名单
        public string GetCancelBlack(long userId)
        {
            string res;

            if (BlackList.Exists(GroupId, userId))
                res = BlackList.Delete(GroupId, userId) == -1
                    ? $"[@:{userId}]{RetryMsg}\n"
                    : $"[@:{userId}]已解除拉黑\n";
            else
                res = $"[@:{userId}]不在黑名单，无需解除\n";

            if (BlackList.IsSystemBlack(userId))
                res += $"[@:{userId}]已被列入官方黑名单\n";
            return res;
        }

        // 黑名单列表
        public string GetGroupBlackList()
        {
            return QueryRes($"SELECT TOP 10 BlackId FROM {BlackList.FullName} WHERE GroupId = {GroupId} ORDER BY Id DESC",
                            "{i} {0}\n") +
                   "已拉黑人数：" + BlackList.CountWhere($"GroupId = {GroupId}") +
                   "\n拉黑 + QQ\n删黑 + QQ";
        }

        //拉黑
        public async Task<string> GetBlackRes()        
        {
            IsCancelProxy = true;

            if (CmdName == "清空黑名单")
                return GetClearBlack();

            if (CmdPara.IsNull())                            
                return GetGroupBlackList();            

            //一次加多个号码进入黑名单
            string res = "";
            CmdName = CmdName.Replace("解除", "取消").Replace("删除", "取消");
            foreach (Match match in CmdPara.Matches(Regexs.Users))
            {                
                long blackUserId = match.Groups["UserId"].Value.AsLong();
                if (CmdName == "拉黑")
                {
                    res += GetAddBlack(blackUserId);
                    await KickOutAsync(SelfId, GroupId, blackUserId);
                }
                else if (CmdName == "取消拉黑")
                    res += GetCancelBlack(blackUserId);
            }            
            return res;
        }

        // 清空黑名单
        public string GetClearBlack()
        {
            if (!IsRobotOwner())
                return OwnerOnlyMsg;

            long blackCount = BlackList.CountKey2(GroupId.ToString());
            if (blackCount == 0)
                return "黑名单已为空，无需清空";

            if (!IsConfirm && blackCount > 10)
                return ConfirmMessage($"清空黑名单 人数{blackCount}");

            return BlackList.DeleteAll(GroupId) == -1
                ? RetryMsg
                : "✅ 黑名单已清空";
        }

        // 拉黑操作
        public string GetAddBlack(long qqBlack)
        {
            string res = "";

            //加入黑名单
            if (BlackList.Exists(GroupId, qqBlack))           
                return $"[@:{qqBlack}] 已被拉黑，无需再次加入\n";            

            if (qqBlack == UserId)
                return "不能拉黑你自己";

            if (BotInfo.IsRobot(qqBlack))
                return "不能拉黑机器人";

            if (Group.RobotOwner == qqBlack)
                return "不能拉黑我主人";

            if (WhiteList.Exists(GroupId, qqBlack))
            {
                if (Group.RobotOwner != UserId && !BotInfo.IsAdmin(SelfId, UserId))
                    return $"您无权拉黑白名单成员";
                res += WhiteList.Delete(GroupId, qqBlack) == -1 
                    ? $"未能将[@:{qqBlack}]从白名单删除" 
                    : $"✅ 已将[@:{qqBlack}]从白名单删除！\n";
            }
            res += BlackList.AddBlackList(SelfId, GroupId, GroupName, UserId, Name, qqBlack, "") == -1
                ? $"[@:{qqBlack}]{RetryMsg}"
                : $"✅ 已拉黑！";
            return res;
        }

        // 加入黑名单
        public int AddBlack(long blackQQ, string blackInfo)
        {
            return BlackList.AddBlackList(SelfId, GroupId, GroupName, UserId, Name, blackQQ, blackInfo);
        }
    }
}
