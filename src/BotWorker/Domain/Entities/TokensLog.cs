using System;
using System.Collections.Generic;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities;

[Table("token_log")]
public class TokensLog
{
    private static ITokensLogRepository Repository => 
        BotMessage.ServiceProvider?.GetRequiredService<ITokensLogRepository>() 
        ?? throw new InvalidOperationException("ITokensLogRepository not registered");

    [Key]
    public long Id { get; set; }
    public long BotUin { get; set; }
    public long GroupId { get; set; }
    public string GroupName { get; set; } = string.Empty;
    public long UserId { get; set; }
    public string UserName { get; set; } = string.Empty;
    public long TokensAdd { get; set; }
    public long TokensValue { get; set; }
    public string TokensInfo { get; set; } = string.Empty;
    public DateTime InsertDate { get; set; } = DateTime.Now;

    //异步增加日志 (支持事务)
    public static async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, long tokensAdd, long tokensValue, string tokensInfo, IDbTransaction? trans = null)
    {
        return await Repository.AddLogAsync(botUin, groupId, groupName, userId, userName, tokensAdd, tokensValue, tokensInfo, trans);
    }
}