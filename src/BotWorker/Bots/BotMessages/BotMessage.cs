using BotWorker.Agents.Entries;
using BotWorker.Common.Exts;
using BotWorker.Bots.Entries;
using System.Diagnostics;
using System.Reflection;
using BotWorker.Bots.Platform;

namespace BotWorker.Bots.BotMessages
{
    public partial class BotMessage
    {        
        public static Core.Plugin.PluginManager? PluginManager { get; set; }
        public static Core.Pipeline.MessagePipeline? Pipeline { get; set; }

        public string MsgGuid { get; set; } = Guid.NewGuid().ToString();
        public long SelfId => SelfInfo.BotUin;
        public string SelfName => SelfInfo.BotName;
        public string Platform => SelfInfo.Platform;
        public string EventType { get; set; } = string.Empty;        
        public string EventMessage { get; set; } = string.Empty;
        public long RealGroupId { get; set; } = 0;
        public string RealGroupName { get; set; } = string.Empty;
        public long GroupId => ParentGroup == null || ParentGroup.Id == 0 ? Group.Id : ParentGroup.Id;
        public string GroupName => ParentGroup == null || ParentGroup.Id == 0 ? Group.GroupName : ParentGroup.GroupName;
        public string GroupOpenid => Group.GroupOpenId;
        public string GuildId { get; set; } = string.Empty;
        public string GuildName { get; set; } = string.Empty;
        public string Message { get; set; } = string.Empty;
        public string CurrentMessage { get; set; } = string.Empty;
        public bool IsGroup => !RealGroupId.In(0, BotInfo.DefaultGroupUinGuild) && !IsPublic;
        public long UserId => User.Id;
        public string Name => User.Name ?? string.Empty;
        public string Card { get; set; } = string.Empty;
        public string Title { get; set; } = string.Empty;
        public string UserOpenId => User.UserOpenId;
        public string DisplayName { get; set; } = string.Empty;
        public string MsgId { get; set; } = string.Empty;
        public bool IsSuperAdmin => BotInfo.IsSuperAdmin(UserId);
        public bool IsAtMe { get; set; } = false;
        public bool IsAtAll { get; set; } = false;
        public bool IsAtOthers { get; set; } = false;
        public bool IsReply { get; set; } = false;
        public string ReplyMsgId { get; set; } = string.Empty;
        public bool IsForward { get; set; } = false;
        public string ForwardMsgId {  get; set; } = string.Empty;
        public bool IsFile { get; set; } = false;
        public bool IsJson { get; set; } = false;
        public bool IsKeyboard { get; set; } = false;
        public bool IsLightApp { get; set; } = false;
        public bool IsLongMsg { get; set; } = false;
        public bool IsMarkdown { get; set; } = false;
        public bool IsStream { get; set; } = false;
        public bool IsImage { get; set; } = false;
        public bool IsFlashImage { get; set; } = false;
        public bool IsVideo { get; set; } = false; 
        public bool IsXml { get; set; } = false;
        public bool IsMusic { get; set; } = false;
        public bool IsPoke { get; set; } = false;
        public bool IsVoice { get; set; } = false;
        public bool IsContactGroup { get; set; } = false;
        public bool IsContactFriend { get; set; } = false;
        public bool IsContactGuild { get; set; } = false;
        public bool IsRefresh { get; set; } = false;     
        public string AppName { get; set; } = string.Empty;
        public string Payload { get; set; } = string.Empty;
        public long Operater { get; set; } = 0;
        public string OperaterName { get; set; } = string.Empty;
        public long InvitorQQ { get; set; } = 0;
        public string InvitorName { get; set; } = string.Empty;
        public string RequestType { get; set; } = string.Empty;
        public string Flag { get; set; } = string.Empty;
        public long Period { get; set; } = 0;
        public int SelfPerm { get; set; } = 2;
        public int UserPerm { get; set; } = 2;
        public bool IsBlack { get; set; } = false;
        public bool IsRobot { get; set; } = false;
        public bool IsWhite { get; set; } = false;
        public bool IsGrey { get; set; } = false;
        public bool IsBlackSystem { get; set; } = false;
        public bool IsGreySystem { get; set; } = false;
        public bool IsCmd { get; set; } = false;
        public string CmdName { get; set; } = string.Empty;
        public string CmdPara { get; set; } = string.Empty;
        public bool IsVip { get; set; } = false;
        public bool IsPublic => Platform == Platforms.Public;    
        public bool IsConfirm { get; set; } = false;         
        public long TargetUin { get; set; }
        public string TargetName { get; set; } = string.Empty;
        public bool IsSet { get; set; } = false;
        public int Accept { get; set; } = 0;                         
        public string Reason { get; set; } = string.Empty;
        public bool IsProxy { get; set; } = false;
        private bool _isCancelProxy;

        public bool IsCancelProxy
        {
            get => _isCancelProxy;
            set
            {
                if (_isCancelProxy == value) return;   // 值没变，不记录

                _isCancelProxy = value;

                // 记录日志
                LogAssignment(nameof(IsCancelProxy), value);
            }
        }
        public long ProxyBotUin { get; set; } = 0;        
        public bool IsProxyInGroup => ProxyBotUin != 0;
        public bool IsDup { get; set; } = false;
        public bool IsAgent => AgentId != AgentInfos.DefaultAgent.Id;
        public long AgentId { get; set; } = AgentInfos.DefaultAgent.Id;
        public bool IsCallAgent { get; set; } = false;
        public string AgentName { get; set; } = string.Empty;
        public int HistoryMessageCount { get; set; } = 3;
        public int ModelId { get; set; } = 0;
        public long InputTokens { get; set; } = 0;
        public long OutputTokens { get; set; } = 0;
        public int TokensTimes { get; set; } = 1;
        public int TokensTimesOutput { get; set; } = 2;
        public long TokensMinus { get; set; } = 0;
        //public bool IsSend { get; set; } = true;
        private bool _isIsSend = true;

        public bool IsSend
        {
            get => _isIsSend;
            set
            {
                if (_isIsSend == value) return;   // 值没变，不记录

                _isIsSend = value;

                // 记录日志
                LogAssignment(nameof(IsSend), value);
            }
        }

        public bool IsNewAnswer { get; set; } = false;
        public float Similarity { get; set; } = 0.00F;
        public long NewQuestionId { get; set; } = 0;
        public string NewQuestion { get; set; } = string.Empty;
        public long AnswerId { get; set; } = 0;
        public string Answer { get; set; } = string.Empty;
        public bool IsAI { get; set; } = false;
        public string AnswerAI { get; set; } = string.Empty;
        public bool IsVoiceReply
        {
            get
            {
                if (!IsNapCat) return false;
                if (!IsGroup) return false;
                if (CmdName == "语音播报") return true;
                if (IsEntirelyInBrackets(Answer)) return false;
                if (IsAgent || IsCallAgent) return CurrentAgent.IsVoice;
                return (ParentGroup == null || ParentGroup.Id == 0 ? Group.IsVoiceReply : ParentGroup.IsVoiceReply && CurrentAgent.IsVoice) && ((AnswerId != 0 && !IsCmd) || IsAI) ;
            }
        }

        public string VoiceId => IsAgent || IsCallAgent ? CurrentAgent.VoiceId : ParentGroup == null || ParentGroup.Id == 0 ? Group.VoiceId : ParentGroup.VoiceId;
        public string VoiceName => VoiceMapUtil.GetVoiceName(VoiceId);
        public double CostTime { get; set; } = 0.00;
        public bool IsQuote { get; set; } = false;
        public int DelayMs { get; set; } = 0; // 延迟发送的毫秒数
        public bool IsRecall { get; set; } = false;
        public int RecallAfterMs { get; set; } = 0; // 撤回消息的毫秒数
        public SongResult? SongResult { get; set; }

        // 判断整条文本是否被括号包住
        bool IsEntirelyInBrackets(string text)
        {
            if (string.IsNullOrWhiteSpace(text)) return false;
            text = text.Trim();
            return text.IsMatch(@"^\(.+\)$");
        }
        private void LogAssignment(string propName, object value)
        {
            try
            {
                var st = new StackTrace(skipFrames: 1, fNeedFileInfo: true);

                MethodBase? caller = null;
                string? file = null;
                int line = 0;

                for (int i = 0; i < st.FrameCount; i++)
                {
                    var frame = st.GetFrame(i);
                    if (frame == null) continue;

                    var mb = frame.GetMethod();
                    if (mb == null) continue;

                    var declaring = mb.DeclaringType;
                    if (declaring == typeof(BotMessage)) continue;

                    caller = mb;
                    file = frame.GetFileName();
                    line = frame.GetFileLineNumber();
                    break;
                }

                string callerInfo = caller != null
                    ? $"{caller.DeclaringType?.FullName}.{caller.Name} ({file}:{line})"
                    : "unknown";

                Console.WriteLine($"[BotMessage] Property '{propName}' set to '{value}' by {callerInfo}");
            }
            catch
            {
                // 忽略日志错误，避免影响主流程
            }
        }

    }
}
