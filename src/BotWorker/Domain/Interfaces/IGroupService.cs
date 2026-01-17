using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Interfaces
{
    public interface IGroupService
    {
        Task<string> SetPowerOnOffAsync(long botUin, long groupId, long userId, string cmdName);
        Task<string> SetAdminRightAsync(long groupId, string cmdPara);
        Task<string> SetRightAsync(long groupId, string cmdPara);
        Task<string> SetTeachRightAsync(long groupId, string cmdPara);
        Task<string> SetBlockMinAsync(long groupId, string cmdPara);
        Task<string> SetJoinGroupAsync(long groupId, string cmdPara);
        Task<string> SetChangHintAsync(long groupId, string cmdPara);
        Task<string> SetWelcomeMsgAsync(long groupId, string cmdPara);
        Task<string> SetSystemPromptAsync(long groupId, string cmdPara);
        Task<string> SetupReplyModeAsync(long groupId, string cmdPara);
        Task<string> GetSetRobotOpenAsync(long groupId, string cmdPara);
        Task<string> GetSetCityAsync(long userId, string cmdPara);
        Task<bool> GetBoolAsync(string field, long groupId);
        Task<string> SetCloudAnswerAsync(long groupId, string cmdPara);
        Task<string> SetExitGroupAsync(long groupId, string cmdPara, GroupInfo group);
        Task<string> SetKickBlackAsync(long groupId, string cmdPara, GroupInfo group);
        Task<string> SetDefaultGroupAsync(long userId, long groupId, bool isGroup, string cmdPara, string botUinDef);
        Task<int> SetValueAsync(string field, object value, long groupId);
        Task<(long groupId, bool isNew)> GetGroupIdAsync(string groupOpenid, string groupName, long userId, long botUin = 0, string botName = "");
        Task<bool> GetIsCreditAsync(long groupId);
        Task<int> SetPowerOffAsync(long groupId);
        Task<int> SetPowerOnAsync(long groupId);
        Task<bool> GetPowerOnAsync(long groupId);
        Task<bool> IsPowerOffAsync(long groupId);
        Task<bool> IsCanTrialAsync(long groupId);
        Task<int> SetInvalidAsync(long groupId, string groupName = "", long groupOwner = 0, long robotOwner = 0);
    }
}
