using System.Data;

namespace BotWorker.Domain.Entities;

public partial class UserInfo : MetaDataGuid<UserInfo>
{
    public static (string, IDataParameter[]) SqlAddTokens(long userId, float tokens)
    {
        return Exists(userId)
            ? SqlPlus("Tokens", tokens, userId)
            : SqlInsert(new List<Cov> {
                            new Cov("UserId", userId),
                                new Cov("Tokens", tokens),
                        });
    }

    public static async Task<(int Result, long TokensValue)> AddTokensAsync(long botUin, long groupId, string groupName, long qq, string name, long tokensAdd, string tokensInfo, IDbTransaction? trans = null)
    {
        await AppendAsync(botUin, groupId, qq, name, GroupInfo.GetGroupOwner(groupId));
        
        bool isNewTrans = false;
        if (trans == null)
        {
            trans = await BeginTransactionAsync();
            isNewTrans = true;
        }

        try
        {
            var (sql1, paras1) = TokensLog.SqlLog(botUin, groupId, groupName, qq, name, tokensAdd, tokensInfo);
            await ExecAsync(sql1, trans, paras1);

            var (sql2, paras2) = SqlPlus("tokens", tokensAdd, qq);
            await ExecAsync(sql2, trans, paras2);

            if (isNewTrans) await trans.CommitAsync();

            var tokensValue = GetTokens(qq);
            SyncCacheField(qq, "Tokens", tokensValue);
            return (0, tokensValue);
        }
        catch
        {
            if (isNewTrans) await trans.RollbackAsync();
            return (-1, 0);
        }
        finally
        {
            if (isNewTrans)
            {
                trans.Connection?.Close();
                trans.Dispose();
            }
        }
    }

    public static async Task<long> GetTokensAsync(long qq)
    {
        return await ExistsAsync(qq)
            ? await GetLongAsync("tokens", qq)
            : 0;
    }

    public static long GetTokens(long qq) => GetTokensAsync(qq).GetAwaiter().GetResult();

    public static async Task<string> GetTokensListAsync(long groupId, long qq, long top, BotData.Platform botType = BotData.Platform.NapCat)
    {
        var format = ((int)botType).In(0, 1) ? "{i} [@:{0}]：{1}\n" : "{i} {0} {1}\n";
        var res = await QueryResAsync($"select {SqlTop(top)} UserId, Tokens from {FullName} WHERE UserId IN (SELECT UserId FROM {GroupMember.FullName} WHERE GroupId = {groupId}) order by Tokens desc {SqlLimit(top)}", format);
        if (!res.Contains(qq.ToString()))
            res += $"{await GetTokensRankingAsync(groupId, qq)} [@:{qq}]：{await GetTokensAsync(qq)}\n";
        return res;
    }

    public static string GetTokensList(long groupId, long qq, long top, BotData.Platform botType = BotData.Platform.NapCat) => GetTokensListAsync(groupId, qq, top, botType).GetAwaiter().GetResult();

    public static async Task<long> GetTokensRankingAsync(long groupId, long qq)
    {
        return await CountWhereAsync($"tokens > {await GetTokensAsync(qq)} and UserId in (SELECT UserId FROM {GroupMember.FullName} WHERE GroupId = {groupId})") + 1;
    }

    public static long GetTokensRanking(long groupId, long qq) => GetTokensRankingAsync(groupId, qq).GetAwaiter().GetResult();

    //消耗算力当天合计（单群）
    public static long GetDayTokensGroup(long groupId, long userId)
    {
        var sql = $"SELECT SUM(TokensAdd) FROM {TokensLog.FullName} WHERE GroupId = {groupId} AND UserId = {userId} " +
                  $"AND ABS({SqlDateDiff("DAY", "InsertDate", SqlDateTime)}) = 0 AND TokensAdd < 0";
        return QueryScalar<long>(sql);
    }

    //消耗算力当天合计（所有）
    public static long GetDayTokens(long userId)
    {
        var sql = $"SELECT SUM(TokensAdd) FROM {TokensLog.FullName} WHERE UserId = {userId} " +
                  $"AND ABS({SqlDateDiff("DAY", "InsertDate", SqlDateTime)}) = 0 AND TokensAdd < 0";
        return QueryScalar<long>(sql);
    }

    public static int AddTokens(long botUin, long groupId, string groupName, long qq, string name, long tokensAdd, string tokensInfo)
    {
        return AddTokensAsync(botUin, groupId, groupName, qq, name, tokensAdd, tokensInfo).GetAwaiter().GetResult().Result;
    }

    public static int MinusTokens(long botUin, long groupId, string groupName, long qq, string name, long minus, string tokensInfo)
    {
        return AddTokens(botUin, groupId, groupName, qq, name, -minus, tokensInfo);
    }

    public static int MinusTokens(BotMessage bm, long minus, string tokensInfo)
    {
        return AddTokens(bm.SelfId, bm.RealGroupId, bm.GroupName, bm.UserId, bm.Name, -minus, tokensInfo);
    }

    public static string MinusTokensRes(long botUin, long groupId, string groupName, long qq, string name, long minus, string tokensInfo)
    {
        return MinusTokens(botUin, groupId, groupName, qq, name, minus, tokensInfo) == -1
            ? ""
            : ""; // $"\n算力:-{minus}，累计：{{TOKENS}}";
    }
}