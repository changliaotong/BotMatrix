using Microsoft.AspNetCore.SignalR;
using Newtonsoft.Json;
using BotWorker.Domain.Models.Messages.BotMessages;

namespace BotWorker.Application.Services
{
    public class RemoteCall(IHubContext<ChatHub> hubContext, RemoteRequest req)
    {
        private readonly IHubContext<ChatHub> hubContext = hubContext;

        public async Task CallAsync(string user, string method, BotMessage context, params object[] args) 
            => await hubContext.Clients.User(user).SendAsync("ReceiveRequestMessage", req.Register(method, context, args), method, JsonConvert.SerializeObject(args));

        public async Task MuteAsync(string user, BotMessage context, long group, long target, int seconds) 
            => await CallAsync(user, "MuteAsync", context, group, target, seconds);

        public async Task KickOutAsync(string user, BotMessage context, long group, long target) 
            => await CallAsync(user, "KickOutAsync", context, group, target);

        public async Task SetTitleAsync(string user, BotMessage context, long group, long target, string title) 
            => await CallAsync(user, "SetTitleAsync", context, group, target, title);

        public async Task ChangeNameAsync(string user, BotMessage context, long group, long target, string name, string boy, string gril, string admin)
            => await CallAsync(user, "ChangeNameAsync", context, group, target, name, boy, gril, admin);

        public async Task ChangeNameAllAsync(string user, BotMessage context, string boy, string girl, string admin) 
            => await CallAsync(user, "ChangeNameAllAsync", context, boy, girl, admin);

        public async Task LeaveAsync(string user, BotMessage context, long group) 
            => await CallAsync(user, "LeaveAsync", context, group);

        public async Task RecallAsync(string user, BotMessage context, string msgGuid, long group, string messageId) 
            => await CallAsync(user, "RecallAsync", context, msgGuid, group, messageId);

        public async Task RecallForwardAsync(string user, BotMessage context, string group, string messageId, string forwardMsgId) 
            => await CallAsync(user, "RecallForward", context, group, messageId, forwardMsgId);
    }

}


