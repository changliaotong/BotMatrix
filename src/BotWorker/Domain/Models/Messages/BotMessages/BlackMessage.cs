using System.Text.RegularExpressions;

namespace BotWorker.Domain.Models.Messages.BotMessages;

//黑名单 blacklist
public partial class BotMessage : MetaData<BotMessage>
{        
        // 解除黑名单
        public string GetCancelBlack(long userId)
        {
            if (BlackList.Exists(GroupId, userId))
            {
                var res = BlackList.Delete(GroupId, userId) == -1
                    ? $"[@:{userId}]{RetryMsg}\n"
                    : $"[@:{userId}]已解除拉黑\n";

                if (BlackList.IsSystemBlack(userId))
                    res += $"[@:{userId}]已被列入官方黑名单\n";
                return res;
            }

            return $"[@:{userId}]不在黑名单，无需解除\n";
        }

        // 黑名单列表
        public async Task<string> GetGroupBlackListAsync()
        {
            var list = await QueryResAsync($"SELECT {SqlTop(10)} BlackId FROM {BlackList.FullName} WHERE GroupId = {GroupId} ORDER BY Id DESC {SqlLimit(10)}",
                            "{i} {0}\n");
            
            var count = await BlackList.CountWhereAsync($"GroupId = {GroupId}");
            
            return (string.IsNullOrEmpty(list) ? "🌑 黑名单暂无记录\n" : $"🌑 黑名单列表 (前10):\n{list}") +
                   $"👥 已拉黑人数：{count}\n" +
                   "📝 命令提示：\n" +
                   "拉黑 + QQ：将用户加入黑名单\n" +
                   "解除拉黑 + QQ：将用户移出黑名单";
        }

        public string GetGroupBlackList() => GetGroupBlackListAsync().GetAwaiter().GetResult();

        //拉黑
        public async Task<string> GetBlackRes()        
        {
            IsCancelProxy = true;

            if (CmdName == "清空黑名单")
                return await GetClearBlackAsync();

            if (CmdPara.IsNull())                            
                return await GetGroupBlackListAsync();            

            //一次加多个号码进入黑名单
            string res = "";
            var isAdd = !CmdName.Contains("取消") && !CmdName.Contains("删除") && !CmdName.Contains("解除");
            
            foreach (Match match in CmdPara.Matches(Regexs.Users))
            {                
                long blackUserId = match.Groups["UserId"].Value.AsLong();
                if (isAdd)
                {
                    res += GetAddBlack(blackUserId);
                    await KickOutAsync(SelfId, GroupId, blackUserId);
                }
                else
                    res += GetCancelBlack(blackUserId);
            }            
            return res;
        }

        // 清空黑名单
        public async Task<string> GetClearBlackAsync()
        {
            if (!IsRobotOwner())
                return OwnerOnlyMsg;

            long blackCount = BlackList.CountKey2(GroupId.ToString());
            if (blackCount == 0)
                return "黑名单已为空，无需清空";

            if (!IsConfirm && blackCount > 10)
                return await ConfirmMessage($"清空黑名单 人数{blackCount}");

            return BlackList.DeleteAll(GroupId) == -1
                ? RetryMsg
                : "✅ 黑名单已清空";
        }

        // 拉黑操作
        public string GetAddBlack(long qqBlack)
        {
            //加入黑名单
            if (BlackList.Exists(GroupId, qqBlack))           
                return $"[@:{qqBlack}] 已被拉黑，无需再次加入\n";            

            if (qqBlack == UserId)
                return "不能拉黑你自己";

            if (BotInfo.IsRobot(qqBlack))
                return "不能拉黑机器人";

            if (Group.RobotOwner == qqBlack)
                return "不能拉黑我主人";

            string res = "";
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
