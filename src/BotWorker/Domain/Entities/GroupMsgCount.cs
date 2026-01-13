namespace BotWorker.Domain.Entities
{
    public class GroupMsgCount : MetaData<GroupMsgCount>
    {        
        public override string TableName => "MsgCount";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "UserId";

        public static async Task<bool> ExistTodayAsync(long groupId, long userId)
        {
            return await ExistsWhereAsync($"GroupId = {groupId} AND UserId = {userId} AND CDate = {SqlDate}");
        }

        public static async Task<int> AppendAsync(long botUin, long groupId, string groupName, long userId, string name)
        {
            return await InsertAsync([
                            new Cov("BotUin", botUin),
                            new Cov("GroupId", groupId),
                            new Cov("GroupName", groupName),
                            new Cov("UserId", userId),
                            new Cov("UserName", name),
                            new Cov("CDate", DateTime.Today),
                            new Cov("MsgDate", DateTime.Now),
                        ]);
        }

        public static int Update(long botUin, long groupId, string groupName, long userId, string name)
            => UpdateAsync(botUin, groupId, groupName, userId, name).GetAwaiter().GetResult();

        public static async Task<int> UpdateAsync(long botUin, long groupId, string groupName, long userId, string name)
        {
            if (!await ExistTodayAsync(groupId, userId))
                return await AppendAsync(botUin, groupId, groupName, userId, name);
            else
                return await UpdateWhereAsync($"MsgDate = {SqlDateTime}, CMsg = CMsg+1", $"GroupId = {groupId} AND UserId = {userId} AND CDate = {SqlDate}");
        }

        // 今日发言次数
        public static async Task<int> GetMsgCountAsync(long groupId, long qq)
        {
            return (await GetWhereAsync("CMsg", $"GroupId = {groupId} and UserId = {qq} and CDate = {SqlDate}")).AsInt();
        }

        // 昨日发言次数
        public static async Task<int> GetMsgCountYAsync(long groupId, long qq)
        {
            return (await GetWhereAsync("CMsg", $"GroupId = {groupId} and UserId = {qq} and CDate = {SqlYesterday}")).AsInt();
        }

        // 今日发言排名
        public static async Task<int> GetCountOrderAsync(long groupId, long userId)
        {
            return await QueryScalarAsync<int>($"select count(Id)+1 as res  from {FullName} " +
                            $"where GroupId = {groupId} and CDate = {SqlDate} " +
                            $"and CMsg > (select {SqlTop(1)}CMsg from {FullName} " +
                            $"where GroupId = {groupId} and UserId = {userId} and CDate = {SqlDate}{SqlLimit(1)})");
        }

        /// 昨日发言排名
        public static async Task<int> GetCountOrderYAsync(long groupId, long userId)
        {
            return await QueryScalarAsync<int>($"select count(Id)+1 from {FullName} " +
                            $"where GroupId = {groupId} and CDate = {SqlYesterday} " +
                            $"and CMsg > (select {SqlTop(1)}CMsg from {FullName} " +
                            $"where GroupId = {groupId} and UserId = {userId} and CDate = {SqlYesterday}{SqlLimit(1)})");
        }

        // 今日发言榜前N名
        public static async Task<string> GetCountListAsync(long botUin, long groupId, long userId, long top)
        {
            if (!await GroupInfo.IsOwnerAsync(groupId, userId) && !BotInfo.IsAdmin(botUin, userId))
                return OwnerOnlyMsg;
            string res = await QueryResAsync($"select {SqlTop(top)}UserId, CMsg from {FullName} " +
                                  $"where GroupId = {groupId} and CDate = {SqlDate} order by CMsg desc{SqlLimit(top)}",
                                  "【第{i}名】 [@:{0}] 发言：{1}\n");
            if (!res.Contains(userId.ToString()))
                res += $"【第{{今日发言排名}}名】 {{你2}} 发言：{{今日发言次数}}";
            res += "\n进入 后台 查看更多内容";
            return res;
        }

        // 昨日发言榜前N名
        public static async Task<string> GetCountListYAsync(long botUin, long groupId, long userId, long top)
        {
            if (!await GroupInfo.IsOwnerAsync(groupId, userId) && !BotInfo.IsAdmin(botUin, userId))
                return OwnerOnlyMsg;
            string res = await QueryResAsync($"select {SqlTop(top)}UserId, CMsg from {FullName} " +
                                  $"where GroupId = {groupId} and CDate = {SqlYesterday} order by CMsg desc{SqlLimit(top)}",
                                  "【第{i}名】 [@:{0}] 发言：{1}\n");
            if (!res.Contains(userId.ToString()))
                res += "【第{{昨日发言排名}}名】 {{你2}} 发言：{{昨日发言次数}}";
            res += "\n进入 后台 查看更多内容";
            return res;
        }
    }
}
