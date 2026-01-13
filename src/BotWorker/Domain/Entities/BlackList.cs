namespace BotWorker.Domain.Entities
{
    public class BlackList : MetaData<BlackList>
    {
        public override string TableName => "BlackList";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "BlackId";

        public const string regexBlack = @"^(?<cmdName>(取消|解除|删除)?(黑名单|拉黑|加黑|删黑))(?<cmdPara>([ ]*(\[?@:?)?[1-9]+\d*(\]?))+)$";

        public static async Task<List<long>> GetSystemBlackListAsync()
        {
            var (sql, parameters) = SqlSelect("BlackId", BotInfo.GroupIdDef);
            return await QueryListAsync<long>(sql, null, parameters);
        }

        public static bool IsSystemBlack(long userId) => IsSystemBlackAsync(userId).GetAwaiter().GetResult();

        public static async Task<bool> IsSystemBlackAsync(long userId)
        {
            return await ExistsAsync(BotInfo.GroupIdDef, userId);
        }

        public static int AddBlackList(long botUin, long groupId, string groupName, long qq, string name, long blackQQ, string blackInfo)
            => AddBlackListAsync(botUin, groupId, groupName, qq, name, blackQQ, blackInfo).GetAwaiter().GetResult();

        // 加入黑名单
        public static async Task<int> AddBlackListAsync(long botUin, long groupId, string groupName, long qq, string name, long blackQQ, string blackInfo)
        {
            return await ExistsAsync(groupId, blackQQ)
                ? 0
                : await InsertAsync([
                            new Cov("BotUin", botUin),
                            new Cov("GroupId", groupId),
                            new Cov("GroupName", groupName),
                            new Cov("UserId", qq),
                            new Cov("UserName", name),
                            new Cov("BlackId", blackQQ),
                            new Cov("BlackInfo", blackInfo),
                        ]);
        }

        /// <summary>
        /// 清空指定群组的黑名单
        /// </summary>
        public static int ClearGroupBlacklist(long groupId) => ClearGroupBlacklistAsync(groupId).GetAwaiter().GetResult();

        public static async Task<int> ClearGroupBlacklistAsync(long groupId)
        {
            return await ExecAsync($"DELETE FROM BlackList WHERE GroupId = {groupId}");
        }
    }
}
