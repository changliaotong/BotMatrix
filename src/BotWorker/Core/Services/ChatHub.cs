using System.Collections.Concurrent;
using System.Diagnostics;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.SignalR;
using Microsoft.AspNetCore.SignalR.Client;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Mirai.Net.Data.Messages.Concretes;
using Newtonsoft.Json;
using sz84.Agents.Plugins;
using sz84.Bots.BotMessages;
using sz84.Bots.Entries;
using sz84.Bots.Groups;
using sz84.Bots.Interfaces;
using sz84.Bots.Users;
using BotWorker.Common.Exts;
using BotWorker.Models;
using sz84.Infrastructure.Background;
using sz84.Infrastructure.Caching;
using sz84.Infrastructure.SignalR;

namespace sz84.Core.Services
{
    public class ChatHub(IServiceProvider provider, ICacheService cache, RemoteRequest remoteRequest, KnowledgeBaseService qaService,
        IUserConnectionManager userConnectionManager, ILogger<MyHub> logger)
        : MyHub(cache, qaService, userConnectionManager, logger)
    {       
        public string StaticMessage { get; set; } = "";
        private static readonly ConcurrentDictionary<string, CancellationTokenSource> _streamCtsMap = new();

        public bool SendResponse(string requestId, string resultJson)
        {
            try
            {
                InfoMessage($"[SignalR] 客户端响应 {requestId} {resultJson} {StaticMessage}");
                remoteRequest.Complete(requestId, resultJson);
                return true;
            }
            catch (Exception ex)
            {
                InfoMessage($"[SignalR] 客户端响应失败 {requestId} {resultJson} {StaticMessage} {ex.Message}");
            }

            return false;
        }

        public Task CancelStream(string msgGuid)
        {
            InfoMessage($"{StaticMessage}");
            if (_streamCtsMap.TryRemove(msgGuid, out var cts))
            {
                cts.Cancel();
                cts.Dispose();
            }
            return Task.CompletedTask;
        }

        private Task ReplyBotMessage(BotMessage ctx)
        {
            return Clients.Client(Context.ConnectionId).SendAsync("ReceiveMessage", ctx.MsgGuid, ctx.Answer);
        }

        private Task ReplyStreamBegin(BotMessage ctx, CancellationToken token)
        {
            return Clients.Client(Context.ConnectionId).SendAsync("ReceiveStreamBeginMessage", ctx.MsgGuid, token);
        }

        private Task ReplyStream(BotMessage ctx, string json, CancellationToken token)
        {
            return Clients.Client(Context.ConnectionId).SendAsync("ReceiveStreamMessage", ctx.MsgGuid, json, token);
        }

        private Task ReplyStreamEnd(BotMessage ctx, CancellationToken token)
        {
            return Clients.Client(Context.ConnectionId).SendAsync("ReceiveStreamEndMessage", ctx.MsgGuid, token);
        }

        public async Task SendStreamUserMessage(string message,
            [FromServices] KnowledgeBaseService knowledgeBaseService)
        {
            try
            {                
                BotMessage? context = JsonConvert.DeserializeObject<BotMessage>(message);
                if (context == null)
                {
                    InfoMessage("[SignalR] [错误] 反序列化结果为空");
                    return;
                }              
                context.KbService = knowledgeBaseService;
                InfoMessage($"{Context.ConnectionId}");
                context.ReplyMessageAsync = () => ReplyBotMessage(context);
                context.ReplyStreamBeginMessageAsync = token => ReplyStreamBegin(context, token);
                context.ReplyStreamMessageAsync = (json, token) => ReplyStream(context, json, token);
                context.ReplyStreamEndMessageAsync = token => ReplyStreamEnd(context, token);
                var cts = new CancellationTokenSource();
                _streamCtsMap[context.MsgGuid] = cts;
                context.CurrentStopwatch = Stopwatch.StartNew();
                await context.StartStreamChatAsync(cts.Token);   
            }
            catch (Exception ex)
            {
                InfoMessage($"[SignalR] 服务端异常 {ex}");
                throw new HubException("[SignalR] SendStreamUserMessage 服务端异常：" + ex.Message);
            }
        }

        public bool SendBotMessage(string guid, string message, 
            [FromServices] IBotHandlerMessage handler, 
            [FromServices] ILogger<ChatHub> logger, 
            [FromServices] KnowledgeBaseService knowledgeBaseService,
            [FromServices] IHubContext<ChatHub> hubContext)
        {
            InfoMessage($"[SignalR] [收到消息] {guid}");

            BotMessage? context = JsonConvert.DeserializeObject<BotMessage>(message);
            if (context == null) return true;

            context.CallerConnectionId = Context.ConnectionId;
            context.KbService = knowledgeBaseService;
            context.ReplyMessageAsync = () => ReplyBotMessage(context);
            context.ReplyBotMessageAsync = async json => await hubContext.Clients.User(context.SelfId.ToString()).SendAsync("ReceiveBotMessage", context.MsgGuid, json);
            context.ReplyProxyMessageAsync = async json =>
            {
                await hubContext.Clients.User(context.ProxyBotUin.ToString()).SendAsync("ReceiveProxyMessage", context.MsgGuid, json);
                await hubContext.Clients.User(context.SelfId.ToString()).SendAsync("ReceiveProxyMessage", context.MsgGuid, json);
            };

            BotTaskHelper.EnqueueBotTask(provider, context, async ctx =>
            {
                ctx.CurrentStopwatch = Stopwatch.StartNew();
                ShowMessage($"[Event] {ctx.EventMessage}");
                ShowMessage($"[Event] 处理中...", ConsoleColor.White);                
                await handler.HandleBotMessageAsync(ctx);
                ctx.CurrentStopwatch.Stop();
                ctx.CostTime = ctx.CurrentStopwatch.Elapsed.TotalSeconds;
                ShowMessage($"[Event] 完成，用时 {ctx.CurrentStopwatch.Elapsed.TotalSeconds:F3} 秒");
                ShowMessage($"{ctx.Reason} {ctx.Answer}", ConsoleColor.Green);
                await ctx.SendMessageAsync();

            }, logger, "处理Bot消息");

            return true;
        }

        public async Task<bool> SendMessage(string guid, string message)
        {                    
            try
            {
                InfoMessage($"{StaticMessage}");
                await Clients.All.SendAsync("ReceiveMessage", guid, message);
                return true;
            }
            catch (Exception ex)
            {
                Console.WriteLine(ex);
            }

            return false;
        }

        public async Task<bool> SendProxyMessage(string userId, string guid, string json)
        {
            try
            {
                InfoMessage($"{StaticMessage} {json}");
                await Clients.User(userId).SendAsync("ReceiveProxyMessage", guid, json);
                return true;
            }
            catch (Exception ex)
            {
                Console.WriteLine(ex);
            }

            return false;
        }

        // 通知艾特官机
        public async Task<bool> SendMentionMessage(string guid, string groupOpenid, long officalBot)
        {
            try
            {
                long groupId = GroupInfo.GetWhere("TargetGroup", $"GroupOpenid = {groupOpenid.Quotes()}").AsLong();
                await Clients.All.SendAsync("ReceiveMentionMessage", guid, groupId, officalBot, GroupInfo.GetLong("BotUin", groupId));
                return true;
            }
            catch (Exception ex)
            { 
                Console.WriteLine(ex); 
            }

            return false;
        }

        public async Task SendPrivateMessage(string userId, string message)
        {
            await Clients.User(userId).SendAsync("ReceiveMessage", message);
        }

        public async Task SendMessageToGroup(string roomId, string message)
        {
            await Clients.Group(roomId).SendAsync("Send", $"{Context.ConnectionId}: {message}");
        }

        /// 广播
        public async Task<bool> BroadCastMessage(string guid, string message)
        {
            try
            {
                await Clients.All.SendAsync("ReceiveMessage", guid, message);
                return true;
            }
            catch (Exception ex)
            {
                Console.WriteLine(ex);
            }

            return false;
        }

        public long GetQQByOpenid(string MemberOpenId)
        {
            InfoMessage($"{StaticMessage}");
            return UserInfo.GetWhere("isnull(TargetUserId, Id)", $"UserOpenid = {MemberOpenId.Quotes()}").AsLong();
        }

        public long GetGroupByOpenid(string GroupOpenid)
        {
            InfoMessage($"{StaticMessage}");
            return GroupInfo.GetWhere("isnull(TargetGroup, 0)", $"GroupOpenid = {GroupOpenid.Quotes()}").AsLong();
        }

        public string GetValue(string field, long id)
        {
            InfoMessage($"{StaticMessage}");
            return BotInfo.GetValue(field, id);
        }

        public int SetValue(string field, string value, long id)
        {
            InfoMessage($"{StaticMessage}");
            return BotInfo.SetValue(field, value, id);
        }

        public int SetIsSend(string msgGuid, int isSend)
        {
            InfoMessage($"{StaticMessage}");
            var sql = $"update {GroupSendMessage.FullName} set is_send = {isSend} where msg_guid = {msgGuid.Quotes()}";
            return Exec(sql);
        }

        public int Debug(string message, string group = "")
        {
            InfoMessage($"{StaticMessage}");
            return DbDebug(message, group);
        }

        public int AppendSendMessage(string message)
        {
            InfoMessage($"{StaticMessage}");
            BotMessage? context = JsonConvert.DeserializeObject<BotMessage>(message);
            if (context == null) return -1;
            return GroupSendMessage.Append(context);
        }
    }
}
