using Microsoft.SemanticKernel.ChatCompletion;
using Newtonsoft.Json;
using QQBot4Sharp.Models;
using BotWorker.Modules.AI.Providers;
using BotWorker.Infrastructure.Utils;
using System.Diagnostics;
using BotWorker.Modules.AI.Plugins;
using BotWorker.Modules.Plugins;
using BotWorker.Application.Messaging.Pipeline;

namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
    [JsonIgnore]
    public PluginManager? PluginManager { get; set; }
    [JsonIgnore]
    public MessagePipeline? Pipeline { get; set; }
    [JsonIgnore]
    public IServiceProvider? ServiceProvider { get; set; }

    [JsonIgnore]
    public KnowledgeBaseService? KbService;
    public GroupInfo Group { get; set; } = new();
    [JsonIgnore]
    public GroupInfo? ParentGroup { get; set; }
    public UserInfo User { get; set; } = new();
    public BotInfo SelfInfo { get; set; } = new();
    [JsonIgnore]
    public Agent CurrentAgent { get; set; } = new();
    [JsonIgnore]
    public Mirai.Net.Data.Events.EventBase? MiraiEvent { get; set; } = null;
    [JsonIgnore]
    public ContextEventArgs? EventArgs { get; set; } = null;
    [JsonIgnore]
    public string? CallerConnectionId = string.Empty;
    [JsonIgnore]
    public Func<string, Task>? ReplyBotMessageAsync { get; set; }
    [JsonIgnore]
    public Func<string, Task>? ReplyProxyMessageAsync { get; set; }
    [JsonIgnore]
    public Func<CancellationToken, Task>? ReplyStreamBeginMessageAsync { get; set; }
    [JsonIgnore]
    public Func<string, CancellationToken, Task>? ReplyStreamMessageAsync { get; set; }
    [JsonIgnore]
    public Func<CancellationToken, Task>? ReplyStreamEndMessageAsync { get; set; }
    [JsonIgnore]
    public Func<Task>? ReplyMessageAsync { get; set; }
    [JsonIgnore]
    public LLMApp? LLMApp { get; set; }
    [JsonIgnore]
    public ChatHistory History { get; set; } = [];
    [JsonIgnore]
    public bool IsWorker => Platform == Platforms.Worker;
    [JsonIgnore]
    public bool IsGuild => Platform == Platforms.QQGuild;
    [JsonIgnore]
    public bool IsWeixin => Platform == Platforms.Weixin;
    [JsonIgnore]
    public bool IsMirai => Platform == Platforms.Mirai;
    [JsonIgnore]
    public bool IsWeb => Platform == Platforms.Web;
    [JsonIgnore]
    public bool IsQQ => Platform == Platforms.QQ;
    public bool IsOnebot => IsQQ || IsGuild || IsWeixin || IsWorker;
    [JsonIgnore]
    public bool IsGroupBound => GroupId < GroupInfo.groupMin;
    [JsonIgnore]
    public bool IsRealProxy => IsProxy && !IsCancelProxy;
    [JsonIgnore]
    public PlaceholderContext Ctx { get; set; } = new(); 
    [JsonIgnore]
    public Stopwatch? CurrentStopwatch { get; set; }

    public async Task<string> GetRecallCountAsync() => (await GroupEvent.GetRecallCountAsync(GroupId)).ToString();
    public async Task<string> GetEventCountAsync(GroupEventType eventType) => (await GroupEvent.GetEventCountAsync(GroupId, eventType.ToString())).ToString();
    public async Task<string> GetMenuResAsync() => await BotMessage.GetMenuTextAsync();
    public async Task<string> GetTestItResAsync() => await this.GetTestItAsync();
    public async Task<string> GetFreeCreditAsync() { await GetCreditMoreAsync(); return ""; }
    public async Task<string> GetCreditListAsync() => await this.GetCreditListAsync(10);
    public async Task<string> GetCreditListAllAsync(long userId) => await this.GetCreditListAllAsync(userId, 10);

    public bool IsEnough() => IsEnoughAsync().GetAwaiter().GetResult();
    public string BatchInsertAgent() => BatchInsertAgentAsync().GetAwaiter().GetResult();
    public async Task<long> GetCreditAsync() => await UserService.GetCreditAsync(SelfId, GroupId, UserId);
    public string MinusCreditRes(long creditMinus, string creditInfo) => MinusCreditResAsync(creditMinus, creditInfo).GetAwaiter().GetResult();
    public bool IsTooFast() => IsTooFastAsync().GetAwaiter().GetResult();
    public string GetCaiquan() => GetCaiquanAsync().GetAwaiter().GetResult();
    public string GetSanggongRes() => GetSanggongResAsync().GetAwaiter().GetResult();
    public string GetLuckyDraw() => GetLuckyDrawAsync().GetAwaiter().GetResult();

    public async Task<string> GetGiftResAsync(long userId, string cmdPara)
    {
        long qqGift = cmdPara.GetQq();
        if (qqGift == 0) qqGift = userId;

        string giftName = cmdPara.RegexReplace(@"\[CQ:at,qq=\d+\]", "").Trim();
        if (giftName == "") giftName = "礼物";

        int giftCount = 1;
        var matchCount = System.Text.RegularExpressions.Regex.Match(cmdPara, @"\s+(\d+)$");
        if (matchCount.Success) giftCount = int.Parse(matchCount.Value);

        var groupGiftService = ServiceProvider!.GetRequiredService<BotWorker.Domain.Interfaces.IGroupGiftService>();
        return await groupGiftService.GetGiftResAsync(SelfId, GroupId, GroupName, UserId, Name, qqGift, giftName, giftCount);
    }
}
