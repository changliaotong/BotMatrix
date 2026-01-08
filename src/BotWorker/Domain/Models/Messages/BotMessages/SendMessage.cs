using Newtonsoft.Json;
using BotWorker.Domain.Entities;
using BotWorker.Infrastructure.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Application.Services;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        private static readonly object consoleLock = new();

        public async Task SendOfficalShareAsync()
        {
            var isCancelProxy = IsCancelProxy;
            IsCancelProxy = true;            
            Answer = "管理拉一下这个号码";
            await SendMessageAsync();            
            Answer = $"[CQ:contact,id={BotInfo.DefaultProxyBotUin},type=qq]";
            await SendMessageAsync();
            IsCancelProxy = isCancelProxy;
        }

        public void ShowBotMessage()
        {
            lock (consoleLock)
            {
                if (!IsBlackSystem)
                {
                    ShowMessage($"{Name}({UserId}) {RealGroupName}({RealGroupId}) {EventMessage}", ConsoleColor.White);
                    ShowMessage($"{SelfName}({SelfId}) {RealGroupName}({RealGroupId})", ConsoleColor.White);
                    InfoMessage($"{(IsSend ? "" : "[未发送]")}{(Answer.IsNull() ? "[无回答]" : Answer)}", ConsoleColor.Green);
                }
            }
        }

        //发送消息
        public async Task SendMessageAsync(bool isDup = false)
        {
            ShowMessage($"[关闭消息] SelfInfo {SelfInfo.BotUin} IsGroup {IsGroup} SelfInfo.IsGroup {SelfInfo.IsGroup} SelfInfo.IsPrivate {SelfInfo.IsPrivate} ");
            if ((IsGroup && !SelfInfo.IsGroup) || (!IsGroup && !SelfInfo.IsPrivate))
            {                
                Reason += IsGroup ? "[群聊关闭]" : "[私聊关闭]";
                IsSend = false;
            }

            if (!isDup)
            {
                await GetFriendlyResAsync();
                GroupSendMessage.Append(this);
            }

            ShowMessage($"{(isDup ? "[重复]" : "")}发送消息 {MsgGuid} {Answer} {IsSend}", ConsoleColor.Green);
            Console.Error.WriteLine($"[SendMessage] IsWeb: {IsWeb}, IsWorker: {IsWorker}, Platform: {Platform}, BotType: {SelfInfo.BotType}");

            if (IsWeb || IsOnebot)
            {
                if (ReplyMessageAsync == null) 
                {
                    Console.Error.WriteLine("[SendMessage] ReplyMessageAsync is NULL!");
                    return;
                }
                Console.Error.WriteLine("[SendMessage] Calling ReplyMessageAsync...");
                await ReplyMessageAsync();
            }
            else
            {
                Console.Error.WriteLine($"[SendMessage] Not Worker/Web. ReplyBotMessageAsync is null? {ReplyBotMessageAsync == null}");
                if (IsRealProxy || ((IsMusic || IsAI) && IsGuild))               
                    SelfInfo = await BotInfo.LoadAsync(ProxyBotUin);

                var json = JsonConvert.SerializeObject(this);
                if (ReplyBotMessageAsync == null) return;
                await ReplyBotMessageAsync(json);
            }            
        }
    }
}
