using System;
using System.Collections.Generic;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities;

[Table("coins")]
public class CoinsLog
{
    private static ICoinsLogRepository Repository => 
        BotMessage.ServiceProvider?.GetRequiredService<ICoinsLogRepository>() 
        ?? throw new InvalidOperationException("ICoinsLogRepository not registered");

    [Key]
    public long Id { get; set; }
    public long BotUin { get; set; }
    public long GroupId { get; set; }
    public string GroupName { get; set; } = string.Empty;
    public long UserId { get; set; }
    public string UserName { get; set; } = string.Empty;
    public int CoinsType { get; set; }
    public long CoinsAdd { get; set; }
    public long CoinsValue { get; set; }
    public string CoinsInfo { get; set; } = string.Empty;
    public DateTime InsertDate { get; set; } = DateTime.Now;

    public enum CoinsTypeEnum { goldCoins, blackCoins, purpleCoins, gameCoins, groupCredit }
    public static List<string> conisFields = ["GoldCoins", "BlackCoins", "PurpleCoins", "GameCoins", "GroupCredit"];
    public static List<string> conisNames = ["金币", "黑金币", "紫币", "游戏币", "本群积分"];

    //异步增加日志 (支持事务)
    public static async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long qq, string name, int coinsType, long coinsAdd, long coinsValue, string coinsInfo, IDbTransaction? trans = null)
    {
        return await Repository.AddLogAsync(botUin, groupId, groupName, qq, name, coinsType, coinsAdd, coinsValue, coinsInfo, trans);
    }
}