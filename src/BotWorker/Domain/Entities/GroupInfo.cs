using System;
using System.Collections.Generic;
using System.Data;
using System.Text;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;
using Newtonsoft.Json;

namespace BotWorker.Domain.Entities
{
    [Table("group_info")]
    public partial class GroupInfo
    {
        public const long groupMin = 990000000000;

        [Write(false)]
        public string GroupOpenId { get; set; } = string.Empty;

        [Write(false)]
        public long TargetGroup { get; set; }

        [ExplicitKey]
        public long Id { get; set; }

        public string GroupName { get; set; } = string.Empty;

        [Write(false)]
        public bool IsValid { get; set; }

        [Write(false)]
        public bool IsProxy { get; set; }

        public string GroupMemo { get; set; } = string.Empty;
        public long GroupOwner { get; set; }
        public string GroupOwnerName { get; set; } = string.Empty;
        public string GroupOwnerNickname { get; set; } = string.Empty;
        public int GroupType { get; set; }

        [Write(false)]
        public long RobotOwner { get; set; }

        public string RobotOwnerName { get; set; } = string.Empty;
        public string WelcomeMessage { get; set; } = string.Empty;
        public int GroupState { get; set; }

        [Write(false)]
        public long BotUin { get; set; }

        public string BotName { get; set; } = string.Empty;

        [JsonIgnore]
        public DateTime LastDate { get; set; }

        [JsonIgnore]
        public int IsInGame { get; set; }

        public bool IsOpen { get; set; }
        public int UseRight { get; set; }
        public int TeachRight { get; set; }
        public int AdminRight { get; set; }

        [Write(false)]
        public DateTime QuietTime { get; set; }

        public bool IsCloseManager { get; set; }
        public int IsAcceptNewMember { get; set; }

        [Write(false)]
        public string CloseRegex { get; set; } = string.Empty;

        public string RegexRequestJoin { get; set; } = string.Empty;
        public string RejectMessage { get; set; } = string.Empty;
        public bool IsWelcomeHint { get; set; }
        public bool IsExitHint { get; set; }
        public bool IsKickHint { get; set; }
        public bool IsChangeHint { get; set; }
        public bool IsRightHint { get; set; }
        public bool IsCloudBlack { get; set; }
        public int IsCloudAnswer { get; set; }
        public bool IsRequirePrefix { get; set; }
        public bool IsSz84 { get; set; }
        public bool IsWarn { get; set; }
        public bool IsBlackExit { get; set; }
        public bool IsBlackKick { get; set; }
        public bool IsBlackShare { get; set; }
        public bool IsChangeEnter { get; set; }
        public bool IsMuteEnter { get; set; }
        public bool IsChangeMessage { get; set; }

        [Write(false)]
        public bool IsSaveRecord { get; set; }

        [Write(false)]
        public bool IsPause { get; set; }

        public string RecallKeyword { get; set; } = string.Empty;
        public string WarnKeyword { get; set; } = string.Empty;
        public string MuteKeyword { get; set; } = string.Empty;
        public string KickKeyword { get; set; } = string.Empty;
        public string BlackKeyword { get; set; } = string.Empty;
        public int MuteEnterCount { get; set; }
        public int MuteKeywordCount { get; set; }
        public int KickCount { get; set; }
        public int BlackCount { get; set; }
        public long ParentGroup { get; set; }
        public string CardNamePrefixBoy { get; set; } = string.Empty;
        public string CardNamePrefixGirl { get; set; } = string.Empty;
        public string CardNamePrefixManager { get; set; } = string.Empty;
        public DateTime UpdatedAt { get; set; }

        [JsonIgnore]
        [Write(false)]
        public string LastAnswer { get; set; } = string.Empty;

        [JsonIgnore]
        public string LastChengyu { get; set; } = string.Empty;

        [JsonIgnore]
        public DateTime LastChengyuDate { get; set; }

        [Write(false)]
        public DateTime TrialStartDate { get; set; }

        [Write(false)]
        public DateTime TrialEndDate { get; set; }

        [JsonIgnore]
        [Write(false)]
        public DateTime LastExitHintDate { get; set; }

        [Write(false)]
        public string BlockRes { get; set; } = string.Empty;

        [Write(false)]
        public int BlockType { get; set; }
        public int BlockMin { get; set; }

        [Write(false)]
        public int BlockFee { get; set; }

        [Write(false)]
        public Guid Guid { get; set; }

        [Write(false)]
        public Guid GroupGuid { get; set; }

        public bool IsBlock { get; set; }
        public bool IsWhite { get; set; }
        public string CityName { get; set; } = string.Empty;
        public bool IsMuteRefresh { get; set; }
        public int MuteRefreshCount { get; set; }
        public bool IsProp { get; set; }
        public bool IsPet { get; set; }
        public bool IsBlackRefresh { get; set; }
        public string FansName { get; set; } = string.Empty;
        public bool IsConfirmNew { get; set; }
        public string WhiteKeyword { get; set; } = string.Empty;
        public string CreditKeyword { get; set; } = string.Empty;
        public bool IsCredit { get; set; }
        public bool IsPowerOn { get; set; }
        public bool IsHintClose { get; set; }
        public int RecallTime { get; set; }
        public bool IsInvite { get; set; }
        public int InviteCredit { get; set; }
        public bool IsReplyImage { get; set; }
        public bool IsReplyRecall { get; set; }
        public bool IsVoiceReply { get; set; }
        public string VoiceId { get; set; } = string.Empty;
        public bool IsAI { get; set; }
        public string SystemPrompt { get; set; } = string.Empty;
        public bool IsOwnerPay { get; set; }
        public int ContextCount { get; set; }
        public bool IsMultAI { get; set; }
        public bool IsAutoSignin { get; set; }
        public bool IsUseKnowledgebase { get; set; }

        [Write(false)]
        public DateTime InsertDate { get; set; }

        public bool IsSendHelpInfo { get; set; }
        public bool IsRecall { get; set; }
        public bool IsCreditSystem { get; set; }
        public string CreditName => "积分";
    }
}
