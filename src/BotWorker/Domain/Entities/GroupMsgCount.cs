using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public class GroupMsgCount : MetaData<GroupMsgCount>
    {        
        public override string TableName => "MsgCount";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "UserId";

        public static bool ExistToday(long groupId, long userId)
        {
            return ExistsWhere($"GroupId = {groupId} AND UserId = {userId} AND CDate = CONVERT(DATE, GETDATE())");
        }

        public static int Append(long botUin, long groupId, string groupName, long userId, string name)
        {
            return Insert([
                            new Cov("BotUin", botUin),
                            new Cov("GroupId", groupId),
                            new Cov("GroupName", groupName),
                            new Cov("UserId", userId),
                            new Cov("UserName", name),
                        ]);
        }

        public static int Update(long botUin, long groupId, string groupName, long userId, string name)
        {
            if (!ExistToday(groupId, userId))
                return Append(botUin, groupId, groupName, userId, name);
            else
                return UpdateWhere($"MsgDate = GETDATE(), CMsg = CMsg+1", $"GroupId = {groupId} AND UserId = {userId} AND CDate = CONVERT(DATE, GETDATE())");
        }

        // 今日发言次数
        public static int GetMsgCount(long groupId, long qq)
        {
            return GetWhere($"CMsg", $"GroupId = {groupId} and UserId = {qq} and CDate = CONVERT(DATE, GETDATE())").AsInt();
        }

        // 昨日发言次数
        public static int GetMsgCountY(long groupId, long qq)
        {
            return GetWhere($"CMsg", $"GroupId = {groupId} and UserId = {qq} and CDate = CONVERT(DATE, GETDATE()-1)").AsInt();            
        }

        // 今日发言排名
        public static int GetCountOrder(long groupId, long userId)
        {
            return Query<int>($"select count(Id)+1 as res  from {FullName} " +
                            $"where GroupId = {groupId} and CDate = Convert(date, GETDATE()) " +
                            $"and CMsg > (select top 1 CMsg from sz84_robot..MsgCount " +
                            $"where GroupId = {groupId} and UserId = {userId} and CDate = CONVERT(DATE, GETDATE()))");
        }

        /// 昨日发言排名
        public static int GetCountOrderY(long groupId, long userId)
        {
            return Query<int>($"select count(Id)+1 from {FullName} " +
                            $"where GroupId = {groupId} and CDate = Convert(date, GETDATE()-1) " +
                            $"and CMsg > (select top 1 CMsg from sz84_robot..MsgCount " +
                            $"where GroupId = {groupId} and UserId = {userId} and CDate = Convert(date, GETDATE()-1))");
        }

        // 今日发言榜前N名
        public static string GetCountList(long botUin, long groupId, long userId, long top)
        {
            if (!UserInfo.IsOwner(groupId, userId) && !BotInfo.IsAdmin(botUin, userId))
                return OwnerOnlyMsg;
            string res = QueryRes($"select top {top} UserId, CMsg from {FullName} " +
                                  $"where GroupId = {groupId} and CDate = Convert(date, GETDATE()) order by CMsg desc", 
                                  "【第{i}名】 [@:{0}] 发言：{1}\n");
            if (!res.Contains(userId.ToString()))  
                res += $"【第{{今日发言排名}}名】 {{你2}} 发言：{{今日发言次数}}";
            res += "\n进入 后台 查看更多内容";
            return res;
        }

        // 昨日发言榜前N名
        public static string GetCountListY(long botUin, long groupId, long userId, long top)
        {
            if (!UserInfo.IsOwner(groupId, userId) && !BotInfo.IsAdmin(botUin, userId))
                return OwnerOnlyMsg;
            string res = QueryRes($"select top {top} UserId, CMsg from {FullName} " +
                                  $"where GroupId = {groupId} and CDate = Convert(date, GETDATE()-1) order by CMsg desc", 
                                  "【第{i}名】 [@:{0}] 发言：{1}\n");
            if (!res.Contains(userId.ToString()))
                res += "【第{{昨日发言排名}}名】 {{你2}} 发言：{{昨日发言次数}}";
            res += "\n进入 后台 查看更多内容";
            return res;
        }
    }
}
