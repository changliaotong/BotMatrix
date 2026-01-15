using System;
using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Entities;

[Table("Balance")]
public class BalanceLog
{
    [Key]
    public long Id { get; set; }
    public long BotUin { get; set; }
    public long GroupId { get; set; }
    public string GroupName { get; set; } = string.Empty;
    public long UserId { get; set; }
    public string UserName { get; set; } = string.Empty;
    public decimal BalanceAdd { get; set; }
    public decimal BalanceValue { get; set; }
    public string BalanceInfo { get; set; } = string.Empty;
    public DateTime InsertDate { get; set; } = DateTime.Now;
}
