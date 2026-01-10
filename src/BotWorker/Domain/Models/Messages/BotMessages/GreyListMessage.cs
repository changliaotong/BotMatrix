using System.Text.RegularExpressions;

namespace BotWorker.Domain.Models.Messages.BotMessages;

// 灰名单 greylist
public partial class BotMessage : MetaData<BotMessage>
{
        // 解除灰名单
        public string GetCancelGrey(long userId)
        {
            string res;

            if (GreyList.Exists(GroupId, userId))
                res = GreyList.Delete(GroupId, userId) == -1
                    ? $"[@:{userId}]{RetryMsg}\n"
                    : $"[@:{userId}]已移出灰名单\n";
            else
                res = $"[@:{userId}]不在灰名单，无需移除\n";

            if (GreyList.IsSystemGrey(userId))
                res += $"[@:{userId}]已被列入官方灰名单\n";

            return res;
        }

        // 灰名单列表
        public string GetGroupGreyList()
        {
            return QueryRes(
                       $"SELECT TOP 10 GreyId FROM {GreyList.FullName} WHERE GroupId = {GroupId} ORDER BY Id DESC",
                       "{i} {0}\n"
                   ) +
                   "灰名单人数：" + GreyList.CountWhere($"GroupId = {GroupId}") +
                   "\n拉灰 + QQ\n删灰 + QQ";
        }

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

            long greyCount = GreyList.CountKey2(GroupId.ToString());
            if (greyCount == 0)
                return "灰名单已为空，无需清空";

            if (!IsConfirm && greyCount >= 7)
                return await ConfirmMessage($"清空灰名单 人数{greyCount}");

            return GreyList.DeleteAll(GroupId) == -1
                ? RetryMsg
                : "✅ 灰名单已清空";
        }

        // 加入灰名单
        public string GetAddGrey(long qqGrey)
        {
            string res = "";

            if (GreyList.Exists(GroupId, qqGrey))
                return $"[@:{qqGrey}] 已在灰名单，无需再次加入\n";

            if (qqGrey == UserId)
                return "不能把自己加入灰名单";

            if (BotInfo.IsRobot(qqGrey))
                return "不能把机器人加入灰名单";

            if (Group.RobotOwner == qqGrey)
                return "不能把我主人加入灰名单";

            if (WhiteList.Exists(GroupId, qqGrey))
            {
                if (Group.RobotOwner != UserId && !BotInfo.IsAdmin(SelfId, UserId))
                    return $"您无权操作白名单成员";

                res += WhiteList.Delete(GroupId, qqGrey) == -1
                    ? $"未能将[@:{qqGrey}]从白名单删除"
                    : $"✅ 已将[@:{qqGrey}]从白名单删除！\n";
            }

            res += GreyList.AddGreyList(SelfId, GroupId, GroupName, UserId, Name, qqGrey, "") == -1
                ? $"[@:{qqGrey}]{RetryMsg}"
                : $"✅ 已加入灰名单！";

            return res;
        }

        // 加入灰名单（外部调用）
        public int AddGrey(long greyQQ, string greyInfo)
        {
            return GreyList.AddGreyList(SelfId, GroupId, GroupName, UserId, Name, greyQQ, greyInfo);
        }
}
