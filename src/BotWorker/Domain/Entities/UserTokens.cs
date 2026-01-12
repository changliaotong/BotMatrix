using System.Data;

namespace BotWorker.Domain.Entities;

public partial class UserInfo : MetaDataGuid<UserInfo>
{
    public static async Task<(string, IDataParameter[])> SqlAddTokensAsync(long userId, float tokens, IDbTransaction? trans = null)
    {
        return await ExistsAsync(userId, trans)
            ? SqlPlus("Tokens", tokens, userId)
            : SqlInsert(new List<Cov> {
                            new Cov("UserId", userId),
                            new Cov("Tokens", tokens),
                        });
    }

    public static async Task<(int Result, long TokensValue, int LogId)> AddTokensAsync(long botUin, long groupId, string groupName, long qq, string name, long tokensAdd, string tokensInfo, IDbTransaction? trans = null)
    {
        try
        {
            // 1. 确保用户存在 (必须使用同一事务)
            await AppendAsync(botUin, groupId, qq, name, await GroupInfo.GetGroupOwnerAsync(groupId), trans: trans);

            // 2. 获取当前准确值并加锁 (UPDLOCK)
            // 这一步是防止死锁的关键：先锁定 User 行，后续所有操作都在此锁保护下
            var tokensValue = await GetTokensForUpdateAsync(qq, trans);

            // 3. 如果是消耗算力，检查是否足够
            if (tokensAdd < 0 && tokensValue < Math.Abs(tokensAdd))
            {
                return (-2, tokensValue, 0); // -2 表示算力不足
            }

            // 4. 记录日志 (直接使用已获取的 tokensValue，避免再次查询)
            int logId = await TokensLog.AddLogAsync(botUin, groupId, groupName, qq, name, tokensAdd, tokensValue, tokensInfo, trans);

            // 5. 更新算力
            var (sql2, paras2) = SqlPlus("tokens", tokensAdd, qq);
            await ExecAsync(sql2, trans, paras2);

            return (0, tokensValue + tokensAdd, logId);
        }
        catch (Exception ex)
        {
            Logger.Error($"[AddTokens Error] {ex.Message}\n{ex.StackTrace}");
            if (trans != null) throw; // 事务嵌套时抛出异常，由外层事务处理回滚
            return (-1, 0, 0);
        }
    }

    public static async Task<(int Result, long TokensValue, int LogId)> AddTokensTransAsync(long botUin, long groupId, string groupName, long qq, string name, long tokensAdd, string tokensInfo, IDbTransaction? trans = null)
    {
        using var wrapper = await BeginTransactionAsync(trans);
        try
        {
            var res = await AddTokensAsync(botUin, groupId, groupName, qq, name, tokensAdd, tokensInfo, wrapper.Transaction);
            wrapper.Commit();

            await SyncTokensCacheAsync(qq, res.TokensValue);
            return res;
        }
        catch (Exception ex)
        {
            Logger.Error($"[AddTokensTrans Error] {ex.Message}");
            wrapper.Rollback();
            if (trans != null) throw;
            return (-1, 0, 0);
        }
    }

    public static async Task SyncTokensCacheAsync(long qq, long newValue)
    {
        SyncCacheField(qq, "Tokens", newValue);
        await Task.CompletedTask;
    }

    public static async Task<long> GetTokensAsync(long qq)
    {
        return await ExistsAsync(qq)
            ? await GetLongAsync("tokens", qq)
            : 0;
    }

    public static async Task<long> GetTokensForUpdateAsync(long qq, IDbTransaction? trans = null)
    {
        return await GetForUpdateAsync<long>("tokens", qq, null, 0, trans);
    }

    public static async Task<string> GetTokensListAsync(long groupId, long qq, long top, BotData.Platform botType = BotData.Platform.QQ)
    {
        return await QueryResAsync(
            $"SELECT {SqlTop((int)top)} [UserId], [Tokens] FROM {FullName} " +
            $"WHERE [UserId] IN (SELECT [UserId] FROM {GroupMember.FullName} WHERE [GroupId] = {0}) " +
            $"ORDER BY [Tokens] DESC {SqlLimit((int)top)}",
            "【第{i}名】 [@:{0}] 算力：{1}\n",
            groupId);
    }

    public static async Task<long> GetTokensRankingAsync(long groupId, long qq)
    {
        return await CountWhereAsync($"tokens > {await GetTokensAsync(qq)} and UserId in (SELECT UserId FROM {GroupMember.FullName} WHERE GroupId = {groupId})") + 1;
    }

    //消耗算力当天合计（单群）
    public static async Task<long> GetDayTokensGroupAsync(long groupId, long userId)
    {
        var sql = $"SELECT SUM(TokensAdd) FROM {TokensLog.FullName} WHERE GroupId = {groupId} AND UserId = {userId} " +
                  $"AND ABS({SqlDateDiff("DAY", "InsertDate", SqlDateTime)}) = 0 AND TokensAdd < 0";
        return await QueryScalarAsync<long>(sql);
    }

    //消耗算力当天合计（所有）
    public static async Task<long> GetDayTokensAsync(long userId)
    {
        var sql = $"SELECT SUM(TokensAdd) FROM {TokensLog.FullName} WHERE UserId = {userId} " +
                  $"AND ABS({SqlDateDiff("DAY", "InsertDate", SqlDateTime)}) = 0 AND TokensAdd < 0";
        return await QueryScalarAsync<long>(sql);
    }

    public static async Task<int> MinusTokensAsync(long botUin, long groupId, string groupName, long qq, string name, long minus, string tokensInfo)
    {
        var res = await AddTokensTransAsync(botUin, groupId, groupName, qq, name, -minus, tokensInfo);
        return res.Result;
    }

    public static async Task<string> MinusTokensResAsync(long botUin, long groupId, string groupName, long qq, string name, long minus, string tokensInfo)
    {
        return await MinusTokensAsync(botUin, groupId, groupName, qq, name, minus, tokensInfo) == -1
            ? ""
            : ""; 
    }
}