using Newtonsoft.Json;

namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
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
            /*
            lock (consoleLock)
            {
                if (!IsBlackSystem)
                {
                    ShowMessage($"{Name}({UserId}) {RealGroupName}({RealGroupId}) {EventMessage}", ConsoleColor.White);
                    ShowMessage($"{SelfName}({SelfId}) {RealGroupName}({RealGroupId})", ConsoleColor.White);
                    InfoMessage($"{(IsSend ? "" : "[未发送]")}{(Answer.IsNull() ? "[无回答]" : Answer)}", ConsoleColor.Green);
                }
            }
            */
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
                await AppendGroupSendMessageAsync();
            }

            if (!string.IsNullOrEmpty(Answer))
            {
                // 将 [@:12345] 替换为 CQ 码 [CQ:at,qq=12345]
                Answer = System.Text.RegularExpressions.Regex.Replace(Answer, @"\[@:(?<UserId>\d+)\]", "[CQ:at,qq=${UserId}]");
            }

            /*
            ShowMessage($"{(isDup ? "[重复]" : "")}发送消息 {MsgGuid} {Answer} {IsSend}", ConsoleColor.Green);
            Console.Error.WriteLine($"[SendMessage] EventId: {MsgId}, IsSend: {IsSend}, IsWeb: {IsWeb}, IsOnebot: {IsOnebot}, Platform: {Platform}, BotType: {SelfInfo.BotType}");
            */

            if (!IsSend)
            {
                // Console.Error.WriteLine($"[SendMessage] Message {MsgId} suppressed: {Reason}");
                return;
            }

            if (IsWeb || IsOnebot)
            {
                if (ReplyMessageAsync == null) 
                {
                    Console.Error.WriteLine($"[SendMessage] ReplyMessageAsync is NULL for event {MsgId}!");
                    return;
                }
                // Console.Error.WriteLine($"[SendMessage] Calling ReplyMessageAsync for event {MsgId}...");
                await ReplyMessageAsync();
                IsSent = true;
                // Console.Error.WriteLine($"[SendMessage] ReplyMessageAsync call completed for event {MsgId}.");
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

        private async Task<int> AppendGroupSendMessageAsync()
        {
            if (User.IsLog) await BotLogRepository.LogAsync($"{GroupName}({GroupId}) {Name}({UserId}) {EventType}：\n{Message}", "处理后", this);
            if (IsBlackSystem && EventType.In("EventPrivateMessage", "EventGroupMessage", "TempMessageEvent")) return 0;

            var entity = new GroupSendMessage
            {
                MsgGuid = MsgGuid,
                BotUin = SelfId,
                GroupId = RealGroupId,
                GroupName = RealGroupName,
                UserId = UserId,
                UserName = Name,
                MsgId = MsgId,
                Question = Message.IsNull() ? EventType : Message,
                Message = IsSend && IsRealProxy && !IsAI && AnswerId == 0 ? $"@{Card.ReplaceInvalid().RemoveUserIds().ReplaceSensitive(Regexs.OfficalRejectWords)}:{Answer}" : Answer,
                AnswerAi = AnswerAI,
                AnswerId = AnswerId,
                IsAI = IsAI,
                AgentId = AgentId,
                AgentName = AgentName,
                IsSend = IsSend,
                IsRealProxy = IsRealProxy,
                Reason = Reason,
                IsCmd = IsCmd,
                InputTokens = InputTokens,
                OutputTokens = OutputTokens,
                TokensMinus = TokensMinus,
                IsVoiceReply = IsVoiceReply,
                VoiceName = VoiceName,
                CostTime = (int?)CostTime,
                IsRecall = IsRecall,
                ReCallAfterMs = RecallAfterMs,
                InsertDate = DateTime.Now
            };

            return await GroupSendMessageRepository.AppendAsync(entity);
        }
    }
}
