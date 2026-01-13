namespace BotWorker.Domain.Entities
{
    public class GroupSignIn : MetaData<GroupSignIn>
    {
        public override string TableName => "robot_weibo";
        public override string KeyField => "weibo_id";

        public static async Task<int> AppendAsync(long botUin, long groupId, long qq, string weiboInfo, System.Data.IDbTransaction? trans = null)
        {
            return await InsertAsync(new List<Cov> {
                            new Cov("robot_qq", botUin),
                            new Cov("weibo_qq", qq),
                            new Cov("weibo_info", weiboInfo),
                            new Cov("weibo_type", 1),
                            new Cov("group_id", groupId),
                            new Cov("insert_date", DateTime.MinValue),
                        }, trans);
        }

        // 今日签到人数
        public static async Task<long> SignCountAsync(long groupId)
        {
            return await CountWhereAsync($"weibo_type = 1 AND group_id = {groupId} AND ABS({SqlDateDiff("DAY", "insert_date", SqlDateTime)}) = 0");
        }

        // 昨日签到人数
        public static async Task<long> SignCountYAsync(long groupId)
        {
            return await CountWhereAsync($"weibo_type = 1 AND group_id = {groupId} AND ABS({SqlDateDiff("DAY", "insert_date", SqlDateTime)}) = 1");
        }

        // 本月签到次数
        public static async Task<long> SignCountThisMonthAsync(long groupId, long qq)
        {
            return await CountWhereAsync($"weibo_type = 1 AND group_id = {groupId} AND weibo_qq = {qq} AND ABS({SqlDateDiff("MONTH", "insert_date", SqlDateTime)}) = 0");
        }
    }
}
