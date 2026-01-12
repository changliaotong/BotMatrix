using System.Text.RegularExpressions;

namespace BotWorker.Domain.Models.BotMessages;

//黑名单 blacklist
public partial class BotMessage : MetaData<BotMessage>
{        
        // 解除黑名单
        public async Task<string> GetCancelBlackAsync(long userId)
        {
            if (await BlackList.ExistsAsync(GroupId, userId))
            {
                var res = await BlackList.DeleteAsync(GroupId, userId) == -1
                    ? $"[@:{userId}]{RetryMsg}\n"
                    : $"[@:{userId}]已解除拉黑\n";

                if (await BlackList.IsSystemBlackAsync(userId))
                    res += $"[@:{userId}]已被列入官方黑名单\n";
                return res;
            }

            return $"[@:{userId}]不在黑名单，无需解除\n";
        }

        public string GetCancelBlack(long userId) => GetCancelBlackAsync(userId).GetAwaiter().GetResult();

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
                    res += await GetAddBlackAsync(blackUserId);
                    await KickOutAsync(SelfId, GroupId, blackUserId);
                }
                else
                    res += await GetCancelBlackAsync(blackUserId);
            }            
            return res;
        }

        // 清空黑名单
        public async Task<string> GetClearBlackAsync()
        {
            if (!IsRobotOwner())
                return OwnerOnlyMsg;

            long blackCount = await BlackList.CountKey2Async(GroupId.ToString());
            if (blackCount == 0)
                return "黑名单已为空，无需清空";

            if (!IsConfirm && blackCount > 10)
                return await ConfirmMessage($"清空黑名单 人数{blackCount}");

            return await BlackList.DeleteAllAsync(GroupId) == -1
                ? RetryMsg
                : "✅ 黑名单已清空";
        }

        // 拉黑操作
        public async Task<string> GetAddBlackAsync(long qqBlack)
        {
            //加入黑名单
            if (await BlackList.ExistsAsync(GroupId, qqBlack))           
                return $"[@:{qqBlack}] 已被拉黑，无需再次加入\n";            

            if (qqBlack == UserId)
                return "不能拉黑你自己";

            if (BotInfo.IsRobot(qqBlack))
                return "不能拉黑机器人";

            if (Group.RobotOwner == qqBlack)
                return "不能拉黑我主人";

            string res = "";
            if (await WhiteList.ExistsAsync(GroupId, qqBlack))
            {
                if (Group.RobotOwner != UserId && !BotInfo.IsAdmin(SelfId, UserId))
                    return $"您无权拉黑白名单成员";
                res += await WhiteList.DeleteAsync(GroupId, qqBlack) == -1 
                    ? $"未能将[@:{qqBlack}]从白名单删除" 
                    : $"✅ 已将[@:{qqBlack}]从白名单删除！\n";
            }
            res += await BlackList.AddBlackListAsync(SelfId, GroupId, GroupName, UserId, Name, qqBlack, "") == -1
                ? $"[@:{qqBlack}]{RetryMsg}"
                : $"✅ 已拉黑！";
            return res;
        }

        public string GetAddBlack(long qqBlack) => GetAddBlackAsync(qqBlack).GetAwaiter().GetResult();

        // 加入黑名单
        public async Task<int> AddBlackAsync(long blackQQ, string blackInfo)
        {
            return await BlackList.AddBlackListAsync(SelfId, GroupId, GroupName, UserId, Name, blackQQ, blackInfo);
        }

        public int AddBlack(long blackQQ, string blackInfo) => AddBlackAsync(blackQQ, blackInfo).GetAwaiter().GetResult();
}
