namespace BotWorker.Application.Services
{
    public class LimiterService : ILimiter
    {
        public async Task<bool> HasUsedAsync(long? groupId, long userId, string actionKey)
        {
            var sql = "SELECT UsedAt FROM LimiterLogs WHERE (GroupId = @GroupId OR (@GroupId IS NULL AND GroupId IS NULL)) AND UserId = @UserId AND ActionKey = @ActionKey";
            var usedAt = await SQLConn.QueryScalarAsync<DateTime?>(sql, 
                SQLConn.CreateParameters(
                    ("@GroupId", (object?)groupId ?? DBNull.Value),
                    ("@UserId", userId),
                    ("@ActionKey", actionKey)
                ));

            return usedAt != null && usedAt.Value.Date == DateTime.Today;
        }

        public async Task MarkUsedAsync(long? groupId, long userId, string actionKey)
        {
            var checkSql = "SELECT Id FROM LimiterLogs WHERE (GroupId = @GroupId OR (@GroupId IS NULL AND GroupId IS NULL)) AND UserId = @UserId AND ActionKey = @ActionKey";
            var id = await SQLConn.QueryScalarAsync<int?>(checkSql,
                SQLConn.CreateParameters(
                    ("@GroupId", (object?)groupId ?? DBNull.Value),
                    ("@UserId", userId),
                    ("@ActionKey", actionKey)
                ));

            if (id == null)
            {
                var insertSql = "INSERT INTO LimiterLogs (GroupId, UserId, ActionKey, UsedAt) VALUES (@GroupId, @UserId, @ActionKey, @UsedAt)";
                await SQLConn.ExecAsync(insertSql,
                    SQLConn.CreateParameters(
                        ("@GroupId", (object?)groupId ?? DBNull.Value),
                        ("@UserId", userId),
                        ("@ActionKey", actionKey),
                        ("@UsedAt", DateTime.Now)
                    ));
            }
            else
            {
                var updateSql = "UPDATE LimiterLogs SET UsedAt = @UsedAt WHERE Id = @Id";
                await SQLConn.ExecAsync(updateSql,
                    SQLConn.CreateParameters(
                        ("@UsedAt", DateTime.Now),
                        ("@Id", id.Value)
                    ));
            }
        }

        public async Task<DateTime?> GetLastUsedAsync(long? groupId, long userId, string actionKey)
        {
            var sql = "SELECT UsedAt FROM LimiterLogs WHERE (GroupId = @GroupId OR (@GroupId IS NULL AND GroupId IS NULL)) AND UserId = @UserId AND ActionKey = @ActionKey";
            return await SQLConn.QueryScalarAsync<DateTime?>(sql,
                SQLConn.CreateParameters(
                    ("@GroupId", (object?)groupId ?? DBNull.Value),
                    ("@UserId", userId),
                    ("@ActionKey", actionKey)
                ));
        }

        public async Task<bool> TryUseAsync(long? groupId, long userId, string actionKey)
        {
            if (await HasUsedAsync(groupId, userId, actionKey))
                return false;

            await MarkUsedAsync(groupId, userId, actionKey);
            return true;
        }
    }
}


