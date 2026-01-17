using System;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;
using Newtonsoft.Json;

namespace BotWorker.Domain.Entities;

[Table("user_info")]
    public partial class UserInfo
    {
        [Key]
    public long Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public string UserOpenId { get; set; } = string.Empty;
    public DateTime InsertDate { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public long Credit { get; set; }
    public long CreditTotal => Credit + SaveCredit;
    [JsonIgnore]
    [HighFrequency]
    public long CreditFreeze { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public long Coins { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public long CoinsFreeze { get; set; }
    public int Sz84Uid { get; set; }
    public string Sz84UserName { get; set; } = string.Empty;
    public int HomeUid { get; set; }
    public string HomeUserName { get; set; } = string.Empty;
    public string HomeRealName { get; set; } = string.Empty;
    public long BotUin { get; set; }
    public int State { get; set; }
    public int IsOpen { get; set; }
    public string SzTong { get; set; } = string.Empty;
    public string CityName { get; set; } = string.Empty;
    public long DefaultGroup { get; set; }
    public bool IsDefaultHint { get; set; }
    public DateTime BindDate { get; set; }
    public DateTime BindDateHome { get; set; }
    public bool IsBlack { get; set; }
    public int TeachLevel { get; set; }
    public bool IsCoins { get; set; }
    public bool XCredit { get; set; }
    public int RCq { get; set; }
    public long RefUserId { get; set; }
    public Guid UserGuid { get; set; }
    public bool IsBlock { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public long SaveCredit { get; set; }
    public DateTime UpgradeDate { get; set; }
    public long PartnerUserId { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public long FreezeCredit { get; set; }
    public DateTime VipStart { get; set; }
    public DateTime VipEnd { get; set; }
    public string LastChengyu { get; set; } = string.Empty;
    [JsonIgnore]
    [HighFrequency]
    public decimal Balance { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public decimal BalanceFreeze { get; set; }
    public long CreditGiving { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public int LvValue { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public decimal SumIncome { get; set; }
    public DateTime SuperDate { get; set; }
    public bool IsSuper { get; set; }
    public bool IsFreeze { get; set; }
    public long GroupId { get; set; }
    public bool IsSz84 { get; set; }
    public DateTime Sz84Date { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public long AnswerId { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public DateTime AnswerDate { get; set; }
    public bool IsTeach { get; set; }
    public bool IsShutup { get; set; }
    public string SystemPrompt { get; set; } = string.Empty;
    [JsonIgnore]
    [HighFrequency]
    public long Tokens { get; set; }
    public DateTime UpdatedAt { get; set; }
    public bool IsAgent { get; set; }
    public bool IsAI { get; set; }
    public bool Xxian { get; set; }
    public long AgentId { get; set; }
    public bool IsSendHelpInfo { get; set; }
    public bool IsLog { get; set; }
    public bool IsMusicLogo { get; set; }
}
