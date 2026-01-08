namespace BotWorker.Domain.Entities
{
    public class GroupSignIn : MetaData<GroupSignIn>
    {
        public override string TableName => "robot_weibo";
        public override string KeyField => "weibo_id";

        public static async Task<int> AppendAsync(long botUin, long groupId, long qq, string weiboInfo)
        {
            return await InsertAsync(new List<Cov> {
                            new Cov("robot_qq", botUin),
                            new Cov("weibo_qq", qq),
                            new Cov("weibo_info", weiboInfo),
                            new Cov("weibo_type", 1),
                            new Cov("group_id", groupId),
                        });
        }

        public static int Append(long botUin, long groupId, long qq, string weiboInfo)
        {
            return AppendAsync(botUin, groupId, qq, weiboInfo).GetAwaiter().GetResult();
        }

        // 今日签到人数
        public static async Task<long> SignCountAsync(long groupId)
        {
            return await CountWhereAsync($"weibo_type = 1 AND group_id = {groupId} AND ABS(DATEDIFF(DAY,insert_date,GETDATE())) = 0");
        }

        public static long SignCount(long groupId)
        {
            return SignCountAsync(groupId).GetAwaiter().GetResult();
        }

        // 昨日签到人数
        public static async Task<long> SignCountYAsync(long groupId)
        {
            return await CountWhereAsync($"weibo_type = 1 AND group_id = {groupId} AND ABS(DATEDIFF(DAY,insert_date,GETDATE())) = 1");
        }

        public static long SignCountY(long groupId)
        {
            return SignCountYAsync(groupId).GetAwaiter().GetResult();
        }

        // 本月签到次数
        public static async Task<long> SignCountThisMonthAsync(long groupId, long qq)
        {
            return await CountWhereAsync($"weibo_type = 1 AND group_id = {groupId} AND weibo_qq = {qq} AND ABS(DATEDIFF(MONTH,insert_date,GETDATE())) = 0");
        }

        public static long SignCountThisMonth(long groupId, long qq)
        {
            return SignCountThisMonthAsync(groupId, qq).GetAwaiter().GetResult();
        }
    }
}
