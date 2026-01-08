using BotWorker.Infrastructure.Persistence.Database;
using Microsoft.Data.SqlClient;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GroupRepository : IGroupRepository
    {
        public async Task<long> GetGroupOwnerAsync(long groupId)
        {
            var sql = "SELECT GroupOwner FROM [Group] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", groupId)]);
            return result == null || result == DBNull.Value ? 0 : Convert.ToInt64(result);
        }

        public async Task<long> GetRobotOwnerAsync(long groupId)
        {
            var sql = "SELECT RobotOwner FROM [Group] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", groupId)]);
            return result == null || result == DBNull.Value ? 0 : Convert.ToInt64(result);
        }

        public async Task<int> SetRobotOwnerAsync(long groupId, long ownerId)
        {
            var sql = "UPDATE [Group] SET RobotOwner = @ownerId WHERE Id = @id";
            return SQLConn.Exec(sql, [new SqlParameter("@ownerId", ownerId), new SqlParameter("@id", groupId)]);
        }

        public async Task<int> SetInGameAsync(long groupId, int isInGame)
        {
            var sql = "UPDATE [Group] SET IsInGame = @isInGame WHERE Id = @id";
            return SQLConn.Exec(sql, [new SqlParameter("@isInGame", isInGame), new SqlParameter("@id", groupId)]);
        }

        public async Task<int> UpdateGroupAsync(long groupId, string name, long selfId, long groupOwner = 0, long robotOwner = 0)
        {
            var sql = "UPDATE [Group] SET GroupName = @name, BotUin = @selfId, LastDate = GETDATE()";
            var parameters = new List<SqlParameter>
            {
                new SqlParameter("@name", name),
                new SqlParameter("@selfId", selfId),
                new SqlParameter("@id", groupId)
            };

            if (groupOwner != 0)
            {
                sql += ", GroupOwner = @groupOwner";
                parameters.Add(new SqlParameter("@groupOwner", groupOwner));
            }

            if (robotOwner != 0)
            {
                sql += ", RobotOwner = @robotOwner";
                parameters.Add(new SqlParameter("@robotOwner", robotOwner));
            }

            sql += " WHERE Id = @id";
            return SQLConn.Exec(sql, parameters.ToArray());
        }

        public async Task<bool> IsOpenAsync(long groupId)
        {
            var sql = "SELECT IsOpen FROM [Group] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", groupId)]);
            return result != null && result != DBNull.Value && Convert.ToBoolean(result);
        }

        public async Task<int> SetOpenStatusAsync(long groupId, bool isOpen)
        {
            var sql = "UPDATE [Group] SET IsOpen = @isOpen WHERE Id = @id";
            return SQLConn.Exec(sql, [new SqlParameter("@isOpen", isOpen ? 1 : 0), new SqlParameter("@id", groupId)]);
        }

        public async Task<int> GetVipRestDaysAsync(long groupId)
        {
            var sql = "SELECT VipRestDays FROM [Group] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", groupId)]);
            return result == null || result == DBNull.Value ? 0 : Convert.ToInt32(result);
        }

        public async Task<bool> IsSz84Async(long groupId)
        {
            var sql = "SELECT IsSz84 FROM [Group] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", groupId)]);
            return result != null && result != DBNull.Value && Convert.ToBoolean(result);
        }
    }
}
