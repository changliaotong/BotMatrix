using System.Threading.Tasks;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public class GroupMsgCount : MetaData<GroupMsgCount>
    {        
        public override string TableName => "MsgCount";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "UserId";

        public static bool ExistToday(long groupId, long userId) => ExistTodayAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<bool> ExistTodayAsync(long groupId, long userId)
        {
            return await ExistsWhereAsync($"GroupId = {groupId} AND UserId = {userId} AND CDate = CONVERT(DATE, GETDATE())");
        }

        public static int Append(long botUin, long groupId, string groupName, long userId, string name) => AppendAsync(botUin, groupId, groupName, userId, name).GetAwaiter().GetResult();

        public static async Task<int> AppendAsync(long botUin, long groupId, string groupName, long userId, string name)
        {
            return await InsertAsync([
                            new Cov("BotUin", botUin),
                            new Cov("GroupId", groupId),
                            new Cov("GroupName", groupName),
                            new Cov("UserId", userId),
                            new Cov("UserName", name),
                        ]);
        }

        public static int Update(long botUin, long groupId, string groupName, long userId, string name) => UpdateAsync(botUin, groupId, groupName, userId, name).GetAwaiter().GetResult();

        public static async Task<int> UpdateAsync(long botUin, long groupId, string groupName, long userId, string name)
        {
            if (!await ExistTodayAsync(groupId, userId))
                return await AppendAsync(botUin, groupId, groupName, userId, name);
            else
                return await UpdateWhereAsync($"MsgDate = GETDATE(), CMsg = CMsg+1", $"GroupId = {groupId} AND UserId = {userId} AND CDate = CONVERT(DATE, GETDATE())");
        }

        // 今日发言次数
        public static int GetMsgCount(long groupId, long qq) => GetMsgCountAsync(groupId, qq).GetAwaiter().GetResult();

        public static async Task<int> GetMsgCountAsync(long groupId, long qq)
        {
            return (await GetWhereAsync("CMsg", $"GroupId = {groupId} and UserId = {qq} and CDate = CONVERT(DATE, GETDATE())")).AsInt();
        }

        // 昨日发言次数
        public static int GetMsgCountY(long groupId, long qq) => GetMsgCountYAsync(groupId, qq).GetAwaiter().GetResult();

        public static async Task<int> GetMsgCountYAsync(long groupId, long qq)
        {
            return (await GetWhereAsync("CMsg", $"GroupId = {groupId} and UserId = {qq} and CDate = CONVERT(DATE, GETDATE()-1)")).AsInt();
        }

        // 今日发言排名
        public static int GetCountOrder(long groupId, long userId) => GetCountOrderAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<int> GetCountOrderAsync(long groupId, long userId)
        {
            return await QueryScalarAsync<int>($"select count(Id)+1 as res  from {FullName} " +
                            $"where GroupId = {groupId} and CDate = Convert(date, GETDATE()) " +
                            $"and CMsg > (select top 1 CMsg from sz84_robot..MsgCount " +
                            $"where GroupId = {groupId} and UserId = {userId} and CDate = CONVERT(DATE, GETDATE()))", null);
        }

        /// 昨日发言排名
        public static int GetCountOrderY(long groupId, long userId) => GetCountOrderYAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<int> GetCountOrderYAsync(long groupId, long userId)
        {
            return await QueryScalarAsync<int>($"select count(Id)+1 from {FullName} " +
                            $"where GroupId = {groupId} and CDate = Convert(date, GETDATE()-1) " +
                            $"and CMsg > (select top 1 CMsg from sz84_robot..MsgCount " +
                            $"where GroupId = {groupId} and UserId = {userId} and CDate = Convert(date, GETDATE()-1))", null);
        }

        // 今日发言榜前N名
        public static string GetCountList(long botUin, long groupId, long userId, long top) => GetCountListAsync(botUin, groupId, userId, top).GetAwaiter().GetResult();

        public static async Task<string> GetCountListAsync(long botUin, long groupId, long userId, long top)
        {
            if (!await GroupInfo.IsOwnerAsync(groupId, userId) && !BotInfo.IsAdmin(botUin, userId))
                return OwnerOnlyMsg;
            string res = await QueryResAsync($"select top {top} UserId, CMsg from {FullName} " +
                                  $"where GroupId = {groupId} and CDate = Convert(date, GETDATE()) order by CMsg desc",
                                  "【第{i}名】 [@:{0}] 发言：{1}\n");
            if (!res.Contains(userId.ToString()))
                res += $"【第{{今日发言排名}}名】 {{你2}} 发言：{{今日发言次数}}";
            res += "\n进入 后台 查看更多内容";
            return res;
        }

        // 昨日发言榜前N名
        public static string GetCountListY(long botUin, long groupId, long userId, long top) => GetCountListYAsync(botUin, groupId, userId, top).GetAwaiter().GetResult();

        public static async Task<string> GetCountListYAsync(long botUin, long groupId, long userId, long top)
        {
            if (!await GroupInfo.IsOwnerAsync(groupId, userId) && !BotInfo.IsAdmin(botUin, userId))
                return OwnerOnlyMsg;
            string res = await QueryResAsync($"select top {top} UserId, CMsg from {FullName} " +
                                  $"where GroupId = {groupId} and CDate = Convert(date, GETDATE()-1) order by CMsg desc",
                                  "【第{i}名】 [@:{0}] 发言：{1}\n");
            if (!res.Contains(userId.ToString()))
                res += "【第{{昨日发言排名}}名】 {{你2}} 发言：{{昨日发言次数}}";
            res += "\n进入 后台 查看更多内容";
            return res;
        }
    }
}
