using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IGroupWarnRepository : IBaseRepository<GroupWarn>
    {
        Task<long> CountByGroupAndUserAsync(long groupId, long userId);
        Task<int> DeleteByGroupAndUserAsync(long groupId, long userId);
        Task<string> GetEditKeywordAsync(long groupID, string message);
        Task<string> GetKeysSetAsync(long group_id, string cmdName = "");
        Task<bool> ExistsKeyAsync(long group_id, string cmdPara, string cmdPara2);
        Task<string> GetClearResAsync(long groupId, string cmdPara);
        Task<string> GetWarnInfoAsync(long groupId, string cmdPara);
        Task<int> AppendWarnAsync(long botUin, long userId, long groupId, string warnInfo, long insertBy);
        string GetCmdName(string cmdName);
        string GetCmdPara(string cmdPara);
    }
}
