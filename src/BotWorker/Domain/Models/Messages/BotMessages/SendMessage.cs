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
            if ((IsGroup && !BotInfo.GetBool("IsGroup", SelfId)) || (!IsGroup && !BotInfo.GetBool("IsPrivate", SelfId)))
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
            Console.Error.WriteLine($"[SendMessage] EventId: {MsgId}, IsSend: {IsSend}, IsWeb: {IsWeb}, IsOnebot: {IsOnebot}, Platform: {Platform}, BotType: {SelfInfo.BotType}");

            if (!IsSend)
            {
                Console.Error.WriteLine($"[SendMessage] Message {MsgId} suppressed: {Reason}");
                return;
            }

            if (IsWeb || IsOnebot)
            {
                if (ReplyMessageAsync == null) 
                {
                    Console.Error.WriteLine($"[SendMessage] ReplyMessageAsync is NULL for event {MsgId}!");
                    return;
                }
                Console.Error.WriteLine($"[SendMessage] Calling ReplyMessageAsync for event {MsgId}...");
                await ReplyMessageAsync();
                IsSent = true;
                Console.Error.WriteLine($"[SendMessage] ReplyMessageAsync call completed for event {MsgId}.");
            }
            else
            {
                Console.Error.WriteLine($"[SendMessage] Not Worker/Web. ReplyBotMessageAsync is null? {ReplyBotMessageAsync == null}");
                if (IsRealProxy || ((IsMusic || IsAI) && IsGuild))               
                    SelfInfo = await BotInfo.LoadAsync(ProxyBotUin) ?? new();

                var json = JsonConvert.SerializeObject(this);
                if (ReplyBotMessageAsync == null) return;
                await ReplyBotMessageAsync(json);
                IsSent = true;
            }            
        }
    }
}
