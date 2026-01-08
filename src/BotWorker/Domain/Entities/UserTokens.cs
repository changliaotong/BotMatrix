using Microsoft.Data.SqlClient;

namespace BotWorker.Domain.Entities;

public partial class UserInfo : MetaDataGuid<UserInfo>
{
    public static (string, SqlParameter[]) SqlAddTokens(long userId, float tokens)
    {
        return Exists(userId)
            ? SqlPlus("Tokens", tokens, userId)
            : SqlInsert([
                            new Cov("UserId", userId),
                                new Cov("Tokens", tokens),
                        ]);
    }

    public static string GetTokensList(long groupId, long qq, long top, BotData.Platform botType = BotData.Platform.NapCat)
    {
        var format = ((int)botType).In(0, 1) ? "{i} [@:{0}]：{1}\n" : "{i} {0} {1}\n";
        var res = QueryWhere($"TOP {top} UserId, Tokens", $"UserId IN (SELECT UserId FROM {GroupMember.FullName} WHERE GroupId = {groupId})",
                             $"Tokens desc", format);
        if (!res.Contains(qq.ToString()))
            res += $"{{算力排名}} [@:{qq}]：{{算力}}\n";
        return res;
    }

    public static long GetTokensRanking(long groupId, long qq)
    {
        return CountWhere($"tokens > {GetTokens(qq)} and UserId in (SELECT UserId FROM {GroupMember.FullName} WHERE GroupId = {groupId})") + 1;
    }

    public static long GetTokens(long qq)
    {
        return Exists(qq)
            ? GetLong("tokens", qq)
            : 0;
    }

    //消耗算力当天合计（单群）
    public static long GetDayTokensGroup(long groupId, long userId)
    {
        var sql = $"SELECT SUM(TokensAdd) FROM {TokensLog.FullName} WHERE GroupId = {groupId} AND UserId = {userId} " +
                  $"AND ABS(DATEDIFF(DAY, InsertDate, GETDATE())) = 0 AND TokensAdd < 0";
        return Query<long>(sql);
    }

    //消耗算力当天合计（所有）
    public static long GetDayTokens(long userId)
    {
        var sql = $"SELECT SUM(TokensAdd) FROM {TokensLog.FullName} WHERE UserId = {userId} " +
                  $"AND ABS(DATEDIFF(DAY, InsertDate, GETDATE())) = 0 AND TokensAdd < 0";
        return Query<long>(sql);
    }

    public static int AddTokens(long botUin, long groupId, string groupName, long qq, string name, long tokensAdd, string tokensInfo)
    {
        Append(botUin, groupId, qq, name, GroupInfo.GetGroupOwner(groupId));
        var tokensValue = GetTokens(qq) + tokensAdd;
        var sql = TokensLog.SqlLog(botUin, groupId, groupName, qq, name, tokensAdd, tokensInfo);
        var sql2 = SqlPlus("tokens", tokensAdd, qq);
        int i = ExecTrans(sql, sql2);
        if (i == 0)
        {
            SyncCacheField(qq, "Tokens", tokensValue);
        }
        return i;
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