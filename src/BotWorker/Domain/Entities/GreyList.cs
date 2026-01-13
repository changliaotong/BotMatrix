namespace BotWorker.Domain.Entities
{
    public class GreyList : MetaData<GreyList>
    {
        public override string TableName => "GreyList";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "GreyId";

        // 灰名单指令：灰、加灰、删灰、取消灰、解除灰名单…
        public const string regexGrey = @"^(?<cmdName>(取消|解除|删除)?(灰名单|灰|加灰|删灰))(?<cmdPara>([ ]*(\[?@:?)?[1-9]+\d*(\]?))+)$";

        // 系统灰名单（通用灰名单）
        public static async Task<List<long>> GetSystemGreyListAsync()
        {
            var (sql, parameters) = SqlSelect("GreyId", BotInfo.GroupIdDef);
            return await QueryListAsync<long>(sql, null, parameters);
        }

        public static bool IsSystemGrey(long userId)
        {
            return Exists(BotInfo.GroupIdDef, userId);
        }

        // 加入灰名单
        public static int AddGreyList(long botUin, long groupId, string groupName, long qq, string name, long greyQQ, string greyInfo)
        {
            return Exists(groupId, greyQQ)
                ? 0
                : Insert([
                            new Cov("BotUin", botUin),
                            new Cov("GroupId", groupId),
                            new Cov("GroupName", groupName),
                            new Cov("UserId", qq),
                            new Cov("UserName", name),
                            new Cov("GreyId", greyQQ),
                            new Cov("GreyInfo", greyInfo),
                        ]);
        }
    }
}
