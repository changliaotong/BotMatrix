using Microsoft.Data.SqlClient;
using BotWorker.Bots.Entries;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Groups
{
    public class GroupSignIn : MetaData<GroupSignIn>
    {
        public override string TableName => "robot_weibo";
        public override string KeyField => "weibo_id";

        public static int Append(long botUin, long groupId, long qq, string weiboInfo)
        {
            return Insert([
                            new Cov("robot_qq", botUin),
                            new Cov("weibo_qq", qq),
                            new Cov("weibo_info", weiboInfo),
                            new Cov("weibo_type", 1),
                            new Cov("group_id", groupId),
                        ]);
        }

        // 今日签到人数
        public static long SignCount(long groupId)
        {
            return CountWhere($"weibo_type = 1 AND group_id = {groupId} AND ABS(DATEDIFF(DAY,insert_date,GETDATE())) = 0");
        }

        // 昨日签到人数
        public static long SignCountY(long groupId)
        {
            return CountWhere($"weibo_type = 1 AND group_id = {groupId} AND ABS(DATEDIFF(DAY,insert_date,GETDATE())) = 1");
        }

        // 本月签到次数
        public static long SignCountThisMonth(long groupId, long qq)
        {
            return CountWhere($"weibo_type = 1 AND group_id = {groupId} AND weibo_qq = {qq} AND ABS(DATEDIFF(MONTH,insert_date,GETDATE())) = 0");
        }
    }

    public partial class GroupMember : MetaData<GroupMember>
    {
        /// 签到榜
        public static string GetSignList(long groupId, int top = 3)
        {
            return QueryRes($"select top {top} UserId, SignTimes from {FullName} where GroupId = {groupId} order by SignTimes desc", "【第{i}名】 {0} {1}\n");
        }

        // 更新连续签到信息
        public static (string, SqlParameter[]) SqlUpdateSignInfo(long groupId, long qq, int signTimes, int signLevel)
        {
            return SqlSetValues($"SignDate=GETDATE(), SignTimes={signTimes}, SignLevel={signLevel}, SignTimesAll=SignTimesAll+1", groupId, qq);
        }

        public static int GetSignDateDiff(long groupId, long qq)
        {
            return GetInt("DATEDIFF(DAY, ISNULL(SignDate, GETDATE()-365), GETDATE())", groupId, qq);
        }

        // 连续签到天数
        public static int GetSignTimes(long groupId, long qq)
        {
            return GetInt("SignTimes", groupId, qq);
        }

        // 签到等级
        public static int GetSignLevel(long groupId, long qq)
        {
            return GetDef("SignLevel", groupId, qq, 1);
        }

        // 当天是否签到过
        public static bool IsSignIn(long groupId, long qq)
        {
            return GetSignDateDiff(groupId, qq) == 0;
        }

    }
}
