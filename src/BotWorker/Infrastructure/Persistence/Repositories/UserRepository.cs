using BotWorker.Infrastructure.Persistence.Database;
using Microsoft.Data.SqlClient;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class UserRepository : IUserRepository
    {
        public async Task<bool> IsBlacklistedAsync(long userId, long groupId)
        {
            var sql = "SELECT IsBlack FROM [User] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", userId)]);
            return result != null && result != DBNull.Value && Convert.ToBoolean(result);
        }

        public async Task<int> AddToBlacklistAsync(long userId, long groupId, string reason)
        {
            var sql = "UPDATE [User] SET IsBlack = 1, BlackReason = @reason WHERE Id = @id";
            return SQLConn.Exec(sql, [new SqlParameter("@reason", reason), new SqlParameter("@id", userId)]);
        }

        public async Task<int> GetPointsAsync(long userId)
        {
            var sql = "SELECT Credit FROM [User] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", userId)]);
            return result == null || result == DBNull.Value ? 0 : Convert.ToInt32(result);
        }

        public async Task<int> UpdatePointsAsync(long userId, int delta)
        {
            var sql = "UPDATE [User] SET Credit = Credit + @delta WHERE Id = @id";
            return SQLConn.Exec(sql, [new SqlParameter("@delta", delta), new SqlParameter("@id", userId)]);
        }

        public async Task<bool> HasSignedIdAsync(long userId, string date)
        {
            // 这里假设有一个签到记录表，简化实现：
            var sql = "SELECT COUNT(1) FROM UserSignLog WHERE UserId = @userId AND SignDate = @date";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@userId", userId), new SqlParameter("@date", date)]);
            return result != null && result != DBNull.Value && Convert.ToInt32(result) > 0;
        }

        // 保留原有的辅助方法，虽然不在 IUserRepository 接口中定义，但可能被内部使用
        public async Task<bool> IsSuperAdminAsync(long userId)
        {
            var sql = "SELECT IsSuper FROM [User] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", userId)]);
            return result != null && result != DBNull.Value && Convert.ToBoolean(result);
        }

        public async Task<int> AddUserAsync(long botUin, long groupId, long userId, string name, long refUserId, string userOpenId = "", string groupOpenId = "")
        {
            var sql = @"INSERT INTO [User] (BotUin, GroupId, Id, Name, RefUserId, UserOpenid, GroupOpenid, Credit, InsertDate) 
                        VALUES (@botUin, @groupId, @id, @name, @refUserId, @userOpenid, @groupOpenid, @credit, GETDATE())";
            
            var parameters = new[]
            {
                new SqlParameter("@botUin", botUin),
                new SqlParameter("@groupId", groupId),
                new SqlParameter("@id", userId),
                new SqlParameter("@name", name),
                new SqlParameter("@refUserId", refUserId),
                new SqlParameter("@userOpenid", userOpenId),
                new SqlParameter("@groupOpenid", groupOpenId),
                new SqlParameter("@credit", string.IsNullOrEmpty(userOpenId) ? 50 : 5000)
            };

            return SQLConn.Exec(sql, parameters);
        }
    }
}
