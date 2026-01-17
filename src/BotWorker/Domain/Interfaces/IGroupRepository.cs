using System;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using System.Collections.Generic;
using System.Data;

namespace BotWorker.Domain.Interfaces
{
    public interface IGroupRepository : IBaseRepository<GroupInfo>
    {
        Task<GroupInfo?> GetByOpenIdAsync(string openId, long botUin);
        Task<long> AddAsync(GroupInfo group);
        Task<bool> UpdateAsync(GroupInfo group);
        Task<long> GetGroupOwnerAsync(long groupId, long def = 0, IDbTransaction? trans = null);
        Task<bool> GetIsCreditAsync(long groupId);
        Task<bool> GetIsPetAsync(long groupId);
        Task<int> SetPowerOnAsync(long groupId, IDbTransaction? trans = null);
        Task<int> SetPowerOffAsync(long groupId, IDbTransaction? trans = null);
        Task<int> StartCyGameAsync(int state, string lastChengyu, long groupId);
        Task<int> StartCyGameAsync(long groupId);
        Task<int> GetChengyuIdleMinutesAsync(long groupId);
        Task<bool> GetPowerOnAsync(long groupId, IDbTransaction? trans = null);
        Task<int> SetRobotOwnerAsync(long groupId, long ownerId, IDbTransaction? trans = null);
        Task<long> GetRobotOwnerAsync(long groupId, long def = 0, IDbTransaction? trans = null);
        Task<bool> IsOwnerAsync(long groupId, long userId, IDbTransaction? trans = null);
        Task<bool> IsPowerOffAsync(long groupId, IDbTransaction? trans = null);
        Task<bool> GetIsValidAsync(long groupId, IDbTransaction? trans = null);
        Task<string> GetRobotOwnerNameAsync(long groupId, string botName = "");
        Task<bool> IsCanTrialAsync(long groupId);
        Task<int> SetInvalidAsync(long groupId, string groupName = "", long groupOwner = 0, long robotOwner = 0);
        Task<int> SetHintDateAsync(long groupId);
        Task<bool> GetIsWhiteAsync(long groupId);
        Task<string> GetIsBlockResAsync(long groupId);
        Task<bool> GetIsBlockAsync(long groupId);
        Task<int> GetIsOpenAsync(long groupId);
        Task<int> GetLastHintTimeAsync(long groupId);
        Task<int> CloudAnswerAsync(long groupId);
        Task<string> CloudAnswerResAsync(long groupId);
        Task<bool> GetIsBlackExitAsync(long groupId);
        Task<bool> GetIsBlackKickAsync(long groupId);
        Task<string> GetClosedFuncAsync(long groupId);
        Task<string> GetClosedRegexAsync(long groupId);
        Task<bool> GetIsExitHintAsync(long groupId);
        Task<bool> GetIsKickHintAsync(long groupId);
        Task<bool> GetIsRequirePrefixAsync(long groupId);
        Task<string> GetJoinResAsync(long groupId);
        Task<string> GetSystemPromptAsync(long groupId);
        Task<string> GetAdminRightResAsync(long groupId);
        Task<string> GetRightResAsync(long groupId);
        Task<string> GetTeachRightResAsync(long groupId);
        Task<int> SetInGameAsync(int isInGame, long groupId);
        Task<string> GetWelcomeResAsync(long groupId);
        Task<string> GetGroupNameAsync(long groupId);
        Task<string> GetGroupOwnerNicknameAsync(long groupId);
        Task<bool> GetIsAIAsync(long groupId);
        Task<bool> GetIsOwnerPayAsync(long groupId);
        Task<int> GetContextCountAsync(long groupId);
        Task<bool> GetIsMultAIAsync(long groupId);
        Task<bool> GetIsUseKnowledgebaseAsync(long groupId);
        Task<int> AppendAsync(long groupId, string name, long selfId, string selfName, long groupOwner = 0, long robotOwner = 0, string openid = "");
        Task<bool> GetIsNoLogAsync(long groupId);
        Task<bool> GetIsNoCheckAsync(long groupId);
        Task<bool> GetIsHintCloseAsync(long groupId);
        Task<long> GetSourceGroupIdAsync(long groupId);
        Task<long> GetSourceGroupIdAsync(long botUin, long groupId);
        Task<int> UpdateGroupAsync(long group, string name, long selfId, long groupOwner = 0, long robotOwner = 0);
        Task<int> SetIsOpenAsync(bool isOpen, long groupId);
        Task<int> SetPowerOnAsync(bool isOpen, long groupId);
        Task<bool> GetPowerOnAsync(long groupId);
        Task<string> GetSystemPromptStatusAsync(long groupId);
        Task<string> GetVipResAsync(long groupId);

        // New methods for GroupSetup.cs refactoring
        Task<int> UpdateIsPowerOnAsync(long groupId, bool isPowerOn, IDbTransaction? trans = null);
        Task<int> UpdateAdminRightAsync(long groupId, int adminRight);
        Task<int> UpdateUseRightAsync(long groupId, int useRight);
        Task<int> UpdateTeachRightAsync(long groupId, int teachRight);
        Task<int> UpdateBlockMinAsync(long groupId, int blockMin);
        Task<int> UpdateJoinGroupSettingsAsync(long groupId, int isAccept, string rejectMessage, string regexRequestJoin);
        Task<int> UpdateIsChangeHintAsync(long groupId, bool isChangeHint);
        Task<int> UpdateWelcomeMessageAsync(long groupId, string message);
        Task<int> UpdateIsWelcomeHintAsync(long groupId, bool isWelcomeHint);
        Task<int> UpdateSystemPromptAsync(long groupId, string systemPrompt);
        Task<int> UpdateReplyModeAsync(long groupId, int modeReply);
        Task<int> UpdateCloseRegexAsync(long groupId, string closeRegex);
        Task<int> UpdateIsCloudAnswerAsync(long groupId, int isCloudAnswer);
        Task<int> UpdateExitGroupSettingsAsync(long groupId, bool isExitHint, bool isBlackExit);
        Task<int> UpdateKickBlackSettingsAsync(long groupId, bool isKickHint, bool isBlackKick);
        Task<IEnumerable<(long GroupId, string GroupName)>> GetOwnedGroupsAsync(long userId, int top = 5);
        Task<int> GetBlockMinAsync(long groupId);
        Task<string> GetIsChangeHintResAsync(long groupId);
        Task<string> GetReplyModeResAsync(long groupId);
    }
}
