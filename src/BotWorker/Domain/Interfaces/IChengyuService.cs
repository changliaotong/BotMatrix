using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Domain.Interfaces
{
    public interface IChengyuService
    {
        Task<long> GetOidAsync(string text);
        Task<bool> ExistsAsync(string text);
        Task<string> PinYinAsync(string text);
        Task<string> PinYinAsciiAsync(string text);
        Task<string> GetCyInfoAsync(string text, long oid = 0);
        Task<Dictionary<string, string>> GetCyInfoAsync(IEnumerable<string> cys);
        Task<string> GetInfoHtmlAsync(string text, long oid = 0);
        Task<Dictionary<string, string>> GetInfoHtmlAsync(IEnumerable<string> cys);
        Task<string> PinYinFirstAsync(string textCy);
        Task<string> PinYinLastAsync(string text);
        Task<string> GetCyResAsync(IPluginContext ctx, string cmdPara);
        Task<string> GetFanChaResAsync(IPluginContext ctx, string cmdPara);
        Task<string> GetRandomAsync(string category);
    }
}
