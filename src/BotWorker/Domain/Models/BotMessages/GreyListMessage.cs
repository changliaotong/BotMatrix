using System.Text.RegularExpressions;

namespace BotWorker.Domain.Models.BotMessages;

// 灰名单 greylist
public partial class BotMessage
{
        // 解除灰名单
        public async Task<string> GetCancelGreyAsync(long userId)
        {
            string res;

            if (await GreyListRepository.IsExistsAsync(GroupId, userId))
                res = await GreyListRepository.DeleteAsync(GroupId, userId) == -1
                    ? $"[@:{userId}]{RetryMsg}\n"
                    : $"[@:{userId}]已移出灰名单\n";
            else
                res = $"[@:{userId}]不在灰名单，无需移除\n";

            if (await GreyListRepository.IsExistsAsync(BotInfo.GroupIdDef, userId))
                res += $"[@:{userId}]已被列入官方灰名单\n";

            return res;
        }

        public string GetCancelGrey(long userId) => GetCancelGreyAsync(userId).GetAwaiter().GetResult();

        // 灰名单列表
        public async Task<string> GetGroupGreyListAsync()
        {
            string res = await QueryResAsync(
                       $"SELECT GreyId FROM grey_list WHERE GroupId = {GroupId} ORDER BY Id DESC limit 10",
                       "{i} {0}\n"
                   );
            
            return res +
                   "灰名单人数：" + await GreyListRepository.CountAsync($"WHERE group_id = {GroupId}") +
                   "\n拉灰 + QQ\n删灰 + QQ";
        }

        public string GetGroupGreyList() => GetGroupGreyListAsync().GetAwaiter().GetResult();

        // 添加灰名单/取消灰名单
        public async Task<string> GetGreyRes()
        {
            IsCancelProxy = true;

            if (CmdName == "清空灰名单")
                return await GetClearGreyAsync();

            if (CmdPara.IsNull())
                return GetGroupGreyList();

            string res = "";
            CmdName = CmdName.Replace("解除", "取消").Replace("删除", "取消");

            foreach (Match match in CmdPara.Matches(Regexs.Users))
            {
                long greyUserId = match.Groups["UserId"].Value.AsLong();
                if (CmdName == "拉灰")
                {
                    res += GetAddGrey(greyUserId);
                    await KickOutAsync(SelfId, GroupId, greyUserId);
                }
                else if (CmdName == "取消拉灰")
                    res += GetCancelGrey(greyUserId);
            }
            return res;
        }

        // 清空灰名单
        public async Task<string> GetClearGreyAsync()
        {
            if (!IsRobotOwner())
                return $"您无权清空灰名单";

            long greyCount = await GreyListRepository.CountAsync($"WHERE group_id = {GroupId}");
            if (greyCount == 0)
                return "灰名单已为空，无需清空";

            if (!IsConfirm && greyCount >= 7)
                return await ConfirmMessage($"清空灰名单 人数{greyCount}");

            return await GreyListRepository.DeleteAsync($"WHERE group_id = {GroupId}") == -1
                ? RetryMsg
                : "✅ 灰名单已清空";
        }

        // 加入灰名单
        public async Task<string> GetAddGreyAsync(long qqGrey)
        {
            string res = "";

            if (await GreyListRepository.IsExistsAsync(GroupId, qqGrey))
                return $"[@:{qqGrey}] 已在灰名单，无需再次加入\n";

            if (qqGrey == UserId)
                return "不能把自己加入灰名单";

            if (BotInfo.IsRobot(qqGrey))
                return "不能把机器人加入灰名单";

            if (Group.RobotOwner == qqGrey)
                return "不能把我主人加入灰名单";

            if (await WhiteListRepository.IsExistsAsync(GroupId, qqGrey))
            {
                if (Group.RobotOwner != UserId && !BotInfo.IsAdmin(SelfId, UserId))
                    return $"您无权操作白名单成员";

                res += await WhiteListRepository.DeleteAsync(GroupId, qqGrey) == -1
                    ? $"未能将[@:{qqGrey}]从白名单删除"
                    : $"✅ 已将[@:{qqGrey}]从白名单删除！\n";
            }

            var greyList = new GreyList
            {
                BotUin = SelfId,
                GroupId = GroupId,
                GroupName = GroupName,
                UserId = UserId,
                UserName = Name,
                GreyId = qqGrey,
                GreyInfo = ""
            };

            res += await GreyListRepository.AddAsync(greyList) == -1
                ? $"[@:{qqGrey}]{RetryMsg}"
                : $"✅ 已加入灰名单！";

            return res;
        }

        public string GetAddGrey(long qqGrey) => GetAddGreyAsync(qqGrey).GetAwaiter().GetResult();

        // 加入灰名单（外部调用）
        public async Task<int> AddGreyAsync(long greyQQ, string greyInfo)
        {
            if (await GreyListRepository.IsExistsAsync(GroupId, greyQQ))
                return 0;

            var greyList = new GreyList
            {
                BotUin = SelfId,
                GroupId = GroupId,
                GroupName = GroupName,
                UserId = UserId,
                UserName = Name,
                GreyId = greyQQ,
                GreyInfo = greyInfo
            };

            return await GreyListRepository.AddAsync(greyList);
        }

        public int AddGrey(long greyQQ, string greyInfo) => AddGreyAsync(greyQQ, greyInfo).GetAwaiter().GetResult();
}
