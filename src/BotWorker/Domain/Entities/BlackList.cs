using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

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

        public static bool IsSystemBlack(long userId)
        {
            return Exists(BotInfo.GroupIdDef, userId);
        }

        // 加入黑名单
        public static int AddBlackList(long botUin, long groupId, string groupName, long qq, string name, long blackQQ, string blackInfo)
        {
            return Exists(groupId, blackQQ)
                ? 0
                : Insert([
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
        public static int ClearGroupBlacklist(long groupId)
        {
            return Exec($"DELETE FROM BlackList WHERE GroupId = {groupId}");
        }
    }
}
