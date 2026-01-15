using System;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities;

public partial class UserInfo
{
    public static async Task<(int Result, long TokensValue, int LogId)> AddTokensAsync(long botUin, long groupId, string groupName, long qq, string name, long tokensAdd, string tokensInfo, IDbTransaction? trans = null)
    {
        try
        {
            // 1. 确保用户存在 (必须使用同一事务)
            await Repository.AppendAsync(botUin, groupId, qq, name, await GroupInfo.GetGroupOwnerAsync(groupId), trans);

            // 2. 获取当前准确值并加锁 (FOR UPDATE)
            var tokensValue = await Repository.GetTokensForUpdateAsync(qq, trans);

            // 3. 如果是消耗算力，检查是否足够
            if (tokensAdd < 0 && tokensValue < Math.Abs(tokensAdd))
            {
                return (-2, tokensValue, 0); // -2 表示算力不足
            }

            // 4. 记录日志 (直接使用已获取的 tokensValue，避免再次查询)
            int logId = await TokensLog.AddLogAsync(botUin, groupId, groupName, qq, name, tokensAdd, tokensValue, tokensInfo, trans);

            // 5. 更新算力
            await Repository.AddTokensAsync(qq, tokensAdd, trans);

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
        using var wrapper = await Repository.BeginTransactionAsync(trans);
        try
        {
            var res = await AddTokensAsync(botUin, groupId, groupName, qq, name, tokensAdd, tokensInfo, wrapper.Transaction);
            await wrapper.CommitAsync();

            await SyncTokensCacheAsync(qq, res.TokensValue);
            return res;
        }
        catch (Exception ex)
        {
            Logger.Error($"[AddTokensTrans Error] {ex.Message}");
            await wrapper.RollbackAsync();
            if (trans != null) throw;
            return (-1, 0, 0);
        }
    }

    public static async Task SyncTokensCacheAsync(long qq, long newValue)
    {
        await Repository.SyncCacheFieldAsync(qq, "tokens", newValue);
    }

    public static async Task<long> GetTokensAsync(long qq)
    {
        return await Repository.GetTokensAsync(qq);
    }

    public static async Task<long> GetTokensForUpdateAsync(long qq, IDbTransaction? trans = null)
    {
        return await Repository.GetTokensForUpdateAsync(qq, trans);
    }

    public static async Task<string> GetTokensListAsync(long groupId, long qq, long top, BotData.Platform botType = BotData.Platform.QQ)
    {
        return await Repository.GetTokensListAsync(groupId, (int)top);
    }

    public static async Task<long> GetTokensRankingAsync(long groupId, long qq)
    {
        return await Repository.GetTokensRankingAsync(groupId, qq);
    }

    //消耗算力当天合计（单群）
    public static async Task<long> GetDayTokensGroupAsync(long groupId, long userId)
    {
        return await Repository.GetDayTokensGroupAsync(groupId, userId);
    }

    //消耗算力当天合计（所有）
    public static async Task<long> GetDayTokensAsync(long userId)
    {
        return await Repository.GetDayTokensAsync(userId);
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
