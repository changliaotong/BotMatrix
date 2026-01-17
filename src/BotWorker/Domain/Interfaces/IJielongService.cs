using System;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Domain.Interfaces
{
    public interface IJielongService
    {
        Task<string> GetJielongResAsync(IPluginContext ctx, string cmdPara);
        Task<bool> InGameAsync(long groupId, long userId);
        Task<int> SetLastChengyuAsync(long groupId, long userId, bool isGroup, string currCy);
        Task<int> StartAsync(long groupId, long userId, bool isGroup, string cmdPara);
        Task<int> GameOverAsync(long groupId, long userId, bool isGroup);
        Task<string> CurrCyAsync(long groupId, long userId, bool isGroup);
        Task<bool> UserInGameAsync(long groupId, long userId, bool isGroup);
        Task<int> AppendAsync(long groupId, long qq, string name, string chengYu, int gameNo);
        Task<bool> IsDupAsync(long groupId, long qq, string chengYu);
        Task<string> GetJielongAsync(long groupId, long UserId, string currCy);
        Task<int> GetMaxIdAsync(long groupId);
        Task<string> GetGameCountStrAsync(long groupId, long userId);
        Task<int> GetCountAsync(long groupId, long userId);
        Task<long> GetCreditAddAsync(long userId);
        Task<string> AddCreditAsync(IPluginContext ctx);
        Task<string> MinusCreditAsync(IPluginContext ctx);
    }
}
