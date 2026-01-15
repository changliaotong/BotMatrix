using System;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;
using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Entities;

[Table("credit_log")]
public class CreditLog
{
    private static ICreditLogRepository Repository => 
        BotMessage.ServiceProvider?.GetRequiredService<ICreditLogRepository>() 
        ?? throw new InvalidOperationException("ICreditLogRepository not registered");

    [Key]
    public long Id { get; set; }
    public long BotUin { get; set; }
    public long GroupId { get; set; }
    public string GroupName { get; set; } = string.Empty;
    public long UserId { get; set; }
    public string UserName { get; set; } = string.Empty;
    public long CreditAdd { get; set; }
    public long CreditValue { get; set; }
    public string CreditInfo { get; set; } = string.Empty;
    public DateTime CreatedAt { get; set; } = DateTime.Now;

    //异步增加日志 (支持事务)
    public static async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, long creditValue, string creditInfo, IDbTransaction? trans = null)
    {
        return await Repository.AddLogAsync(botUin, groupId, groupName, qq, name, creditAdd, creditValue, creditInfo, trans);
    }

    public static async Task<int> CreditCountAsync(long userId, string creditInfo, int second = 60)
    {
        return await Repository.CreditCountAsync(userId, creditInfo, second);
    }
}