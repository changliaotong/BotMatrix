using System.Diagnostics;
using Microsoft.SemanticKernel.ChatCompletion;
using Newtonsoft.Json;
using QQBot4Sharp.Models;
using BotWorker.Agents.Entries;
using BotWorker.Agents.Plugins;
using BotWorker.Agents.Providers;
using BotWorker.Bots.Entries;
using BotWorker.Bots.Platform;
using BotWorker.Bots.Users;
using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;
using BotWorker.Core.Services;
using BotWorker.Infrastructure.Utils;

namespace BotWorker.Bots.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
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
    public static LLMApp LLMApp { get; set; } = new();
    [JsonIgnore]
    public ChatHistory History { get; set; } = [];
    [JsonIgnore]
    public bool IsWorker => Platform == Platforms.Worker;
    [JsonIgnore]
    public bool IsGuild => Platform == Platforms.QQGuild;
    [JsonIgnore]
    public bool IsWeixin => Platform == Platforms.Weixin;
    [JsonIgnore]
    public bool IsWeb => Platform == Platforms.Web;
    [JsonIgnore]
    public bool IsNapCat => Platform == Platforms.NapCat;
    public bool IsOnebot => IsNapCat || IsGuild || IsWeixin || IsWorker;
    [JsonIgnore]
    public bool IsGroupBound => GroupId < GroupInfo.groupMin;
    [JsonIgnore]
    public bool IsRealProxy => IsProxy && !IsCancelProxy;
    [JsonIgnore]
    public PlaceholderContext Ctx { get; set; } = new(); 
    [JsonIgnore]
    public Stopwatch? CurrentStopwatch { get; set; }
}
