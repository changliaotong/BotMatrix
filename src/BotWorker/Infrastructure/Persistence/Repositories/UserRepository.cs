using System;
using System.Threading.Tasks;
using System.Data;
using Microsoft.Data.SqlClient;
using BotWorker.Core.Database;

namespace BotWorker.Core.Repositories
{
    public class UserRepository : IUserRepository
    {
        public async Task<bool> IsSuperAdminAsync(long userId)
        {
            var sql = "SELECT IsSuper FROM [User] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", userId)]);
            return result != null && result != DBNull.Value && Convert.ToBoolean(result);
        }

        public async Task<bool> IsBlackAsync(long userId)
        {
            var sql = "SELECT IsBlack FROM [User] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", userId)]);
            return result != null && result != DBNull.Value && Convert.ToBoolean(result);
        }

        public async Task<int> SetIsBlackAsync(long userId, bool isBlack, string reason = "")
        {
            var sql = "UPDATE [User] SET IsBlack = @isBlack WHERE Id = @id";
            return SQLConn.Exec(sql, [new SqlParameter("@isBlack", isBlack ? 1 : 0), new SqlParameter("@id", userId)]);
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

        public async Task<int> SetStateAsync(long userId, int state)
        {
            var sql = "UPDATE [User] SET State = @state WHERE Id = @id";
            return SQLConn.Exec(sql, [new SqlParameter("@state", state), new SqlParameter("@id", userId)]);
        }

        public async Task<int> GetStateAsync(long userId)
        {
            var sql = "SELECT State FROM [User] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", userId)]);
            return result == null || result == DBNull.Value ? 0 : Convert.ToInt32(result);
        }

        public async Task<bool> ExistsAsync(long userId)
        {
            var sql = "SELECT COUNT(1) FROM [User] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", userId)]);
            return result != null && result != DBNull.Value && Convert.ToInt32(result) > 0;
        }

        public async Task<long> GetCreditAsync(long groupId, long userId)
        {
            var sql = groupId == 0 
                ? "SELECT Credit FROM [User] WHERE Id = @id" 
                : "SELECT Credit FROM GroupMember WHERE GroupId = @groupId AND UserId = @userId";
            
            var parameters = groupId == 0 
                ? new[] { new SqlParameter("@id", userId) }
                : new[] { new SqlParameter("@groupId", groupId), new SqlParameter("@userId", userId) };

            var result = SQLConn.ExecScalar(sql, parameters);
            return result == null || result == DBNull.Value ? 0 : Convert.ToInt64(result);
        }

        public async Task<int> AddCreditAsync(long botUin, long groupId, long userId, long credit, string reason)
        {
            // 这里应该包含更新积分和插入历史记录的逻辑
            // 简单实现：
            var sql = groupId == 0 
                ? "UPDATE [User] SET Credit = Credit + @credit WHERE Id = @id"
                : "UPDATE GroupMember SET Credit = Credit + @credit WHERE GroupId = @groupId AND UserId = @userId";

            var parameters = groupId == 0 
                ? new[] { new SqlParameter("@credit", credit), new SqlParameter("@id", userId) }
                : new[] { new SqlParameter("@credit", credit), new SqlParameter("@groupId", groupId), new SqlParameter("@userId", userId) };

            return SQLConn.Exec(sql, parameters);
        }

        public async Task<long> GetSaveCreditAsync(long userId)
        {
            var sql = "SELECT CreditSave FROM [User] WHERE Id = @id";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@id", userId)]);
            return result == null || result == DBNull.Value ? 0 : Convert.ToInt64(result);
        }

        public async Task<int> AddSaveCreditAsync(long userId, long credit, string reason)
        {
            var sql = "UPDATE [User] SET CreditSave = CreditSave + @credit WHERE Id = @id";
            return SQLConn.Exec(sql, [new SqlParameter("@credit", credit), new SqlParameter("@id", userId)]);
        }

        public async Task<long> GetCoinsAsync(int coinsType, long groupId, long userId)
        {
            // 根据 coinsType 获取对应的列名，这部分逻辑建议以后优化
            string[] coinsFields = { "GoldCoins", "BlackGoldCoins", "PurpleCoins", "GameCoins", "Credit" };
            if (coinsType < 0 || coinsType >= coinsFields.Length) return 0;

            var field = coinsFields[coinsType];
            var sql = $"SELECT {field} FROM GroupMember WHERE GroupId = @groupId AND UserId = @userId";
            var result = SQLConn.ExecScalar(sql, [new SqlParameter("@groupId", groupId), new SqlParameter("@userId", userId)]);
            return result == null || result == DBNull.Value ? 0 : Convert.ToInt64(result);
        }

        public async Task<int> AddCoinsAsync(int coinsType, long coinsValue, long groupId, long userId, string reason)
        {
            if (coinsType < 0 || coinsType >= CoinsLog.conisFields.Count)
                return -1;

            string fieldName = CoinsLog.conisFields[coinsType];
            
            // 1. 确保用户�?GroupMember 表中存在
            if (!Groups.GroupMember.Exists(groupId, userId))
            {
                Groups.GroupMember.Append(groupId, userId, "");
            }

            // 2. 获取基本信息（用于日志）
            var group = await Groups.GroupInfo.GetByKeyAsync(groupId);
            string groupName = group?.GroupName ?? "";
            long botUin = group?.BotUin ?? 0;

            // 3. 构建原子更新任务
            // 使用 TaskPlus 代替手动 SQL，这样可以利用我们的 ReSync 机制更新局部缓�?            var updateTask = Groups.GroupMember.TaskPlus(fieldName, coinsValue, groupId, userId);
            
            // 4. 构建日志 SQL
            long lastValue = 0;
            var (logSql, logParams) = CoinsLog.SqlCoins(botUin, groupId, groupName, userId, "", coinsType, coinsValue, ref lastValue, reason);

            // 5. 开启事务执�?            // ExecTrans 会自动检�?SqlTask 中的 NeedsReSync 并在成功后重读数据库同步 Redis
            return Groups.GroupMember.ExecTrans(updateTask, (logSql, logParams));
        }

        public async Task<System.Collections.Generic.IEnumerable<(long UserId, long Credit)>> GetCreditRankAsync(long groupId, int top = 10)
        {
            using var db = _dbFactory.CreateConnection();
            // 简化实现：默认�?UserInfo 表中查询
            string sql = $@"SELECT TOP (@top) Id AS UserId, Credit 
                            FROM UserInfo 
                            WHERE Id IN (SELECT UserId FROM GroupMember WHERE GroupId = @groupId)
                            ORDER BY Credit DESC";
            
            var result = await db.QueryAsync<(long UserId, long Credit)>(sql, new { groupId, top });
            return result;
        }
    }
}


