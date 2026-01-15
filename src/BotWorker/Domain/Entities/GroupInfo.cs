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

        // Static Methods Bridge
        private static IGroupRepository Repo => GlobalConfig.ServiceProvider!.GetRequiredService<IGroupRepository>();

        public static async Task<bool> GetIsCreditAsync(long groupId, IDbTransaction? trans = null) => await Repo.GetIsCreditAsync(groupId);

        public static async Task<bool> GetIsPetAsync(long groupId, IDbTransaction? trans = null) => await Repo.GetIsPetAsync(groupId);

        public static async Task<int> SetPowerOffAsync(long groupId, IDbTransaction? trans = null) => await Repo.SetIsOpenAsync(false, groupId);

        public static async Task<int> SetPowerOnAsync(long groupId, IDbTransaction? trans = null) => await Repo.SetIsOpenAsync(true, groupId);

        public static async Task<bool> GetPowerOnAsync(long groupId, IDbTransaction? trans = null) => await Repo.GetPowerOnAsync(groupId);

        public static async Task<long> GetGroupOwnerAsync(long groupId, long def = 0, IDbTransaction? trans = null) => await Repo.GetGroupOwnerAsync(groupId, def, trans);

        public static int SetRobotOwner(long groupId, long ownerId) => SetRobotOwnerAsync(groupId, ownerId).GetAwaiter().GetResult();

        public static async Task<int> SetRobotOwnerAsync(long groupId, long ownerId, IDbTransaction? trans = null) => await Repo.SetRobotOwnerAsync(groupId, ownerId);

        public static long GetRobotOwner(long groupId) => GetRobotOwnerAsync(groupId).GetAwaiter().GetResult();

        public static async Task<long> GetRobotOwnerAsync(long groupId, long def = 0, IDbTransaction? trans = null) => await Repo.GetRobotOwnerAsync(groupId, def, trans);

        public static bool IsOwner(long groupId, long userId) => IsOwnerAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<bool> IsOwnerAsync(long groupId, long userId, IDbTransaction? trans = null) => userId == await GetRobotOwnerAsync(groupId, 0, trans);

        public static bool IsPowerOff(long groupId) => IsPowerOffAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> IsPowerOffAsync(long groupId, IDbTransaction? trans = null) => !await GetPowerOnAsync(groupId, trans);

        public static async Task<bool> GetIsValidAsync(long groupId, IDbTransaction? trans = null) => await Repo.GetIsValidAsync(groupId);

        public static async Task<string> GetRobotOwnerNameAsync(long groupId) => await Repo.GetRobotOwnerNameAsync(groupId);

        public static async Task<string> GetRobotOwnerNameAsync(long groupId, string botName) => await Repo.GetRobotOwnerNameAsync(groupId, botName);

        // Keeping the signature but ignoring enum for now to avoid dependency issue if possible, or mapping it.
        // Assuming BotData is available if it was before.
        public static async Task<string> GetRobotOwnerNameAsync(long groupId, BotWorker.Domain.Models.BotData.Platform botType) 
        {
             return await Repo.GetRobotOwnerNameAsync(groupId);
        }

        public static bool IsCanTrial(long groupId) => IsCanTrialAsync(groupId).GetAwaiter().GetResult();

        public static async Task<GroupInfo?> LoadAsync(long id) => await Repo.GetAsync(id);

        public static void Append(long groupId, string name, long selfId, string selfName, long userId, long userId2)
        {
             Repo.AppendAsync(groupId, name, selfId, selfName, userId, userId2).GetAwaiter().GetResult();
        }

        public static async Task<bool> IsCanTrialAsync(long groupId) => await Repo.IsCanTrialAsync(groupId);

        public static async Task<int> SetInvalidAsync(long groupId, string groupName = "", long groupOwner = 0, long robotOwner = 0) => await Repo.SetInvalidAsync(groupId, groupName, groupOwner, robotOwner);

        public static async Task<int> SetHintDateAsync(long groupId) => await Repo.SetHintDateAsync(groupId);

        public static async Task<bool> GetIsWhiteAsync(long groupId) => await Repo.GetIsWhiteAsync(groupId);

        public static async Task<string> GetIsBlockResAsync(long groupId) => await Repo.GetIsBlockAsync(groupId) ? "已开启" : "已关闭";

        public static async Task<bool> GetIsBlockAsync(long groupId) => await Repo.GetIsBlockAsync(groupId);

        public static async Task<int> GetIsOpenAsync(long groupId) => await Repo.GetIsOpenAsync(groupId);

        public static async Task<int> GetLastHintTimeAsync(long groupId) => await Repo.GetLastHintTimeAsync(groupId);

        public static async Task<int> CloudAnswerAsync(long groupId) => await Repo.CloudAnswerAsync(groupId);

        public static async Task<string> CloudAnswerResAsync(long groupId) => await Repo.CloudAnswerResAsync(groupId);

        public static async Task<bool> GetIsBlackExitAsync(long groupId) => await Repo.GetIsBlackExitAsync(groupId);

        public static async Task<bool> GetIsBlackKickAsync(long groupId) => await Repo.GetIsBlackKickAsync(groupId);

        public static string GetClosedFunc(long groupId) => GetClosedFuncAsync(groupId).GetAwaiter().GetResult();

        public static async Task<string> GetClosedFuncAsync(long groupId) => await Repo.GetClosedFuncAsync(groupId);

        public static string GetClosedRegex(long groupId) => GetClosedRegexAsync(groupId).GetAwaiter().GetResult();

        public static async Task<string> GetClosedRegexAsync(long groupId) => await Repo.GetClosedRegexAsync(groupId);

        public static async Task<bool> GetIsExitHintAsync(long groupId) => await Repo.GetIsExitHintAsync(groupId);

        public static async Task<bool> GetIsKickHintAsync(long groupId) => await Repo.GetIsKickHintAsync(groupId);

        public static async Task<bool> GetIsRequirePrefixAsync(long groupId) => await Repo.GetIsRequirePrefixAsync(groupId);

        public static async Task<string> GetJoinResAsync(long groupId) => await Repo.GetJoinResAsync(groupId);

        public static async Task<string> GetSystemPromptAsync(long groupId) => await Repo.GetSystemPromptAsync(groupId);

        public static async Task<string> GetAdminRightResAsync(long groupId) => await Repo.GetAdminRightResAsync(groupId);

        public static async Task<string> GetRightResAsync(long groupId) => await Repo.GetRightResAsync(groupId);

        public static async Task<string> GetTeachRightResAsync(long groupId) => await Repo.GetTeachRightResAsync(groupId);

        public static async Task<int> SetInGameAsync(int isInGame, long groupId) => await Repo.SetInGameAsync(isInGame, groupId);

        public static async Task<int> StartCyGameAsync(int state, string lastChengyu, long groupId) => await Repo.StartCyGameAsync(state, lastChengyu, groupId);

        public static async Task<int> StartCyGameAsync(long groupId) => await Repo.StartCyGameAsync(groupId);

        public static async Task<int> GetChengyuIdleMinutesAsync(long groupId) => await Repo.GetChengyuIdleMinutesAsync(groupId);

        public static int SetPowerOn(long groupId) => SetPowerOnAsync(groupId).GetAwaiter().GetResult();
        public static int SetPowerOff(long groupId) => SetPowerOffAsync(groupId).GetAwaiter().GetResult();
        public static int SetInGame(int isInGame, long groupId) => SetInGameAsync(isInGame, groupId).GetAwaiter().GetResult();
        public static int StartCyGame(int state, string lastChengyu, long groupId) => StartCyGameAsync(state, lastChengyu, groupId).GetAwaiter().GetResult();
        public static int StartCyGame(long groupId) => StartCyGameAsync(groupId).GetAwaiter().GetResult();
        public static int GetChengyuIdleMinutes(long groupId) => GetChengyuIdleMinutesAsync(groupId).GetAwaiter().GetResult();

        public static int SetPowerOn(bool isOpen, long groupId) => SetPowerOnAsync(isOpen, groupId).GetAwaiter().GetResult();
        public static int SetIsOpen(bool isOpen, long groupId) => Repo.SetIsOpenAsync(isOpen, groupId).GetAwaiter().GetResult();

        public static int GetLastHintTime(long groupId) => GetLastHintTimeAsync(groupId).GetAwaiter().GetResult();
        public static int SetHintDate(long groupId) => SetHintDateAsync(groupId).GetAwaiter().GetResult();
        public static int SetInvalid(long groupId, string groupName = "", long groupOwner = 0, long robotOwner = 0) => SetInvalidAsync(groupId, groupName, groupOwner, robotOwner).GetAwaiter().GetResult();
        public static bool IsVip(long groupId) => GroupVip.IsVipAsync(groupId).GetAwaiter().GetResult();

        public static string GetWelcomeRes(long groupId) => GetWelcomeResAsync(groupId).GetAwaiter().GetResult();
        public static async Task<string> GetWelcomeResAsync(long groupId) => await Repo.GetWelcomeResAsync(groupId);

        public static async Task<int> AppendAsync(long group, string name, long selfId, string selfName, long groupOwner = 0, long robotOwner = 0, string openid = "")
        {
             return await Repo.AppendAsync(group, name, selfId, selfName, groupOwner, robotOwner, openid);
        }
        
        // Missing UpdateGroupAsync static wrapper if used
        public static async Task<int> UpdateGroupAsync(long group, string name, long selfId, long groupOwner = 0, long robotOwner = 0)
        {
             return await Repo.UpdateGroupAsync(group, name, selfId, groupOwner, robotOwner);
        }

        // Missing GetSourceGroupIdAsync static wrappers
        public static async Task<long> GetSourceGroupIdAsync(long groupId) => await Repo.GetSourceGroupIdAsync(groupId);
        public static async Task<long> GetSourceGroupIdAsync(long botUin, long groupId) => await Repo.GetSourceGroupIdAsync(botUin, groupId);

        public static bool GetBool(string key, long id)
        {
             if (key == "IsHintClose") return Repo.GetIsHintCloseAsync(id).GetAwaiter().GetResult();
             return false;
        }
    }
}