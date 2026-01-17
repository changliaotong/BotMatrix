using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IGroupWarnService
    {
        Task<string> GetEditKeywordAsync(long groupId, string message);
        Task<string> GetClearResAsync(long groupId, string cmdPara);
        Task<string> GetWarnInfoAsync(long groupId, string cmdPara);
        string GetCmdName(string cmdName);
        Task<string> GetKeysSetAsync(long groupId, string cmdName = "");
        Task<bool> ExistsKeyAsync(long groupId, string cmdPara, string cmdPara2);
        string RegexReplaceKeyword(string keyword);
        string RegexRemove(string regexKey, string keyToRemove);
    }
}
