using System.Text.RegularExpressions;

namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        public async Task<bool> IsWhiteListAsync(long userId)
        {
            return await WhiteListRepository.IsExistsAsync(GroupId, userId)
                || UserPerm == 0
                || Group.RobotOwner == userId
                || (Group.IsWhite && UserId == userId && UserPerm < 2);
        }

        public bool IsWhiteList(long userId) => IsWhiteListAsync(userId).GetAwaiter().GetResult();

        public async Task<bool> IsWhiteListAsync()
        {
           return await IsWhiteListAsync(UserId);
        }

        public bool IsWhiteList() => IsWhiteListAsync().GetAwaiter().GetResult();

        // 白名单人数
        public async Task<long> CountWhiteListAsync()
        {
            return await WhiteListRepository.CountAsync($"WHERE group_id = {GroupId}");
        }

        public long CountWhiteList() => CountWhiteListAsync().GetAwaiter().GetResult();

        // 管理员是否有白名单权限
        public string IsWhiteListRes()
        {
            return Group.IsWhite ? "\n管理员已加白" : "";
        }    

        // 白名单列表
        public async Task<string> GetGroupWhiteListAsync()
        {
            string res = await QueryResAsync($"select top 9 white_id from white_list where group_id = {GroupId} order by id desc", "{i}    [@:{0}]\n");
            return $"{(res.IsNull() ? "" : $"{res}\n")}白名单人数：{await CountWhiteListAsync()}\n白名单 + QQ\n取消白名单 + QQ{IsWhiteListRes()}";
        }

        public string GetGroupWhiteList() => GetGroupWhiteListAsync().GetAwaiter().GetResult();

        public async Task<int> AddWhiteAsync(long userId)
        {
            var whiteList = new WhiteList
            {
                BotUin = SelfId,
                GroupId = GroupId,
                GroupName = GroupName,
                UserId = UserId,
                UserName = Name,
                WhiteId = userId
            };
            return await WhiteListRepository.AddAsync(whiteList);
        }

        public int AddWhite(long userId) => AddWhiteAsync(userId).GetAwaiter().GetResult();

        public async Task<string> GetWhiteResAsync()
        {
            IsCancelProxy = true;

            if (!IsRobotOwner() && !BotInfo.IsAdmin(SelfId, UserId))
                return OwnerOnlyMsg;

            if (CmdName == "清空白名单")
                return await GetClearWhiteAsync();

            if (CmdPara == "")
                return await GetGroupWhiteListAsync();

            string res = "";
            if (CmdPara == "管理员")
            {
                var isWhite = CmdName == "白名单";

                if (await GroupRepository.SetValueAsync("IsWhite", isWhite, GroupId) == -1)
                    return RetryMsg;

                return isWhite
                    ? "✅ 管理员已加入白名单"
                    : "✅ 已取消管理员的白名单";
            }

            CmdName = CmdName.Replace("解除", "取消").Replace("删除", "取消");
            foreach (Match match in CmdPara.Matches(Regexs.Users))
            {
                var qqWhite = match.Groups["UserId"].Value.AsLong();
                if (CmdName == "白名单")
                {
                    if (await BlackListRepository.IsExistsAsync(GroupId, qqWhite))
                        res += await BlackListRepository.DeleteAsync(GroupId, qqWhite) == -1
                            ? $"将[@:{qqWhite}]从黑名单删除{RetryMsg}\n"
                            : $"✅ 已成功将[@:{qqWhite}]从黑名单删除！\n";

                    res += await WhiteListRepository.IsExistsAsync(GroupId, qqWhite)
                        ? $"[@:{qqWhite}]已经在白名单里，无需再次加入。\n"
                        : await AddWhiteAsync(qqWhite) == -1
                            ? $"[@:{qqWhite}]加入白名单{RetryMsg}\n"
                            : $"✅ 已成功将[@:{qqWhite}]加入白名单！\n";
                }
                else if (CmdName == "取消白名单")
                {
                    if (IsRobotOwner(qqWhite))
                    {
                        res += "不能取消主人的白名单";
                        continue;
                    }

                    res += !await WhiteListRepository.IsExistsAsync(GroupId, qqWhite)
                        ? $"[@:{qqWhite}]不在白名单中，无需删除!\n"
                        : await WhiteListRepository.DeleteAsync(GroupId, qqWhite) == -1
                            ? $"[@:{qqWhite}]{RetryMsg}\n"
                            : $"✅ [@:{qqWhite}]已经从白名单中删除!\n";
                }
            }
            return res;
        }

        public string GetWhiteRes() => GetWhiteResAsync().GetAwaiter().GetResult(); 

        // 清空白名单
        public async Task<string> GetClearWhiteAsync()
        {
            if (!IsRobotOwner() && !BotInfo.IsAdmin(SelfId, UserId))
                return OwnerOnlyMsg;

            var whiteCount = await CountWhiteListAsync();
            if (whiteCount == 0)
                return "✅ 白名单本就是空的";

            if (whiteCount > 10 && !IsConfirm)
                return await ConfirmMessage($"清空群{GroupId}白名单 数量：{whiteCount}");

            return await WhiteListRepository.DeleteAsync($"WHERE group_id = {GroupId}") == -1
                ? RetryMsg
                : "✅ 白名单已清空";
        }
}
