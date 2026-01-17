using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Interfaces
{
    public interface IGame2048Service
    {
        Task<string> GetGameResAsync(long groupId, long qq, string cmdPara);
    }
}
